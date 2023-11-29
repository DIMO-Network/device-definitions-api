// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
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

const (
	IntegrationService_GetCompatibilitiesByMake_FullMethodName           = "/grpc.IntegrationService/GetCompatibilitiesByMake"
	IntegrationService_GetCompatibilityByDeviceDefinition_FullMethodName = "/grpc.IntegrationService/GetCompatibilityByDeviceDefinition"
	IntegrationService_GetCompatibilityByDeviceArray_FullMethodName      = "/grpc.IntegrationService/GetCompatibilityByDeviceArray"
	IntegrationService_GetIntegrationFeatureByID_FullMethodName          = "/grpc.IntegrationService/GetIntegrationFeatureByID"
	IntegrationService_GetIntegrationFeatures_FullMethodName             = "/grpc.IntegrationService/GetIntegrationFeatures"
	IntegrationService_GetIntegrationOptions_FullMethodName              = "/grpc.IntegrationService/GetIntegrationOptions"
	IntegrationService_CreateIntegrationFeature_FullMethodName           = "/grpc.IntegrationService/CreateIntegrationFeature"
	IntegrationService_UpdateIntegrationFeature_FullMethodName           = "/grpc.IntegrationService/UpdateIntegrationFeature"
	IntegrationService_DeleteIntegrationFeature_FullMethodName           = "/grpc.IntegrationService/DeleteIntegrationFeature"
)

// IntegrationServiceClient is the client API for IntegrationService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IntegrationServiceClient interface {
	// GetCompatibilitiesByMake for explorer makes page, get by makeId
	GetCompatibilitiesByMake(ctx context.Context, in *GetCompatibilitiesByMakeRequest, opts ...grpc.CallOption) (*GetCompatibilitiesByMakeResponse, error)
	// GetCompatibilityByDeviceDefinition for explorer models page, get by ddid
	GetCompatibilityByDeviceDefinition(ctx context.Context, in *GetCompatibilityByDeviceDefinitionRequest, opts ...grpc.CallOption) (*GetDeviceCompatibilitiesResponse, error)
	// GetCompatibilityByDeviceArray for models endpoint, returns all model compatability levels
	GetCompatibilityByDeviceArray(ctx context.Context, in *GetCompatibilityByDeviceArrayRequest, opts ...grpc.CallOption) (*GetCompatibilityByDeviceArrayResponse, error)
	GetIntegrationFeatureByID(ctx context.Context, in *GetIntegrationFeatureByIDRequest, opts ...grpc.CallOption) (*GetIntegrationFeatureResponse, error)
	GetIntegrationFeatures(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetIntegrationFeatureListResponse, error)
	// GetIntegrationOptions for dropdowns in explorer makes page, get by makeId
	GetIntegrationOptions(ctx context.Context, in *GetIntegrationOptionsRequest, opts ...grpc.CallOption) (*GetIntegrationOptionsResponse, error)
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

func (c *integrationServiceClient) GetCompatibilitiesByMake(ctx context.Context, in *GetCompatibilitiesByMakeRequest, opts ...grpc.CallOption) (*GetCompatibilitiesByMakeResponse, error) {
	out := new(GetCompatibilitiesByMakeResponse)
	err := c.cc.Invoke(ctx, IntegrationService_GetCompatibilitiesByMake_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *integrationServiceClient) GetCompatibilityByDeviceDefinition(ctx context.Context, in *GetCompatibilityByDeviceDefinitionRequest, opts ...grpc.CallOption) (*GetDeviceCompatibilitiesResponse, error) {
	out := new(GetDeviceCompatibilitiesResponse)
	err := c.cc.Invoke(ctx, IntegrationService_GetCompatibilityByDeviceDefinition_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *integrationServiceClient) GetCompatibilityByDeviceArray(ctx context.Context, in *GetCompatibilityByDeviceArrayRequest, opts ...grpc.CallOption) (*GetCompatibilityByDeviceArrayResponse, error) {
	out := new(GetCompatibilityByDeviceArrayResponse)
	err := c.cc.Invoke(ctx, IntegrationService_GetCompatibilityByDeviceArray_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *integrationServiceClient) GetIntegrationFeatureByID(ctx context.Context, in *GetIntegrationFeatureByIDRequest, opts ...grpc.CallOption) (*GetIntegrationFeatureResponse, error) {
	out := new(GetIntegrationFeatureResponse)
	err := c.cc.Invoke(ctx, IntegrationService_GetIntegrationFeatureByID_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *integrationServiceClient) GetIntegrationFeatures(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*GetIntegrationFeatureListResponse, error) {
	out := new(GetIntegrationFeatureListResponse)
	err := c.cc.Invoke(ctx, IntegrationService_GetIntegrationFeatures_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *integrationServiceClient) GetIntegrationOptions(ctx context.Context, in *GetIntegrationOptionsRequest, opts ...grpc.CallOption) (*GetIntegrationOptionsResponse, error) {
	out := new(GetIntegrationOptionsResponse)
	err := c.cc.Invoke(ctx, IntegrationService_GetIntegrationOptions_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *integrationServiceClient) CreateIntegrationFeature(ctx context.Context, in *CreateOrUpdateIntegrationFeatureRequest, opts ...grpc.CallOption) (*IntegrationBaseResponse, error) {
	out := new(IntegrationBaseResponse)
	err := c.cc.Invoke(ctx, IntegrationService_CreateIntegrationFeature_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *integrationServiceClient) UpdateIntegrationFeature(ctx context.Context, in *CreateOrUpdateIntegrationFeatureRequest, opts ...grpc.CallOption) (*IntegrationBaseResponse, error) {
	out := new(IntegrationBaseResponse)
	err := c.cc.Invoke(ctx, IntegrationService_UpdateIntegrationFeature_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *integrationServiceClient) DeleteIntegrationFeature(ctx context.Context, in *DeleteIntegrationFeatureRequest, opts ...grpc.CallOption) (*IntegrationBaseResponse, error) {
	out := new(IntegrationBaseResponse)
	err := c.cc.Invoke(ctx, IntegrationService_DeleteIntegrationFeature_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IntegrationServiceServer is the server API for IntegrationService service.
// All implementations must embed UnimplementedIntegrationServiceServer
// for forward compatibility
type IntegrationServiceServer interface {
	// GetCompatibilitiesByMake for explorer makes page, get by makeId
	GetCompatibilitiesByMake(context.Context, *GetCompatibilitiesByMakeRequest) (*GetCompatibilitiesByMakeResponse, error)
	// GetCompatibilityByDeviceDefinition for explorer models page, get by ddid
	GetCompatibilityByDeviceDefinition(context.Context, *GetCompatibilityByDeviceDefinitionRequest) (*GetDeviceCompatibilitiesResponse, error)
	// GetCompatibilityByDeviceArray for models endpoint, returns all model compatability levels
	GetCompatibilityByDeviceArray(context.Context, *GetCompatibilityByDeviceArrayRequest) (*GetCompatibilityByDeviceArrayResponse, error)
	GetIntegrationFeatureByID(context.Context, *GetIntegrationFeatureByIDRequest) (*GetIntegrationFeatureResponse, error)
	GetIntegrationFeatures(context.Context, *emptypb.Empty) (*GetIntegrationFeatureListResponse, error)
	// GetIntegrationOptions for dropdowns in explorer makes page, get by makeId
	GetIntegrationOptions(context.Context, *GetIntegrationOptionsRequest) (*GetIntegrationOptionsResponse, error)
	CreateIntegrationFeature(context.Context, *CreateOrUpdateIntegrationFeatureRequest) (*IntegrationBaseResponse, error)
	UpdateIntegrationFeature(context.Context, *CreateOrUpdateIntegrationFeatureRequest) (*IntegrationBaseResponse, error)
	DeleteIntegrationFeature(context.Context, *DeleteIntegrationFeatureRequest) (*IntegrationBaseResponse, error)
	mustEmbedUnimplementedIntegrationServiceServer()
}

// UnimplementedIntegrationServiceServer must be embedded to have forward compatible implementations.
type UnimplementedIntegrationServiceServer struct {
}

func (UnimplementedIntegrationServiceServer) GetCompatibilitiesByMake(context.Context, *GetCompatibilitiesByMakeRequest) (*GetCompatibilitiesByMakeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCompatibilitiesByMake not implemented")
}
func (UnimplementedIntegrationServiceServer) GetCompatibilityByDeviceDefinition(context.Context, *GetCompatibilityByDeviceDefinitionRequest) (*GetDeviceCompatibilitiesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCompatibilityByDeviceDefinition not implemented")
}
func (UnimplementedIntegrationServiceServer) GetCompatibilityByDeviceArray(context.Context, *GetCompatibilityByDeviceArrayRequest) (*GetCompatibilityByDeviceArrayResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCompatibilityByDeviceArray not implemented")
}
func (UnimplementedIntegrationServiceServer) GetIntegrationFeatureByID(context.Context, *GetIntegrationFeatureByIDRequest) (*GetIntegrationFeatureResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetIntegrationFeatureByID not implemented")
}
func (UnimplementedIntegrationServiceServer) GetIntegrationFeatures(context.Context, *emptypb.Empty) (*GetIntegrationFeatureListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetIntegrationFeatures not implemented")
}
func (UnimplementedIntegrationServiceServer) GetIntegrationOptions(context.Context, *GetIntegrationOptionsRequest) (*GetIntegrationOptionsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetIntegrationOptions not implemented")
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

func _IntegrationService_GetCompatibilitiesByMake_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCompatibilitiesByMakeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntegrationServiceServer).GetCompatibilitiesByMake(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IntegrationService_GetCompatibilitiesByMake_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntegrationServiceServer).GetCompatibilitiesByMake(ctx, req.(*GetCompatibilitiesByMakeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntegrationService_GetCompatibilityByDeviceDefinition_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCompatibilityByDeviceDefinitionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntegrationServiceServer).GetCompatibilityByDeviceDefinition(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IntegrationService_GetCompatibilityByDeviceDefinition_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntegrationServiceServer).GetCompatibilityByDeviceDefinition(ctx, req.(*GetCompatibilityByDeviceDefinitionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IntegrationService_GetCompatibilityByDeviceArray_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCompatibilityByDeviceArrayRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IntegrationServiceServer).GetCompatibilityByDeviceArray(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IntegrationService_GetCompatibilityByDeviceArray_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntegrationServiceServer).GetCompatibilityByDeviceArray(ctx, req.(*GetCompatibilityByDeviceArrayRequest))
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
		FullMethod: IntegrationService_GetIntegrationFeatureByID_FullMethodName,
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
		FullMethod: IntegrationService_GetIntegrationFeatures_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IntegrationServiceServer).GetIntegrationFeatures(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
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
		FullMethod: IntegrationService_CreateIntegrationFeature_FullMethodName,
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
		FullMethod: IntegrationService_UpdateIntegrationFeature_FullMethodName,
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
		FullMethod: IntegrationService_DeleteIntegrationFeature_FullMethodName,
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
			MethodName: "GetCompatibilitiesByMake",
			Handler:    _IntegrationService_GetCompatibilitiesByMake_Handler,
		},
		{
			MethodName: "GetCompatibilityByDeviceDefinition",
			Handler:    _IntegrationService_GetCompatibilityByDeviceDefinition_Handler,
		},
		{
			MethodName: "GetCompatibilityByDeviceArray",
			Handler:    _IntegrationService_GetCompatibilityByDeviceArray_Handler,
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
			MethodName: "GetIntegrationOptions",
			Handler:    _IntegrationService_GetIntegrationOptions_Handler,
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
