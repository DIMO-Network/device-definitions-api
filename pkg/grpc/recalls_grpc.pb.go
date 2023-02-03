// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.11
// source: pkg/grpc/recalls.proto

package grpc

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

// RecallsServiceClient is the client API for RecallsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type RecallsServiceClient interface {
	// Deprecated: Do not use.
	GetRecallsByMake(ctx context.Context, in *GetRecallsByMakeRequest, opts ...grpc.CallOption) (*GetRecallsResponse, error)
	GetStreamRecallsByMake(ctx context.Context, in *GetRecallsByMakeRequest, opts ...grpc.CallOption) (RecallsService_GetStreamRecallsByMakeClient, error)
	GetRecallsByModel(ctx context.Context, in *GetRecallsByModelRequest, opts ...grpc.CallOption) (*GetRecallsResponse, error)
}

type recallsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewRecallsServiceClient(cc grpc.ClientConnInterface) RecallsServiceClient {
	return &recallsServiceClient{cc}
}

// Deprecated: Do not use.
func (c *recallsServiceClient) GetRecallsByMake(ctx context.Context, in *GetRecallsByMakeRequest, opts ...grpc.CallOption) (*GetRecallsResponse, error) {
	out := new(GetRecallsResponse)
	err := c.cc.Invoke(ctx, "/grpc.RecallsService/GetRecallsByMake", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *recallsServiceClient) GetStreamRecallsByMake(ctx context.Context, in *GetRecallsByMakeRequest, opts ...grpc.CallOption) (RecallsService_GetStreamRecallsByMakeClient, error) {
	stream, err := c.cc.NewStream(ctx, &RecallsService_ServiceDesc.Streams[0], "/grpc.RecallsService/GetStreamRecallsByMake", opts...)
	if err != nil {
		return nil, err
	}
	x := &recallsServiceGetStreamRecallsByMakeClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type RecallsService_GetStreamRecallsByMakeClient interface {
	Recv() (*RecallItem, error)
	grpc.ClientStream
}

type recallsServiceGetStreamRecallsByMakeClient struct {
	grpc.ClientStream
}

func (x *recallsServiceGetStreamRecallsByMakeClient) Recv() (*RecallItem, error) {
	m := new(RecallItem)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *recallsServiceClient) GetRecallsByModel(ctx context.Context, in *GetRecallsByModelRequest, opts ...grpc.CallOption) (*GetRecallsResponse, error) {
	out := new(GetRecallsResponse)
	err := c.cc.Invoke(ctx, "/grpc.RecallsService/GetRecallsByModel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// RecallsServiceServer is the server API for RecallsService service.
// All implementations must embed UnimplementedRecallsServiceServer
// for forward compatibility
type RecallsServiceServer interface {
	// Deprecated: Do not use.
	GetRecallsByMake(context.Context, *GetRecallsByMakeRequest) (*GetRecallsResponse, error)
	GetStreamRecallsByMake(*GetRecallsByMakeRequest, RecallsService_GetStreamRecallsByMakeServer) error
	GetRecallsByModel(context.Context, *GetRecallsByModelRequest) (*GetRecallsResponse, error)
	mustEmbedUnimplementedRecallsServiceServer()
}

// UnimplementedRecallsServiceServer must be embedded to have forward compatible implementations.
type UnimplementedRecallsServiceServer struct {
}

func (UnimplementedRecallsServiceServer) GetRecallsByMake(context.Context, *GetRecallsByMakeRequest) (*GetRecallsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRecallsByMake not implemented")
}
func (UnimplementedRecallsServiceServer) GetStreamRecallsByMake(*GetRecallsByMakeRequest, RecallsService_GetStreamRecallsByMakeServer) error {
	return status.Errorf(codes.Unimplemented, "method GetStreamRecallsByMake not implemented")
}
func (UnimplementedRecallsServiceServer) GetRecallsByModel(context.Context, *GetRecallsByModelRequest) (*GetRecallsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRecallsByModel not implemented")
}
func (UnimplementedRecallsServiceServer) mustEmbedUnimplementedRecallsServiceServer() {}

// UnsafeRecallsServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to RecallsServiceServer will
// result in compilation errors.
type UnsafeRecallsServiceServer interface {
	mustEmbedUnimplementedRecallsServiceServer()
}

func RegisterRecallsServiceServer(s grpc.ServiceRegistrar, srv RecallsServiceServer) {
	s.RegisterService(&RecallsService_ServiceDesc, srv)
}

func _RecallsService_GetRecallsByMake_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRecallsByMakeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RecallsServiceServer).GetRecallsByMake(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.RecallsService/GetRecallsByMake",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RecallsServiceServer).GetRecallsByMake(ctx, req.(*GetRecallsByMakeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _RecallsService_GetStreamRecallsByMake_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetRecallsByMakeRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(RecallsServiceServer).GetStreamRecallsByMake(m, &recallsServiceGetStreamRecallsByMakeServer{stream})
}

type RecallsService_GetStreamRecallsByMakeServer interface {
	Send(*RecallItem) error
	grpc.ServerStream
}

type recallsServiceGetStreamRecallsByMakeServer struct {
	grpc.ServerStream
}

func (x *recallsServiceGetStreamRecallsByMakeServer) Send(m *RecallItem) error {
	return x.ServerStream.SendMsg(m)
}

func _RecallsService_GetRecallsByModel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRecallsByModelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(RecallsServiceServer).GetRecallsByModel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.RecallsService/GetRecallsByModel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(RecallsServiceServer).GetRecallsByModel(ctx, req.(*GetRecallsByModelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// RecallsService_ServiceDesc is the grpc.ServiceDesc for RecallsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var RecallsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.RecallsService",
	HandlerType: (*RecallsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetRecallsByMake",
			Handler:    _RecallsService_GetRecallsByMake_Handler,
		},
		{
			MethodName: "GetRecallsByModel",
			Handler:    _RecallsService_GetRecallsByModel_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetStreamRecallsByMake",
			Handler:       _RecallsService_GetStreamRecallsByMake_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "pkg/grpc/recalls.proto",
}
