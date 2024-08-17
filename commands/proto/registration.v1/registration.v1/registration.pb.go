// to generate go files protoc --go_out=. registration.proto

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v3.12.4
// source: registration.v1/registration.proto

package registration_v1

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

type RegistrationMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Email    string `protobuf:"bytes,1,opt,name=Email,proto3" json:"Email,omitempty"`
	FullName string `protobuf:"bytes,2,opt,name=FullName,proto3" json:"FullName,omitempty"`
}

func (x *RegistrationMessage) Reset() {
	*x = RegistrationMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_registration_v1_registration_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RegistrationMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegistrationMessage) ProtoMessage() {}

func (x *RegistrationMessage) ProtoReflect() protoreflect.Message {
	mi := &file_registration_v1_registration_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegistrationMessage.ProtoReflect.Descriptor instead.
func (*RegistrationMessage) Descriptor() ([]byte, []int) {
	return file_registration_v1_registration_proto_rawDescGZIP(), []int{0}
}

func (x *RegistrationMessage) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

func (x *RegistrationMessage) GetFullName() string {
	if x != nil {
		return x.FullName
	}
	return ""
}

var File_registration_v1_registration_proto protoreflect.FileDescriptor

var file_registration_v1_registration_proto_rawDesc = []byte{
	0x0a, 0x22, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x76,
	0x31, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x2e, 0x76, 0x31, 0x22, 0x47, 0x0a, 0x13, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x14, 0x0a, 0x05,
	0x45, 0x6d, 0x61, 0x69, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x45, 0x6d, 0x61,
	0x69, 0x6c, 0x12, 0x1a, 0x0a, 0x08, 0x46, 0x75, 0x6c, 0x6c, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x46, 0x75, 0x6c, 0x6c, 0x4e, 0x61, 0x6d, 0x65, 0x42, 0x13,
	0x5a, 0x11, 0x2e, 0x2f, 0x72, 0x65, 0x67, 0x69, 0x73, 0x74, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2e, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_registration_v1_registration_proto_rawDescOnce sync.Once
	file_registration_v1_registration_proto_rawDescData = file_registration_v1_registration_proto_rawDesc
)

func file_registration_v1_registration_proto_rawDescGZIP() []byte {
	file_registration_v1_registration_proto_rawDescOnce.Do(func() {
		file_registration_v1_registration_proto_rawDescData = protoimpl.X.CompressGZIP(file_registration_v1_registration_proto_rawDescData)
	})
	return file_registration_v1_registration_proto_rawDescData
}

var file_registration_v1_registration_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_registration_v1_registration_proto_goTypes = []any{
	(*RegistrationMessage)(nil), // 0: Registration.v1.RegistrationMessage
}
var file_registration_v1_registration_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_registration_v1_registration_proto_init() }
func file_registration_v1_registration_proto_init() {
	if File_registration_v1_registration_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_registration_v1_registration_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*RegistrationMessage); i {
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
			RawDescriptor: file_registration_v1_registration_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_registration_v1_registration_proto_goTypes,
		DependencyIndexes: file_registration_v1_registration_proto_depIdxs,
		MessageInfos:      file_registration_v1_registration_proto_msgTypes,
	}.Build()
	File_registration_v1_registration_proto = out.File
	file_registration_v1_registration_proto_rawDesc = nil
	file_registration_v1_registration_proto_goTypes = nil
	file_registration_v1_registration_proto_depIdxs = nil
}