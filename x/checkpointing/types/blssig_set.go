package types

import (
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/tendermint/tendermint/libs/bits"
	"github.com/tendermint/tendermint/libs/bytes"
)

type BlsSigSet struct {
	epoch          uint16
	lastCommitHash bytes.HexBytes
	validators     types.Validators

	sum          uint64
	sigsBySigner map[string]*BlsSig
	sigsBitArray *bits.BitArray
}

// NewBlsSigSet constructs a new BlsSigSet struct used to accumulate bls sigs for a given epoch
func NewBlsSigSet(epoch uint16, lastCommitHash bytes.HexBytes, validators types.Validators) *BlsSigSet {
	return &BlsSigSet{
		epoch:          epoch,
		lastCommitHash: lastCommitHash,
		validators:     validators,
		sum:            0,
		sigsBySigner:   make(map[string]*BlsSig, validators.Len()),
		sigsBitArray:   bits.NewBitArray(validators.Len()),
	}
}

func (bs *BlsSigSet) AddBlsSig(sig *BlsSig) (bool, error) {
	if bs == nil {
		panic("AddVote() on nil BlsSigSet")
	}
	return bs.addBlsSig(sig)
}

func (bs *BlsSigSet) addBlsSig(sig *BlsSig) (bool, error) {
	panic("implement this!")
}

func (bs *BlsSigSet) MakeRawCheckpoint() *RawCheckpoint {
	panic("implement this!")
}
