// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.7
// source: pkg/grpc/decoder.proto

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

// VINDecoderServiceClient is the client API for VINDecoderService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type VINDecoderServiceClient interface {
	DecodeVIN(ctx context.Context, in *DecodeVINRequest, opts ...grpc.CallOption) (*DecodeVINResponse, error)
}

type vINDecoderServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewVINDecoderServiceClient(cc grpc.ClientConnInterface) VINDecoderServiceClient {
	return &vINDecoderServiceClient{cc}
}

func (c *vINDecoderServiceClient) DecodeVIN(ctx context.Context, in *DecodeVINRequest, opts ...grpc.CallOption) (*DecodeVINResponse, error) {
	out := new(DecodeVINResponse)
	err := c.cc.Invoke(ctx, "/grpc.VINDecoderService/DecodeVIN", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VINDecoderServiceServer is the server API for VINDecoderService service.
// All implementations must embed UnimplementedVINDecoderServiceServer
// for forward compatibility
type VINDecoderServiceServer interface {
	DecodeVIN(context.Context, *DecodeVINRequest) (*DecodeVINResponse, error)
	mustEmbedUnimplementedVINDecoderServiceServer()
}

// UnimplementedVINDecoderServiceServer must be embedded to have forward compatible implementations.
type UnimplementedVINDecoderServiceServer struct {
}

func (UnimplementedVINDecoderServiceServer) DecodeVIN(context.Context, *DecodeVINRequest) (*DecodeVINResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DecodeVIN not implemented")
}
func (UnimplementedVINDecoderServiceServer) mustEmbedUnimplementedVINDecoderServiceServer() {}

// UnsafeVINDecoderServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to VINDecoderServiceServer will
// result in compilation errors.
type UnsafeVINDecoderServiceServer interface {
	mustEmbedUnimplementedVINDecoderServiceServer()
}

func RegisterVINDecoderServiceServer(s grpc.ServiceRegistrar, srv VINDecoderServiceServer) {
	s.RegisterService(&VINDecoderService_ServiceDesc, srv)
}

func _VINDecoderService_DecodeVIN_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DecodeVINRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VINDecoderServiceServer).DecodeVIN(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.VINDecoderService/DecodeVIN",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VINDecoderServiceServer).DecodeVIN(ctx, req.(*DecodeVINRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// VINDecoderService_ServiceDesc is the grpc.ServiceDesc for VINDecoderService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var VINDecoderService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.VINDecoderService",
	HandlerType: (*VINDecoderServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DecodeVIN",
			Handler:    _VINDecoderService_DecodeVIN_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/grpc/decoder.proto",
}
