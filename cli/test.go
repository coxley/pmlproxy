// You're responsible for managing the pool
package main

import (
	"fmt"
	"os"

	"github.com/coxley/pmlproxy/server"
)

func main() {
	server.DefaultHandler.GroupCache = true
	srv := &server.Server{Addr: "localhost:9000"}
	err := srv.ListenAndServe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "caught error in server: %v", err)
	}
}
