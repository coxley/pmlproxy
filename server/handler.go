package server

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/coxley/pmlproxy/pb"
	"github.com/golang/glog"
	"github.com/golang/groupcache"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler interface {
	pb.PlantUMLServer
	ManageWorkers(ctx context.Context)
}

type handler struct {
	pb.PlantUMLServer

	// How many PlantUML sub-processes should we create to handle the rendering?
	//
	// Each worker spins up a goroutine that manages and reads from the sub-process.
	Workers int

	// Command-line args to Java and PlantUML (default: h.MakeWorkerArgs())
	//
	// Crafting your own arguments instead of amending the default ones may
	// break assumptions. Create a Handler, then set this field ho
	// h.MakeWorkerArgs() + your changes.
	WorkerArgs []string

	// Upper bound to wait for a digram to render (default: 10s)
	//
	// Set high enough to only cancel problematic
	// payloads. Value will vary based on configured workers and available
	// CPU.
	RenderTimeout time.Duration

	// Delimeter between diagrams over the PlantUML pipe (default: XXXPUMLXXX)
	//
	// Pick something that won't appear in your diagram text (more important for SVG).
	PipeDelimiter string

	// Where will PlantUML look to import local themes, plugins, etc?
	//
	// We will ensure the path exists, failing-hard without the correct privileges.
	SearchPath string

	// Path to the Java binary. (default: "java" in $PATH)
	JavaExe string

	// Path to the Java binary. (default: "/usr/share/java/plantuml/plantuml.jar")
	PlantUMLPath string

	// Will store render results in the cache if set.
	//
	// Note: You MUST take care of managing the groupcache pool yourself. See
	// github.com/mailgun/groupcache for a valid example.
	GroupCache      bool
	GroupCacheBytes int64
	renderGroup     *groupcache.Group

	workerCh chan workerReq
}

var DefaultHandler = handler{
	Workers:         runtime.NumCPU(),
	RenderTimeout:   time.Second * 10,
	JavaExe:         "java",
	PlantUMLPath:    "/usr/share/java/plantuml/plantuml.jar",
	SearchPath:      ".",
	PipeDelimiter:   "XXXPUMLXXX",
	GroupCacheBytes: 10000000, // 10MB
	workerCh:        make(chan workerReq),
}

func (h *handler) GetWorkerArgs() []string {
	if len(h.WorkerArgs) > 0 {
		return h.WorkerArgs
	}
	return h.MakeWorkerArgs()
}

func (h *handler) MakeWorkerArgs() []string {
	return []string{
		fmt.Sprintf(`-Dplantuml.include.path="%s"`, h.SearchPath),
		"-jar",
		h.PlantUMLPath,
		"-headless",
		"-pipe",
		"-pipedelimitor",
		h.PipeDelimiter,
	}
}

func (h *handler) makeRenderCache() *groupcache.Group {
	group := groupcache.NewGroup("render", h.GroupCacheBytes, groupcache.GetterFunc(
		func(ctx context.Context, id string, dest groupcache.Sink) error {
			glog.Infof("cache getter: %v", id)
			// id example: SVG:encodedtext
			s := strings.SplitN(id, ":", 2)
			if len(s) != 2 {
				return fmt.Errorf("cache key has wrong format: %v", id)
			}

			fstr, short := s[0], s[1]
			resp, err := h.directRender(ctx, &pb.RenderRequest{
				Diagram: &pb.Diagram{Short: short},
				Format:  pb.Format(pb.Format_value[fstr]),
			})
			if err != nil {
				return err
			}
			if err := dest.SetProto(resp); err != nil {
				return err
			}
			return nil
		},
	))
	return group
}

// ManageWorkers initiates and maintains the right number of ManageWorkers
//
// It creates a groupcache group for rendering, if toggled.
//
// Logs from workers are prefixed with their number. This may be larger than
// max workers as the ID increases after crashes.
//
// Returns only when ctx is done.
func (h *handler) ManageWorkers(ctx context.Context) {
	if h.GroupCache {
		h.renderGroup = h.makeRenderCache()
	}
	if h.workerCh == nil {
		h.workerCh = make(chan workerReq)
	}
	// Start as many workers as able, new ones spinning up as old ones exit.
	var i int
	sem := make(chan struct{}, h.Workers)
	for {
		select {
		case <-ctx.Done():
			return
		case sem <- struct{}{}:
			go func(i int) {
				// TODO: Expose worker counters
				h.worker(ctx, i)
				// drain so we can spawn another
				<-sem
			}(i)
			i++
		}
	}
}

// WorkerRender is a convenience function to hide the return channel.
func (h *handler) WorkerRender(text string, format pb.Format) ([][]byte, error) {
	ch := make(chan workerRes)
	h.workerCh <- workerReq{text: text, format: format, result: ch}
	result := <-ch
	return result.data, result.err
}

func (h *handler) Render(ctx context.Context, req *pb.RenderRequest) (*pb.RenderResponse, error) {
	glog.Info("hitting render")
	if !h.GroupCache {
		return h.directRender(ctx, req)
	}

	enc := req.Diagram.Short
	if enc == "" && req.Diagram.Full == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"full or short diagram must be set",
		)
	}

	if enc == "" {
		e, err := ToShort(req.Diagram.Full)
		if err != nil {
			return nil, err
		}
		enc = e
	}

	var resp pb.RenderResponse
	key := fmt.Sprintf("%s:%s", req.Format.String(), enc)
	glog.Info("doing a cache lookup")
	if err := h.renderGroup.Get(ctx, key, groupcache.ProtoSink(&resp)); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Raw render without hitting the cache
func (h *handler) directRender(ctx context.Context, req *pb.RenderRequest) (*pb.RenderResponse, error) {
	glog.Infof("request to render")

	if req.Format == pb.Format_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "must give a valid Format")
	}

	if req.Diagram.Short == "" && req.Diagram.Full == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"full or short diagram must be set",
		)
	}

	text := req.Diagram.Full
	if req.Diagram.Short != "" && text == "" {
		t, err := FromShort(req.Diagram.Short)
		if err != nil {
			return nil, status.Error(
				codes.InvalidArgument,
				"unable to decode diagram: "+err.Error(),
			)
		}
		text = t
	}
	res, err := h.WorkerRender(text, req.Format)
	return &pb.RenderResponse{Data: res}, err
}

func (h *handler) Shorten(ctx context.Context, req *pb.ShortenRequest) (*pb.ShortenResponse, error) {
	enc, err := ToShort(req.Value)
	return &pb.ShortenResponse{Short: enc}, err
}

func (h *handler) Expand(ctx context.Context, req *pb.ExpandRequest) (*pb.ExpandResponse, error) {
	full, err := FromShort(req.Value)
	return &pb.ExpandResponse{Full: full}, err
}

func (h *handler) Extract(ctx context.Context, req *pb.ExtractRequest) (*pb.ExtractResponse, error) {
	metadata := ExtractFromImage(req.Data)
	if metadata == "" {
		return nil, status.Error(
			codes.NotFound,
			"couldn't extract metadata from either SVG or PNG parsers",
		)
	}

	original, unpacked := divideMetadata(metadata)
	text := original
	if req.ExpandMacros {
		text = unpacked
	}
	text = normalizeText(text)

	short, err := ToShort(text)
	if err != nil {
		return nil, err
	}

	return &pb.ExtractResponse{Diagram: &pb.Diagram{Full: text, Short: short}}, nil
}
