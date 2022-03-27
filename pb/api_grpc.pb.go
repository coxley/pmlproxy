// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package pb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// PlantUMLClient is the client API for PlantUML service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PlantUMLClient interface {
	// Render diagram into an image
	Render(ctx context.Context, in *RenderRequest, opts ...grpc.CallOption) (*RenderResponse, error)
	// Shorten diagram source or expand encoded text.
	//
	// Implemented server-side to avoid penalty of proxying to plantuml
	Shorten(ctx context.Context, in *ShortenRequest, opts ...grpc.CallOption) (*ShortenResponse, error)
	Expand(ctx context.Context, in *ExpandRequest, opts ...grpc.CallOption) (*ExpandResponse, error)
	// Extract the diagram source from a rendered image
	//
	// Works for both PNG and SVG. Format is auto-detected.
	//
	// NOTE: Whitespace isn't guaranteed to be stable so the encoded result
	// before -> after may differ.
	Extract(ctx context.Context, in *ExtractRequest, opts ...grpc.CallOption) (*ExtractResponse, error)
}

type plantUMLClient struct {
	cc grpc.ClientConnInterface
}

func NewPlantUMLClient(cc grpc.ClientConnInterface) PlantUMLClient {
	return &plantUMLClient{cc}
}

func (c *plantUMLClient) Render(ctx context.Context, in *RenderRequest, opts ...grpc.CallOption) (*RenderResponse, error) {
	out := new(RenderResponse)
	err := c.cc.Invoke(ctx, "/pb.PlantUML/Render", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *plantUMLClient) Shorten(ctx context.Context, in *ShortenRequest, opts ...grpc.CallOption) (*ShortenResponse, error) {
	out := new(ShortenResponse)
	err := c.cc.Invoke(ctx, "/pb.PlantUML/Shorten", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *plantUMLClient) Expand(ctx context.Context, in *ExpandRequest, opts ...grpc.CallOption) (*ExpandResponse, error) {
	out := new(ExpandResponse)
	err := c.cc.Invoke(ctx, "/pb.PlantUML/Expand", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *plantUMLClient) Extract(ctx context.Context, in *ExtractRequest, opts ...grpc.CallOption) (*ExtractResponse, error) {
	out := new(ExtractResponse)
	err := c.cc.Invoke(ctx, "/pb.PlantUML/Extract", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PlantUMLServer is the server API for PlantUML service.
// All implementations must embed UnimplementedPlantUMLServer
// for forward compatibility
type PlantUMLServer interface {
	// Render diagram into an image
	Render(context.Context, *RenderRequest) (*RenderResponse, error)
	// Shorten diagram source or expand encoded text.
	//
	// Implemented server-side to avoid penalty of proxying to plantuml
	Shorten(context.Context, *ShortenRequest) (*ShortenResponse, error)
	Expand(context.Context, *ExpandRequest) (*ExpandResponse, error)
	// Extract the diagram source from a rendered image
	//
	// Works for both PNG and SVG. Format is auto-detected.
	//
	// NOTE: Whitespace isn't guaranteed to be stable so the encoded result
	// before -> after may differ.
	Extract(context.Context, *ExtractRequest) (*ExtractResponse, error)
	mustEmbedUnimplementedPlantUMLServer()
}

// UnimplementedPlantUMLServer must be embedded to have forward compatible implementations.
type UnimplementedPlantUMLServer struct {
}

func (UnimplementedPlantUMLServer) Render(context.Context, *RenderRequest) (*RenderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Render not implemented")
}
func (UnimplementedPlantUMLServer) Shorten(context.Context, *ShortenRequest) (*ShortenResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Shorten not implemented")
}
func (UnimplementedPlantUMLServer) Expand(context.Context, *ExpandRequest) (*ExpandResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Expand not implemented")
}
func (UnimplementedPlantUMLServer) Extract(context.Context, *ExtractRequest) (*ExtractResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Extract not implemented")
}
func (UnimplementedPlantUMLServer) mustEmbedUnimplementedPlantUMLServer() {}

// UnsafePlantUMLServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PlantUMLServer will
// result in compilation errors.
type UnsafePlantUMLServer interface {
	mustEmbedUnimplementedPlantUMLServer()
}

func RegisterPlantUMLServer(s grpc.ServiceRegistrar, srv PlantUMLServer) {
	s.RegisterService(&PlantUML_ServiceDesc, srv)
}

func _PlantUML_Render_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RenderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PlantUMLServer).Render(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.PlantUML/Render",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PlantUMLServer).Render(ctx, req.(*RenderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PlantUML_Shorten_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ShortenRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PlantUMLServer).Shorten(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.PlantUML/Shorten",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PlantUMLServer).Shorten(ctx, req.(*ShortenRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PlantUML_Expand_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExpandRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PlantUMLServer).Expand(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.PlantUML/Expand",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PlantUMLServer).Expand(ctx, req.(*ExpandRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _PlantUML_Extract_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExtractRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PlantUMLServer).Extract(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.PlantUML/Extract",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PlantUMLServer).Extract(ctx, req.(*ExtractRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// PlantUML_ServiceDesc is the grpc.ServiceDesc for PlantUML service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var PlantUML_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.PlantUML",
	HandlerType: (*PlantUMLServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Render",
			Handler:    _PlantUML_Render_Handler,
		},
		{
			MethodName: "Shorten",
			Handler:    _PlantUML_Shorten_Handler,
		},
		{
			MethodName: "Expand",
			Handler:    _PlantUML_Expand_Handler,
		},
		{
			MethodName: "Extract",
			Handler:    _PlantUML_Extract_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pb/api.proto",
}
