package cli

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/coxley/pmlproxy/pb"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "shorten [file]",
		Args:  cobra.MaximumNArgs(1),
		Run:   shortenRun,
		Short: "encode diagram text into a shorter, portable string",
		Long:  "read from stdin if no file is provided",
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "edit [file]",
		Args:  cobra.MaximumNArgs(1),
		Run:   editRun,
		Short: "Print a URL to edit the diagram interactively",
		Long:  "read from stdin if no file is provided",
	})
	rootCmd.AddCommand(&cobra.Command{
		Use:   "expand [shortcode]",
		Args:  cobra.MaximumNArgs(1),
		Run:   expandRun,
		Short: "decode the short string into original diagram text",
		Long:  "read from stdin if no file is provided",
	})
}

// encodePrint has the logic shared between the logic for shorten and edit.
// They both need to get a file or stdin and encoded it.
// Pass in the transform function to make it do the custom logic for the two commands.
func encodedPrint(args []string, transform func(string) string) {
	var content []byte
	var err error
	if len(args) == 1 {
		content, err = ioutil.ReadFile(args[0])
		if err != nil {
			fatalf("failed reading file: %v", err)
		}
	} else {
		content, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			fatalf("failed reading stdin: %v", err)
		}
	}

	text := string(content)

	client, err := getClient()
	if err != nil {
		fatalf("unable to connect to server: %v", err)
	}

	resp, err := client.Shorten(context.Background(), &pb.ShortenRequest{Source: text})
	if err != nil {
		fatalf("failed to shorten: %v", err)
	}
	fmt.Print(transform(resp.Encoded))
}

func shortenRun(cmd *cobra.Command, args []string) {
	// transform is a no-op; it Just passes the value through.
	encodedPrint(args, func(s string) string { return s })
}

// editRunTransform takes the encoded diagram and outputs an editor link
func editRunTransform(encoded string) string {
	return "this command isn't implemented just yet — come back later"
}

func editRun(cmd *cobra.Command, args []string) {
	encodedPrint(args, editRunTransform)
}

func expandRun(cmd *cobra.Command, args []string) {
	// TODO: When the editor is launched, support having editor links pasted
	var shortcode string
	if len(args) == 1 {
		shortcode = args[0]
	} else {
		content, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fatalf("failed reading stdin: %v", err)
		}
		shortcode = string(content)
	}

	client, err := getClient()
	if err != nil {
		fatalf("unable to connect to server: %v", err)
	}

	resp, err := client.Expand(context.Background(), &pb.ExpandRequest{Encoded: shortcode})
	if err != nil {
		fatalf("failed to expand: %v", err)
	}
	fmt.Print(resp.Source)
}
