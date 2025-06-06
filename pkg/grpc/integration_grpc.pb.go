// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.26.1
// source: pkg/grpc/integration.proto

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
	IntegrationService_GetIntegrationOptions_FullMethodName = "/grpc.IntegrationService/GetIntegrationOptions"
)

// IntegrationServiceClient is the client API for IntegrationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IntegrationServiceClient interface {
	// GetIntegrationOptions for dropdowns in explorer makes page, get by makeId
	GetIntegrationOptions(ctx context.Context, in *GetIntegrationOptionsRequest, opts ...grpc.CallOption) (*GetIntegrationOptionsResponse, error)
}

type integrationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewIntegrationServiceClient(cc grpc.ClientConnInterface) IntegrationServiceClient {
	return &integrationServiceClient{cc}
}

func (c *integrationServiceClient) GetIntegrationOptions(ctx context.Context, in *GetIntegrationOptionsRequest, opts ...grpc.CallOption) (*GetIntegrationOptionsResponse, error) {
	out := new(GetIntegrationOptionsResponse)
	err := c.cc.Invoke(ctx, IntegrationService_GetIntegrationOptions_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IntegrationServiceServer is the server API for IntegrationService service.
// All implementations must embed UnimplementedIntegrationServiceServer
// for forward compatibility
type IntegrationServiceServer interface {
	// GetIntegrationOptions for dropdowns in explorer makes page, get by makeId
	GetIntegrationOptions(context.Context, *GetIntegrationOptionsRequest) (*GetIntegrationOptionsResponse, error)
	mustEmbedUnimplementedIntegrationServiceServer()
}

// UnimplementedIntegrationServiceServer must be embedded to have forward compatible implementations.
type UnimplementedIntegrationServiceServer struct {
}

func (UnimplementedIntegrationServiceServer) GetIntegrationOptions(context.Context, *GetIntegrationOptionsRequest) (*GetIntegrationOptionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetIntegrationOptions not implemented")
}
func (UnimplementedIntegrationServiceServer) mustEmbedUnimplementedIntegrationServiceServer() {}

// UnsafeIntegrationServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to IntegrationServiceServer will
// result in compilation errors.
type UnsafeIntegrationServiceServer interface {
	mustEmbedUnimplementedIntegrationServiceServer()
}

func RegisterIntegrationServiceServer(s grpc.ServiceRegistrar, srv IntegrationServiceServer) {
	s.RegisterService(&IntegrationService_ServiceDesc, srv)
}

func _IntegrationService_GetIntegrationOptions_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetIntegrationOptionsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntegrationServiceServer).GetIntegrationOptions(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IntegrationService_GetIntegrationOptions_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntegrationServiceServer).GetIntegrationOptions(ctx, req.(*GetIntegrationOptionsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// IntegrationService_ServiceDesc is the grpc.ServiceDesc for IntegrationService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var IntegrationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.IntegrationService",
	HandlerType: (*IntegrationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetIntegrationOptions",
			Handler:    _IntegrationService_GetIntegrationOptions_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/grpc/integration.proto",
}
