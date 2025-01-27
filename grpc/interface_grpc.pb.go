// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package DistributedMutualExclusion

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

// DistributedMutualExclusionClient is the client API for DistributedMutualExclusion service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DistributedMutualExclusionClient interface {
	Request(ctx context.Context, in *Request, opts ...grpc.CallOption) (*RequestBack, error)
	Reply(ctx context.Context, in *Reply, opts ...grpc.CallOption) (*ReplyBack, error)
}

type distributedMutualExclusionClient struct {
	cc grpc.ClientConnInterface
}

func NewDistributedMutualExclusionClient(cc grpc.ClientConnInterface) DistributedMutualExclusionClient {
	return &distributedMutualExclusionClient{cc}
}

func (c *distributedMutualExclusionClient) Request(ctx context.Context, in *Request, opts ...grpc.CallOption) (*RequestBack, error) {
	out := new(RequestBack)
	err := c.cc.Invoke(ctx, "/DistributedMutualExclusion.DistributedMutualExclusion/request", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *distributedMutualExclusionClient) Reply(ctx context.Context, in *Reply, opts ...grpc.CallOption) (*ReplyBack, error) {
	out := new(ReplyBack)
	err := c.cc.Invoke(ctx, "/DistributedMutualExclusion.DistributedMutualExclusion/reply", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DistributedMutualExclusionServer is the server API for DistributedMutualExclusion service.
// All implementations must embed UnimplementedDistributedMutualExclusionServer
// for forward compatibility
type DistributedMutualExclusionServer interface {
	Request(context.Context, *Request) (*RequestBack, error)
	Reply(context.Context, *Reply) (*ReplyBack, error)
	mustEmbedUnimplementedDistributedMutualExclusionServer()
}

// UnimplementedDistributedMutualExclusionServer must be embedded to have forward compatible implementations.
type UnimplementedDistributedMutualExclusionServer struct {
}

func (UnimplementedDistributedMutualExclusionServer) Request(context.Context, *Request) (*RequestBack, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Request not implemented")
}
func (UnimplementedDistributedMutualExclusionServer) Reply(context.Context, *Reply) (*ReplyBack, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Reply not implemented")
}
func (UnimplementedDistributedMutualExclusionServer) mustEmbedUnimplementedDistributedMutualExclusionServer() {
}

// UnsafeDistributedMutualExclusionServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DistributedMutualExclusionServer will
// result in compilation errors.
type UnsafeDistributedMutualExclusionServer interface {
	mustEmbedUnimplementedDistributedMutualExclusionServer()
}

func RegisterDistributedMutualExclusionServer(s grpc.ServiceRegistrar, srv DistributedMutualExclusionServer) {
	s.RegisterService(&DistributedMutualExclusion_ServiceDesc, srv)
}

func _DistributedMutualExclusion_Request_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DistributedMutualExclusionServer).Request(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DistributedMutualExclusion.DistributedMutualExclusion/request",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DistributedMutualExclusionServer).Request(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _DistributedMutualExclusion_Reply_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Reply)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DistributedMutualExclusionServer).Reply(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/DistributedMutualExclusion.DistributedMutualExclusion/reply",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DistributedMutualExclusionServer).Reply(ctx, req.(*Reply))
	}
	return interceptor(ctx, in, info, handler)
}

// DistributedMutualExclusion_ServiceDesc is the grpc.ServiceDesc for DistributedMutualExclusion service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DistributedMutualExclusion_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "DistributedMutualExclusion.DistributedMutualExclusion",
	HandlerType: (*DistributedMutualExclusionServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "request",
			Handler:    _DistributedMutualExclusion_Request_Handler,
		},
		{
			MethodName: "reply",
			Handler:    _DistributedMutualExclusion_Reply_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "grpc/interface.proto",
}
