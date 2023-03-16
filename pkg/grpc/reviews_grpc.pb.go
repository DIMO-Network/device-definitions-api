// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.21.12
// source: pkg/grpc/reviews.proto

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
	ReviewsService_GetReviewsByDeviceDefinitionID_FullMethodName = "/grpc.ReviewsService/GetReviewsByDeviceDefinitionID"
	ReviewsService_GetReviews_FullMethodName                     = "/grpc.ReviewsService/GetReviews"
	ReviewsService_GetReviewByID_FullMethodName                  = "/grpc.ReviewsService/GetReviewByID"
	ReviewsService_CreateReview_FullMethodName                   = "/grpc.ReviewsService/CreateReview"
	ReviewsService_UpdateReview_FullMethodName                   = "/grpc.ReviewsService/UpdateReview"
	ReviewsService_ApproveReview_FullMethodName                  = "/grpc.ReviewsService/ApproveReview"
	ReviewsService_DeleteReview_FullMethodName                   = "/grpc.ReviewsService/DeleteReview"
)

// ReviewsServiceClient is the client API for ReviewsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ReviewsServiceClient interface {
	GetReviewsByDeviceDefinitionID(ctx context.Context, in *GetReviewsByDeviceDefinitionRequest, opts ...grpc.CallOption) (*GetReviewsResponse, error)
	// GetReviews for dimo admin page, get by makeId, model, years
	GetReviews(ctx context.Context, in *GetReviewFilterRequest, opts ...grpc.CallOption) (*GetReviewsResponse, error)
	GetReviewByID(ctx context.Context, in *GetReviewRequest, opts ...grpc.CallOption) (*DeviceReview, error)
	CreateReview(ctx context.Context, in *CreateReviewRequest, opts ...grpc.CallOption) (*ReviewBaseResponse, error)
	UpdateReview(ctx context.Context, in *UpdateReviewRequest, opts ...grpc.CallOption) (*ReviewBaseResponse, error)
	ApproveReview(ctx context.Context, in *ApproveReviewRequest, opts ...grpc.CallOption) (*ReviewBaseResponse, error)
	DeleteReview(ctx context.Context, in *DeleteReviewRequest, opts ...grpc.CallOption) (*ReviewBaseResponse, error)
}

type reviewsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewReviewsServiceClient(cc grpc.ClientConnInterface) ReviewsServiceClient {
	return &reviewsServiceClient{cc}
}

func (c *reviewsServiceClient) GetReviewsByDeviceDefinitionID(ctx context.Context, in *GetReviewsByDeviceDefinitionRequest, opts ...grpc.CallOption) (*GetReviewsResponse, error) {
	out := new(GetReviewsResponse)
	err := c.cc.Invoke(ctx, ReviewsService_GetReviewsByDeviceDefinitionID_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *reviewsServiceClient) GetReviews(ctx context.Context, in *GetReviewFilterRequest, opts ...grpc.CallOption) (*GetReviewsResponse, error) {
	out := new(GetReviewsResponse)
	err := c.cc.Invoke(ctx, ReviewsService_GetReviews_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *reviewsServiceClient) GetReviewByID(ctx context.Context, in *GetReviewRequest, opts ...grpc.CallOption) (*DeviceReview, error) {
	out := new(DeviceReview)
	err := c.cc.Invoke(ctx, ReviewsService_GetReviewByID_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *reviewsServiceClient) CreateReview(ctx context.Context, in *CreateReviewRequest, opts ...grpc.CallOption) (*ReviewBaseResponse, error) {
	out := new(ReviewBaseResponse)
	err := c.cc.Invoke(ctx, ReviewsService_CreateReview_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *reviewsServiceClient) UpdateReview(ctx context.Context, in *UpdateReviewRequest, opts ...grpc.CallOption) (*ReviewBaseResponse, error) {
	out := new(ReviewBaseResponse)
	err := c.cc.Invoke(ctx, ReviewsService_UpdateReview_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *reviewsServiceClient) ApproveReview(ctx context.Context, in *ApproveReviewRequest, opts ...grpc.CallOption) (*ReviewBaseResponse, error) {
	out := new(ReviewBaseResponse)
	err := c.cc.Invoke(ctx, ReviewsService_ApproveReview_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *reviewsServiceClient) DeleteReview(ctx context.Context, in *DeleteReviewRequest, opts ...grpc.CallOption) (*ReviewBaseResponse, error) {
	out := new(ReviewBaseResponse)
	err := c.cc.Invoke(ctx, ReviewsService_DeleteReview_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ReviewsServiceServer is the server API for ReviewsService service.
// All implementations must embed UnimplementedReviewsServiceServer
// for forward compatibility
type ReviewsServiceServer interface {
	GetReviewsByDeviceDefinitionID(context.Context, *GetReviewsByDeviceDefinitionRequest) (*GetReviewsResponse, error)
	// GetReviews for dimo admin page, get by makeId, model, years
	GetReviews(context.Context, *GetReviewFilterRequest) (*GetReviewsResponse, error)
	GetReviewByID(context.Context, *GetReviewRequest) (*DeviceReview, error)
	CreateReview(context.Context, *CreateReviewRequest) (*ReviewBaseResponse, error)
	UpdateReview(context.Context, *UpdateReviewRequest) (*ReviewBaseResponse, error)
	ApproveReview(context.Context, *ApproveReviewRequest) (*ReviewBaseResponse, error)
	DeleteReview(context.Context, *DeleteReviewRequest) (*ReviewBaseResponse, error)
	mustEmbedUnimplementedReviewsServiceServer()
}

// UnimplementedReviewsServiceServer must be embedded to have forward compatible implementations.
type UnimplementedReviewsServiceServer struct {
}

func (UnimplementedReviewsServiceServer) GetReviewsByDeviceDefinitionID(context.Context, *GetReviewsByDeviceDefinitionRequest) (*GetReviewsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetReviewsByDeviceDefinitionID not implemented")
}
func (UnimplementedReviewsServiceServer) GetReviews(context.Context, *GetReviewFilterRequest) (*GetReviewsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetReviews not implemented")
}
func (UnimplementedReviewsServiceServer) GetReviewByID(context.Context, *GetReviewRequest) (*DeviceReview, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetReviewByID not implemented")
}
func (UnimplementedReviewsServiceServer) CreateReview(context.Context, *CreateReviewRequest) (*ReviewBaseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateReview not implemented")
}
func (UnimplementedReviewsServiceServer) UpdateReview(context.Context, *UpdateReviewRequest) (*ReviewBaseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateReview not implemented")
}
func (UnimplementedReviewsServiceServer) ApproveReview(context.Context, *ApproveReviewRequest) (*ReviewBaseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ApproveReview not implemented")
}
func (UnimplementedReviewsServiceServer) DeleteReview(context.Context, *DeleteReviewRequest) (*ReviewBaseResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteReview not implemented")
}
func (UnimplementedReviewsServiceServer) mustEmbedUnimplementedReviewsServiceServer() {}

// UnsafeReviewsServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ReviewsServiceServer will
// result in compilation errors.
type UnsafeReviewsServiceServer interface {
	mustEmbedUnimplementedReviewsServiceServer()
}

func RegisterReviewsServiceServer(s grpc.ServiceRegistrar, srv ReviewsServiceServer) {
	s.RegisterService(&ReviewsService_ServiceDesc, srv)
}

func _ReviewsService_GetReviewsByDeviceDefinitionID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetReviewsByDeviceDefinitionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReviewsServiceServer).GetReviewsByDeviceDefinitionID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReviewsService_GetReviewsByDeviceDefinitionID_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReviewsServiceServer).GetReviewsByDeviceDefinitionID(ctx, req.(*GetReviewsByDeviceDefinitionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReviewsService_GetReviews_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetReviewFilterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReviewsServiceServer).GetReviews(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReviewsService_GetReviews_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReviewsServiceServer).GetReviews(ctx, req.(*GetReviewFilterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReviewsService_GetReviewByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetReviewRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReviewsServiceServer).GetReviewByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReviewsService_GetReviewByID_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReviewsServiceServer).GetReviewByID(ctx, req.(*GetReviewRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReviewsService_CreateReview_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateReviewRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReviewsServiceServer).CreateReview(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReviewsService_CreateReview_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReviewsServiceServer).CreateReview(ctx, req.(*CreateReviewRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReviewsService_UpdateReview_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateReviewRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReviewsServiceServer).UpdateReview(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReviewsService_UpdateReview_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReviewsServiceServer).UpdateReview(ctx, req.(*UpdateReviewRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReviewsService_ApproveReview_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ApproveReviewRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReviewsServiceServer).ApproveReview(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReviewsService_ApproveReview_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReviewsServiceServer).ApproveReview(ctx, req.(*ApproveReviewRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReviewsService_DeleteReview_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteReviewRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReviewsServiceServer).DeleteReview(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ReviewsService_DeleteReview_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReviewsServiceServer).DeleteReview(ctx, req.(*DeleteReviewRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ReviewsService_ServiceDesc is the grpc.ServiceDesc for ReviewsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ReviewsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "grpc.ReviewsService",
	HandlerType: (*ReviewsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetReviewsByDeviceDefinitionID",
			Handler:    _ReviewsService_GetReviewsByDeviceDefinitionID_Handler,
		},
		{
			MethodName: "GetReviews",
			Handler:    _ReviewsService_GetReviews_Handler,
		},
		{
			MethodName: "GetReviewByID",
			Handler:    _ReviewsService_GetReviewByID_Handler,
		},
		{
			MethodName: "CreateReview",
			Handler:    _ReviewsService_CreateReview_Handler,
		},
		{
			MethodName: "UpdateReview",
			Handler:    _ReviewsService_UpdateReview_Handler,
		},
		{
			MethodName: "ApproveReview",
			Handler:    _ReviewsService_ApproveReview_Handler,
		},
		{
			MethodName: "DeleteReview",
			Handler:    _ReviewsService_DeleteReview_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/grpc/reviews.proto",
}
