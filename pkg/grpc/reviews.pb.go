// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v5.27.1
// source: pkg/grpc/reviews.proto

package grpc

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type DeviceReview struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DeviceDefinitionId string `protobuf:"bytes,1,opt,name=device_definition_id,json=deviceDefinitionId,proto3" json:"device_definition_id,omitempty"`
	Url                string `protobuf:"bytes,2,opt,name=url,proto3" json:"url,omitempty"`
	ImageURL           string `protobuf:"bytes,3,opt,name=imageURL,proto3" json:"imageURL,omitempty"`
	Channel            string `protobuf:"bytes,4,opt,name=channel,proto3" json:"channel,omitempty"`
	Approved           bool   `protobuf:"varint,5,opt,name=approved,proto3" json:"approved,omitempty"`
	ApprovedBy         string `protobuf:"bytes,6,opt,name=approved_by,json=approvedBy,proto3" json:"approved_by,omitempty"`
	Comments           string `protobuf:"bytes,7,opt,name=comments,proto3" json:"comments,omitempty"`
	Id                 string `protobuf:"bytes,8,opt,name=id,proto3" json:"id,omitempty"`
	Name               string `protobuf:"bytes,9,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *DeviceReview) Reset() {
	*x = DeviceReview{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_reviews_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeviceReview) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeviceReview) ProtoMessage() {}

func (x *DeviceReview) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_reviews_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeviceReview.ProtoReflect.Descriptor instead.
func (*DeviceReview) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_reviews_proto_rawDescGZIP(), []int{0}
}

func (x *DeviceReview) GetDeviceDefinitionId() string {
	if x != nil {
		return x.DeviceDefinitionId
	}
	return ""
}

func (x *DeviceReview) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *DeviceReview) GetImageURL() string {
	if x != nil {
		return x.ImageURL
	}
	return ""
}

func (x *DeviceReview) GetChannel() string {
	if x != nil {
		return x.Channel
	}
	return ""
}

func (x *DeviceReview) GetApproved() bool {
	if x != nil {
		return x.Approved
	}
	return false
}

func (x *DeviceReview) GetApprovedBy() string {
	if x != nil {
		return x.ApprovedBy
	}
	return ""
}

func (x *DeviceReview) GetComments() string {
	if x != nil {
		return x.Comments
	}
	return ""
}

func (x *DeviceReview) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *DeviceReview) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type GetReviewsByDeviceDefinitionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DeviceDefinitionId string `protobuf:"bytes,1,opt,name=device_definition_id,json=deviceDefinitionId,proto3" json:"device_definition_id,omitempty"`
}

func (x *GetReviewsByDeviceDefinitionRequest) Reset() {
	*x = GetReviewsByDeviceDefinitionRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_reviews_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetReviewsByDeviceDefinitionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetReviewsByDeviceDefinitionRequest) ProtoMessage() {}

func (x *GetReviewsByDeviceDefinitionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_reviews_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetReviewsByDeviceDefinitionRequest.ProtoReflect.Descriptor instead.
func (*GetReviewsByDeviceDefinitionRequest) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_reviews_proto_rawDescGZIP(), []int{1}
}

func (x *GetReviewsByDeviceDefinitionRequest) GetDeviceDefinitionId() string {
	if x != nil {
		return x.DeviceDefinitionId
	}
	return ""
}

type GetReviewsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Reviews []*DeviceReview `protobuf:"bytes,1,rep,name=reviews,proto3" json:"reviews,omitempty"`
}

func (x *GetReviewsResponse) Reset() {
	*x = GetReviewsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_reviews_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetReviewsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetReviewsResponse) ProtoMessage() {}

func (x *GetReviewsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_reviews_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetReviewsResponse.ProtoReflect.Descriptor instead.
func (*GetReviewsResponse) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_reviews_proto_rawDescGZIP(), []int{2}
}

func (x *GetReviewsResponse) GetReviews() []*DeviceReview {
	if x != nil {
		return x.Reviews
	}
	return nil
}

type CreateReviewRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DeviceDefinitionId string `protobuf:"bytes,1,opt,name=device_definition_id,json=deviceDefinitionId,proto3" json:"device_definition_id,omitempty"`
	Url                string `protobuf:"bytes,2,opt,name=url,proto3" json:"url,omitempty"`
	ImageURL           string `protobuf:"bytes,3,opt,name=imageURL,proto3" json:"imageURL,omitempty"`
	Channel            string `protobuf:"bytes,4,opt,name=channel,proto3" json:"channel,omitempty"`
	Comments           string `protobuf:"bytes,5,opt,name=comments,proto3" json:"comments,omitempty"`
}

func (x *CreateReviewRequest) Reset() {
	*x = CreateReviewRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_reviews_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateReviewRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateReviewRequest) ProtoMessage() {}

func (x *CreateReviewRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_reviews_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateReviewRequest.ProtoReflect.Descriptor instead.
func (*CreateReviewRequest) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_reviews_proto_rawDescGZIP(), []int{3}
}

func (x *CreateReviewRequest) GetDeviceDefinitionId() string {
	if x != nil {
		return x.DeviceDefinitionId
	}
	return ""
}

func (x *CreateReviewRequest) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *CreateReviewRequest) GetImageURL() string {
	if x != nil {
		return x.ImageURL
	}
	return ""
}

func (x *CreateReviewRequest) GetChannel() string {
	if x != nil {
		return x.Channel
	}
	return ""
}

func (x *CreateReviewRequest) GetComments() string {
	if x != nil {
		return x.Comments
	}
	return ""
}

type UpdateReviewRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id       string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Url      string `protobuf:"bytes,2,opt,name=url,proto3" json:"url,omitempty"`
	ImageURL string `protobuf:"bytes,3,opt,name=imageURL,proto3" json:"imageURL,omitempty"`
	Channel  string `protobuf:"bytes,4,opt,name=channel,proto3" json:"channel,omitempty"`
	Comments string `protobuf:"bytes,5,opt,name=comments,proto3" json:"comments,omitempty"`
}

func (x *UpdateReviewRequest) Reset() {
	*x = UpdateReviewRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_reviews_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateReviewRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateReviewRequest) ProtoMessage() {}

func (x *UpdateReviewRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_reviews_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateReviewRequest.ProtoReflect.Descriptor instead.
func (*UpdateReviewRequest) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_reviews_proto_rawDescGZIP(), []int{4}
}

func (x *UpdateReviewRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *UpdateReviewRequest) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *UpdateReviewRequest) GetImageURL() string {
	if x != nil {
		return x.ImageURL
	}
	return ""
}

func (x *UpdateReviewRequest) GetChannel() string {
	if x != nil {
		return x.Channel
	}
	return ""
}

func (x *UpdateReviewRequest) GetComments() string {
	if x != nil {
		return x.Comments
	}
	return ""
}

type ReviewBaseResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *ReviewBaseResponse) Reset() {
	*x = ReviewBaseResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_reviews_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReviewBaseResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReviewBaseResponse) ProtoMessage() {}

func (x *ReviewBaseResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_reviews_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReviewBaseResponse.ProtoReflect.Descriptor instead.
func (*ReviewBaseResponse) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_reviews_proto_rawDescGZIP(), []int{5}
}

func (x *ReviewBaseResponse) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type ApproveReviewRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id         string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	ApprovedBy string `protobuf:"bytes,2,opt,name=approved_by,json=approvedBy,proto3" json:"approved_by,omitempty"`
}

func (x *ApproveReviewRequest) Reset() {
	*x = ApproveReviewRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_reviews_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ApproveReviewRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ApproveReviewRequest) ProtoMessage() {}

func (x *ApproveReviewRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_reviews_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ApproveReviewRequest.ProtoReflect.Descriptor instead.
func (*ApproveReviewRequest) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_reviews_proto_rawDescGZIP(), []int{6}
}

func (x *ApproveReviewRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ApproveReviewRequest) GetApprovedBy() string {
	if x != nil {
		return x.ApprovedBy
	}
	return ""
}

type DeleteReviewRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *DeleteReviewRequest) Reset() {
	*x = DeleteReviewRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_reviews_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteReviewRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteReviewRequest) ProtoMessage() {}

func (x *DeleteReviewRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_reviews_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteReviewRequest.ProtoReflect.Descriptor instead.
func (*DeleteReviewRequest) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_reviews_proto_rawDescGZIP(), []int{7}
}

func (x *DeleteReviewRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetReviewRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetReviewRequest) Reset() {
	*x = GetReviewRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_reviews_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetReviewRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetReviewRequest) ProtoMessage() {}

func (x *GetReviewRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_reviews_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetReviewRequest.ProtoReflect.Descriptor instead.
func (*GetReviewRequest) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_reviews_proto_rawDescGZIP(), []int{8}
}

func (x *GetReviewRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetReviewFilterRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MakeId             string `protobuf:"bytes,1,opt,name=make_id,json=makeId,proto3" json:"make_id,omitempty"`
	Year               int32  `protobuf:"varint,2,opt,name=year,proto3" json:"year,omitempty"`
	Model              string `protobuf:"bytes,3,opt,name=model,proto3" json:"model,omitempty"`
	DeviceDefinitionId string `protobuf:"bytes,4,opt,name=device_definition_id,json=deviceDefinitionId,proto3" json:"device_definition_id,omitempty"`
	Approved           bool   `protobuf:"varint,5,opt,name=approved,proto3" json:"approved,omitempty"`
	PageIndex          int32  `protobuf:"varint,6,opt,name=page_index,json=pageIndex,proto3" json:"page_index,omitempty"`
	PageSize           int32  `protobuf:"varint,7,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
}

func (x *GetReviewFilterRequest) Reset() {
	*x = GetReviewFilterRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_reviews_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetReviewFilterRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetReviewFilterRequest) ProtoMessage() {}

func (x *GetReviewFilterRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_reviews_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetReviewFilterRequest.ProtoReflect.Descriptor instead.
func (*GetReviewFilterRequest) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_reviews_proto_rawDescGZIP(), []int{9}
}

func (x *GetReviewFilterRequest) GetMakeId() string {
	if x != nil {
		return x.MakeId
	}
	return ""
}

func (x *GetReviewFilterRequest) GetYear() int32 {
	if x != nil {
		return x.Year
	}
	return 0
}

func (x *GetReviewFilterRequest) GetModel() string {
	if x != nil {
		return x.Model
	}
	return ""
}

func (x *GetReviewFilterRequest) GetDeviceDefinitionId() string {
	if x != nil {
		return x.DeviceDefinitionId
	}
	return ""
}

func (x *GetReviewFilterRequest) GetApproved() bool {
	if x != nil {
		return x.Approved
	}
	return false
}

func (x *GetReviewFilterRequest) GetPageIndex() int32 {
	if x != nil {
		return x.PageIndex
	}
	return 0
}

func (x *GetReviewFilterRequest) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

var File_pkg_grpc_reviews_proto protoreflect.FileDescriptor

var file_pkg_grpc_reviews_proto_rawDesc = []byte{
	0x0a, 0x16, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x72, 0x65, 0x76, 0x69, 0x65,
	0x77, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x67, 0x72, 0x70, 0x63, 0x22, 0x85,
	0x02, 0x0a, 0x0c, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x12,
	0x30, 0x0a, 0x14, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69,
	0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x64,
	0x65, 0x76, 0x69, 0x63, 0x65, 0x44, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x49,
	0x64, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x75, 0x72, 0x6c, 0x12, 0x1a, 0x0a, 0x08, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x55, 0x52, 0x4c, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x55, 0x52, 0x4c, 0x12,
	0x18, 0x0a, 0x07, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x12, 0x1a, 0x0a, 0x08, 0x61, 0x70, 0x70,
	0x72, 0x6f, 0x76, 0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x61, 0x70, 0x70,
	0x72, 0x6f, 0x76, 0x65, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x61, 0x70, 0x70, 0x72, 0x6f, 0x76, 0x65,
	0x64, 0x5f, 0x62, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x61, 0x70, 0x70, 0x72,
	0x6f, 0x76, 0x65, 0x64, 0x42, 0x79, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e,
	0x74, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e,
	0x74, 0x73, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x57, 0x0a, 0x23, 0x47, 0x65, 0x74, 0x52, 0x65, 0x76,
	0x69, 0x65, 0x77, 0x73, 0x42, 0x79, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x44, 0x65, 0x66, 0x69,
	0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x30, 0x0a,
	0x14, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69,
	0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x64, 0x65, 0x76,
	0x69, 0x63, 0x65, 0x44, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x22,
	0x42, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2c, 0x0a, 0x07, 0x72, 0x65, 0x76, 0x69, 0x65, 0x77, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x44, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x52, 0x07, 0x72, 0x65, 0x76, 0x69,
	0x65, 0x77, 0x73, 0x22, 0xab, 0x01, 0x0a, 0x13, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65,
	0x76, 0x69, 0x65, 0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x30, 0x0a, 0x14, 0x64,
	0x65, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x64, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x44, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x10, 0x0a,
	0x03, 0x75, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12,
	0x1a, 0x0a, 0x08, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x55, 0x52, 0x4c, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x55, 0x52, 0x4c, 0x12, 0x18, 0x0a, 0x07, 0x63,
	0x68, 0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x68,
	0x61, 0x6e, 0x6e, 0x65, 0x6c, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74,
	0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74,
	0x73, 0x22, 0x89, 0x01, 0x0a, 0x13, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x76, 0x69,
	0x65, 0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x72, 0x6c,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x1a, 0x0a, 0x08, 0x69,
	0x6d, 0x61, 0x67, 0x65, 0x55, 0x52, 0x4c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x69,
	0x6d, 0x61, 0x67, 0x65, 0x55, 0x52, 0x4c, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x68, 0x61, 0x6e, 0x6e,
	0x65, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x68, 0x61, 0x6e, 0x6e, 0x65,
	0x6c, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x6f, 0x6d, 0x6d, 0x65, 0x6e, 0x74, 0x73, 0x22, 0x24, 0x0a,
	0x12, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x42, 0x61, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x69, 0x64, 0x22, 0x47, 0x0a, 0x14, 0x41, 0x70, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x52, 0x65,
	0x76, 0x69, 0x65, 0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x61,
	0x70, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x64, 0x5f, 0x62, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x61, 0x70, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x64, 0x42, 0x79, 0x22, 0x25, 0x0a, 0x13,
	0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x69, 0x64, 0x22, 0x22, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0xe5, 0x01, 0x0a, 0x16, 0x47, 0x65, 0x74, 0x52,
	0x65, 0x76, 0x69, 0x65, 0x77, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x6d, 0x61, 0x6b, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x61, 0x6b, 0x65, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x79,
	0x65, 0x61, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x79, 0x65, 0x61, 0x72, 0x12,
	0x14, 0x0a, 0x05, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x12, 0x30, 0x0a, 0x14, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5f,
	0x64, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x12, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x44, 0x65, 0x66, 0x69, 0x6e,
	0x69, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x61, 0x70, 0x70, 0x72, 0x6f,
	0x76, 0x65, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x61, 0x70, 0x70, 0x72, 0x6f,
	0x76, 0x65, 0x64, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x69, 0x6e, 0x64, 0x65,
	0x78, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x70, 0x61, 0x67, 0x65, 0x49, 0x6e, 0x64,
	0x65, 0x78, 0x12, 0x1b, 0x0a, 0x09, 0x70, 0x61, 0x67, 0x65, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x32,
	0x90, 0x04, 0x0a, 0x0e, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x73, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x65, 0x0a, 0x1e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x73,
	0x42, 0x79, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x44, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69,
	0x6f, 0x6e, 0x49, 0x44, 0x12, 0x29, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x52,
	0x65, 0x76, 0x69, 0x65, 0x77, 0x73, 0x42, 0x79, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x44, 0x65,
	0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x18, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x44, 0x0a, 0x0a, 0x47, 0x65, 0x74,
	0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x73, 0x12, 0x1c, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x47,
	0x65, 0x74, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x46, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74,
	0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x3b, 0x0a, 0x0d, 0x47, 0x65, 0x74, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x42, 0x79, 0x49, 0x44,
	0x12, 0x16, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x76, 0x69, 0x65,
	0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x12, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e,
	0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x12, 0x43, 0x0a, 0x0c,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x12, 0x19, 0x2e, 0x67,
	0x72, 0x70, 0x63, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x52,
	0x65, 0x76, 0x69, 0x65, 0x77, 0x42, 0x61, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x43, 0x0a, 0x0c, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x65, 0x76, 0x69, 0x65,
	0x77, 0x12, 0x19, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52,
	0x65, 0x76, 0x69, 0x65, 0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x67,
	0x72, 0x70, 0x63, 0x2e, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x42, 0x61, 0x73, 0x65, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x45, 0x0a, 0x0d, 0x41, 0x70, 0x70, 0x72, 0x6f, 0x76,
	0x65, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x12, 0x1a, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x41,
	0x70, 0x70, 0x72, 0x6f, 0x76, 0x65, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x52, 0x65, 0x76, 0x69, 0x65,
	0x77, 0x42, 0x61, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x43, 0x0a,
	0x0c, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x12, 0x19, 0x2e,
	0x67, 0x72, 0x70, 0x63, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x76, 0x69, 0x65,
	0x77, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e,
	0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x42, 0x61, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x42, 0x39, 0x5a, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x44, 0x49, 0x4d, 0x4f, 0x2d, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2f, 0x64, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x2d, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2d, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_grpc_reviews_proto_rawDescOnce sync.Once
	file_pkg_grpc_reviews_proto_rawDescData = file_pkg_grpc_reviews_proto_rawDesc
)

func file_pkg_grpc_reviews_proto_rawDescGZIP() []byte {
	file_pkg_grpc_reviews_proto_rawDescOnce.Do(func() {
		file_pkg_grpc_reviews_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_grpc_reviews_proto_rawDescData)
	})
	return file_pkg_grpc_reviews_proto_rawDescData
}

var file_pkg_grpc_reviews_proto_msgTypes = make([]protoimpl.MessageInfo, 10)
var file_pkg_grpc_reviews_proto_goTypes = []interface{}{
	(*DeviceReview)(nil),                        // 0: grpc.DeviceReview
	(*GetReviewsByDeviceDefinitionRequest)(nil), // 1: grpc.GetReviewsByDeviceDefinitionRequest
	(*GetReviewsResponse)(nil),                  // 2: grpc.GetReviewsResponse
	(*CreateReviewRequest)(nil),                 // 3: grpc.CreateReviewRequest
	(*UpdateReviewRequest)(nil),                 // 4: grpc.UpdateReviewRequest
	(*ReviewBaseResponse)(nil),                  // 5: grpc.ReviewBaseResponse
	(*ApproveReviewRequest)(nil),                // 6: grpc.ApproveReviewRequest
	(*DeleteReviewRequest)(nil),                 // 7: grpc.DeleteReviewRequest
	(*GetReviewRequest)(nil),                    // 8: grpc.GetReviewRequest
	(*GetReviewFilterRequest)(nil),              // 9: grpc.GetReviewFilterRequest
}
var file_pkg_grpc_reviews_proto_depIdxs = []int32{
	0, // 0: grpc.GetReviewsResponse.reviews:type_name -> grpc.DeviceReview
	1, // 1: grpc.ReviewsService.GetReviewsByDeviceDefinitionID:input_type -> grpc.GetReviewsByDeviceDefinitionRequest
	9, // 2: grpc.ReviewsService.GetReviews:input_type -> grpc.GetReviewFilterRequest
	8, // 3: grpc.ReviewsService.GetReviewByID:input_type -> grpc.GetReviewRequest
	3, // 4: grpc.ReviewsService.CreateReview:input_type -> grpc.CreateReviewRequest
	4, // 5: grpc.ReviewsService.UpdateReview:input_type -> grpc.UpdateReviewRequest
	6, // 6: grpc.ReviewsService.ApproveReview:input_type -> grpc.ApproveReviewRequest
	7, // 7: grpc.ReviewsService.DeleteReview:input_type -> grpc.DeleteReviewRequest
	2, // 8: grpc.ReviewsService.GetReviewsByDeviceDefinitionID:output_type -> grpc.GetReviewsResponse
	2, // 9: grpc.ReviewsService.GetReviews:output_type -> grpc.GetReviewsResponse
	0, // 10: grpc.ReviewsService.GetReviewByID:output_type -> grpc.DeviceReview
	5, // 11: grpc.ReviewsService.CreateReview:output_type -> grpc.ReviewBaseResponse
	5, // 12: grpc.ReviewsService.UpdateReview:output_type -> grpc.ReviewBaseResponse
	5, // 13: grpc.ReviewsService.ApproveReview:output_type -> grpc.ReviewBaseResponse
	5, // 14: grpc.ReviewsService.DeleteReview:output_type -> grpc.ReviewBaseResponse
	8, // [8:15] is the sub-list for method output_type
	1, // [1:8] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_pkg_grpc_reviews_proto_init() }
func file_pkg_grpc_reviews_proto_init() {
	if File_pkg_grpc_reviews_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_grpc_reviews_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeviceReview); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_grpc_reviews_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetReviewsByDeviceDefinitionRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_grpc_reviews_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetReviewsResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_grpc_reviews_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateReviewRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_grpc_reviews_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateReviewRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_grpc_reviews_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReviewBaseResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_grpc_reviews_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ApproveReviewRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_grpc_reviews_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteReviewRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_grpc_reviews_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetReviewRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_grpc_reviews_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetReviewFilterRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pkg_grpc_reviews_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   10,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_grpc_reviews_proto_goTypes,
		DependencyIndexes: file_pkg_grpc_reviews_proto_depIdxs,
		MessageInfos:      file_pkg_grpc_reviews_proto_msgTypes,
	}.Build()
	File_pkg_grpc_reviews_proto = out.File
	file_pkg_grpc_reviews_proto_rawDesc = nil
	file_pkg_grpc_reviews_proto_goTypes = nil
	file_pkg_grpc_reviews_proto_depIdxs = nil
}
