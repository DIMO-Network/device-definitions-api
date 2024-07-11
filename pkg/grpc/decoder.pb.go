// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.1
// source: pkg/grpc/decoder.proto

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

type DecodeVinRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Vin                string `protobuf:"bytes,1,opt,name=vin,proto3" json:"vin,omitempty"`
	Country            string `protobuf:"bytes,2,opt,name=country,proto3" json:"country,omitempty"`
	KnownModel         string `protobuf:"bytes,3,opt,name=knownModel,proto3" json:"knownModel,omitempty"`
	KnownYear          int32  `protobuf:"varint,4,opt,name=knownYear,proto3" json:"knownYear,omitempty"`
	DeviceDefinitionId string `protobuf:"bytes,5,opt,name=device_definition_id,json=deviceDefinitionId,proto3" json:"device_definition_id,omitempty"`
}

func (x *DecodeVinRequest) Reset() {
	*x = DecodeVinRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_decoder_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DecodeVinRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DecodeVinRequest) ProtoMessage() {}

func (x *DecodeVinRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_decoder_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DecodeVinRequest.ProtoReflect.Descriptor instead.
func (*DecodeVinRequest) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_decoder_proto_rawDescGZIP(), []int{0}
}

func (x *DecodeVinRequest) GetVin() string {
	if x != nil {
		return x.Vin
	}
	return ""
}

func (x *DecodeVinRequest) GetCountry() string {
	if x != nil {
		return x.Country
	}
	return ""
}

func (x *DecodeVinRequest) GetKnownModel() string {
	if x != nil {
		return x.KnownModel
	}
	return ""
}

func (x *DecodeVinRequest) GetKnownYear() int32 {
	if x != nil {
		return x.KnownYear
	}
	return 0
}

func (x *DecodeVinRequest) GetDeviceDefinitionId() string {
	if x != nil {
		return x.DeviceDefinitionId
	}
	return ""
}

type DecodeVinResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DeviceMakeId       string `protobuf:"bytes,1,opt,name=device_make_id,json=deviceMakeId,proto3" json:"device_make_id,omitempty"`
	DeviceDefinitionId string `protobuf:"bytes,2,opt,name=device_definition_id,json=deviceDefinitionId,proto3" json:"device_definition_id,omitempty"`
	DeviceStyleId      string `protobuf:"bytes,3,opt,name=device_style_id,json=deviceStyleId,proto3" json:"device_style_id,omitempty"`
	Year               int32  `protobuf:"varint,4,opt,name=year,proto3" json:"year,omitempty"`
	Source             string `protobuf:"bytes,5,opt,name=source,proto3" json:"source,omitempty"`
	Powertrain         string `protobuf:"bytes,6,opt,name=powertrain,proto3" json:"powertrain,omitempty"`
	NameSlug           string `protobuf:"bytes,7,opt,name=name_slug,json=nameSlug,proto3" json:"name_slug,omitempty"`
}

func (x *DecodeVinResponse) Reset() {
	*x = DecodeVinResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_grpc_decoder_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DecodeVinResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DecodeVinResponse) ProtoMessage() {}

func (x *DecodeVinResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_grpc_decoder_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DecodeVinResponse.ProtoReflect.Descriptor instead.
func (*DecodeVinResponse) Descriptor() ([]byte, []int) {
	return file_pkg_grpc_decoder_proto_rawDescGZIP(), []int{1}
}

func (x *DecodeVinResponse) GetDeviceMakeId() string {
	if x != nil {
		return x.DeviceMakeId
	}
	return ""
}

func (x *DecodeVinResponse) GetDeviceDefinitionId() string {
	if x != nil {
		return x.DeviceDefinitionId
	}
	return ""
}

func (x *DecodeVinResponse) GetDeviceStyleId() string {
	if x != nil {
		return x.DeviceStyleId
	}
	return ""
}

func (x *DecodeVinResponse) GetYear() int32 {
	if x != nil {
		return x.Year
	}
	return 0
}

func (x *DecodeVinResponse) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *DecodeVinResponse) GetPowertrain() string {
	if x != nil {
		return x.Powertrain
	}
	return ""
}

func (x *DecodeVinResponse) GetNameSlug() string {
	if x != nil {
		return x.NameSlug
	}
	return ""
}

var File_pkg_grpc_decoder_proto protoreflect.FileDescriptor

var file_pkg_grpc_decoder_proto_rawDesc = []byte{
	0x0a, 0x16, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x64, 0x65, 0x63, 0x6f, 0x64,
	0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x67, 0x72, 0x70, 0x63, 0x22, 0xae,
	0x01, 0x0a, 0x10, 0x44, 0x65, 0x63, 0x6f, 0x64, 0x65, 0x56, 0x69, 0x6e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x76, 0x69, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x76, 0x69, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x1e, 0x0a, 0x0a, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0a, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x4d, 0x6f, 0x64, 0x65, 0x6c, 0x12,
	0x1c, 0x0a, 0x09, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x59, 0x65, 0x61, 0x72, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x09, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x59, 0x65, 0x61, 0x72, 0x12, 0x30, 0x0a,
	0x14, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69,
	0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x64, 0x65, 0x76,
	0x69, 0x63, 0x65, 0x44, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x22,
	0xfc, 0x01, 0x0a, 0x11, 0x44, 0x65, 0x63, 0x6f, 0x64, 0x65, 0x56, 0x69, 0x6e, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x24, 0x0a, 0x0e, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5f,
	0x6d, 0x61, 0x6b, 0x65, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x64,
	0x65, 0x76, 0x69, 0x63, 0x65, 0x4d, 0x61, 0x6b, 0x65, 0x49, 0x64, 0x12, 0x30, 0x0a, 0x14, 0x64,
	0x65, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e,
	0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x64, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x44, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x12, 0x26, 0x0a,
	0x0f, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x5f, 0x73, 0x74, 0x79, 0x6c, 0x65, 0x5f, 0x69, 0x64,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x53, 0x74,
	0x79, 0x6c, 0x65, 0x49, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x79, 0x65, 0x61, 0x72, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x04, 0x79, 0x65, 0x61, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x6f, 0x77, 0x65, 0x72, 0x74, 0x72, 0x61, 0x69, 0x6e, 0x18,
	0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x70, 0x6f, 0x77, 0x65, 0x72, 0x74, 0x72, 0x61, 0x69,
	0x6e, 0x12, 0x1b, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x5f, 0x73, 0x6c, 0x75, 0x67, 0x18, 0x07,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6e, 0x61, 0x6d, 0x65, 0x53, 0x6c, 0x75, 0x67, 0x32, 0x51,
	0x0a, 0x11, 0x56, 0x69, 0x6e, 0x44, 0x65, 0x63, 0x6f, 0x64, 0x65, 0x72, 0x53, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x12, 0x3c, 0x0a, 0x09, 0x44, 0x65, 0x63, 0x6f, 0x64, 0x65, 0x56, 0x69, 0x6e,
	0x12, 0x16, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x44, 0x65, 0x63, 0x6f, 0x64, 0x65, 0x56, 0x69,
	0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e,
	0x44, 0x65, 0x63, 0x6f, 0x64, 0x65, 0x56, 0x69, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x42, 0x39, 0x5a, 0x37, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x44, 0x49, 0x4d, 0x4f, 0x2d, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2f, 0x64, 0x65, 0x76,
	0x69, 0x63, 0x65, 0x2d, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2d,
	0x61, 0x70, 0x69, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_grpc_decoder_proto_rawDescOnce sync.Once
	file_pkg_grpc_decoder_proto_rawDescData = file_pkg_grpc_decoder_proto_rawDesc
)

func file_pkg_grpc_decoder_proto_rawDescGZIP() []byte {
	file_pkg_grpc_decoder_proto_rawDescOnce.Do(func() {
		file_pkg_grpc_decoder_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_grpc_decoder_proto_rawDescData)
	})
	return file_pkg_grpc_decoder_proto_rawDescData
}

var file_pkg_grpc_decoder_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_pkg_grpc_decoder_proto_goTypes = []any{
	(*DecodeVinRequest)(nil),  // 0: grpc.DecodeVinRequest
	(*DecodeVinResponse)(nil), // 1: grpc.DecodeVinResponse
}
var file_pkg_grpc_decoder_proto_depIdxs = []int32{
	0, // 0: grpc.VinDecoderService.DecodeVin:input_type -> grpc.DecodeVinRequest
	1, // 1: grpc.VinDecoderService.DecodeVin:output_type -> grpc.DecodeVinResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pkg_grpc_decoder_proto_init() }
func file_pkg_grpc_decoder_proto_init() {
	if File_pkg_grpc_decoder_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_grpc_decoder_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*DecodeVinRequest); i {
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
		file_pkg_grpc_decoder_proto_msgTypes[1].Exporter = func(v any, i int) any {
			switch v := v.(*DecodeVinResponse); i {
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
			RawDescriptor: file_pkg_grpc_decoder_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_grpc_decoder_proto_goTypes,
		DependencyIndexes: file_pkg_grpc_decoder_proto_depIdxs,
		MessageInfos:      file_pkg_grpc_decoder_proto_msgTypes,
	}.Build()
	File_pkg_grpc_decoder_proto = out.File
	file_pkg_grpc_decoder_proto_rawDesc = nil
	file_pkg_grpc_decoder_proto_goTypes = nil
	file_pkg_grpc_decoder_proto_depIdxs = nil
}
