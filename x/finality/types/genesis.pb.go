// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: babylon/finality/v1/genesis.proto

package types

import (
	fmt "fmt"
	github_com_babylonchain_babylon_types "github.com/babylonchain/babylon/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
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

// GenesisState defines the finality module's genesis state.
type GenesisState struct {
	// params the current params of the state.
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	// indexed_blocks all the btc blocks and if their status are finalized.
	IndexedBlocks []*IndexedBlock `protobuf:"bytes,2,rep,name=indexed_blocks,json=indexedBlocks,proto3" json:"indexed_blocks,omitempty"`
	// evidences all the evidences ever registered.
	Evidences []*Evidence `protobuf:"bytes,3,rep,name=evidences,proto3" json:"evidences,omitempty"`
	// votes_sigs contains all the votes of finality providers ever registered.
	VoteSigs []*VoteSig `protobuf:"bytes,4,rep,name=vote_sigs,json=voteSigs,proto3" json:"vote_sigs,omitempty"`
	// pub_rand_commit contains all the public randomness commitment ever commited from the finality providers.
	PubRandCommit []*PubRandCommitWithPK `protobuf:"bytes,5,rep,name=pub_rand_commit,json=pubRandCommit,proto3" json:"pub_rand_commit,omitempty"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_52dc577f74d797d1, []int{0}
}
func (m *GenesisState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisState.Merge(m, src)
}
func (m *GenesisState) XXX_Size() int {
	return m.Size()
}
func (m *GenesisState) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisState.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisState proto.InternalMessageInfo

func (m *GenesisState) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func (m *GenesisState) GetIndexedBlocks() []*IndexedBlock {
	if m != nil {
		return m.IndexedBlocks
	}
	return nil
}

func (m *GenesisState) GetEvidences() []*Evidence {
	if m != nil {
		return m.Evidences
	}
	return nil
}

func (m *GenesisState) GetVoteSigs() []*VoteSig {
	if m != nil {
		return m.VoteSigs
	}
	return nil
}

func (m *GenesisState) GetPubRandCommit() []*PubRandCommitWithPK {
	if m != nil {
		return m.PubRandCommit
	}
	return nil
}

// VoteSig the vote of an finality provider
// with the block of the vote, the finality provider btc public key and the vote signature.
type VoteSig struct {
	// block_height is the height of the voted block.
	BlockHeight uint64 `protobuf:"varint,1,opt,name=block_height,json=blockHeight,proto3" json:"block_height,omitempty"`
	// fp_btc_pk is the BTC PK of the finality provider that casts this vote
	FpBtcPk *github_com_babylonchain_babylon_types.BIP340PubKey `protobuf:"bytes,2,opt,name=fp_btc_pk,json=fpBtcPk,proto3,customtype=github.com/babylonchain/babylon/types.BIP340PubKey" json:"fp_btc_pk,omitempty"`
	// finality_sig is the finality signature to this block
	// where finality signature is an EOTS signature, i.e.
	FinalitySig *github_com_babylonchain_babylon_types.SchnorrEOTSSig `protobuf:"bytes,3,opt,name=finality_sig,json=finalitySig,proto3,customtype=github.com/babylonchain/babylon/types.SchnorrEOTSSig" json:"finality_sig,omitempty"`
}

func (m *VoteSig) Reset()         { *m = VoteSig{} }
func (m *VoteSig) String() string { return proto.CompactTextString(m) }
func (*VoteSig) ProtoMessage()    {}
func (*VoteSig) Descriptor() ([]byte, []int) {
	return fileDescriptor_52dc577f74d797d1, []int{1}
}
func (m *VoteSig) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *VoteSig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_VoteSig.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *VoteSig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VoteSig.Merge(m, src)
}
func (m *VoteSig) XXX_Size() int {
	return m.Size()
}
func (m *VoteSig) XXX_DiscardUnknown() {
	xxx_messageInfo_VoteSig.DiscardUnknown(m)
}

var xxx_messageInfo_VoteSig proto.InternalMessageInfo

func (m *VoteSig) GetBlockHeight() uint64 {
	if m != nil {
		return m.BlockHeight
	}
	return 0
}

// PubRandCommitWithPK is the public randomness commitment with the finality provider's BTC public key
type PubRandCommitWithPK struct {
	// fp_btc_pk is the BTC PK of the finality provider that commits the public randomness
	FpBtcPk *github_com_babylonchain_babylon_types.BIP340PubKey `protobuf:"bytes,1,opt,name=fp_btc_pk,json=fpBtcPk,proto3,customtype=github.com/babylonchain/babylon/types.BIP340PubKey" json:"fp_btc_pk,omitempty"`
	// pub_rand_commit is the public randomness commitment
	PubRandCommit *PubRandCommit `protobuf:"bytes,2,opt,name=pub_rand_commit,json=pubRandCommit,proto3" json:"pub_rand_commit,omitempty"`
}

func (m *PubRandCommitWithPK) Reset()         { *m = PubRandCommitWithPK{} }
func (m *PubRandCommitWithPK) String() string { return proto.CompactTextString(m) }
func (*PubRandCommitWithPK) ProtoMessage()    {}
func (*PubRandCommitWithPK) Descriptor() ([]byte, []int) {
	return fileDescriptor_52dc577f74d797d1, []int{2}
}
func (m *PubRandCommitWithPK) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *PubRandCommitWithPK) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_PubRandCommitWithPK.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *PubRandCommitWithPK) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PubRandCommitWithPK.Merge(m, src)
}
func (m *PubRandCommitWithPK) XXX_Size() int {
	return m.Size()
}
func (m *PubRandCommitWithPK) XXX_DiscardUnknown() {
	xxx_messageInfo_PubRandCommitWithPK.DiscardUnknown(m)
}

var xxx_messageInfo_PubRandCommitWithPK proto.InternalMessageInfo

func (m *PubRandCommitWithPK) GetPubRandCommit() *PubRandCommit {
	if m != nil {
		return m.PubRandCommit
	}
	return nil
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "babylon.finality.v1.GenesisState")
	proto.RegisterType((*VoteSig)(nil), "babylon.finality.v1.VoteSig")
	proto.RegisterType((*PubRandCommitWithPK)(nil), "babylon.finality.v1.PubRandCommitWithPK")
}

func init() { proto.RegisterFile("babylon/finality/v1/genesis.proto", fileDescriptor_52dc577f74d797d1) }

var fileDescriptor_52dc577f74d797d1 = []byte{
	// 490 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x93, 0x4f, 0x6f, 0xd3, 0x30,
	0x18, 0xc6, 0x9b, 0xae, 0x6c, 0xd4, 0xed, 0x40, 0xf2, 0x38, 0x44, 0x05, 0xd2, 0x36, 0xa7, 0x9e,
	0x92, 0xad, 0x9b, 0x10, 0x13, 0xb7, 0xa0, 0x89, 0xfd, 0x39, 0x10, 0x39, 0x08, 0x24, 0x38, 0x44,
	0x71, 0xe2, 0x26, 0x56, 0x5b, 0x3b, 0x8a, 0xdd, 0x68, 0xfd, 0x16, 0x7c, 0x19, 0xbe, 0xc3, 0x8e,
	0x3b, 0xa2, 0x49, 0xab, 0x50, 0xfb, 0x45, 0x50, 0x9d, 0x94, 0x0d, 0x88, 0x34, 0x0e, 0xdc, 0xec,
	0x37, 0xcf, 0xfb, 0xcb, 0xe3, 0xe7, 0xb5, 0x41, 0x1f, 0x07, 0x78, 0x3e, 0xe1, 0xcc, 0x1e, 0x51,
	0x16, 0x4c, 0xa8, 0x9c, 0xdb, 0xf9, 0x81, 0x1d, 0x13, 0x46, 0x04, 0x15, 0x56, 0x9a, 0x71, 0xc9,
	0xe1, 0x5e, 0x29, 0xb1, 0x36, 0x12, 0x2b, 0x3f, 0xe8, 0x3c, 0x8b, 0x79, 0xcc, 0xd5, 0x77, 0x7b,
	0xbd, 0x2a, 0xa4, 0x9d, 0x5e, 0x15, 0x2d, 0x0d, 0xb2, 0x60, 0x5a, 0xc2, 0x3a, 0x66, 0x95, 0xe2,
	0x17, 0x58, 0x69, 0xcc, 0xdb, 0x3a, 0x68, 0xbf, 0x2b, 0x2c, 0x78, 0x32, 0x90, 0x04, 0x1e, 0x83,
	0xed, 0x02, 0xa2, 0x6b, 0x3d, 0x6d, 0xd0, 0x1a, 0x3e, 0xb7, 0x2a, 0x2c, 0x59, 0xae, 0x92, 0x38,
	0x8d, 0xab, 0x45, 0xb7, 0x86, 0xca, 0x06, 0x78, 0x0a, 0x9e, 0x50, 0x16, 0x91, 0x4b, 0x12, 0xf9,
	0x78, 0xc2, 0xc3, 0xb1, 0xd0, 0xeb, 0xbd, 0xad, 0x41, 0x6b, 0xd8, 0xaf, 0x44, 0x9c, 0x15, 0x52,
	0x67, 0xad, 0x44, 0xbb, 0xf4, 0xde, 0x4e, 0xc0, 0x37, 0xa0, 0x49, 0x72, 0x1a, 0x11, 0x16, 0x12,
	0xa1, 0x6f, 0x29, 0xc8, 0xcb, 0x4a, 0xc8, 0x49, 0xa9, 0x42, 0x77, 0x7a, 0x78, 0x0c, 0x9a, 0x39,
	0x97, 0xc4, 0x17, 0x34, 0x16, 0x7a, 0x43, 0x35, 0xbf, 0xa8, 0x6c, 0xfe, 0xc8, 0x25, 0xf1, 0x68,
	0x8c, 0x1e, 0xe7, 0xc5, 0x42, 0x40, 0x17, 0x3c, 0x4d, 0x67, 0xd8, 0xcf, 0x02, 0x16, 0xf9, 0x21,
	0x9f, 0x4e, 0xa9, 0xd4, 0x1f, 0x29, 0xc0, 0xa0, 0x3a, 0x85, 0x19, 0x46, 0x01, 0x8b, 0xde, 0x2a,
	0xe5, 0x27, 0x2a, 0x13, 0xf7, 0x02, 0xed, 0xa6, 0xf7, 0x8b, 0xe6, 0xad, 0x06, 0x76, 0xca, 0xff,
	0xc0, 0x3e, 0x68, 0xab, 0x5c, 0xfc, 0x84, 0xd0, 0x38, 0x91, 0x2a, 0xe0, 0x06, 0x6a, 0xa9, 0xda,
	0xa9, 0x2a, 0x41, 0x04, 0x9a, 0xa3, 0xd4, 0xc7, 0x32, 0xf4, 0xd3, 0xb1, 0x5e, 0xef, 0x69, 0x83,
	0xb6, 0xf3, 0xea, 0x66, 0xd1, 0x1d, 0xc6, 0x54, 0x26, 0x33, 0x6c, 0x85, 0x7c, 0x6a, 0x97, 0x46,
	0xc2, 0x24, 0xa0, 0x6c, 0xb3, 0xb1, 0xe5, 0x3c, 0x25, 0xc2, 0x72, 0xce, 0xdc, 0xc3, 0xa3, 0x7d,
	0x77, 0x86, 0x2f, 0xc8, 0x1c, 0xed, 0x8c, 0x52, 0x47, 0x86, 0xee, 0x18, 0x7e, 0x01, 0xed, 0x8d,
	0xe9, 0x75, 0x26, 0xfa, 0x96, 0xc2, 0xbe, 0xbe, 0x59, 0x74, 0x8f, 0xfe, 0x0d, 0xeb, 0x85, 0x09,
	0xe3, 0x59, 0x76, 0xf2, 0xfe, 0x83, 0xb7, 0x8e, 0xab, 0xb5, 0xa1, 0x79, 0x34, 0x36, 0xbf, 0x69,
	0x60, 0xaf, 0x22, 0x86, 0xdf, 0x0f, 0xa2, 0xfd, 0x9f, 0x83, 0x9c, 0xff, 0x3d, 0x9d, 0xba, 0xba,
	0xa3, 0xe6, 0xc3, 0xd3, 0xf9, 0x63, 0x2e, 0xce, 0xf9, 0xd5, 0xd2, 0xd0, 0xae, 0x97, 0x86, 0xf6,
	0x63, 0x69, 0x68, 0x5f, 0x57, 0x46, 0xed, 0x7a, 0x65, 0xd4, 0xbe, 0xaf, 0x8c, 0xda, 0xe7, 0xfd,
	0x87, 0x2c, 0x5e, 0xde, 0xbd, 0x27, 0xe5, 0x16, 0x6f, 0xab, 0xa7, 0x74, 0xf8, 0x33, 0x00, 0x00,
	0xff, 0xff, 0x79, 0x47, 0xe2, 0xe5, 0xe0, 0x03, 0x00, 0x00,
}

func (m *GenesisState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.PubRandCommit) > 0 {
		for iNdEx := len(m.PubRandCommit) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.PubRandCommit[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x2a
		}
	}
	if len(m.VoteSigs) > 0 {
		for iNdEx := len(m.VoteSigs) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.VoteSigs[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x22
		}
	}
	if len(m.Evidences) > 0 {
		for iNdEx := len(m.Evidences) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Evidences[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x1a
		}
	}
	if len(m.IndexedBlocks) > 0 {
		for iNdEx := len(m.IndexedBlocks) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.IndexedBlocks[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintGenesis(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *VoteSig) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *VoteSig) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *VoteSig) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.FinalitySig != nil {
		{
			size := m.FinalitySig.Size()
			i -= size
			if _, err := m.FinalitySig.MarshalTo(dAtA[i:]); err != nil {
				return 0, err
			}
			i = encodeVarintGenesis(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x1a
	}
	if m.FpBtcPk != nil {
		{
			size := m.FpBtcPk.Size()
			i -= size
			if _, err := m.FpBtcPk.MarshalTo(dAtA[i:]); err != nil {
				return 0, err
			}
			i = encodeVarintGenesis(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if m.BlockHeight != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.BlockHeight))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *PubRandCommitWithPK) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *PubRandCommitWithPK) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *PubRandCommitWithPK) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.PubRandCommit != nil {
		{
			size, err := m.PubRandCommit.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintGenesis(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if m.FpBtcPk != nil {
		{
			size := m.FpBtcPk.Size()
			i -= size
			if _, err := m.FpBtcPk.MarshalTo(dAtA[i:]); err != nil {
				return 0, err
			}
			i = encodeVarintGenesis(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintGenesis(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenesis(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovGenesis(uint64(l))
	if len(m.IndexedBlocks) > 0 {
		for _, e := range m.IndexedBlocks {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.Evidences) > 0 {
		for _, e := range m.Evidences {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.VoteSigs) > 0 {
		for _, e := range m.VoteSigs {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	if len(m.PubRandCommit) > 0 {
		for _, e := range m.PubRandCommit {
			l = e.Size()
			n += 1 + l + sovGenesis(uint64(l))
		}
	}
	return n
}

func (m *VoteSig) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.BlockHeight != 0 {
		n += 1 + sovGenesis(uint64(m.BlockHeight))
	}
	if m.FpBtcPk != nil {
		l = m.FpBtcPk.Size()
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.FinalitySig != nil {
		l = m.FinalitySig.Size()
		n += 1 + l + sovGenesis(uint64(l))
	}
	return n
}

func (m *PubRandCommitWithPK) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.FpBtcPk != nil {
		l = m.FpBtcPk.Size()
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.PubRandCommit != nil {
		l = m.PubRandCommit.Size()
		n += 1 + l + sovGenesis(uint64(l))
	}
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GenesisState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: GenesisState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IndexedBlocks", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.IndexedBlocks = append(m.IndexedBlocks, &IndexedBlock{})
			if err := m.IndexedBlocks[len(m.IndexedBlocks)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Evidences", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Evidences = append(m.Evidences, &Evidence{})
			if err := m.Evidences[len(m.Evidences)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field VoteSigs", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.VoteSigs = append(m.VoteSigs, &VoteSig{})
			if err := m.VoteSigs[len(m.VoteSigs)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PubRandCommit", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PubRandCommit = append(m.PubRandCommit, &PubRandCommitWithPK{})
			if err := m.PubRandCommit[len(m.PubRandCommit)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func (m *VoteSig) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: VoteSig: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: VoteSig: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BlockHeight", wireType)
			}
			m.BlockHeight = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BlockHeight |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FpBtcPk", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			var v github_com_babylonchain_babylon_types.BIP340PubKey
			m.FpBtcPk = &v
			if err := m.FpBtcPk.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FinalitySig", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			var v github_com_babylonchain_babylon_types.SchnorrEOTSSig
			m.FinalitySig = &v
			if err := m.FinalitySig.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func (m *PubRandCommitWithPK) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
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
			return fmt.Errorf("proto: PubRandCommitWithPK: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PubRandCommitWithPK: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field FpBtcPk", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			var v github_com_babylonchain_babylon_types.BIP340PubKey
			m.FpBtcPk = &v
			if err := m.FpBtcPk.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PubRandCommit", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
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
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.PubRandCommit == nil {
				m.PubRandCommit = &PubRandCommit{}
			}
			if err := m.PubRandCommit.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
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
func skipGenesis(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
					return 0, ErrIntOverflowGenesis
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
				return 0, ErrInvalidLengthGenesis
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenesis
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenesis
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenesis        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenesis          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenesis = fmt.Errorf("proto: unexpected end of group")
)
