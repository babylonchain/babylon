package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto/tmhash"
)

func (m BlsSig) Hash() BlsSigHash {
	fields := [][]byte{
		sdk.Uint64ToBigEndian(m.EpochNum),
		m.LastCommitHash.MustMarshal(),
		m.BlsSig.MustMarshal(),
		[]byte(m.SignerAddress),
	}
	return hash(fields)
}

func (m RawCheckpoint) Hash() RawCkptHash {
	fields := [][]byte{
		sdk.Uint64ToBigEndian(m.EpochNum),
		m.LastCommitHash.MustMarshal(),
		m.BlsMultiSig.MustMarshal(),
		m.Bitmap,
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

func (m BlsSigHash) Bytes() []byte {
	return m
}

func (m RawCkptHash) Bytes() []byte {
	return m
}
