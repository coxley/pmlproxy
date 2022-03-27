package server

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

// 89 50 4E 47 0D 0A 1A 0A
const pngHeader = "\x89\x50\x4E\x47\x0D\x0A\x1A\x0A"

// Technically, offset is minimum of 5-bytes. PlantUML doesn't set language or
// translated values, so we should be OK with this naive check.
//   * http://www.libpng.org/pub/png/spec/1.2/PNG-Chunks.html#C.iTXt
const offset = 5

const versionDelim = "PlantUML version "

var keyword = []byte("plantuml")

// ImageMetadata represents what PlantUML stores alongside rendered images
type ImageMetadata struct {
	Text string
	// TODO: Include other fields such as server version
}

func ExtractFromImage(img []byte) string {
	rdr := bytes.NewReader(img)
	metadata, err := FromPNG(rdr)
	if err == nil {
		return metadata.Text
	}

	// try SVG
	rdr.Reset(img)
	metadata, err = FromSVG(rdr)
	if err == nil {
		return metadata.Text
	}

	return ""
}

// FromPNG reads the source PlantUML diagram from a PNG image
//
// PNG images are implemented as a series of "TLV" + CRC chunks. There's a
// category of chunks called "Ancillary chunks".
//
// PlantUML includes the diagram text (+ other metadata), zlib-compressed, in
// a iTXt chunk type.  It's prefixed with the "plantuml" keyword.
//
// Read the type & length of every chunk until we reach iEND or find the
// plantuml iTXt. For all other chunks, seek past `length+CRC` to avoid reading
// any data into memory.
//
// Returns error if unable to locate.
//
// For more information on the PNG spec, see:
//   * http://www.libpng.org/pub/png/spec/1.2/PNG-Structure.html
func FromPNG(r io.ReadSeeker) (*ImageMetadata, error) {
	hdr := make([]byte, 8)
	if _, err := io.ReadFull(r, hdr); err != nil {
		return nil, err
	}
	if string(hdr) != pngHeader {
		return nil, fmt.Errorf("png header not found: %v", hdr)
	}

	// All 4-bytes, except data which is length-bytes
	// LENGTH CHUNK_TYPE DATA CRC
	buf := make([]byte, 4)
	var metadata string
Loop:
	for {
		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, fmt.Errorf("failed to read length: %v", err)
		}
		length := int64(binary.BigEndian.Uint32(buf))

		if _, err := io.ReadFull(r, buf); err != nil {
			return nil, fmt.Errorf("failed to read ctype: %v", err)
		}
		ctype := string(buf)
		switch ctype {
		case "iTXt":
			tmp := make([]byte, length)
			if _, err := io.ReadFull(r, tmp); err != nil {
				return nil, fmt.Errorf("failed to read data: %v", err)
			}

			if !bytes.HasPrefix(tmp, keyword) {
				// not our iTXt chunk
				continue
			}

			tmp = tmp[len(keyword)+offset:]
			zr, err := zlib.NewReader(bytes.NewReader(tmp))
			if err != nil {
				return nil, fmt.Errorf("failed to create zlib reader: %v", err)
			}
			enflated, err := ioutil.ReadAll(zr)
			if err != nil {
				return nil, fmt.Errorf("failed to decode metadata: %v", err)
			}
			metadata = string(enflated)
			break Loop
		case "iEND":
			break Loop
		default:
			// Skip data & CRC from current offset
			r.Seek(length+4, io.SeekCurrent)
		}
	}

	if metadata == "" {
		return nil, fmt.Errorf("plantuml iTXt not found in png")
	}

	text := extractMetadata(metadata)
	if text == "" {
		return nil, fmt.Errorf("unexpected output: %v", metadata)
	}
	return &ImageMetadata{text}, nil
}

// FromSVG reads the source PlantUML diagram from an SVG image
//
// Format of PlantUML SVG is roughly:
//
// <svg><g>
//   <...>
//   <!--MD5=[...]
// @startuml
// ...
// @enduml
//
// rest of metadata
// --></g></svg>
//
// Returns error if unable to locate.
func FromSVG(r io.ReadSeeker) (*ImageMetadata, error) {
	type svg struct {
		Groups []struct {
			Comment string `xml:",comment"`
		} `xml:"g"`
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	v := svg{}
	err = xml.Unmarshal(b, &v)
	if err != nil {
		return nil, err
	}
	if len(v.Groups) != 1 {
		return nil, fmt.Errorf("expected only one <g> in SVG: %+v", v)
	}

	// We want the last comment, which starts with @startXXX. The XML lib
	// doesn't support exporting comments as []string.
	tmp := strings.SplitN(v.Groups[0].Comment, "@start", 2)
	if len(tmp) != 2 {
		return nil, fmt.Errorf("@start not found in SVG comments: %v", v.Groups[0].Comment)
	}

	// add @start back from earlier splitting
	text := extractMetadata("@start" + tmp[1])
	if text == "" {
		return nil, fmt.Errorf("unexpected output: %v", tmp[1])
	}

	// https://forum.plantuml.net/11453/svg-metadata-space-between-hyphens
	// While-loop because strings.ReplaceAll avoids overlapping characters
	//
	// Eg: "- - -" must become "---", not "-- -"
	for strings.Contains(text, "- -") {
		text = strings.ReplaceAll(text, "- -", "--")
	}

	return &ImageMetadata{text}, nil
}

func extractMetadata(s string) string {
	split := strings.Split(s, versionDelim)
	if len(split) != 2 {
		return ""
	}
	return strings.TrimSpace(split[0])
}

// divideMetadata into up-to two diagrams
//
// The first diagram is written as it was submitted. The second, if present, is
// the version with pre-processing macros expanded. See the below post:
//   - https://forum.plantuml.net/11488/metadata-can-we-get-the-pre-processed-version
//
// @startuml
// ...
// @enduml
// @startuml
// ...
// @enduml
func divideMetadata(s string) (string, string) {
	if strings.Count(s, "\n@end") == 1 {
		return s, ""
	}

	// Locate and split on the last line of the first diagram
	lastLine := strings.Index(s, "\n@end")
	var lastLineEnd int
	if i := strings.Index(s[lastLine:], "\n"); i == 0 {
		lastLineEnd = lastLine + strings.Index(s[lastLine+1:], "\n") + 2
	} else {
		lastLineEnd = lastLine + i + 1
	}
	return s[:lastLineEnd], s[lastLineEnd:]
}
