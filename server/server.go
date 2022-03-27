package server

import (
	"context"
	"net"

	"github.com/coxley/pmlproxy/pb"
	"github.com/golang/glog"
	"google.golang.org/grpc"
)

// Server wraps around our handler and gRPC server
//
// Handler can be used without the Server for applications that want more
// control. To customize your gRPC server options, override MakeGRPC.
// Server.ListenAndServe() will take care of registering it.
type Server struct {
	*grpc.Server // set by ListenAndServe
	Handler      Handler
	Addr         string
}

var MakeGRPC = func() *grpc.Server {
	return grpc.NewServer()
}

func (s *Server) ListenAndServe() error {
	lis, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	if s.Handler == nil {
		s.Handler = &DefaultHandler
	}
	s.Server = MakeGRPC()
	pb.RegisterPlantUMLServer(s.Server, s.Handler)

	// Initialize workers and shut them down on-exit
	ctx, cancel := context.WithCancel(context.Background())
	go s.Handler.ManageWorkers(ctx)
	defer cancel()

	glog.Infof("Starting server on %s", s.Addr)
	return s.Server.Serve(lis)
}
