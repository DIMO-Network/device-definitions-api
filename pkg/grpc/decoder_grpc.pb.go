// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.12
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

const (
	VinDecoderService_DecodeVin_FullMethodName = "/grpc.VinDecoderService/DecodeVin"
)

// VinDecoderServiceClient is the client API for VinDecoderService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type VinDecoderServiceClient interface {
	DecodeVin(ctx context.Context, in *DecodeVinRequest, opts ...grpc.CallOption) (*DecodeVinResponse, error)
}

type vinDecoderServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewVinDecoderServiceClient(cc grpc.ClientConnInterface) VinDecoderServiceClient {
	return &vinDecoderServiceClient{cc}
}

func (c *vinDecoderServiceClient) DecodeVin(ctx context.Context, in *DecodeVinRequest, opts ...grpc.CallOption) (*DecodeVinResponse, error) {
	out := new(DecodeVinResponse)
	err := c.cc.Invoke(ctx, VinDecoderService_DecodeVin_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VinDecoderServiceServer is the server API for VinDecoderService service.
// All implementations must embed UnimplementedVinDecoderServiceServer
// for forward compatibility
type VinDecoderServiceServer interface {
	DecodeVin(context.Context, *DecodeVinRequest) (*DecodeVinResponse, error)
	mustEmbedUnimplementedVinDecoderServiceServer()
}

// UnimplementedVinDecoderServiceServer must be embedded to have forward compatible implementations.
type UnimplementedVinDecoderServiceServer struct {
}

func (UnimplementedVinDecoderServiceServer) DecodeVin(context.Context, *DecodeVinRequest) (*DecodeVinResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DecodeVin not implemented")
}
func (UnimplementedVinDecoderServiceServer) mustEmbedUnimplementedVinDecoderServiceServer() {}

// UnsafeVinDecoderServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to VinDecoderServiceServer will
// result in compilation errors.
type UnsafeVinDecoderServiceServer interface {
	mustEmbedUnimplementedVinDecoderServiceServer()
}

func RegisterVinDecoderServiceServer(s grpc.ServiceRegistrar, srv VinDecoderServiceServer) {
	s.RegisterService(&VinDecoderService_ServiceDesc, srv)
}

func _VinDecoderService_DecodeVin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DecodeVinRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VinDecoderServiceServer).DecodeVin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: VinDecoderService_DecodeVin_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VinDecoderServiceServer).DecodeVin(ctx, req.(*DecodeVinRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// VinDecoderService_ServiceDesc is the grpc.ServiceDesc for VinDecoderService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var VinDecoderService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.VinDecoderService",
	HandlerType: (*VinDecoderServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DecodeVin",
			Handler:    _VinDecoderService_DecodeVin_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/grpc/decoder.proto",
}
