// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.1
// 	protoc        (unknown)
// source: jimmy/v1/config.proto

package jimmyv1

import (
	_ "buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
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

// The .jimmy.yml configuration file.
type Config struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The location of the migrations directory.
	Path string `protobuf:"bytes,1,opt,name=path,proto3" json:"path,omitempty"`
	// The Google project ID.
	ProjectId string `protobuf:"bytes,2,opt,name=project_id,json=projectId,proto3" json:"project_id,omitempty"`
	// The Spanner instance ID.
	InstanceId string `protobuf:"bytes,3,opt,name=instance_id,json=instanceId,proto3" json:"instance_id,omitempty"`
	// The Spanner database ID.
	DatabaseId string `protobuf:"bytes,4,opt,name=database_id,json=databaseId,proto3" json:"database_id,omitempty"`
	// The migration table.
	Table string `protobuf:"bytes,5,opt,name=table,proto3" json:"table,omitempty"`
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_jimmy_v1_config_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_jimmy_v1_config_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Config.ProtoReflect.Descriptor instead.
func (*Config) Descriptor() ([]byte, []int) {
	return file_jimmy_v1_config_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *Config) GetProjectId() string {
	if x != nil {
		return x.ProjectId
	}
	return ""
}

func (x *Config) GetInstanceId() string {
	if x != nil {
		return x.InstanceId
	}
	return ""
}

func (x *Config) GetDatabaseId() string {
	if x != nil {
		return x.DatabaseId
	}
	return ""
}

func (x *Config) GetTable() string {
	if x != nil {
		return x.Table
	}
	return ""
}

var File_jimmy_v1_config_proto protoreflect.FileDescriptor

var file_jimmy_v1_config_proto_rawDesc = []byte{
	0x0a, 0x15, 0x6a, 0x69, 0x6d, 0x6d, 0x79, 0x2f, 0x76, 0x31, 0x2f, 0x63, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x6a, 0x69, 0x6d, 0x6d, 0x79, 0x2e, 0x76,
	0x31, 0x1a, 0x1b, 0x62, 0x75, 0x66, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f,
	0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc9,
	0x01, 0x0a, 0x06, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x1a, 0x0a, 0x04, 0x70, 0x61, 0x74,
	0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x06, 0xba, 0x48, 0x03, 0xc8, 0x01, 0x01, 0x52,
	0x04, 0x70, 0x61, 0x74, 0x68, 0x12, 0x1d, 0x0a, 0x0a, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63, 0x74,
	0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70, 0x72, 0x6f, 0x6a, 0x65,
	0x63, 0x74, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65,
	0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x69, 0x6e, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x64, 0x61, 0x74, 0x61, 0x62, 0x61, 0x73,
	0x65, 0x5f, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x64, 0x61, 0x74, 0x61,
	0x62, 0x61, 0x73, 0x65, 0x49, 0x64, 0x12, 0x42, 0x0a, 0x05, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x09, 0x42, 0x2c, 0xba, 0x48, 0x29, 0xc8, 0x01, 0x01, 0x72, 0x24, 0x32,
	0x22, 0x5e, 0x5b, 0x61, 0x2d, 0x7a, 0x41, 0x2d, 0x5a, 0x5d, 0x5b, 0x61, 0x2d, 0x7a, 0x41, 0x2d,
	0x5a, 0x30, 0x2d, 0x39, 0x5f, 0x5d, 0x2a, 0x5b, 0x61, 0x2d, 0x7a, 0x41, 0x2d, 0x5a, 0x30, 0x2d,
	0x39, 0x5d, 0x24, 0x52, 0x05, 0x74, 0x61, 0x62, 0x6c, 0x65, 0x42, 0x91, 0x01, 0x0a, 0x0c, 0x63,
	0x6f, 0x6d, 0x2e, 0x6a, 0x69, 0x6d, 0x6d, 0x79, 0x2e, 0x76, 0x31, 0x42, 0x0b, 0x43, 0x6f, 0x6e,
	0x66, 0x69, 0x67, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x33, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x69, 0x6c, 0x61, 0x73, 0x2f, 0x6a, 0x69, 0x6d,
	0x6d, 0x79, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x6e, 0x61, 0x6c, 0x2f, 0x70, 0x62, 0x2f, 0x6a,
	0x69, 0x6d, 0x6d, 0x79, 0x2f, 0x76, 0x31, 0x3b, 0x6a, 0x69, 0x6d, 0x6d, 0x79, 0x76, 0x31, 0xa2,
	0x02, 0x03, 0x4a, 0x58, 0x58, 0xaa, 0x02, 0x08, 0x4a, 0x69, 0x6d, 0x6d, 0x79, 0x2e, 0x56, 0x31,
	0xca, 0x02, 0x08, 0x4a, 0x69, 0x6d, 0x6d, 0x79, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x14, 0x4a, 0x69,
	0x6d, 0x6d, 0x79, 0x5c, 0x56, 0x31, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0xea, 0x02, 0x09, 0x4a, 0x69, 0x6d, 0x6d, 0x79, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_jimmy_v1_config_proto_rawDescOnce sync.Once
	file_jimmy_v1_config_proto_rawDescData = file_jimmy_v1_config_proto_rawDesc
)

func file_jimmy_v1_config_proto_rawDescGZIP() []byte {
	file_jimmy_v1_config_proto_rawDescOnce.Do(func() {
		file_jimmy_v1_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_jimmy_v1_config_proto_rawDescData)
	})
	return file_jimmy_v1_config_proto_rawDescData
}

var file_jimmy_v1_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_jimmy_v1_config_proto_goTypes = []any{
	(*Config)(nil), // 0: jimmy.v1.Config
}
var file_jimmy_v1_config_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_jimmy_v1_config_proto_init() }
func file_jimmy_v1_config_proto_init() {
	if File_jimmy_v1_config_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_jimmy_v1_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_jimmy_v1_config_proto_goTypes,
		DependencyIndexes: file_jimmy_v1_config_proto_depIdxs,
		MessageInfos:      file_jimmy_v1_config_proto_msgTypes,
	}.Build()
	File_jimmy_v1_config_proto = out.File
	file_jimmy_v1_config_proto_rawDesc = nil
	file_jimmy_v1_config_proto_goTypes = nil
	file_jimmy_v1_config_proto_depIdxs = nil
}
