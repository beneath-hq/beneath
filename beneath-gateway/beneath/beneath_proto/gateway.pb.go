// Code generated by protoc-gen-go. DO NOT EDIT.
// source: gateway.proto

package beneath_proto

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	math "math"
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

type WriteEncodedRecordsResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *WriteEncodedRecordsResponse) Reset()         { *m = WriteEncodedRecordsResponse{} }
func (m *WriteEncodedRecordsResponse) String() string { return proto.CompactTextString(m) }
func (*WriteEncodedRecordsResponse) ProtoMessage()    {}
func (*WriteEncodedRecordsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_f1a937782ebbded5, []int{0}
}

func (m *WriteEncodedRecordsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_WriteEncodedRecordsResponse.Unmarshal(m, b)
}
func (m *WriteEncodedRecordsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_WriteEncodedRecordsResponse.Marshal(b, m, deterministic)
}
func (m *WriteEncodedRecordsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_WriteEncodedRecordsResponse.Merge(m, src)
}
func (m *WriteEncodedRecordsResponse) XXX_Size() int {
	return xxx_messageInfo_WriteEncodedRecordsResponse.Size(m)
}
func (m *WriteEncodedRecordsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_WriteEncodedRecordsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_WriteEncodedRecordsResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*WriteEncodedRecordsResponse)(nil), "beneath_proto.WriteEncodedRecordsResponse")
}

func init() { proto.RegisterFile("gateway.proto", fileDescriptor_f1a937782ebbded5) }

var fileDescriptor_f1a937782ebbded5 = []byte{
	// 159 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4d, 0x4f, 0x2c, 0x49,
	0x2d, 0x4f, 0xac, 0xd4, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x4d, 0x4a, 0xcd, 0x4b, 0x4d,
	0x2c, 0xc9, 0x88, 0x07, 0x73, 0xa5, 0xf8, 0x0a, 0x72, 0x12, 0x4b, 0xd2, 0xf2, 0x8b, 0x72, 0x21,
	0xd2, 0x4a, 0xb2, 0x5c, 0xd2, 0xe1, 0x45, 0x99, 0x25, 0xa9, 0xae, 0x79, 0xc9, 0xf9, 0x29, 0xa9,
	0x29, 0x41, 0xa9, 0xc9, 0xf9, 0x45, 0x29, 0xc5, 0x41, 0xa9, 0xc5, 0x05, 0xf9, 0x79, 0xc5, 0xa9,
	0x46, 0x95, 0x5c, 0xec, 0xee, 0x10, 0xe3, 0x84, 0xf2, 0xb8, 0x84, 0xb1, 0xa8, 0x14, 0xd2, 0xd4,
	0x43, 0xb1, 0x40, 0x0f, 0xab, 0x69, 0x85, 0xa5, 0xa9, 0xc5, 0x25, 0x52, 0x5a, 0xc4, 0x28, 0x85,
	0x58, 0xac, 0xc4, 0xe0, 0xa4, 0xce, 0x25, 0x9a, 0x97, 0x5a, 0x52, 0x9e, 0x5f, 0x94, 0x0d, 0xd3,
	0x06, 0x71, 0xb2, 0x13, 0x8f, 0x13, 0x84, 0x1b, 0x00, 0xe2, 0x05, 0x30, 0x26, 0xb1, 0x81, 0x85,
	0x8d, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0x76, 0xec, 0xa4, 0x4e, 0xf9, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// GatewayClient is the client API for Gateway service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type GatewayClient interface {
	WriteEncodedRecords(ctx context.Context, in *WriteEncodedRecordsRequest, opts ...grpc.CallOption) (*WriteEncodedRecordsResponse, error)
}

type gatewayClient struct {
	cc *grpc.ClientConn
}

func NewGatewayClient(cc *grpc.ClientConn) GatewayClient {
	return &gatewayClient{cc}
}

func (c *gatewayClient) WriteEncodedRecords(ctx context.Context, in *WriteEncodedRecordsRequest, opts ...grpc.CallOption) (*WriteEncodedRecordsResponse, error) {
	out := new(WriteEncodedRecordsResponse)
	err := c.cc.Invoke(ctx, "/beneath_proto.Gateway/WriteEncodedRecords", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GatewayServer is the server API for Gateway service.
type GatewayServer interface {
	WriteEncodedRecords(context.Context, *WriteEncodedRecordsRequest) (*WriteEncodedRecordsResponse, error)
}

func RegisterGatewayServer(s *grpc.Server, srv GatewayServer) {
	s.RegisterService(&_Gateway_serviceDesc, srv)
}

func _Gateway_WriteEncodedRecords_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WriteEncodedRecordsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GatewayServer).WriteEncodedRecords(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/beneath_proto.Gateway/WriteEncodedRecords",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GatewayServer).WriteEncodedRecords(ctx, req.(*WriteEncodedRecordsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Gateway_serviceDesc = grpc.ServiceDesc{
	ServiceName: "beneath_proto.Gateway",
	HandlerType: (*GatewayServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "WriteEncodedRecords",
			Handler:    _Gateway_WriteEncodedRecords_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "gateway.proto",
}