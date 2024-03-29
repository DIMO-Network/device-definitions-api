// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v4.25.2
// source: pkg/grpc/recalls.proto

package grpc

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type RecallItem struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DeviceDefinitionId string                 `protobuf:"bytes,1,opt,name=device_definition_id,json=deviceDefinitionId,proto3" json:"device_definition_id,omitempty"`
	Name               string                 `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Year               int32                  `protobuf:"varint,3,opt,name=year,proto3" json:"year,omitempty"`
	Description        string                 `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	ComponentName      string                 `protobuf:"bytes,5,opt,name=componentName,proto3" json:"componentName,omitempty"`
	ConsequenceDefect  string                 `protobuf:"bytes,6,opt,name=consequenceDefect,proto3" json:"consequenceDefect,omitempty"`
	ManufactureCampNo  string                 `protobuf:"bytes,7,opt,name=manufactureCampNo,proto3" json:"manufactureCampNo,omitempty"`
	Date               *timestamppb.Timestamp `protobuf:"bytes,8,opt,name=date,proto3" json:"date,omitempty"`
}

func (x *RecallItem) Reset() {
	*x = RecallItem{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_recalls_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecallItem) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecallItem) ProtoMessage() {}

func (x *RecallItem) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_recalls_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecallItem.ProtoReflect.Descriptor instead.
func (*RecallItem) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_recalls_proto_rawDescGZIP(), []int{0}
}

func (x *RecallItem) GetDeviceDefinitionId() string {
	if x != nil {
		return x.DeviceDefinitionId
	}
	return ""
}

func (x *RecallItem) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *RecallItem) GetYear() int32 {
	if x != nil {
		return x.Year
	}
	return 0
}

func (x *RecallItem) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *RecallItem) GetComponentName() string {
	if x != nil {
		return x.ComponentName
	}
	return ""
}

func (x *RecallItem) GetConsequenceDefect() string {
	if x != nil {
		return x.ConsequenceDefect
	}
	return ""
}

func (x *RecallItem) GetManufactureCampNo() string {
	if x != nil {
		return x.ManufactureCampNo
	}
	return ""
}

func (x *RecallItem) GetDate() *timestamppb.Timestamp {
	if x != nil {
		return x.Date
	}
	return nil
}

type GetRecallsByMakeRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MakeId string `protobuf:"bytes,1,opt,name=make_id,json=makeId,proto3" json:"make_id,omitempty"`
}

func (x *GetRecallsByMakeRequest) Reset() {
	*x = GetRecallsByMakeRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_recalls_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRecallsByMakeRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRecallsByMakeRequest) ProtoMessage() {}

func (x *GetRecallsByMakeRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_recalls_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRecallsByMakeRequest.ProtoReflect.Descriptor instead.
func (*GetRecallsByMakeRequest) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_recalls_proto_rawDescGZIP(), []int{1}
}

func (x *GetRecallsByMakeRequest) GetMakeId() string {
	if x != nil {
		return x.MakeId
	}
	return ""
}

type GetRecallsByModelRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DeviceDefinitionId string `protobuf:"bytes,1,opt,name=device_definition_id,json=deviceDefinitionId,proto3" json:"device_definition_id,omitempty"`
}

func (x *GetRecallsByModelRequest) Reset() {
	*x = GetRecallsByModelRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_recalls_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRecallsByModelRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRecallsByModelRequest) ProtoMessage() {}

func (x *GetRecallsByModelRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_recalls_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRecallsByModelRequest.ProtoReflect.Descriptor instead.
func (*GetRecallsByModelRequest) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_recalls_proto_rawDescGZIP(), []int{2}
}

func (x *GetRecallsByModelRequest) GetDeviceDefinitionId() string {
	if x != nil {
		return x.DeviceDefinitionId
	}
	return ""
}

type GetRecallsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Recalls []*RecallItem `protobuf:"bytes,1,rep,name=recalls,proto3" json:"recalls,omitempty"`
}

func (x *GetRecallsResponse) Reset() {
	*x = GetRecallsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_recalls_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetRecallsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRecallsResponse) ProtoMessage() {}

func (x *GetRecallsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_recalls_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetRecallsResponse.ProtoReflect.Descriptor instead.
func (*GetRecallsResponse) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_recalls_proto_rawDescGZIP(), []int{3}
}

func (x *GetRecallsResponse) GetRecalls() []*RecallItem {
	if x != nil {
		return x.Recalls
	}
	return nil
}

var File_pkg_grpc_recalls_proto protoreflect.FileDescriptor

var file_pkg_grpc_recalls_proto_rawDesc = []byte{
	0x0a, 0x16, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x72, 0x65, 0x63, 0x61, 0x6c,
	0x6c, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x67, 0x72, 0x70, 0x63, 0x1a, 0x1f,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0xba, 0x02, 0x0a, 0x0a, 0x52, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x49, 0x74, 0x65, 0x6d, 0x12, 0x30,
	0x0a, 0x14, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74,
	0x69, 0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x64, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x44, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64,
	0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x79, 0x65, 0x61, 0x72, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x04, 0x79, 0x65, 0x61, 0x72, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63,
	0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x64,
	0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x24, 0x0a, 0x0d, 0x63, 0x6f,
	0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0d, 0x63, 0x6f, 0x6d, 0x70, 0x6f, 0x6e, 0x65, 0x6e, 0x74, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x2c, 0x0a, 0x11, 0x63, 0x6f, 0x6e, 0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x44,
	0x65, 0x66, 0x65, 0x63, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11, 0x63, 0x6f, 0x6e,
	0x73, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x65, 0x44, 0x65, 0x66, 0x65, 0x63, 0x74, 0x12, 0x2c,
	0x0a, 0x11, 0x6d, 0x61, 0x6e, 0x75, 0x66, 0x61, 0x63, 0x74, 0x75, 0x72, 0x65, 0x43, 0x61, 0x6d,
	0x70, 0x4e, 0x6f, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11, 0x6d, 0x61, 0x6e, 0x75, 0x66,
	0x61, 0x63, 0x74, 0x75, 0x72, 0x65, 0x43, 0x61, 0x6d, 0x70, 0x4e, 0x6f, 0x12, 0x2e, 0x0a, 0x04,
	0x64, 0x61, 0x74, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d,
	0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x04, 0x64, 0x61, 0x74, 0x65, 0x22, 0x32, 0x0a, 0x17,
	0x47, 0x65, 0x74, 0x52, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x73, 0x42, 0x79, 0x4d, 0x61, 0x6b, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x17, 0x0a, 0x07, 0x6d, 0x61, 0x6b, 0x65, 0x5f,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x61, 0x6b, 0x65, 0x49, 0x64,
	0x22, 0x4c, 0x0a, 0x18, 0x47, 0x65, 0x74, 0x52, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x73, 0x42, 0x79,
	0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x30, 0x0a, 0x14,
	0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f,
	0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x64, 0x65, 0x76, 0x69,
	0x63, 0x65, 0x44, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x22, 0x40,
	0x0a, 0x12, 0x47, 0x65, 0x74, 0x52, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x73, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2a, 0x0a, 0x07, 0x72, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x73, 0x18,
	0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x52, 0x65, 0x63,
	0x61, 0x6c, 0x6c, 0x49, 0x74, 0x65, 0x6d, 0x52, 0x07, 0x72, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x73,
	0x32, 0xfe, 0x01, 0x0a, 0x0e, 0x52, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x73, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x50, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x52, 0x65, 0x63, 0x61, 0x6c, 0x6c,
	0x73, 0x42, 0x79, 0x4d, 0x61, 0x6b, 0x65, 0x12, 0x1d, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x47,
	0x65, 0x74, 0x52, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x73, 0x42, 0x79, 0x4d, 0x61, 0x6b, 0x65, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65,
	0x74, 0x52, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x03, 0x88, 0x02, 0x01, 0x12, 0x4b, 0x0a, 0x16, 0x47, 0x65, 0x74, 0x53, 0x74, 0x72, 0x65,
	0x61, 0x6d, 0x52, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x73, 0x42, 0x79, 0x4d, 0x61, 0x6b, 0x65, 0x12,
	0x1d, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x47, 0x65, 0x74, 0x52, 0x65, 0x63, 0x61, 0x6c, 0x6c,
	0x73, 0x42, 0x79, 0x4d, 0x61, 0x6b, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x10,
	0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x52, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x49, 0x74, 0x65, 0x6d,
	0x30, 0x01, 0x12, 0x4d, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x52, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x73,
	0x42, 0x79, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x12, 0x1e, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x47,
	0x65, 0x74, 0x52, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x73, 0x42, 0x79, 0x4d, 0x6f, 0x64, 0x65, 0x6c,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x47,
	0x65, 0x74, 0x52, 0x65, 0x63, 0x61, 0x6c, 0x6c, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x42, 0x39, 0x5a, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x44, 0x49, 0x4d, 0x4f, 0x2d, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2f, 0x64, 0x65, 0x76,
	0x69, 0x63, 0x65, 0x2d, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2d,
	0x61, 0x70, 0x69, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_grpc_recalls_proto_rawDescOnce sync.Once
	file_pkg_grpc_recalls_proto_rawDescData = file_pkg_grpc_recalls_proto_rawDesc
)

func file_pkg_grpc_recalls_proto_rawDescGZIP() []byte {
	file_pkg_grpc_recalls_proto_rawDescOnce.Do(func() {
		file_pkg_grpc_recalls_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_grpc_recalls_proto_rawDescData)
	})
	return file_pkg_grpc_recalls_proto_rawDescData
}

var file_pkg_grpc_recalls_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_pkg_grpc_recalls_proto_goTypes = []interface{}{
	(*RecallItem)(nil),               // 0: grpc.RecallItem
	(*GetRecallsByMakeRequest)(nil),  // 1: grpc.GetRecallsByMakeRequest
	(*GetRecallsByModelRequest)(nil), // 2: grpc.GetRecallsByModelRequest
	(*GetRecallsResponse)(nil),       // 3: grpc.GetRecallsResponse
	(*timestamppb.Timestamp)(nil),    // 4: google.protobuf.Timestamp
}
var file_pkg_grpc_recalls_proto_depIdxs = []int32{
	4, // 0: grpc.RecallItem.date:type_name -> google.protobuf.Timestamp
	0, // 1: grpc.GetRecallsResponse.recalls:type_name -> grpc.RecallItem
	1, // 2: grpc.RecallsService.GetRecallsByMake:input_type -> grpc.GetRecallsByMakeRequest
	1, // 3: grpc.RecallsService.GetStreamRecallsByMake:input_type -> grpc.GetRecallsByMakeRequest
	2, // 4: grpc.RecallsService.GetRecallsByModel:input_type -> grpc.GetRecallsByModelRequest
	3, // 5: grpc.RecallsService.GetRecallsByMake:output_type -> grpc.GetRecallsResponse
	0, // 6: grpc.RecallsService.GetStreamRecallsByMake:output_type -> grpc.RecallItem
	3, // 7: grpc.RecallsService.GetRecallsByModel:output_type -> grpc.GetRecallsResponse
	5, // [5:8] is the sub-list for method output_type
	2, // [2:5] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_pkg_grpc_recalls_proto_init() }
func file_pkg_grpc_recalls_proto_init() {
	if File_pkg_grpc_recalls_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_grpc_recalls_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RecallItem); i {
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
		file_pkg_grpc_recalls_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRecallsByMakeRequest); i {
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
		file_pkg_grpc_recalls_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRecallsByModelRequest); i {
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
		file_pkg_grpc_recalls_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetRecallsResponse); i {
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
			RawDescriptor: file_pkg_grpc_recalls_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_grpc_recalls_proto_goTypes,
		DependencyIndexes: file_pkg_grpc_recalls_proto_depIdxs,
		MessageInfos:      file_pkg_grpc_recalls_proto_msgTypes,
	}.Build()
	File_pkg_grpc_recalls_proto = out.File
	file_pkg_grpc_recalls_proto_rawDesc = nil
	file_pkg_grpc_recalls_proto_goTypes = nil
	file_pkg_grpc_recalls_proto_depIdxs = nil
}
