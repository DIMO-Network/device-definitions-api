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
	GetDeviceDefinitionByID(ctx context.Context, in *GetDeviceDefinitionRequest, opts ...grpc.CallOption) (*GetDeviceDefinitionResponse, error)
	GetDeviceDefinitionByMMY(ctx context.Context, in *GetDeviceDefinitionByMMYRequest, opts ...grpc.CallOption) (*GetDeviceDefinitionItemResponse, error)
	GetIntegrations(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*GetIntegrationResponse, error)
	GetDeviceDefinitionIntegration(ctx context.Context, in *GetDeviceDefinitionIntegrationRequest, opts ...grpc.CallOption) (*GetDeviceDefinitionIntegrationResponse, error)
	CreateDeviceDefinition(ctx context.Context, in *CreateDeviceDefinitionRequest, opts ...grpc.CallOption) (*CreateDeviceDefinitionResponse, error)
	CreateDeviceIntegration(ctx context.Context, in *CreateDeviceIntegrationRequest, opts ...grpc.CallOption) (*CreateDeviceIntegrationResponse, error)
	UpdateDeviceDefinition(ctx context.Context, in *UpdateDeviceDefinitionRequest, opts ...grpc.CallOption) (*UpdateDeviceDefinitionResponse, error)
	SetDeviceDefinitionImage(ctx context.Context, in *UpdateDeviceDefinitionImageRequest, opts ...grpc.CallOption) (*UpdateDeviceDefinitionResponse, error)
	GetDeviceDefinitionAll(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*GetDeviceDefinitionAllResponse, error)
}

type deviceDefinitionServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDeviceDefinitionServiceClient(cc grpc.ClientConnInterface) DeviceDefinitionServiceClient {
	return &deviceDefinitionServiceClient{cc}
}

func (c *deviceDefinitionServiceClient) GetDeviceDefinitionByID(ctx context.Context, in *GetDeviceDefinitionRequest, opts ...grpc.CallOption) (*GetDeviceDefinitionResponse, error) {
	out := new(GetDeviceDefinitionResponse)
	err := c.cc.Invoke(ctx, "/grpc.DeviceDefinitionService/GetDeviceDefinitionByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceDefinitionServiceClient) GetDeviceDefinitionByMMY(ctx context.Context, in *GetDeviceDefinitionByMMYRequest, opts ...grpc.CallOption) (*GetDeviceDefinitionItemResponse, error) {
	out := new(GetDeviceDefinitionItemResponse)
	err := c.cc.Invoke(ctx, "/grpc.DeviceDefinitionService/GetDeviceDefinitionByMMY", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceDefinitionServiceClient) GetIntegrations(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*GetIntegrationResponse, error) {
	out := new(GetIntegrationResponse)
	err := c.cc.Invoke(ctx, "/grpc.DeviceDefinitionService/GetIntegrations", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceDefinitionServiceClient) GetDeviceDefinitionIntegration(ctx context.Context, in *GetDeviceDefinitionIntegrationRequest, opts ...grpc.CallOption) (*GetDeviceDefinitionIntegrationResponse, error) {
	out := new(GetDeviceDefinitionIntegrationResponse)
	err := c.cc.Invoke(ctx, "/grpc.DeviceDefinitionService/GetDeviceDefinitionIntegration", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceDefinitionServiceClient) CreateDeviceDefinition(ctx context.Context, in *CreateDeviceDefinitionRequest, opts ...grpc.CallOption) (*CreateDeviceDefinitionResponse, error) {
	out := new(CreateDeviceDefinitionResponse)
	err := c.cc.Invoke(ctx, "/grpc.DeviceDefinitionService/CreateDeviceDefinition", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceDefinitionServiceClient) CreateDeviceIntegration(ctx context.Context, in *CreateDeviceIntegrationRequest, opts ...grpc.CallOption) (*CreateDeviceIntegrationResponse, error) {
	out := new(CreateDeviceIntegrationResponse)
	err := c.cc.Invoke(ctx, "/grpc.DeviceDefinitionService/CreateDeviceIntegration", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceDefinitionServiceClient) UpdateDeviceDefinition(ctx context.Context, in *UpdateDeviceDefinitionRequest, opts ...grpc.CallOption) (*UpdateDeviceDefinitionResponse, error) {
	out := new(UpdateDeviceDefinitionResponse)
	err := c.cc.Invoke(ctx, "/grpc.DeviceDefinitionService/UpdateDeviceDefinition", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceDefinitionServiceClient) SetDeviceDefinitionImage(ctx context.Context, in *UpdateDeviceDefinitionImageRequest, opts ...grpc.CallOption) (*UpdateDeviceDefinitionResponse, error) {
	out := new(UpdateDeviceDefinitionResponse)
	err := c.cc.Invoke(ctx, "/grpc.DeviceDefinitionService/SetDeviceDefinitionImage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *deviceDefinitionServiceClient) GetDeviceDefinitionAll(ctx context.Context, in *EmptyRequest, opts ...grpc.CallOption) (*GetDeviceDefinitionAllResponse, error) {
	out := new(GetDeviceDefinitionAllResponse)
	err := c.cc.Invoke(ctx, "/grpc.DeviceDefinitionService/GetDeviceDefinitionAll", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DeviceDefinitionServiceServer is the server API for DeviceDefinitionService service.
// All implementations must embed UnimplementedDeviceDefinitionServiceServer
// for forward compatibility
type DeviceDefinitionServiceServer interface {
	GetDeviceDefinitionByID(context.Context, *GetDeviceDefinitionRequest) (*GetDeviceDefinitionResponse, error)
	GetDeviceDefinitionByMMY(context.Context, *GetDeviceDefinitionByMMYRequest) (*GetDeviceDefinitionItemResponse, error)
	GetIntegrations(context.Context, *EmptyRequest) (*GetIntegrationResponse, error)
	GetDeviceDefinitionIntegration(context.Context, *GetDeviceDefinitionIntegrationRequest) (*GetDeviceDefinitionIntegrationResponse, error)
	CreateDeviceDefinition(context.Context, *CreateDeviceDefinitionRequest) (*CreateDeviceDefinitionResponse, error)
	CreateDeviceIntegration(context.Context, *CreateDeviceIntegrationRequest) (*CreateDeviceIntegrationResponse, error)
	UpdateDeviceDefinition(context.Context, *UpdateDeviceDefinitionRequest) (*UpdateDeviceDefinitionResponse, error)
	SetDeviceDefinitionImage(context.Context, *UpdateDeviceDefinitionImageRequest) (*UpdateDeviceDefinitionResponse, error)
	GetDeviceDefinitionAll(context.Context, *EmptyRequest) (*GetDeviceDefinitionAllResponse, error)
	mustEmbedUnimplementedDeviceDefinitionServiceServer()
}

// UnimplementedDeviceDefinitionServiceServer must be embedded to have forward compatible implementations.
type UnimplementedDeviceDefinitionServiceServer struct {
}

func (UnimplementedDeviceDefinitionServiceServer) GetDeviceDefinitionByID(context.Context, *GetDeviceDefinitionRequest) (*GetDeviceDefinitionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDeviceDefinitionByID not implemented")
}
func (UnimplementedDeviceDefinitionServiceServer) GetDeviceDefinitionByMMY(context.Context, *GetDeviceDefinitionByMMYRequest) (*GetDeviceDefinitionItemResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDeviceDefinitionByMMY not implemented")
}
func (UnimplementedDeviceDefinitionServiceServer) GetIntegrations(context.Context, *EmptyRequest) (*GetIntegrationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetIntegrations not implemented")
}
func (UnimplementedDeviceDefinitionServiceServer) GetDeviceDefinitionIntegration(context.Context, *GetDeviceDefinitionIntegrationRequest) (*GetDeviceDefinitionIntegrationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDeviceDefinitionIntegration not implemented")
}
func (UnimplementedDeviceDefinitionServiceServer) CreateDeviceDefinition(context.Context, *CreateDeviceDefinitionRequest) (*CreateDeviceDefinitionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateDeviceDefinition not implemented")
}
func (UnimplementedDeviceDefinitionServiceServer) CreateDeviceIntegration(context.Context, *CreateDeviceIntegrationRequest) (*CreateDeviceIntegrationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateDeviceIntegration not implemented")
}
func (UnimplementedDeviceDefinitionServiceServer) UpdateDeviceDefinition(context.Context, *UpdateDeviceDefinitionRequest) (*UpdateDeviceDefinitionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateDeviceDefinition not implemented")
}
func (UnimplementedDeviceDefinitionServiceServer) SetDeviceDefinitionImage(context.Context, *UpdateDeviceDefinitionImageRequest) (*UpdateDeviceDefinitionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetDeviceDefinitionImage not implemented")
}
func (UnimplementedDeviceDefinitionServiceServer) GetDeviceDefinitionAll(context.Context, *EmptyRequest) (*GetDeviceDefinitionAllResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDeviceDefinitionAll not implemented")
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

func _DeviceDefinitionService_GetDeviceDefinitionByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDeviceDefinitionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceDefinitionServiceServer).GetDeviceDefinitionByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.DeviceDefinitionService/GetDeviceDefinitionByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceDefinitionServiceServer).GetDeviceDefinitionByID(ctx, req.(*GetDeviceDefinitionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceDefinitionService_GetDeviceDefinitionByMMY_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDeviceDefinitionByMMYRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceDefinitionServiceServer).GetDeviceDefinitionByMMY(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.DeviceDefinitionService/GetDeviceDefinitionByMMY",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceDefinitionServiceServer).GetDeviceDefinitionByMMY(ctx, req.(*GetDeviceDefinitionByMMYRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceDefinitionService_GetIntegrations_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceDefinitionServiceServer).GetIntegrations(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.DeviceDefinitionService/GetIntegrations",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceDefinitionServiceServer).GetIntegrations(ctx, req.(*EmptyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceDefinitionService_GetDeviceDefinitionIntegration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDeviceDefinitionIntegrationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceDefinitionServiceServer).GetDeviceDefinitionIntegration(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.DeviceDefinitionService/GetDeviceDefinitionIntegration",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceDefinitionServiceServer).GetDeviceDefinitionIntegration(ctx, req.(*GetDeviceDefinitionIntegrationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceDefinitionService_CreateDeviceDefinition_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateDeviceDefinitionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceDefinitionServiceServer).CreateDeviceDefinition(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.DeviceDefinitionService/CreateDeviceDefinition",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceDefinitionServiceServer).CreateDeviceDefinition(ctx, req.(*CreateDeviceDefinitionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceDefinitionService_CreateDeviceIntegration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateDeviceIntegrationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceDefinitionServiceServer).CreateDeviceIntegration(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.DeviceDefinitionService/CreateDeviceIntegration",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceDefinitionServiceServer).CreateDeviceIntegration(ctx, req.(*CreateDeviceIntegrationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceDefinitionService_UpdateDeviceDefinition_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateDeviceDefinitionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceDefinitionServiceServer).UpdateDeviceDefinition(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.DeviceDefinitionService/UpdateDeviceDefinition",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceDefinitionServiceServer).UpdateDeviceDefinition(ctx, req.(*UpdateDeviceDefinitionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceDefinitionService_SetDeviceDefinitionImage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateDeviceDefinitionImageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceDefinitionServiceServer).SetDeviceDefinitionImage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.DeviceDefinitionService/SetDeviceDefinitionImage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceDefinitionServiceServer).SetDeviceDefinitionImage(ctx, req.(*UpdateDeviceDefinitionImageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DeviceDefinitionService_GetDeviceDefinitionAll_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EmptyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DeviceDefinitionServiceServer).GetDeviceDefinitionAll(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.DeviceDefinitionService/GetDeviceDefinitionAll",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DeviceDefinitionServiceServer).GetDeviceDefinitionAll(ctx, req.(*EmptyRequest))
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
			MethodName: "GetDeviceDefinitionByID",
			Handler:    _DeviceDefinitionService_GetDeviceDefinitionByID_Handler,
		},
		{
			MethodName: "GetDeviceDefinitionByMMY",
			Handler:    _DeviceDefinitionService_GetDeviceDefinitionByMMY_Handler,
		},
		{
			MethodName: "GetIntegrations",
			Handler:    _DeviceDefinitionService_GetIntegrations_Handler,
		},
		{
			MethodName: "GetDeviceDefinitionIntegration",
			Handler:    _DeviceDefinitionService_GetDeviceDefinitionIntegration_Handler,
		},
		{
			MethodName: "CreateDeviceDefinition",
			Handler:    _DeviceDefinitionService_CreateDeviceDefinition_Handler,
		},
		{
			MethodName: "CreateDeviceIntegration",
			Handler:    _DeviceDefinitionService_CreateDeviceIntegration_Handler,
		},
		{
			MethodName: "UpdateDeviceDefinition",
			Handler:    _DeviceDefinitionService_UpdateDeviceDefinition_Handler,
		},
		{
			MethodName: "SetDeviceDefinitionImage",
			Handler:    _DeviceDefinitionService_SetDeviceDefinitionImage_Handler,
		},
		{
			MethodName: "GetDeviceDefinitionAll",
			Handler:    _DeviceDefinitionService_GetDeviceDefinitionAll_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/grpc/device_definition.proto",
}
