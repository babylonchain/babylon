package types

import (
	"bytes"
	"github.com/cosmos/cosmos-sdk/codec"
)

type LastCommitHash []byte

//type ValidatorAddress string

type BlsSigHash []byte

type RawCkptHash []byte

//
//func (addr ValidatorAddress) Marshal() ([]byte, error) {
//	return []byte(addr), nil
//}
//
//func (addr ValidatorAddress) MustMarshal() []byte {
//	bz, err := addr.Marshal()
//	if err != nil {
//		panic(err)
//	}
//
//	return bz
//}
//
//func (addr ValidatorAddress) MarshalTo(data []byte) (int, error) {
//	copy(data, addr)
//	return len(data), nil
//}
//
//func (addr ValidatorAddress) Size() int {
//	bz := addr.MustMarshal()
//	return len(bz)
//}
//
//func (addr *ValidatorAddress) Unmarshal(data []byte) error {
//	*addr = ValidatorAddress(data)
//	return nil
//}
//
//func (addr ValidatorAddress) Byte() []byte {
//	return []byte(addr)
//}
//
//func (addr ValidatorAddress) Equal(s ValidatorAddress) bool {
//	return addr == s
//}

func NewCheckpoint(epochNum uint64, lch LastCommitHash) *RawCheckpoint {
	return &RawCheckpoint{
		EpochNum:       epochNum,
		LastCommitHash: lch,
		Bitmap:         nil,
		BlsMultiSig:    nil,
	}
}

func NewCheckpointWithMeta(ckpt *RawCheckpoint, status CheckpointStatus) *RawCheckpointWithMeta {
	return &RawCheckpointWithMeta{
		Ckpt:   ckpt,
		Status: status,
	}
}

func RawCkptToBytes(cdc codec.BinaryCodec, ckpt *RawCheckpoint) []byte {
	return cdc.MustMarshal(ckpt)
}

func BytesToRawCkpt(cdc codec.BinaryCodec, bz []byte) (*RawCheckpoint, error) {
	ckpt := new(RawCheckpoint)
	err := cdc.Unmarshal(bz, ckpt)
	return ckpt, err
}

func CkptWithMetaToBytes(cdc codec.BinaryCodec, ckptWithMeta *RawCheckpointWithMeta) []byte {
	return cdc.MustMarshal(ckptWithMeta)
}

func BytesToCkptWithMeta(cdc codec.BinaryCodec, bz []byte) (*RawCheckpointWithMeta, error) {
	ckptWithMeta := new(RawCheckpointWithMeta)
	err := cdc.Unmarshal(bz, ckptWithMeta)
	return ckptWithMeta, err
}

func BlsSigToBytes(cdc codec.BinaryCodec, blsSig *BlsSig) []byte {
	return cdc.MustMarshal(blsSig)
}

func BytesToBlsSig(cdc codec.BinaryCodec, bz []byte) (*BlsSig, error) {
	blsSig := new(BlsSig)
	err := cdc.Unmarshal(bz, blsSig)
	return blsSig, err
}

func (m RawCkptHash) Equals(h RawCkptHash) bool {
	if bytes.Compare(m.Bytes(), h.Bytes()) == 0 {
		return true
	}
	return false
}
