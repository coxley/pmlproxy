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
		Short: "encode full diagram into a shorter, portable string",
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

func shortenRun(cmd *cobra.Command, args []string) {
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

	resp, err := client.Shorten(context.Background(), &pb.ShortenRequest{Value: text})
	if err != nil {
		fatalf("failed to shorten: %v", err)
	}
	fmt.Print(resp.Short)
}

func expandRun(cmd *cobra.Command, args []string) {
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

	resp, err := client.Expand(context.Background(), &pb.ExpandRequest{Value: shortcode})
	if err != nil {
		fatalf("failed to expand: %v", err)
	}
	fmt.Print(resp.Full)
}
