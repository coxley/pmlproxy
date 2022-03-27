package server

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/coxley/pmlproxy/pb"
	"github.com/golang/glog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type workerReq struct {
	text   string
	format pb.Format
	result chan workerRes
}

type workerRes struct {
	data [][]byte
	err  error
	// TODO: Capture both pre and post processed text for comparing.
	// TODO: Recreate the issue with paged diagrams.
	// TODO: Maybe the above two can help signal that an error vs. original diagram is rendered.
}

// Spin up a PlantUML process and stream diagrams to it from jobs
//
// Putting the process into -pipe mode (assumption from args) avoids the JVM
// start-up penalty for every render. We lose exit-code as a diagnostic, but
// errors are rendered as images.
//
// Image format is specified by prepending @@@format <type> before the diagram.
//   - https://forum.plantuml.net/10808/is-there-a-way-to-use-multiple-output-formats-with-pipe
func (h *handler) worker(ctx context.Context, id int) {
	glog.Infof("[%d] starting worker", id)
	cctx, cancelCmd := context.WithCancel(ctx)
	cmd := exec.CommandContext(cctx, h.JavaExe, h.GetWorkerArgs()...)

	// Clean-up process on return
	defer cancelCmd()
	defer func() {
		if err := cmd.Process.Kill(); err != nil {
			// Should only reach in exceptional cases
			glog.Errorf("[%d] failed to kill java proc: %v", id, err)
		}
		cmd.Wait()
	}()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		glog.Errorf("[%d] worker failed to bind stderr: %v", id, err)
	}
	go workerLogger(id, stderr)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		glog.Errorf("[%d] worker failed to bind stdin: %v", id, err)
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		glog.Errorf("[%d] worker failed to bind stdout: %v", id, err)
		return
	}

	if err := cmd.Start(); err != nil {
		glog.Errorf("[%d] worker failed to start process: %v", id, err)
		return
	}

	glog.Infof("[%d] plantuml process started: %v", id, cmd)

	// PlantUML outputs the delimiter followed by a newline.
	splitOn := []byte(h.PipeDelimiter + "\n")
	buf := make([]byte, 4)

	for j := range h.workerCh {
		normalized := normalizeText(j.text)
		pageCnt, err := validate(normalized)
		if err != nil {
			j.result <- workerRes{
				data: nil,
				err:  err,
			}
			continue
		}

		deadline := time.AfterFunc(h.RenderTimeout, func() {
			glog.Errorf("[%d] aborting worker, diagram took over %s", id, h.RenderTimeout)
			cancelCmd() // will close the buffer we're reading from
		})
		data, err := func() ([][]byte, error) {
			// TODO: Expose counter for busy vs. free worker
			defer deadline.Stop()
			short, _ := ToShort(j.text)
			glog.Infof("[%d] rendering %d diagram(s): %s", id, pageCnt, short)
			fmt.Fprint(stdin, addFormatSpec(normalized, j.format))

			// TODO: Log about multiple pages
			var res [][]byte
			var cur []byte
			var found int
			for found < pageCnt {
				if bytes.HasSuffix(cur, splitOn) {
					cur = cur[:len(cur)-len(splitOn)]
					res = append(res, cur)
					found++
					cur = nil
					glog.Infof(
						"[%d] found separator, diagram %d/%d size: %d bytes",
						id, found, pageCnt, len(res[found-1]),
					)
					continue
				}
				n, err := stdout.Read(buf)
				if err != nil {
					return nil, fmt.Errorf("error reading diagram: %v", err)
				}
				cur = append(cur, buf[:n]...)
			}
			return res, nil
		}()
		res := workerRes{
			data: data,
			err:  err,
		}
		j.result <- res

		// Errors seen after data is sent to the sub-process make it hard to
		// know what state PlantUML is in.
		if err != nil {
			glog.Errorf("[%d] exiting worker due to error: %v", id, err)
			return
		}
	}
}

func workerLogger(id int, stderr io.Reader) {
	scanErr := bufio.NewScanner(stderr)
	for scanErr.Scan() {
		glog.Errorf("[%d] plantuml stderr: %v", id, scanErr.Text())
	}
	if err := scanErr.Err(); err != nil {
		glog.Errorf("[%d] closing stderr scanner: %v", id, err)
	}
}

// normalizeText prepares a user-provided diagram for being piped into PlantUML
func normalizeText(s string) string {
	// Normalize newlines to UNIX
	s = strings.ReplaceAll(s, "\r\n", "\n")
	// Remove all leading and trailing whitespace
	s = strings.TrimSpace(s)

	// If the @enduml line has space preceding it, PlantUML won't respond over
	// the pipe.
	//
	// From:
	// @startuml
	//     Bob -> Alice
	//     @enduml
	//
	// To:
	// @startuml
	//     Bob -> Alice
	// @enduml
	if i := strings.LastIndex(s, "\n"); i != -1 {
		s = s[:i] + "\n" + strings.TrimSpace(s[i:])
	}
	return s
}

// addFormatSpec tells PlantUML which format to render the image in
func addFormatSpec(s string, format pb.Format) string {
	var formatSpec string
	switch format {
	case pb.Format_SVG:
		formatSpec = "@@@format svg"
	case pb.Format_PNG:
		formatSpec = "@@@format png"
	default:
		glog.Fatalf("unknown format specifier: %s", format.String())
	}

	// We also need to finish with a newline
	return formatSpec + "\n" + s + "\n"
}

// validate if diagram text looks OK, returning the number of diagrams to expect
//
// PlantUML in pipe mode will wait until it sees @startXYZ. Without this, we
// will hang until the abortTimer fires.
//
// Assumes s has been trimmed of whitespace, and that @enduml has nothing preceding it.
//
// PlantUML supports multiple pairs of @startXYZ and @endXYZ in the same source
// text. It just returns multiple images. We should know how many images to
// expect before parsing.
func validate(s string) (int, error) {

	// To make the count deterministic
	s = "\n" + s
	startCount := strings.Count(s, "\n@start")
	endCount := strings.Count(s, "\n@end")

	if startCount == 0 || endCount == 0 {
		return 0, status.Error(
			codes.InvalidArgument,
			"diagram is missing required @startXYZ/@endXYZ: https://plantuml.com/faq#ef810064611fabdc",
		)
	}

	if startCount != endCount {
		return 0, status.Error(
			codes.InvalidArgument,
			"diagram has mismatched count of @startXYZ and @endXYZ pairs: https://plantuml.com/faq#ef810064611fabdc",
		)
	}
	return startCount, nil
}
