package types

import (
	"encoding/binary"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

func (m BlsSig) Hash() BlsSigHash {
	fields := [][]byte{
		m.LastCommitHash,
		m.BlsSig,
		sdk.Uint64ToBigEndian(m.EpochNum),
		[]byte(m.SignerAddress),
	}
	return hash(fields)
}

func (m RawCheckpoint) Hash() RawCkptHash {
	fields := [][]byte{
		m.LastCommitHash,
		sdk.Uint64ToBigEndian(m.EpochNum),
		m.Bitmap,
		m.BlsMultiSig,
	}
	return hash(fields)
}

func hash(fields [][]byte) []byte {
	var bz []byte
	for _, b := range fields {
		bz = append(bz, b...)
	}
	return tmhash.Sum(bz)
}

func BlsSigHashToBytes(h BlsSigHash) []byte {
	return h
}

func Uint32ToBitEndian(i uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, i)
	return b
}

func BigEndianToUint32(bz []byte) uint32 {
	if len(bz) == 0 {
		return 0
	}

	return binary.BigEndian.Uint32(bz)
}
