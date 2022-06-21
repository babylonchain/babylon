// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: babylon/btccheckpoint/tx.proto

package types

import (
	context "context"
	fmt "fmt"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Consider we have a Merkle tree with following structure:
//            ROOT
//           /    \
//      H1234      H5555
//     /     \       \
//   H12     H34      H55
//  /  \    /  \     /
// H1  H2  H3  H4  H5
// L1  L2  L3  L4  L5
// To prove L3 was part of ROOT we need:
// - btc_transaction_index = 2 which in binary is 010
// (where 0 means going left, 1 means going right in the tree)
// - merkle_nodes we'd have H4 || H12 || H5555
// By looking at 010 we would know that H4 is a right sibling,
// H12 is left, H5555 is right again.
type BTCSpvProof struct {
	// Valid bitcoin transaction containing OP_RETURN opcode.
	BtcTransaction []byte `protobuf:"bytes,1,opt,name=btc_transaction,json=btcTransaction,proto3" json:"btc_transaction,omitempty"`
	// Index of transaction within the block. Index is needed to determine if
	// currently hashed node is left or right.
	BtcTransactionIndex uint32 `protobuf:"varint,2,opt,name=btc_transaction_index,json=btcTransactionIndex,proto3" json:"btc_transaction_index,omitempty"`
	// List of concatenated intermediate merkle tree nodes, without root node and leaf node
	// against which we calculate the proof.
	// Each node has 32 byte length.
	// Example proof can look like: 32_bytes_of_node1 || 32_bytes_of_node2 ||  32_bytes_of_node3
	// so the length of the proof will always be divisible by 32.
	MerkleNodes []byte `protobuf:"bytes,3,opt,name=merkle_nodes,json=merkleNodes,proto3" json:"merkle_nodes,omitempty"`
	// Valid btc header which confirms btc_transaction.
	// Should have exactly 80 bytes
	ConfirmingBtcHeader []byte `protobuf:"bytes,4,opt,name=confirming_btc_header,json=confirmingBtcHeader,proto3" json:"confirming_btc_header,omitempty"`
}

func (m *BTCSpvProof) Reset()         { *m = BTCSpvProof{} }
func (m *BTCSpvProof) String() string { return proto.CompactTextString(m) }
func (*BTCSpvProof) ProtoMessage()    {}
func (*BTCSpvProof) Descriptor() ([]byte, []int) {
	return fileDescriptor_aeec89810b39ea83, []int{0}
}
func (m *BTCSpvProof) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *BTCSpvProof) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_BTCSpvProof.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *BTCSpvProof) XXX_Merge(src proto.Message) {
	xxx_messageInfo_BTCSpvProof.Merge(m, src)
}
func (m *BTCSpvProof) XXX_Size() int {
	return m.Size()
}
func (m *BTCSpvProof) XXX_DiscardUnknown() {
	xxx_messageInfo_BTCSpvProof.DiscardUnknown(m)
}

var xxx_messageInfo_BTCSpvProof proto.InternalMessageInfo

func (m *BTCSpvProof) GetBtcTransaction() []byte {
	if m != nil {
		return m.BtcTransaction
	}
	return nil
}

func (m *BTCSpvProof) GetBtcTransactionIndex() uint32 {
	if m != nil {
		return m.BtcTransactionIndex
	}
	return 0
}

func (m *BTCSpvProof) GetMerkleNodes() []byte {
	if m != nil {
		return m.MerkleNodes
	}
	return nil
}

func (m *BTCSpvProof) GetConfirmingBtcHeader() []byte {
	if m != nil {
		return m.ConfirmingBtcHeader
	}
	return nil
}

type InsertBTCSpvProofRequest struct {
	Proofs []*BTCSpvProof `protobuf:"bytes,1,rep,name=proofs,proto3" json:"proofs,omitempty"`
}

func (m *InsertBTCSpvProofRequest) Reset()         { *m = InsertBTCSpvProofRequest{} }
func (m *InsertBTCSpvProofRequest) String() string { return proto.CompactTextString(m) }
func (*InsertBTCSpvProofRequest) ProtoMessage()    {}
func (*InsertBTCSpvProofRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_aeec89810b39ea83, []int{1}
}
func (m *InsertBTCSpvProofRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *InsertBTCSpvProofRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_InsertBTCSpvProofRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *InsertBTCSpvProofRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InsertBTCSpvProofRequest.Merge(m, src)
}
func (m *InsertBTCSpvProofRequest) XXX_Size() int {
	return m.Size()
}
func (m *InsertBTCSpvProofRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_InsertBTCSpvProofRequest.DiscardUnknown(m)
}

var xxx_messageInfo_InsertBTCSpvProofRequest proto.InternalMessageInfo

func (m *InsertBTCSpvProofRequest) GetProofs() []*BTCSpvProof {
	if m != nil {
		return m.Proofs
	}
	return nil
}

type InsertBTCSpvProofResponse struct {
}

func (m *InsertBTCSpvProofResponse) Reset()         { *m = InsertBTCSpvProofResponse{} }
func (m *InsertBTCSpvProofResponse) String() string { return proto.CompactTextString(m) }
func (*InsertBTCSpvProofResponse) ProtoMessage()    {}
func (*InsertBTCSpvProofResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_aeec89810b39ea83, []int{2}
}
func (m *InsertBTCSpvProofResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *InsertBTCSpvProofResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_InsertBTCSpvProofResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *InsertBTCSpvProofResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InsertBTCSpvProofResponse.Merge(m, src)
}
func (m *InsertBTCSpvProofResponse) XXX_Size() int {
	return m.Size()
}
func (m *InsertBTCSpvProofResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_InsertBTCSpvProofResponse.DiscardUnknown(m)
}

var xxx_messageInfo_InsertBTCSpvProofResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*BTCSpvProof)(nil), "babylonchain.babylon.btccheckpoint.BTCSpvProof")
	proto.RegisterType((*InsertBTCSpvProofRequest)(nil), "babylonchain.babylon.btccheckpoint.InsertBTCSpvProofRequest")
	proto.RegisterType((*InsertBTCSpvProofResponse)(nil), "babylonchain.babylon.btccheckpoint.InsertBTCSpvProofResponse")
}

func init() { proto.RegisterFile("babylon/btccheckpoint/tx.proto", fileDescriptor_aeec89810b39ea83) }

var fileDescriptor_aeec89810b39ea83 = []byte{
	// 341 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x52, 0x3d, 0x4f, 0x02, 0x31,
	0x18, 0xa6, 0x62, 0x18, 0x0a, 0x6a, 0x2c, 0x31, 0x39, 0x35, 0x69, 0xf0, 0x16, 0x99, 0xee, 0x12,
	0x8c, 0x9b, 0x2e, 0x38, 0x28, 0x83, 0x1f, 0x41, 0x26, 0x97, 0xcb, 0xb5, 0x14, 0xae, 0x01, 0xda,
	0xb3, 0x7d, 0x31, 0xf0, 0x2f, 0x18, 0xfd, 0x3b, 0x6e, 0x8e, 0x8c, 0x8e, 0x06, 0xfe, 0x88, 0xb9,
	0x03, 0xc3, 0x87, 0x1a, 0x8d, 0xe3, 0xf3, 0xf1, 0x3e, 0x4f, 0xdf, 0xe6, 0xc5, 0x94, 0x85, 0x6c,
	0xd8, 0xd5, 0xca, 0x67, 0xc0, 0x79, 0x24, 0x78, 0x27, 0xd6, 0x52, 0x81, 0x0f, 0x03, 0x2f, 0x36,
	0x1a, 0x34, 0x71, 0xe7, 0x3a, 0x8f, 0x42, 0xa9, 0xbc, 0x39, 0xf0, 0x56, 0xcc, 0xee, 0x0b, 0xc2,
	0xf9, 0x6a, 0xe3, 0xe2, 0x3e, 0x7e, 0xba, 0x33, 0x5a, 0xb7, 0xc8, 0x31, 0xde, 0x61, 0xc0, 0x03,
	0x30, 0xa1, 0xb2, 0x21, 0x07, 0xa9, 0x95, 0x83, 0x4a, 0xa8, 0x5c, 0xa8, 0x6f, 0x33, 0xe0, 0x8d,
	0x05, 0x4b, 0x2a, 0x78, 0x6f, 0xcd, 0x18, 0x48, 0xd5, 0x14, 0x03, 0x67, 0xa3, 0x84, 0xca, 0x5b,
	0xf5, 0xe2, 0xaa, 0xbd, 0x96, 0x48, 0xe4, 0x08, 0x17, 0x7a, 0xc2, 0x74, 0xba, 0x22, 0x50, 0xba,
	0x29, 0xac, 0x93, 0x4d, 0x93, 0xf3, 0x33, 0xee, 0x26, 0xa1, 0x92, 0x58, 0xae, 0x55, 0x4b, 0x9a,
	0x9e, 0x54, 0xed, 0x20, 0x69, 0x88, 0x44, 0xd8, 0x14, 0xc6, 0xd9, 0x4c, 0xbd, 0xc5, 0x85, 0x58,
	0x05, 0x7e, 0x95, 0x4a, 0x2e, 0xc7, 0x4e, 0x4d, 0x59, 0x61, 0x60, 0x69, 0x91, 0xba, 0x78, 0xec,
	0x0b, 0x0b, 0xe4, 0x12, 0xe7, 0xe2, 0x04, 0x5b, 0x07, 0x95, 0xb2, 0xe5, 0x7c, 0xc5, 0xf7, 0x7e,
	0xff, 0x14, 0x6f, 0x39, 0x67, 0x3e, 0xee, 0x1e, 0xe2, 0xfd, 0x6f, 0x4a, 0x6c, 0xac, 0x95, 0x15,
	0x95, 0x67, 0x84, 0xb3, 0xd7, 0xb6, 0x4d, 0x46, 0x08, 0xef, 0x7e, 0x71, 0x91, 0xb3, 0xbf, 0x74,
	0xfe, 0xb4, 0xc1, 0xc1, 0xf9, 0x3f, 0xa7, 0x67, 0x4f, 0xab, 0xde, 0xbe, 0x4e, 0x28, 0x1a, 0x4f,
	0x28, 0x7a, 0x9f, 0x50, 0x34, 0x9a, 0xd2, 0xcc, 0x78, 0x4a, 0x33, 0x6f, 0x53, 0x9a, 0x79, 0x38,
	0x6d, 0x4b, 0x88, 0xfa, 0xcc, 0xe3, 0xba, 0xe7, 0x2f, 0x57, 0x7c, 0x02, 0x7f, 0xb0, 0x7e, 0x58,
	0xc3, 0x58, 0x58, 0x96, 0x4b, 0x8f, 0xeb, 0xe4, 0x23, 0x00, 0x00, 0xff, 0xff, 0xe4, 0xa6, 0x96,
	0x82, 0x7e, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	InsertBTCSpvProof(ctx context.Context, in *InsertBTCSpvProofRequest, opts ...grpc.CallOption) (*InsertBTCSpvProofResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) InsertBTCSpvProof(ctx context.Context, in *InsertBTCSpvProofRequest, opts ...grpc.CallOption) (*InsertBTCSpvProofResponse, error) {
	out := new(InsertBTCSpvProofResponse)
	err := c.cc.Invoke(ctx, "/babylonchain.babylon.btccheckpoint.Msg/InsertBTCSpvProof", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	InsertBTCSpvProof(context.Context, *InsertBTCSpvProofRequest) (*InsertBTCSpvProofResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) InsertBTCSpvProof(ctx context.Context, req *InsertBTCSpvProofRequest) (*InsertBTCSpvProofResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InsertBTCSpvProof not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_InsertBTCSpvProof_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InsertBTCSpvProofRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).InsertBTCSpvProof(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/babylonchain.babylon.btccheckpoint.Msg/InsertBTCSpvProof",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).InsertBTCSpvProof(ctx, req.(*InsertBTCSpvProofRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "babylonchain.babylon.btccheckpoint.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "InsertBTCSpvProof",
			Handler:    _Msg_InsertBTCSpvProof_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "babylon/btccheckpoint/tx.proto",
}

func (m *BTCSpvProof) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *BTCSpvProof) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *BTCSpvProof) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ConfirmingBtcHeader) > 0 {
		i -= len(m.ConfirmingBtcHeader)
		copy(dAtA[i:], m.ConfirmingBtcHeader)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ConfirmingBtcHeader)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.MerkleNodes) > 0 {
		i -= len(m.MerkleNodes)
		copy(dAtA[i:], m.MerkleNodes)
		i = encodeVarintTx(dAtA, i, uint64(len(m.MerkleNodes)))
		i--
		dAtA[i] = 0x1a
	}
	if m.BtcTransactionIndex != 0 {
		i = encodeVarintTx(dAtA, i, uint64(m.BtcTransactionIndex))
		i--
		dAtA[i] = 0x10
	}
	if len(m.BtcTransaction) > 0 {
		i -= len(m.BtcTransaction)
		copy(dAtA[i:], m.BtcTransaction)
		i = encodeVarintTx(dAtA, i, uint64(len(m.BtcTransaction)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *InsertBTCSpvProofRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *InsertBTCSpvProofRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *InsertBTCSpvProofRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Proofs) > 0 {
		for iNdEx := len(m.Proofs) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Proofs[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTx(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *InsertBTCSpvProofResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *InsertBTCSpvProofResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *InsertBTCSpvProofResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *BTCSpvProof) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.BtcTransaction)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if m.BtcTransactionIndex != 0 {
		n += 1 + sovTx(uint64(m.BtcTransactionIndex))
	}
	l = len(m.MerkleNodes)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ConfirmingBtcHeader)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	return n
}

func (m *InsertBTCSpvProofRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Proofs) > 0 {
		for _, e := range m.Proofs {
			l = e.Size()
			n += 1 + l + sovTx(uint64(l))
		}
	}
	return n
}

func (m *InsertBTCSpvProofResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *BTCSpvProof) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: BTCSpvProof: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: BTCSpvProof: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field BtcTransaction", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.BtcTransaction = append(m.BtcTransaction[:0], dAtA[iNdEx:postIndex]...)
			if m.BtcTransaction == nil {
				m.BtcTransaction = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BtcTransactionIndex", wireType)
			}
			m.BtcTransactionIndex = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BtcTransactionIndex |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MerkleNodes", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MerkleNodes = append(m.MerkleNodes[:0], dAtA[iNdEx:postIndex]...)
			if m.MerkleNodes == nil {
				m.MerkleNodes = []byte{}
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ConfirmingBtcHeader", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ConfirmingBtcHeader = append(m.ConfirmingBtcHeader[:0], dAtA[iNdEx:postIndex]...)
			if m.ConfirmingBtcHeader == nil {
				m.ConfirmingBtcHeader = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *InsertBTCSpvProofRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: InsertBTCSpvProofRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: InsertBTCSpvProofRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Proofs", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Proofs = append(m.Proofs, &BTCSpvProof{})
			if err := m.Proofs[len(m.Proofs)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *InsertBTCSpvProofResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: InsertBTCSpvProofResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: InsertBTCSpvProofResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTx
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTx
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
