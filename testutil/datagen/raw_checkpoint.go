package datagen

import (
	"github.com/babylonchain/babylon/x/checkpointing/types"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"math/rand"
)

func GenRandomRawCheckpointWithMeta() *types.RawCheckpointWithMeta {
	ckptWithMeta := &types.RawCheckpointWithMeta{
		Ckpt:     GenRandomRawCheckpoint(),
		Status:   types.Accumulating,
		PowerSum: 0,
	}
	return ckptWithMeta
}

func GenRandomRawCheckpoint() *types.RawCheckpoint {
	randomHashBytes := GenRandomLastCommitHash()
	return &types.RawCheckpoint{
		EpochNum:       GenRandomEpochNum(),
		LastCommitHash: &randomHashBytes,
	}
}

func GenRandomEpochNum() uint64 {
	return rand.Uint64()
}

func GenRandomLastCommitHash() types.LastCommitHash {
	return tmrand.Bytes(types.HashSize)
}
