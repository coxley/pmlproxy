package cli

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/coxley/pmlproxy/pb"
	"github.com/spf13/cobra"
)

var extractShort bool

func init() {
	cmd := cobra.Command{
		Use:   "extract [file]",
		Args:  cobra.ExactArgs(1),
		Run:   extractRun,
		Short: "extract original diagram text from image",
		Long:  "PlantUML embeds images with the original text — either SVG comment or PNG metadata",
	}
	cmd.Flags().BoolVarP(&extractShort, "short", "s", false, "display shorter, compressed diagram instead of full text")
	rootCmd.AddCommand(&cmd)
}

func extractRun(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		args = []string{"-"}
	}
	var data []byte
	var err error
	switch args[0] {
	case "-":
		data, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			fatalf("unable to read stream: %v", err)
		}
	default:
		data, err = ioutil.ReadFile(args[0])
		if err != nil {
			fatalf("unable to read file: %v", err)
		}
	}

	client, err := getClient()
	if err != nil {
		fatalf("unable to connect to server: %v", err)
	}

	resp, err := client.Extract(context.Background(), &pb.ExtractRequest{Data: data})
	if err != nil {
		fatalf("failed to extract: %v", err)
	}

	if extractShort {
		fmt.Print(resp.Diagram.Full)
	} else {
		fmt.Print(resp.Diagram.Short)
	}
}
