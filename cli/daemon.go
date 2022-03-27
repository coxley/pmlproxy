package cli

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/coxley/pmlproxy/server"
	"github.com/golang/groupcache"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/golang/glog"

	"github.com/spf13/cobra"
)

var (
	daemonPprof  string
	cacheAddr    string
	groupMembers []string
)

var handler = server.DefaultHandler

func init() {
	cmd := &cobra.Command{
		Use:   "daemon",
		Args:  cobra.ExactArgs(0),
		Run:   daemonRun,
		Short: "Run the PlantUML proxy server",
	}
	rootCmd.AddCommand(cmd)
	flags := cmd.Flags()
	flags.IntVar(&handler.Workers, "workers", handler.Workers, "number of plantuml processes used for rendering")
	flags.StringVar(&daemonPprof, "pprof", "", "enable pprof and listen on addr (eg: :6060")
	flags.StringVar(&handler.JavaExe, "java-path", handler.JavaExe, "path to java")
	flags.StringVar(&handler.PipeDelimiter, "pipe-delimiter", handler.PipeDelimiter, "used by plantuml to separate image results. only need to override if it may be found in your user's diagrams")
	flags.StringVar(&handler.PlantUMLPath, "plantuml-path", handler.PlantUMLPath, "path to plantuml jar")
	flags.StringVar(&handler.SearchPath, "search-path", handler.SearchPath, "path for plantuml to search for modules/themes that we create on start")
	flags.DurationVar(&handler.RenderTimeout, "render-timeout", handler.RenderTimeout, "max time for server to wait on diagram rendering before killing the request")

	flags.StringVarP(&cacheAddr, "cache-addr", "c", "", "Enables groupcache and configures HTTP socket to listen on")
	flags.StringSliceVarP(&groupMembers, "group-member", "g", []string{}, "other participant in the group cache — can specify multiple times")
}

func setupPprof(addr string) {
	glog.Infof("starting pprof server on %s", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		glog.Errorf("failed to start pprof server: %v", err)
	}
}

func setupCache(localAddr string, peers []string) *http.Server {
	pool := groupcache.NewHTTPPoolOpts("http://"+localAddr, &groupcache.HTTPPoolOptions{})
	var peerURLs []string
	// for _, a := range append([]string{localAddr}, peers...) {
	for _, a := range append(peers, localAddr) {
		peerURLs = append(peerURLs, fmt.Sprintf("http://%s", a))
	}
	pool.Set(peerURLs...)
	server := http.Server{Addr: localAddr, Handler: pool}
	go func() {
		glog.Infof("starting groupcache server with these peers: %v", peerURLs)
		if err := server.ListenAndServe(); err != nil {
			glog.Fatal(err)
		}
	}()
	return &server
}

func daemonRun(cmd *cobra.Command, args []string) {
	// Call after handling everything else — and only in paths that use glog —
	// because otherwise it overrides the --help docs
	flag.Parse()

	if daemonPprof != "" {
		go setupPprof(daemonPprof)
	}
	if cacheAddr != "" {
		cacheSrv := setupCache(cacheAddr, groupMembers)
		defer cacheSrv.Shutdown(context.Background())
	}

	server.MakeGRPC = func() *grpc.Server {
		s := grpc.NewServer()
		reflection.Register(s)
		return s
	}
	handler.GroupCache = true
	srv := &server.Server{
		Addr:    addr, // global flag
		Handler: &handler,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, os.Interrupt)

	serverErr := make(chan error, 1)
	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			glog.Error(err)
		}
		serverErr <- err
	}()

	select {
	case err := <-serverErr:
		glog.Infof("caught error in server, shutting down: %v", err)
	case signal := <-sig:
		glog.Infof("caught signal, waiting for requests to finish: %v", signal)
		srv.GracefulStop()
		glog.Infof("all requests finished, shutting down")
	}

	<-serverErr
}
