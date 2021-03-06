// Code generated by protoc-gen-go. DO NOT EDIT.
// source: infra/engine/proto/engine.proto

package proto

import (
	fmt "fmt"
	math "math"

	proto1 "github.com/beneath-hq/beneath/server/data/grpc/proto"
	proto "github.com/golang/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type WriteRequest struct {
	WriteId              []byte                    `protobuf:"bytes,1,opt,name=write_id,json=writeId,proto3" json:"write_id,omitempty"`
	InstanceRecords      []*proto1.InstanceRecords `protobuf:"bytes,2,rep,name=instance_records,json=instanceRecords,proto3" json:"instance_records,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}

func (m *WriteRequest) Reset()         { *m = WriteRequest{} }
func (m *WriteRequest) String() string { return proto.CompactTextString(m) }
func (*WriteRequest) ProtoMessage()    {}
func (*WriteRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_0e27bf340d39c716, []int{0}
}

func (m *WriteRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WriteRequest.Unmarshal(m, b)
}
func (m *WriteRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WriteRequest.Marshal(b, m, deterministic)
}
func (m *WriteRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WriteRequest.Merge(m, src)
}
func (m *WriteRequest) XXX_Size() int {
	return xxx_messageInfo_WriteRequest.Size(m)
}
func (m *WriteRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_WriteRequest.DiscardUnknown(m)
}

var xxx_messageInfo_WriteRequest proto.InternalMessageInfo

func (m *WriteRequest) GetWriteId() []byte {
	if m != nil {
		return m.WriteId
	}
	return nil
}

func (m *WriteRequest) GetInstanceRecords() []*proto1.InstanceRecords {
	if m != nil {
		return m.InstanceRecords
	}
	return nil
}

type WriteReport struct {
	WriteId              []byte   `protobuf:"bytes,1,opt,name=write_id,json=writeId,proto3" json:"write_id,omitempty"`
	InstanceId           []byte   `protobuf:"bytes,2,opt,name=instance_id,json=instanceId,proto3" json:"instance_id,omitempty"`
	RecordsCount         int32    `protobuf:"varint,3,opt,name=records_count,json=recordsCount,proto3" json:"records_count,omitempty"`
	BytesTotal           int32    `protobuf:"varint,4,opt,name=bytes_total,json=bytesTotal,proto3" json:"bytes_total,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *WriteReport) Reset()         { *m = WriteReport{} }
func (m *WriteReport) String() string { return proto.CompactTextString(m) }
func (*WriteReport) ProtoMessage()    {}
func (*WriteReport) Descriptor() ([]byte, []int) {
	return fileDescriptor_0e27bf340d39c716, []int{1}
}

func (m *WriteReport) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WriteReport.Unmarshal(m, b)
}
func (m *WriteReport) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WriteReport.Marshal(b, m, deterministic)
}
func (m *WriteReport) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WriteReport.Merge(m, src)
}
func (m *WriteReport) XXX_Size() int {
	return xxx_messageInfo_WriteReport.Size(m)
}
func (m *WriteReport) XXX_DiscardUnknown() {
	xxx_messageInfo_WriteReport.DiscardUnknown(m)
}

var xxx_messageInfo_WriteReport proto.InternalMessageInfo

func (m *WriteReport) GetWriteId() []byte {
	if m != nil {
		return m.WriteId
	}
	return nil
}

func (m *WriteReport) GetInstanceId() []byte {
	if m != nil {
		return m.InstanceId
	}
	return nil
}

func (m *WriteReport) GetRecordsCount() int32 {
	if m != nil {
		return m.RecordsCount
	}
	return 0
}

func (m *WriteReport) GetBytesTotal() int32 {
	if m != nil {
		return m.BytesTotal
	}
	return 0
}

func init() {
	proto.RegisterType((*WriteRequest)(nil), "engine.WriteRequest")
	proto.RegisterType((*WriteReport)(nil), "engine.WriteReport")
}

func init() {
	proto.RegisterFile("infra/engine/proto/engine.proto", fileDescriptor_0e27bf340d39c716)
}

var fileDescriptor_0e27bf340d39c716 = []byte{
	// 271 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x90, 0x41, 0x4b, 0xc3, 0x30,
	0x14, 0xc7, 0xe9, 0xa6, 0x53, 0xd2, 0x8a, 0xd2, 0x53, 0xd5, 0x83, 0x65, 0x7a, 0xe8, 0xc5, 0x06,
	0xf5, 0x24, 0xde, 0x14, 0x84, 0x5e, 0x8b, 0x20, 0x78, 0x29, 0x69, 0xf2, 0xec, 0x02, 0x33, 0x69,
	0x93, 0xd7, 0x8d, 0x7d, 0x0a, 0xbf, 0xb2, 0x34, 0x4d, 0x85, 0x5d, 0x76, 0x4a, 0xfe, 0xbf, 0xfc,
	0x78, 0xff, 0xf0, 0x48, 0x26, 0xd5, 0xb7, 0x61, 0x16, 0x4d, 0xcf, 0xb1, 0x37, 0x40, 0x41, 0x35,
	0x52, 0x01, 0x6d, 0x8d, 0x46, 0xed, 0x43, 0xee, 0x42, 0xbc, 0x18, 0xd3, 0xd5, 0x9d, 0x05, 0xb3,
	0x01, 0x43, 0x05, 0x43, 0x46, 0x1b, 0xd3, 0x72, 0x2f, 0x37, 0x0c, 0x61, 0xcb, 0x76, 0xa3, 0xbd,
	0xec, 0x48, 0xf4, 0x69, 0x24, 0x42, 0x09, 0x5d, 0x0f, 0x16, 0xe3, 0x4b, 0x72, 0xba, 0x1d, 0x72,
	0x25, 0x45, 0x12, 0xa4, 0x41, 0x16, 0x95, 0x27, 0x2e, 0x17, 0x22, 0x7e, 0x27, 0x17, 0x52, 0x59,
	0x64, 0x8a, 0x43, 0x65, 0x80, 0x6b, 0x23, 0x6c, 0x32, 0x4b, 0xe7, 0x59, 0xf8, 0x78, 0x9d, 0x4f,
	0x43, 0x37, 0x0f, 0x79, 0xe1, 0x9d, 0x72, 0x54, 0xca, 0x73, 0xb9, 0x0f, 0x96, 0xbf, 0x01, 0x09,
	0x7d, 0x67, 0xab, 0xcd, 0xc1, 0xca, 0x1b, 0x12, 0xfe, 0x57, 0x4a, 0x91, 0xcc, 0xdc, 0x2b, 0x99,
	0x50, 0x21, 0xe2, 0x5b, 0x72, 0xe6, 0xbf, 0x52, 0x71, 0xdd, 0x2b, 0x4c, 0xe6, 0x69, 0x90, 0x1d,
	0x97, 0x91, 0x87, 0x6f, 0x03, 0x1b, 0xa6, 0xd4, 0x3b, 0x04, 0x5b, 0xa1, 0x46, 0xb6, 0x4e, 0x8e,
	0x9c, 0x42, 0x1c, 0xfa, 0x18, 0xc8, 0xeb, 0xcb, 0xd7, 0x73, 0x23, 0x71, 0xcd, 0xea, 0x9c, 0xeb,
	0x1f, 0x5a, 0x83, 0x02, 0x86, 0xab, 0xfb, 0x55, 0x37, 0x5d, 0xe9, 0x81, 0xe5, 0xd7, 0x0b, 0x77,
	0x3c, 0xfd, 0x05, 0x00, 0x00, 0xff, 0xff, 0x0e, 0x9f, 0xbf, 0xd4, 0xa2, 0x01, 0x00, 0x00,
}
