package main

import (
	"fmt"
	"os"

	"github.com/coxley/pmlproxy/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
