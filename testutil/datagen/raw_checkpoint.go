package datagen

import (
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/boljen/go-bitmap"
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
	randomBLSSig := GenRandomBlsMultiSig()
	return &types.RawCheckpoint{
		EpochNum:       GenRandomEpochNum(),
		LastCommitHash: &randomHashBytes,
		Bitmap:         bitmap.New(13),
		BlsMultiSig:    &randomBLSSig,
	}
}

func GenRandomEpochNum() uint64 {
	return rand.Uint64()
}

func GenRandomLastCommitHash() types.LastCommitHash {
	return GenRandomByteArray(types.HashSize)
}

func GenRandomBlsMultiSig() bls12381.Signature {
	return GenRandomByteArray(bls12381.SignatureSize)
}
