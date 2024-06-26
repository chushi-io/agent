// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.1
// 	protoc        (unknown)
// source: agent/v1/logs.proto

package agentv1

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

type StreamLogsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RunId   string `protobuf:"bytes,1,opt,name=run_id,json=runId,proto3" json:"run_id,omitempty"`
	Content string `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *StreamLogsRequest) Reset() {
	*x = StreamLogsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_agent_v1_logs_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StreamLogsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamLogsRequest) ProtoMessage() {}

func (x *StreamLogsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_agent_v1_logs_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamLogsRequest.ProtoReflect.Descriptor instead.
func (*StreamLogsRequest) Descriptor() ([]byte, []int) {
	return file_agent_v1_logs_proto_rawDescGZIP(), []int{0}
}

func (x *StreamLogsRequest) GetRunId() string {
	if x != nil {
		return x.RunId
	}
	return ""
}

func (x *StreamLogsRequest) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

type StreamLogsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *StreamLogsResponse) Reset() {
	*x = StreamLogsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_agent_v1_logs_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StreamLogsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StreamLogsResponse) ProtoMessage() {}

func (x *StreamLogsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_agent_v1_logs_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StreamLogsResponse.ProtoReflect.Descriptor instead.
func (*StreamLogsResponse) Descriptor() ([]byte, []int) {
	return file_agent_v1_logs_proto_rawDescGZIP(), []int{1}
}

func (x *StreamLogsResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type UploadLogsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RunId   string `protobuf:"bytes,1,opt,name=run_id,json=runId,proto3" json:"run_id,omitempty"`
	Content string `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *UploadLogsRequest) Reset() {
	*x = UploadLogsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_agent_v1_logs_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadLogsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadLogsRequest) ProtoMessage() {}

func (x *UploadLogsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_agent_v1_logs_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadLogsRequest.ProtoReflect.Descriptor instead.
func (*UploadLogsRequest) Descriptor() ([]byte, []int) {
	return file_agent_v1_logs_proto_rawDescGZIP(), []int{2}
}

func (x *UploadLogsRequest) GetRunId() string {
	if x != nil {
		return x.RunId
	}
	return ""
}

func (x *UploadLogsRequest) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

type UploadLogsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *UploadLogsResponse) Reset() {
	*x = UploadLogsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_agent_v1_logs_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadLogsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadLogsResponse) ProtoMessage() {}

func (x *UploadLogsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_agent_v1_logs_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadLogsResponse.ProtoReflect.Descriptor instead.
func (*UploadLogsResponse) Descriptor() ([]byte, []int) {
	return file_agent_v1_logs_proto_rawDescGZIP(), []int{3}
}

func (x *UploadLogsResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

var File_agent_v1_logs_proto protoreflect.FileDescriptor

var file_agent_v1_logs_proto_rawDesc = []byte{
	0x0a, 0x13, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2f, 0x76, 0x31, 0x2f, 0x6c, 0x6f, 0x67, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x76, 0x31, 0x22,
	0x44, 0x0a, 0x11, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x4c, 0x6f, 0x67, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x15, 0x0a, 0x06, 0x72, 0x75, 0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x72, 0x75, 0x6e, 0x49, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6f,
	0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0x2e, 0x0a, 0x12, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x4c,
	0x6f, 0x67, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73,
	0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x22, 0x44, 0x0a, 0x11, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x4c,
	0x6f, 0x67, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x15, 0x0a, 0x06, 0x72, 0x75,
	0x6e, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x72, 0x75, 0x6e, 0x49,
	0x64, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0x2e, 0x0a, 0x12, 0x55,
	0x70, 0x6c, 0x6f, 0x61, 0x64, 0x4c, 0x6f, 0x67, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x32, 0x9c, 0x01, 0x0a, 0x04,
	0x4c, 0x6f, 0x67, 0x73, 0x12, 0x49, 0x0a, 0x0a, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x4c, 0x6f,
	0x67, 0x73, 0x12, 0x1b, 0x2e, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74,
	0x72, 0x65, 0x61, 0x6d, 0x4c, 0x6f, 0x67, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x1c, 0x2e, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x74, 0x72, 0x65, 0x61,
	0x6d, 0x4c, 0x6f, 0x67, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12,
	0x49, 0x0a, 0x0a, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x4c, 0x6f, 0x67, 0x73, 0x12, 0x1b, 0x2e,
	0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x4c,
	0x6f, 0x67, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x61, 0x67, 0x65,
	0x6e, 0x74, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x4c, 0x6f, 0x67, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x8b, 0x01, 0x0a, 0x0c, 0x63,
	0x6f, 0x6d, 0x2e, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x76, 0x31, 0x42, 0x09, 0x4c, 0x6f, 0x67,
	0x73, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x63, 0x68, 0x75, 0x73, 0x68, 0x69, 0x2d, 0x69, 0x6f, 0x2f, 0x61,
	0x67, 0x65, 0x6e, 0x74, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2f, 0x76,
	0x31, 0x3b, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x41, 0x58, 0x58, 0xaa,
	0x02, 0x08, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x08, 0x41, 0x67, 0x65,
	0x6e, 0x74, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x14, 0x41, 0x67, 0x65, 0x6e, 0x74, 0x5c, 0x56, 0x31,
	0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x09, 0x41,
	0x67, 0x65, 0x6e, 0x74, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_agent_v1_logs_proto_rawDescOnce sync.Once
	file_agent_v1_logs_proto_rawDescData = file_agent_v1_logs_proto_rawDesc
)

func file_agent_v1_logs_proto_rawDescGZIP() []byte {
	file_agent_v1_logs_proto_rawDescOnce.Do(func() {
		file_agent_v1_logs_proto_rawDescData = protoimpl.X.CompressGZIP(file_agent_v1_logs_proto_rawDescData)
	})
	return file_agent_v1_logs_proto_rawDescData
}

var file_agent_v1_logs_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_agent_v1_logs_proto_goTypes = []interface{}{
	(*StreamLogsRequest)(nil),  // 0: agent.v1.StreamLogsRequest
	(*StreamLogsResponse)(nil), // 1: agent.v1.StreamLogsResponse
	(*UploadLogsRequest)(nil),  // 2: agent.v1.UploadLogsRequest
	(*UploadLogsResponse)(nil), // 3: agent.v1.UploadLogsResponse
}
var file_agent_v1_logs_proto_depIdxs = []int32{
	0, // 0: agent.v1.Logs.StreamLogs:input_type -> agent.v1.StreamLogsRequest
	2, // 1: agent.v1.Logs.UploadLogs:input_type -> agent.v1.UploadLogsRequest
	1, // 2: agent.v1.Logs.StreamLogs:output_type -> agent.v1.StreamLogsResponse
	3, // 3: agent.v1.Logs.UploadLogs:output_type -> agent.v1.UploadLogsResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_agent_v1_logs_proto_init() }
func file_agent_v1_logs_proto_init() {
	if File_agent_v1_logs_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_agent_v1_logs_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StreamLogsRequest); i {
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
		file_agent_v1_logs_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StreamLogsResponse); i {
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
		file_agent_v1_logs_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadLogsRequest); i {
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
		file_agent_v1_logs_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadLogsResponse); i {
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
			RawDescriptor: file_agent_v1_logs_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_agent_v1_logs_proto_goTypes,
		DependencyIndexes: file_agent_v1_logs_proto_depIdxs,
		MessageInfos:      file_agent_v1_logs_proto_msgTypes,
	}.Build()
	File_agent_v1_logs_proto = out.File
	file_agent_v1_logs_proto_rawDesc = nil
	file_agent_v1_logs_proto_goTypes = nil
	file_agent_v1_logs_proto_depIdxs = nil
}
