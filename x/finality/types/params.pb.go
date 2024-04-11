// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: babylon/finality/v1/params.proto

package types

import (
	cosmossdk_io_math "cosmossdk.io/math"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	github_com_cosmos_gogoproto_types "github.com/cosmos/gogoproto/types"
	_ "google.golang.org/protobuf/types/known/durationpb"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Params defines the parameters for the module.
type Params struct {
	// signed_blocks_window defines the size of the sliding window for tracking finality provider liveness
	SignedBlocksWindow int64 `protobuf:"varint,1,opt,name=signed_blocks_window,json=signedBlocksWindow,proto3" json:"signed_blocks_window,omitempty"`
	// min_signed_per_window defines the minimum number of blocks that a finality provider is required to sign
	// within the sliding window to avoid being jailed
	MinSignedPerWindow cosmossdk_io_math.LegacyDec `protobuf:"bytes,2,opt,name=min_signed_per_window,json=minSignedPerWindow,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"min_signed_per_window"`
	// jail_duration defines the duration after which the finality provider can unjail
	JailDuration time.Duration `protobuf:"bytes,3,opt,name=jail_duration,json=jailDuration,proto3,stdduration" json:"jail_duration"`
}

func (m *Params) Reset()      { *m = Params{} }
func (*Params) ProtoMessage() {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_25539c9a61c72ee9, []int{0}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetSignedBlocksWindow() int64 {
	if m != nil {
		return m.SignedBlocksWindow
	}
	return 0
}

func (m *Params) GetJailDuration() time.Duration {
	if m != nil {
		return m.JailDuration
	}
	return 0
}

func init() {
	proto.RegisterType((*Params)(nil), "babylon.finality.v1.Params")
}

func init() { proto.RegisterFile("babylon/finality/v1/params.proto", fileDescriptor_25539c9a61c72ee9) }

var fileDescriptor_25539c9a61c72ee9 = []byte{
	// 370 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x44, 0x91, 0xb1, 0x4e, 0xeb, 0x30,
	0x14, 0x86, 0xe3, 0xdb, 0xab, 0x0e, 0xb9, 0xed, 0x70, 0x43, 0x91, 0xda, 0x22, 0x25, 0x11, 0x53,
	0x85, 0x84, 0xdd, 0x82, 0xc4, 0xc0, 0x18, 0x75, 0x42, 0x20, 0x55, 0x65, 0x40, 0x62, 0x89, 0x9c,
	0xc4, 0x4d, 0x4d, 0x13, 0x3b, 0x4a, 0xd2, 0x96, 0xbc, 0x05, 0x63, 0x47, 0x46, 0x46, 0x06, 0x1e,
	0xa2, 0x63, 0xc5, 0x84, 0x18, 0x0a, 0x6a, 0x07, 0xde, 0x81, 0x09, 0xd5, 0x76, 0xc4, 0x12, 0xe5,
	0xf8, 0xfb, 0xcf, 0x39, 0xff, 0xaf, 0xa3, 0xdb, 0x1e, 0xf6, 0x8a, 0x88, 0x33, 0x34, 0xa2, 0x0c,
	0x47, 0x34, 0x2f, 0xd0, 0xac, 0x87, 0x12, 0x9c, 0xe2, 0x38, 0x83, 0x49, 0xca, 0x73, 0x6e, 0xec,
	0x29, 0x05, 0x2c, 0x15, 0x70, 0xd6, 0x6b, 0x37, 0x42, 0x1e, 0x72, 0xc1, 0xd1, 0xee, 0x4f, 0x4a,
	0xdb, 0xff, 0x71, 0x4c, 0x19, 0x47, 0xe2, 0xab, 0x9e, 0xcc, 0x90, 0xf3, 0x30, 0x22, 0x48, 0x54,
	0xde, 0x74, 0x84, 0x82, 0x69, 0x8a, 0x73, 0xca, 0x99, 0xe2, 0x2d, 0x9f, 0x67, 0x31, 0xcf, 0x5c,
	0x39, 0x4b, 0x16, 0x12, 0x1d, 0x7e, 0x03, 0xbd, 0x3a, 0x10, 0x4e, 0x8c, 0xae, 0xde, 0xc8, 0x68,
	0xc8, 0x48, 0xe0, 0x7a, 0x11, 0xf7, 0x27, 0x99, 0x3b, 0xa7, 0x2c, 0xe0, 0xf3, 0x26, 0xb0, 0x41,
	0xa7, 0x32, 0x34, 0x24, 0x73, 0x04, 0xba, 0x11, 0xc4, 0xa0, 0xfa, 0x7e, 0x4c, 0x99, 0xab, 0xba,
	0x12, 0x92, 0x96, 0x2d, 0x7f, 0x6c, 0xd0, 0xa9, 0x39, 0x67, 0xcb, 0xb5, 0xa5, 0xbd, 0xaf, 0xad,
	0x03, 0xb9, 0x31, 0x0b, 0x26, 0x90, 0x72, 0x14, 0xe3, 0x7c, 0x0c, 0x2f, 0x49, 0x88, 0xfd, 0xa2,
	0x4f, 0xfc, 0xd7, 0x97, 0x63, 0x5d, 0x19, 0xea, 0x13, 0xff, 0xe9, 0xeb, 0xf9, 0x08, 0x0c, 0x8d,
	0x98, 0xb2, 0x6b, 0x31, 0x73, 0x40, 0x52, 0xb5, 0xea, 0x4a, 0xaf, 0xdf, 0x61, 0x1a, 0xb9, 0x65,
	0xb2, 0x66, 0xc5, 0x06, 0x9d, 0x7f, 0x27, 0x2d, 0x28, 0xa3, 0xc3, 0x32, 0x3a, 0xec, 0x2b, 0x81,
	0x53, 0xdf, 0x6d, 0x5f, 0x7c, 0x58, 0x40, 0x0e, 0xad, 0xed, 0xda, 0x4b, 0x78, 0xfe, 0x77, 0xf1,
	0x68, 0x69, 0xce, 0xc5, 0x72, 0x63, 0x82, 0xd5, 0xc6, 0x04, 0x9f, 0x1b, 0x13, 0x3c, 0x6c, 0x4d,
	0x6d, 0xb5, 0x35, 0xb5, 0xb7, 0xad, 0xa9, 0xdd, 0x76, 0x43, 0x9a, 0x8f, 0xa7, 0x1e, 0xf4, 0x79,
	0x8c, 0xd4, 0x69, 0xfc, 0x31, 0xa6, 0xac, 0x2c, 0xd0, 0xfd, 0xef, 0x2d, 0xf3, 0x22, 0x21, 0x99,
	0x57, 0x15, 0x0e, 0x4e, 0x7f, 0x02, 0x00, 0x00, 0xff, 0xff, 0xa7, 0x20, 0x28, 0x14, 0xec, 0x01,
	0x00, 0x00,
}

func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	n1, err1 := github_com_cosmos_gogoproto_types.StdDurationMarshalTo(m.JailDuration, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.JailDuration):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintParams(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0x1a
	{
		size := m.MinSignedPerWindow.Size()
		i -= size
		if _, err := m.MinSignedPerWindow.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if m.SignedBlocksWindow != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.SignedBlocksWindow))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintParams(dAtA []byte, offset int, v uint64) int {
	offset -= sovParams(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.SignedBlocksWindow != 0 {
		n += 1 + sovParams(uint64(m.SignedBlocksWindow))
	}
	l = m.MinSignedPerWindow.Size()
	n += 1 + l + sovParams(uint64(l))
	l = github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.JailDuration)
	n += 1 + l + sovParams(uint64(l))
	return n
}

func sovParams(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozParams(x uint64) (n int) {
	return sovParams(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowParams
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
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SignedBlocksWindow", wireType)
			}
			m.SignedBlocksWindow = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SignedBlocksWindow |= int64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinSignedPerWindow", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.MinSignedPerWindow.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field JailDuration", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdDurationUnmarshal(&m.JailDuration, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipParams(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthParams
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
func skipParams(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowParams
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
					return 0, ErrIntOverflowParams
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
					return 0, ErrIntOverflowParams
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
				return 0, ErrInvalidLengthParams
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupParams
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthParams
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthParams        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowParams          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupParams = fmt.Errorf("proto: unexpected end of group")
)
