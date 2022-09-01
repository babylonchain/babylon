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
		Status:   GenRandomStatus(),
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

func GenRandomSequenceRawCheckpointsWithMeta() []*types.RawCheckpointWithMeta {
	topEpoch := GenRandomEpochNum()
	var checkpoints []*types.RawCheckpointWithMeta
	for i := uint64(0); i <= topEpoch; i++ {
		ckpt := GenRandomRawCheckpointWithMeta()
		ckpt.Ckpt.EpochNum = i
		checkpoints = append(checkpoints, ckpt)
	}

	return checkpoints
}

func GenRandomEpochNum() uint64 {
	epochNum := rand.Int63n(100)
	return uint64(epochNum)
}

func GenRandomLastCommitHash() types.LastCommitHash {
	return GenRandomByteArray(types.HashSize)
}

func GenRandomBlsMultiSig() bls12381.Signature {
	return GenRandomByteArray(bls12381.SignatureSize)
}

func GenRandomStatus() types.CheckpointStatus {
	return types.CheckpointStatus(rand.Int31n(int32(len(types.CheckpointStatus_name))))
}
