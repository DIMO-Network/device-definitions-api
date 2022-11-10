// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.7
// source: pkg/grpc/integration.proto

package grpc

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// IntegrationServiceClient is the client API for IntegrationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IntegrationServiceClient interface {
	GetDeviceCompatibilities(ctx context.Context, in *GetDeviceCompatibilityListRequest, opts ...grpc.CallOption) (*GetDeviceCompatibilityListResponse, error)
	GetIntegrationFeatureByID(ctx context.Context, in *GetIntegrationFeatureByIDRequest, opts ...grpc.CallOption) (*GetIntegrationFeatureResponse, error)
	GetIntegrationFeatures(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetIntegrationFeatureListResponse, error)
	CreateIntegrationFeature(ctx context.Context, in *CreateOrUpdateIntegrationFeatureRequest, opts ...grpc.CallOption) (*IntegrationBaseResponse, error)
	UpdateIntegrationFeature(ctx context.Context, in *CreateOrUpdateIntegrationFeatureRequest, opts ...grpc.CallOption) (*IntegrationBaseResponse, error)
	DeleteIntegrationFeature(ctx context.Context, in *DeleteIntegrationFeatureRequest, opts ...grpc.CallOption) (*IntegrationBaseResponse, error)
}

type integrationServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewIntegrationServiceClient(cc grpc.ClientConnInterface) IntegrationServiceClient {
	return &integrationServiceClient{cc}
}

func (c *integrationServiceClient) GetDeviceCompatibilities(ctx context.Context, in *GetDeviceCompatibilityListRequest, opts ...grpc.CallOption) (*GetDeviceCompatibilityListResponse, error) {
	out := new(GetDeviceCompatibilityListResponse)
	err := c.cc.Invoke(ctx, "/grpc.IntegrationService/GetDeviceCompatibilities", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *integrationServiceClient) GetIntegrationFeatureByID(ctx context.Context, in *GetIntegrationFeatureByIDRequest, opts ...grpc.CallOption) (*GetIntegrationFeatureResponse, error) {
	out := new(GetIntegrationFeatureResponse)
	err := c.cc.Invoke(ctx, "/grpc.IntegrationService/GetIntegrationFeatureByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *integrationServiceClient) GetIntegrationFeatures(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetIntegrationFeatureListResponse, error) {
	out := new(GetIntegrationFeatureListResponse)
	err := c.cc.Invoke(ctx, "/grpc.IntegrationService/GetIntegrationFeatures", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *integrationServiceClient) CreateIntegrationFeature(ctx context.Context, in *CreateOrUpdateIntegrationFeatureRequest, opts ...grpc.CallOption) (*IntegrationBaseResponse, error) {
	out := new(IntegrationBaseResponse)
	err := c.cc.Invoke(ctx, "/grpc.IntegrationService/CreateIntegrationFeature", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *integrationServiceClient) UpdateIntegrationFeature(ctx context.Context, in *CreateOrUpdateIntegrationFeatureRequest, opts ...grpc.CallOption) (*IntegrationBaseResponse, error) {
	out := new(IntegrationBaseResponse)
	err := c.cc.Invoke(ctx, "/grpc.IntegrationService/UpdateIntegrationFeature", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *integrationServiceClient) DeleteIntegrationFeature(ctx context.Context, in *DeleteIntegrationFeatureRequest, opts ...grpc.CallOption) (*IntegrationBaseResponse, error) {
	out := new(IntegrationBaseResponse)
	err := c.cc.Invoke(ctx, "/grpc.IntegrationService/DeleteIntegrationFeature", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IntegrationServiceServer is the server API for IntegrationService service.
// All implementations must embed UnimplementedIntegrationServiceServer
// for forward compatibility
type IntegrationServiceServer interface {
	GetDeviceCompatibilities(context.Context, *GetDeviceCompatibilityListRequest) (*GetDeviceCompatibilityListResponse, error)
	GetIntegrationFeatureByID(context.Context, *GetIntegrationFeatureByIDRequest) (*GetIntegrationFeatureResponse, error)
	GetIntegrationFeatures(context.Context, *emptypb.Empty) (*GetIntegrationFeatureListResponse, error)
	CreateIntegrationFeature(context.Context, *CreateOrUpdateIntegrationFeatureRequest) (*IntegrationBaseResponse, error)
	UpdateIntegrationFeature(context.Context, *CreateOrUpdateIntegrationFeatureRequest) (*IntegrationBaseResponse, error)
	DeleteIntegrationFeature(context.Context, *DeleteIntegrationFeatureRequest) (*IntegrationBaseResponse, error)
	mustEmbedUnimplementedIntegrationServiceServer()
}

// UnimplementedIntegrationServiceServer must be embedded to have forward compatible implementations.
type UnimplementedIntegrationServiceServer struct {
}

func (UnimplementedIntegrationServiceServer) GetDeviceCompatibilities(context.Context, *GetDeviceCompatibilityListRequest) (*GetDeviceCompatibilityListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDeviceCompatibilities not implemented")
}
func (UnimplementedIntegrationServiceServer) GetIntegrationFeatureByID(context.Context, *GetIntegrationFeatureByIDRequest) (*GetIntegrationFeatureResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetIntegrationFeatureByID not implemented")
}
func (UnimplementedIntegrationServiceServer) GetIntegrationFeatures(context.Context, *emptypb.Empty) (*GetIntegrationFeatureListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetIntegrationFeatures not implemented")
}
func (UnimplementedIntegrationServiceServer) CreateIntegrationFeature(context.Context, *CreateOrUpdateIntegrationFeatureRequest) (*IntegrationBaseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateIntegrationFeature not implemented")
}
func (UnimplementedIntegrationServiceServer) UpdateIntegrationFeature(context.Context, *CreateOrUpdateIntegrationFeatureRequest) (*IntegrationBaseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateIntegrationFeature not implemented")
}
func (UnimplementedIntegrationServiceServer) DeleteIntegrationFeature(context.Context, *DeleteIntegrationFeatureRequest) (*IntegrationBaseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteIntegrationFeature not implemented")
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

func _IntegrationService_GetDeviceCompatibilities_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDeviceCompatibilityListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntegrationServiceServer).GetDeviceCompatibilities(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.IntegrationService/GetDeviceCompatibilities",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntegrationServiceServer).GetDeviceCompatibilities(ctx, req.(*GetDeviceCompatibilityListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntegrationService_GetIntegrationFeatureByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetIntegrationFeatureByIDRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntegrationServiceServer).GetIntegrationFeatureByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.IntegrationService/GetIntegrationFeatureByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntegrationServiceServer).GetIntegrationFeatureByID(ctx, req.(*GetIntegrationFeatureByIDRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntegrationService_GetIntegrationFeatures_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntegrationServiceServer).GetIntegrationFeatures(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.IntegrationService/GetIntegrationFeatures",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntegrationServiceServer).GetIntegrationFeatures(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntegrationService_CreateIntegrationFeature_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateOrUpdateIntegrationFeatureRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntegrationServiceServer).CreateIntegrationFeature(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.IntegrationService/CreateIntegrationFeature",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntegrationServiceServer).CreateIntegrationFeature(ctx, req.(*CreateOrUpdateIntegrationFeatureRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntegrationService_UpdateIntegrationFeature_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateOrUpdateIntegrationFeatureRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntegrationServiceServer).UpdateIntegrationFeature(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.IntegrationService/UpdateIntegrationFeature",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntegrationServiceServer).UpdateIntegrationFeature(ctx, req.(*CreateOrUpdateIntegrationFeatureRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntegrationService_DeleteIntegrationFeature_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteIntegrationFeatureRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntegrationServiceServer).DeleteIntegrationFeature(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/grpc.IntegrationService/DeleteIntegrationFeature",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntegrationServiceServer).DeleteIntegrationFeature(ctx, req.(*DeleteIntegrationFeatureRequest))
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
			MethodName: "GetDeviceCompatibilities",
			Handler:    _IntegrationService_GetDeviceCompatibilities_Handler,
		},
		{
			MethodName: "GetIntegrationFeatureByID",
			Handler:    _IntegrationService_GetIntegrationFeatureByID_Handler,
		},
		{
			MethodName: "GetIntegrationFeatures",
			Handler:    _IntegrationService_GetIntegrationFeatures_Handler,
		},
		{
			MethodName: "CreateIntegrationFeature",
			Handler:    _IntegrationService_CreateIntegrationFeature_Handler,
		},
		{
			MethodName: "UpdateIntegrationFeature",
			Handler:    _IntegrationService_UpdateIntegrationFeature_Handler,
		},
		{
			MethodName: "DeleteIntegrationFeature",
			Handler:    _IntegrationService_DeleteIntegrationFeature_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/grpc/integration.proto",
}
