// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.5
// source: pkg/grpc/device_definition.proto

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

// DeviceDefinitionServiceClient is the client API for DeviceDefinitionService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DeviceDefinitionServiceClient interface {
	GetDeviceDefinitionById(ctx context.Context, in *GetDeviceDefinitionRequest, opts ...grpc.CallOption) (*GetDeviceDefinitionResponse, error)
}

type deviceDefinitionServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDeviceDefinitionServiceClient(cc grpc.ClientConnInterface) DeviceDefinitionServiceClient {
	return &deviceDefinitionServiceClient{cc}
}

func (c *deviceDefinitionServiceClient) GetDeviceDefinitionById(ctx context.Context, in *GetDeviceDefinitionRequest, opts ...grpc.CallOption) (*GetDeviceDefinitionResponse, error) {
	out := new(GetDeviceDefinitionResponse)
	err := c.cc.Invoke(ctx, "/grpc.DeviceDefinitionService/GetDeviceDefinitionById", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DeviceDefinitionServiceServer is the server API for DeviceDefinitionService service.
// All implementations must embed UnimplementedDeviceDefinitionServiceServer
// for forward compatibility
type DeviceDefinitionServiceServer interface {
	GetDeviceDefinitionById(context.Context, *GetDeviceDefinitionRequest) (*GetDeviceDefinitionResponse, error)
	mustEmbedUnimplementedDeviceDefinitionServiceServer()
}

// UnimplementedDeviceDefinitionServiceServer must be embedded to have forward compatible implementations.
type UnimplementedDeviceDefinitionServiceServer struct {
}

func (UnimplementedDeviceDefinitionServiceServer) GetDeviceDefinitionById(context.Context, *GetDeviceDefinitionRequest) (*GetDeviceDefinitionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDeviceDefinitionById not implemented")
}
func (UnimplementedDeviceDefinitionServiceServer) mustEmbedUnimplementedDeviceDefinitionServiceServer() {
}

// UnsafeDeviceDefinitionServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DeviceDefinitionServiceServer will
// result in compilation errors.
type UnsafeDeviceDefinitionServiceServer interface {
	mustEmbedUnimplementedDeviceDefinitionServiceServer()
}

func RegisterDeviceDefinitionServiceServer(s grpc.ServiceRegistrar, srv DeviceDefinitionServiceServer) {
	s.RegisterService(&DeviceDefinitionService_ServiceDesc, srv)
}

func _DeviceDefinitionService_GetDeviceDefinitionById_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDeviceDefinitionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceDefinitionServiceServer).GetDeviceDefinitionById(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.DeviceDefinitionService/GetDeviceDefinitionById",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceDefinitionServiceServer).GetDeviceDefinitionById(ctx, req.(*GetDeviceDefinitionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DeviceDefinitionService_ServiceDesc is the grpc.ServiceDesc for DeviceDefinitionService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DeviceDefinitionService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.DeviceDefinitionService",
	HandlerType: (*DeviceDefinitionServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetDeviceDefinitionById",
			Handler:    _DeviceDefinitionService_GetDeviceDefinitionById_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/grpc/device_definition.proto",
}
