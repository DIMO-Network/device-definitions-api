// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.7
// source: pkg/grpc/integration.proto

package grpc

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type IntegrationBaseResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *IntegrationBaseResponse) Reset() {
	*x = IntegrationBaseResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_integration_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IntegrationBaseResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IntegrationBaseResponse) ProtoMessage() {}

func (x *IntegrationBaseResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_integration_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IntegrationBaseResponse.ProtoReflect.Descriptor instead.
func (*IntegrationBaseResponse) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_integration_proto_rawDescGZIP(), []int{0}
}

func (x *IntegrationBaseResponse) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetIntegrationFeatureByIDRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *GetIntegrationFeatureByIDRequest) Reset() {
	*x = GetIntegrationFeatureByIDRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_integration_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetIntegrationFeatureByIDRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetIntegrationFeatureByIDRequest) ProtoMessage() {}

func (x *GetIntegrationFeatureByIDRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_integration_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetIntegrationFeatureByIDRequest.ProtoReflect.Descriptor instead.
func (*GetIntegrationFeatureByIDRequest) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_integration_proto_rawDescGZIP(), []int{1}
}

func (x *GetIntegrationFeatureByIDRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type CreateOrUpdateIntegrationFeatureRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id              string  `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	ElasticProperty string  `protobuf:"bytes,2,opt,name=elastic_property,json=elasticProperty,proto3" json:"elastic_property,omitempty"`
	DisplayName     string  `protobuf:"bytes,3,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
	CssIcon         string  `protobuf:"bytes,4,opt,name=css_icon,json=cssIcon,proto3" json:"css_icon,omitempty"`
	FeatureWeight   float32 `protobuf:"fixed32,5,opt,name=feature_weight,json=featureWeight,proto3" json:"feature_weight,omitempty"`
}

func (x *CreateOrUpdateIntegrationFeatureRequest) Reset() {
	*x = CreateOrUpdateIntegrationFeatureRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_integration_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateOrUpdateIntegrationFeatureRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateOrUpdateIntegrationFeatureRequest) ProtoMessage() {}

func (x *CreateOrUpdateIntegrationFeatureRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_integration_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateOrUpdateIntegrationFeatureRequest.ProtoReflect.Descriptor instead.
func (*CreateOrUpdateIntegrationFeatureRequest) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_integration_proto_rawDescGZIP(), []int{2}
}

func (x *CreateOrUpdateIntegrationFeatureRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *CreateOrUpdateIntegrationFeatureRequest) GetElasticProperty() string {
	if x != nil {
		return x.ElasticProperty
	}
	return ""
}

func (x *CreateOrUpdateIntegrationFeatureRequest) GetDisplayName() string {
	if x != nil {
		return x.DisplayName
	}
	return ""
}

func (x *CreateOrUpdateIntegrationFeatureRequest) GetCssIcon() string {
	if x != nil {
		return x.CssIcon
	}
	return ""
}

func (x *CreateOrUpdateIntegrationFeatureRequest) GetFeatureWeight() float32 {
	if x != nil {
		return x.FeatureWeight
	}
	return 0
}

type DeleteIntegrationFeatureRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *DeleteIntegrationFeatureRequest) Reset() {
	*x = DeleteIntegrationFeatureRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_integration_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteIntegrationFeatureRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteIntegrationFeatureRequest) ProtoMessage() {}

func (x *DeleteIntegrationFeatureRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_integration_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteIntegrationFeatureRequest.ProtoReflect.Descriptor instead.
func (*DeleteIntegrationFeatureRequest) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_integration_proto_rawDescGZIP(), []int{3}
}

func (x *DeleteIntegrationFeatureRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type GetIntegrationFeatureListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	IntegrationFeatures []*GetIntegrationFeatureResponse `protobuf:"bytes,2,rep,name=integration_features,json=integrationFeatures,proto3" json:"integration_features,omitempty"`
}

func (x *GetIntegrationFeatureListResponse) Reset() {
	*x = GetIntegrationFeatureListResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_integration_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetIntegrationFeatureListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetIntegrationFeatureListResponse) ProtoMessage() {}

func (x *GetIntegrationFeatureListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_integration_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetIntegrationFeatureListResponse.ProtoReflect.Descriptor instead.
func (*GetIntegrationFeatureListResponse) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_integration_proto_rawDescGZIP(), []int{4}
}

func (x *GetIntegrationFeatureListResponse) GetIntegrationFeatures() []*GetIntegrationFeatureResponse {
	if x != nil {
		return x.IntegrationFeatures
	}
	return nil
}

type GetIntegrationFeatureResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	FeatureKey      string  `protobuf:"bytes,1,opt,name=feature_key,json=featureKey,proto3" json:"feature_key,omitempty"`
	ElasticProperty string  `protobuf:"bytes,2,opt,name=elastic_property,json=elasticProperty,proto3" json:"elastic_property,omitempty"`
	DisplayName     string  `protobuf:"bytes,3,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
	CssIcon         string  `protobuf:"bytes,4,opt,name=css_icon,json=cssIcon,proto3" json:"css_icon,omitempty"`
	FeatureWeight   float32 `protobuf:"fixed32,5,opt,name=feature_weight,json=featureWeight,proto3" json:"feature_weight,omitempty"`
}

func (x *GetIntegrationFeatureResponse) Reset() {
	*x = GetIntegrationFeatureResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_integration_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetIntegrationFeatureResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetIntegrationFeatureResponse) ProtoMessage() {}

func (x *GetIntegrationFeatureResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_integration_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetIntegrationFeatureResponse.ProtoReflect.Descriptor instead.
func (*GetIntegrationFeatureResponse) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_integration_proto_rawDescGZIP(), []int{5}
}

func (x *GetIntegrationFeatureResponse) GetFeatureKey() string {
	if x != nil {
		return x.FeatureKey
	}
	return ""
}

func (x *GetIntegrationFeatureResponse) GetElasticProperty() string {
	if x != nil {
		return x.ElasticProperty
	}
	return ""
}

func (x *GetIntegrationFeatureResponse) GetDisplayName() string {
	if x != nil {
		return x.DisplayName
	}
	return ""
}

func (x *GetIntegrationFeatureResponse) GetCssIcon() string {
	if x != nil {
		return x.CssIcon
	}
	return ""
}

func (x *GetIntegrationFeatureResponse) GetFeatureWeight() float32 {
	if x != nil {
		return x.FeatureWeight
	}
	return 0
}

type GetDeviceCompatibilityListRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MakeId        string `protobuf:"bytes,1,opt,name=make_id,json=makeId,proto3" json:"make_id,omitempty"`
	IntegrationId string `protobuf:"bytes,2,opt,name=integration_id,json=integrationId,proto3" json:"integration_id,omitempty"`
	Region        string `protobuf:"bytes,3,opt,name=region,proto3" json:"region,omitempty"`
	Size          int64  `protobuf:"varint,4,opt,name=size,proto3" json:"size,omitempty"`
	Cursor        string `protobuf:"bytes,5,opt,name=cursor,proto3" json:"cursor,omitempty"`
}

func (x *GetDeviceCompatibilityListRequest) Reset() {
	*x = GetDeviceCompatibilityListRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_integration_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDeviceCompatibilityListRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDeviceCompatibilityListRequest) ProtoMessage() {}

func (x *GetDeviceCompatibilityListRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_integration_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDeviceCompatibilityListRequest.ProtoReflect.Descriptor instead.
func (*GetDeviceCompatibilityListRequest) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_integration_proto_rawDescGZIP(), []int{6}
}

func (x *GetDeviceCompatibilityListRequest) GetMakeId() string {
	if x != nil {
		return x.MakeId
	}
	return ""
}

func (x *GetDeviceCompatibilityListRequest) GetIntegrationId() string {
	if x != nil {
		return x.IntegrationId
	}
	return ""
}

func (x *GetDeviceCompatibilityListRequest) GetRegion() string {
	if x != nil {
		return x.Region
	}
	return ""
}

func (x *GetDeviceCompatibilityListRequest) GetSize() int64 {
	if x != nil {
		return x.Size
	}
	return 0
}

func (x *GetDeviceCompatibilityListRequest) GetCursor() string {
	if x != nil {
		return x.Cursor
	}
	return ""
}

type GetDeviceCompatibilityListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Models []*DeviceCompatibilityList `protobuf:"bytes,1,rep,name=models,proto3" json:"models,omitempty"`
	Cursor string                     `protobuf:"bytes,2,opt,name=cursor,proto3" json:"cursor,omitempty"`
}

func (x *GetDeviceCompatibilityListResponse) Reset() {
	*x = GetDeviceCompatibilityListResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_integration_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetDeviceCompatibilityListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDeviceCompatibilityListResponse) ProtoMessage() {}

func (x *GetDeviceCompatibilityListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_integration_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDeviceCompatibilityListResponse.ProtoReflect.Descriptor instead.
func (*GetDeviceCompatibilityListResponse) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_integration_proto_rawDescGZIP(), []int{7}
}

func (x *GetDeviceCompatibilityListResponse) GetModels() []*DeviceCompatibilityList {
	if x != nil {
		return x.Models
	}
	return nil
}

func (x *GetDeviceCompatibilityListResponse) GetCursor() string {
	if x != nil {
		return x.Cursor
	}
	return ""
}

type DeviceCompatibilityList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name  string                   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Years []*DeviceCompatibilities `protobuf:"bytes,2,rep,name=years,proto3" json:"years,omitempty"`
}

func (x *DeviceCompatibilityList) Reset() {
	*x = DeviceCompatibilityList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_integration_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeviceCompatibilityList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeviceCompatibilityList) ProtoMessage() {}

func (x *DeviceCompatibilityList) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_integration_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeviceCompatibilityList.ProtoReflect.Descriptor instead.
func (*DeviceCompatibilityList) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_integration_proto_rawDescGZIP(), []int{8}
}

func (x *DeviceCompatibilityList) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *DeviceCompatibilityList) GetYears() []*DeviceCompatibilities {
	if x != nil {
		return x.Years
	}
	return nil
}

type DeviceCompatibilities struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Year     int32      `protobuf:"varint,1,opt,name=year,proto3" json:"year,omitempty"`
	Features []*Feature `protobuf:"bytes,2,rep,name=features,proto3" json:"features,omitempty"`
	Level    string     `protobuf:"bytes,3,opt,name=level,proto3" json:"level,omitempty"`
}

func (x *DeviceCompatibilities) Reset() {
	*x = DeviceCompatibilities{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_integration_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeviceCompatibilities) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeviceCompatibilities) ProtoMessage() {}

func (x *DeviceCompatibilities) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_integration_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeviceCompatibilities.ProtoReflect.Descriptor instead.
func (*DeviceCompatibilities) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_integration_proto_rawDescGZIP(), []int{9}
}

func (x *DeviceCompatibilities) GetYear() int32 {
	if x != nil {
		return x.Year
	}
	return 0
}

func (x *DeviceCompatibilities) GetFeatures() []*Feature {
	if x != nil {
		return x.Features
	}
	return nil
}

func (x *DeviceCompatibilities) GetLevel() string {
	if x != nil {
		return x.Level
	}
	return ""
}

type Feature struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key          string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	SupportLevel int32  `protobuf:"varint,2,opt,name=support_level,json=supportLevel,proto3" json:"support_level,omitempty"`
	CssIcon      string `protobuf:"bytes,3,opt,name=css_icon,json=cssIcon,proto3" json:"css_icon,omitempty"`
}

func (x *Feature) Reset() {
	*x = Feature{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_integration_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Feature) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Feature) ProtoMessage() {}

func (x *Feature) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_integration_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Feature.ProtoReflect.Descriptor instead.
func (*Feature) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_integration_proto_rawDescGZIP(), []int{10}
}

func (x *Feature) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Feature) GetSupportLevel() int32 {
	if x != nil {
		return x.SupportLevel
	}
	return 0
}

func (x *Feature) GetCssIcon() string {
	if x != nil {
		return x.CssIcon
	}
	return ""
}

var File_pkg_grpc_integration_proto protoreflect.FileDescriptor

var file_pkg_grpc_integration_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x67,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x67, 0x72,
	0x70, 0x63, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x29, 0x0a, 0x17, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x61,
	0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x32, 0x0a, 0x20, 0x47, 0x65,
	0x74, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x65, 0x61, 0x74,
	0x75, 0x72, 0x65, 0x42, 0x79, 0x49, 0x44, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0xc9,
	0x01, 0x0a, 0x27, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4f, 0x72, 0x55, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x65, 0x61, 0x74,
	0x75, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x29, 0x0a, 0x10, 0x65, 0x6c,
	0x61, 0x73, 0x74, 0x69, 0x63, 0x5f, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x79, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x65, 0x6c, 0x61, 0x73, 0x74, 0x69, 0x63, 0x50, 0x72, 0x6f,
	0x70, 0x65, 0x72, 0x74, 0x79, 0x12, 0x21, 0x0a, 0x0c, 0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79,
	0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64, 0x69, 0x73,
	0x70, 0x6c, 0x61, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x63, 0x73, 0x73, 0x5f,
	0x69, 0x63, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x73, 0x73, 0x49,
	0x63, 0x6f, 0x6e, 0x12, 0x25, 0x0a, 0x0e, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x5f, 0x77,
	0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x02, 0x52, 0x0d, 0x66, 0x65, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x57, 0x65, 0x69, 0x67, 0x68, 0x74, 0x22, 0x31, 0x0a, 0x1f, 0x44, 0x65,
	0x6c, 0x65, 0x74, 0x65, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46,
	0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x7b, 0x0a,
	0x21, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46,
	0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x56, 0x0a, 0x14, 0x69, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x5f, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x23, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x74, 0x65, 0x67,
	0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x52, 0x13, 0x69, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x22, 0xd0, 0x01, 0x0a, 0x1d, 0x47,
	0x65, 0x74, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x65, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1f, 0x0a, 0x0b,
	0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0a, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x4b, 0x65, 0x79, 0x12, 0x29, 0x0a,
	0x10, 0x65, 0x6c, 0x61, 0x73, 0x74, 0x69, 0x63, 0x5f, 0x70, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74,
	0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x65, 0x6c, 0x61, 0x73, 0x74, 0x69, 0x63,
	0x50, 0x72, 0x6f, 0x70, 0x65, 0x72, 0x74, 0x79, 0x12, 0x21, 0x0a, 0x0c, 0x64, 0x69, 0x73, 0x70,
	0x6c, 0x61, 0x79, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b,
	0x64, 0x69, 0x73, 0x70, 0x6c, 0x61, 0x79, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x63,
	0x73, 0x73, 0x5f, 0x69, 0x63, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63,
	0x73, 0x73, 0x49, 0x63, 0x6f, 0x6e, 0x12, 0x25, 0x0a, 0x0e, 0x66, 0x65, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x5f, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x02, 0x52, 0x0d,
	0x66, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x57, 0x65, 0x69, 0x67, 0x68, 0x74, 0x22, 0xa7, 0x01,
	0x0a, 0x21, 0x47, 0x65, 0x74, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x43, 0x6f, 0x6d, 0x70, 0x61,
	0x74, 0x69, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x6d, 0x61, 0x6b, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x61, 0x6b, 0x65, 0x49, 0x64, 0x12, 0x25, 0x0a, 0x0e,
	0x69, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x69, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x12, 0x12, 0x0a, 0x04, 0x73,
	0x69, 0x7a, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x73, 0x69, 0x7a, 0x65, 0x12,
	0x16, 0x0a, 0x06, 0x63, 0x75, 0x72, 0x73, 0x6f, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x63, 0x75, 0x72, 0x73, 0x6f, 0x72, 0x22, 0x73, 0x0a, 0x22, 0x47, 0x65, 0x74, 0x44, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x43, 0x6f, 0x6d, 0x70, 0x61, 0x74, 0x69, 0x62, 0x69, 0x6c, 0x69, 0x74,
	0x79, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x35, 0x0a,
	0x06, 0x6d, 0x6f, 0x64, 0x65, 0x6c, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e,
	0x67, 0x72, 0x70, 0x63, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x43, 0x6f, 0x6d, 0x70, 0x61,
	0x74, 0x69, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x06, 0x6d, 0x6f,
	0x64, 0x65, 0x6c, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x63, 0x75, 0x72, 0x73, 0x6f, 0x72, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x63, 0x75, 0x72, 0x73, 0x6f, 0x72, 0x22, 0x60, 0x0a, 0x17,
	0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x43, 0x6f, 0x6d, 0x70, 0x61, 0x74, 0x69, 0x62, 0x69, 0x6c,
	0x69, 0x74, 0x79, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x31, 0x0a, 0x05, 0x79,
	0x65, 0x61, 0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x67, 0x72, 0x70,
	0x63, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x43, 0x6f, 0x6d, 0x70, 0x61, 0x74, 0x69, 0x62,
	0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x52, 0x05, 0x79, 0x65, 0x61, 0x72, 0x73, 0x22, 0x6c,
	0x0a, 0x15, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x43, 0x6f, 0x6d, 0x70, 0x61, 0x74, 0x69, 0x62,
	0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x79, 0x65, 0x61, 0x72, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x79, 0x65, 0x61, 0x72, 0x12, 0x29, 0x0a, 0x08, 0x66,
	0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0d, 0x2e,
	0x67, 0x72, 0x70, 0x63, 0x2e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x52, 0x08, 0x66, 0x65,
	0x61, 0x74, 0x75, 0x72, 0x65, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x22, 0x5b, 0x0a, 0x07,
	0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x23, 0x0a, 0x0d, 0x73, 0x75, 0x70,
	0x70, 0x6f, 0x72, 0x74, 0x5f, 0x6c, 0x65, 0x76, 0x65, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x0c, 0x73, 0x75, 0x70, 0x70, 0x6f, 0x72, 0x74, 0x4c, 0x65, 0x76, 0x65, 0x6c, 0x12, 0x19,
	0x0a, 0x08, 0x63, 0x73, 0x73, 0x5f, 0x69, 0x63, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x63, 0x73, 0x73, 0x49, 0x63, 0x6f, 0x6e, 0x32, 0xfe, 0x04, 0x0a, 0x12, 0x49, 0x6e,
	0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65,
	0x12, 0x6d, 0x0a, 0x18, 0x47, 0x65, 0x74, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x43, 0x6f, 0x6d,
	0x70, 0x61, 0x74, 0x69, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x69, 0x65, 0x73, 0x12, 0x27, 0x2e, 0x67,
	0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x43, 0x6f, 0x6d,
	0x70, 0x61, 0x74, 0x69, 0x62, 0x69, 0x6c, 0x69, 0x74, 0x79, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x28, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74,
	0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x43, 0x6f, 0x6d, 0x70, 0x61, 0x74, 0x69, 0x62, 0x69, 0x6c,
	0x69, 0x74, 0x79, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12,
	0x68, 0x0a, 0x19, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x42, 0x79, 0x49, 0x44, 0x12, 0x26, 0x2e, 0x67,
	0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x42, 0x79, 0x49, 0x44, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x23, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x49,
	0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x59, 0x0a, 0x16, 0x47, 0x65, 0x74,
	0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x65, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x73, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x27, 0x2e, 0x67, 0x72,
	0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x68, 0x0a, 0x18, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x49, 0x6e,
	0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65,
	0x12, 0x2d, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4f, 0x72,
	0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x1d, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x42, 0x61, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x68,
	0x0a, 0x18, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x46, 0x65, 0x61, 0x74, 0x75, 0x72, 0x65, 0x12, 0x2d, 0x2e, 0x67, 0x72, 0x70,
	0x63, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x4f, 0x72, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x65, 0x61, 0x74, 0x75,
	0x72, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x67, 0x72, 0x70, 0x63,
	0x2e, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x61, 0x73, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x60, 0x0a, 0x18, 0x44, 0x65, 0x6c, 0x65,
	0x74, 0x65, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x65, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x12, 0x25, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x44, 0x65, 0x6c, 0x65,
	0x74, 0x65, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x65, 0x61,
	0x74, 0x75, 0x72, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1d, 0x2e, 0x67, 0x72,
	0x70, 0x63, 0x2e, 0x49, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x42, 0x61,
	0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x39, 0x5a, 0x37, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x44, 0x49, 0x4d, 0x4f, 0x2d, 0x4e, 0x65,
	0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2f, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x2d, 0x64, 0x65, 0x66,
	0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2d, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x6b, 0x67,
	0x2f, 0x67, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_grpc_integration_proto_rawDescOnce sync.Once
	file_pkg_grpc_integration_proto_rawDescData = file_pkg_grpc_integration_proto_rawDesc
)

func file_pkg_grpc_integration_proto_rawDescGZIP() []byte {
	file_pkg_grpc_integration_proto_rawDescOnce.Do(func() {
		file_pkg_grpc_integration_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_grpc_integration_proto_rawDescData)
	})
	return file_pkg_grpc_integration_proto_rawDescData
}

var file_pkg_grpc_integration_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_pkg_grpc_integration_proto_goTypes = []interface{}{
	(*IntegrationBaseResponse)(nil),                 // 0: grpc.IntegrationBaseResponse
	(*GetIntegrationFeatureByIDRequest)(nil),        // 1: grpc.GetIntegrationFeatureByIDRequest
	(*CreateOrUpdateIntegrationFeatureRequest)(nil), // 2: grpc.CreateOrUpdateIntegrationFeatureRequest
	(*DeleteIntegrationFeatureRequest)(nil),         // 3: grpc.DeleteIntegrationFeatureRequest
	(*GetIntegrationFeatureListResponse)(nil),       // 4: grpc.GetIntegrationFeatureListResponse
	(*GetIntegrationFeatureResponse)(nil),           // 5: grpc.GetIntegrationFeatureResponse
	(*GetDeviceCompatibilityListRequest)(nil),       // 6: grpc.GetDeviceCompatibilityListRequest
	(*GetDeviceCompatibilityListResponse)(nil),      // 7: grpc.GetDeviceCompatibilityListResponse
	(*DeviceCompatibilityList)(nil),                 // 8: grpc.DeviceCompatibilityList
	(*DeviceCompatibilities)(nil),                   // 9: grpc.DeviceCompatibilities
	(*Feature)(nil),                                 // 10: grpc.Feature
	(*emptypb.Empty)(nil),                           // 11: google.protobuf.Empty
}
var file_pkg_grpc_integration_proto_depIdxs = []int32{
	5,  // 0: grpc.GetIntegrationFeatureListResponse.integration_features:type_name -> grpc.GetIntegrationFeatureResponse
	8,  // 1: grpc.GetDeviceCompatibilityListResponse.models:type_name -> grpc.DeviceCompatibilityList
	9,  // 2: grpc.DeviceCompatibilityList.years:type_name -> grpc.DeviceCompatibilities
	10, // 3: grpc.DeviceCompatibilities.features:type_name -> grpc.Feature
	6,  // 4: grpc.IntegrationService.GetDeviceCompatibilities:input_type -> grpc.GetDeviceCompatibilityListRequest
	1,  // 5: grpc.IntegrationService.GetIntegrationFeatureByID:input_type -> grpc.GetIntegrationFeatureByIDRequest
	11, // 6: grpc.IntegrationService.GetIntegrationFeatures:input_type -> google.protobuf.Empty
	2,  // 7: grpc.IntegrationService.CreateIntegrationFeature:input_type -> grpc.CreateOrUpdateIntegrationFeatureRequest
	2,  // 8: grpc.IntegrationService.UpdateIntegrationFeature:input_type -> grpc.CreateOrUpdateIntegrationFeatureRequest
	3,  // 9: grpc.IntegrationService.DeleteIntegrationFeature:input_type -> grpc.DeleteIntegrationFeatureRequest
	7,  // 10: grpc.IntegrationService.GetDeviceCompatibilities:output_type -> grpc.GetDeviceCompatibilityListResponse
	5,  // 11: grpc.IntegrationService.GetIntegrationFeatureByID:output_type -> grpc.GetIntegrationFeatureResponse
	4,  // 12: grpc.IntegrationService.GetIntegrationFeatures:output_type -> grpc.GetIntegrationFeatureListResponse
	0,  // 13: grpc.IntegrationService.CreateIntegrationFeature:output_type -> grpc.IntegrationBaseResponse
	0,  // 14: grpc.IntegrationService.UpdateIntegrationFeature:output_type -> grpc.IntegrationBaseResponse
	0,  // 15: grpc.IntegrationService.DeleteIntegrationFeature:output_type -> grpc.IntegrationBaseResponse
	10, // [10:16] is the sub-list for method output_type
	4,  // [4:10] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
}

func init() { file_pkg_grpc_integration_proto_init() }
func file_pkg_grpc_integration_proto_init() {
	if File_pkg_grpc_integration_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_grpc_integration_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IntegrationBaseResponse); i {
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
		file_pkg_grpc_integration_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetIntegrationFeatureByIDRequest); i {
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
		file_pkg_grpc_integration_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateOrUpdateIntegrationFeatureRequest); i {
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
		file_pkg_grpc_integration_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteIntegrationFeatureRequest); i {
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
		file_pkg_grpc_integration_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetIntegrationFeatureListResponse); i {
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
		file_pkg_grpc_integration_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetIntegrationFeatureResponse); i {
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
		file_pkg_grpc_integration_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDeviceCompatibilityListRequest); i {
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
		file_pkg_grpc_integration_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetDeviceCompatibilityListResponse); i {
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
		file_pkg_grpc_integration_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeviceCompatibilityList); i {
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
		file_pkg_grpc_integration_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeviceCompatibilities); i {
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
		file_pkg_grpc_integration_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Feature); i {
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
			RawDescriptor: file_pkg_grpc_integration_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_grpc_integration_proto_goTypes,
		DependencyIndexes: file_pkg_grpc_integration_proto_depIdxs,
		MessageInfos:      file_pkg_grpc_integration_proto_msgTypes,
	}.Build()
	File_pkg_grpc_integration_proto = out.File
	file_pkg_grpc_integration_proto_rawDesc = nil
	file_pkg_grpc_integration_proto_goTypes = nil
	file_pkg_grpc_integration_proto_depIdxs = nil
}
