package cli

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/coxley/pmlproxy/pb"
	"github.com/spf13/cobra"
)

var (
	ErrFileNoExist     = fmt.Errorf("file doesn't exist")
	ErrFileIsDir       = fmt.Errorf("cannot pass directory")
	renderFormat       string
	renderOutputToDisk bool
	renderOutputFname  string = "diagram"
	renderOutputSep    string = "---PMLPROXY---"
)

func init() {
	cmd := &cobra.Command{
		Use:   "render [file|shortcode]",
		Args:  cobra.MaximumNArgs(1),
		Run:   renderRun,
		Short: "render diagram(s) as images (PNG or SVG)",
		Long: `Data is read from stdin when no argument is provided

Diagram MUST be wrapped with @startXXX and @endXXX (@startuml, @startditaa, @startgantt, ...)

One "diagram" can contain multiple definitions back-to-back. You can dump the images
as different files or stdout with a custom separator.

Stdin can either be the source text or shortcode.
`,
		Example: `
pml render diagram.puml
pml render SyfFKj2rKt3CoKnELR1Io4ZDoSa70000
echo -e '@startuml\nBob->Alice\n@enduml' | pml render

# Multiple diagrams
cat <<EOF | pml render -o
@startuml
rectangle Foo
@enduml

@startuml
rectangle Bar
@enduml
EOF

# Example extracting
pml extract diagram-0.png
# @startuml
# rectangle Foo
# @enduml
`,
	}
	rootCmd.AddCommand(cmd)
	flags := cmd.Flags()
	flags.StringVarP(&renderFormat, "format", "f", "png", "format to render diagram as, see help for options")
	flags.BoolVarP(&renderOutputToDisk, "output-to-disk", "o", renderOutputToDisk, "writes diagram(s) to disk when set")
	flags.StringVarP(&renderOutputFname, "output-name", "n", renderOutputFname, "name of files to write, sans ext — appended with ordered numbers if multiple diagrams in source")
	flags.StringVar(&renderOutputSep, "sep", renderOutputSep, "string to write between multiple diagrams when not writing to disk")

}

func getFormatOptions() string {
	var opts string
	for k, v := range pb.Format_value {
		if v == 0 {
			continue
		}
		opts += fmt.Sprintf("* %s\n", strings.ToLower(k))
	}
	return opts
}

func renderRun(cmd *cobra.Command, args []string) {
	diagram := pb.Diagram{}

	// Work out where to get input from
	if len(args) == 1 {
		content, err := fileContents(args[0])
		if err == nil {
			diagram.Source = content
		} else if err == ErrFileNoExist {
			// Assume shortcode
			diagram.Encoded = args[0]
		} else {
			fatalf("unable to read %s: %v", args[0], err)
		}
	} else {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fatalf("failed reading stdin: %v", err)
		}
		content := string(b)
		if strings.Contains(content, "@start") && strings.Contains(content, "@end") {
			diagram.Source = content
		} else {
			diagram.Encoded = content
		}
	}

	_, ok := pb.Format_value[strings.ToUpper(renderFormat)]
	if !ok {
		fatalf("invalid format type: %s", renderFormat)
	}
	format := pb.Format(pb.Format_value[strings.ToUpper(renderFormat)])
	req := &pb.RenderRequest{Diagram: &diagram, Format: format}

	client, err := getClient()
	if err != nil {
		fatalf("unable to connect to plantuml server: %v", err)
	}

	ctx := context.Background()
	resp, err := client.Render(ctx, req)
	if err != nil {
		fatalf("unexpected failure: %v\n", err)
	}

	for num, img := range resp.Data {
		dest := getDest(
			renderOutputFname, num, strings.ToLower(renderFormat), renderOutputToDisk,
		)
		if num > 0 && !renderOutputToDisk {
			dest.WriteString(renderOutputSep)
		}
		dest.Write(img)
	}
}

func getDest(fname string, num int, ext string, writeToDisk bool) *os.File {
	if writeToDisk {
		dest, err := os.Create(fmt.Sprintf("%s-%d.%s", fname, num, ext))
		if err != nil {
			fatalf("couldn't create output file: %v", err)
		}
		return dest
	}
	return os.Stdout
}

func fileContents(name string) (string, error) {
	info, err := os.Stat(name)
	if os.IsNotExist(err) {
		return "", ErrFileNoExist
	}
	if info.IsDir() {
		return "", ErrFileIsDir
	}
	content, err := ioutil.ReadFile(name)
	if err != nil {
		return "", fmt.Errorf("failed reading file: %v", err)
	}
	return string(content), nil
}
