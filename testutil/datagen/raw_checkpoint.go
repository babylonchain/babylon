package datagen

import (
	"math/rand"

	"github.com/babylonchain/babylon/btctxformatter"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/boljen/go-bitmap"
)

func GetRandomRawBtcCheckpoint() *btctxformatter.RawBtcCheckpoint {
	rawCkpt := GenRandomRawCheckpoint()
	return &btctxformatter.RawBtcCheckpoint{
		Epoch:            rawCkpt.EpochNum,
		LastCommitHash:   *rawCkpt.LastCommitHash,
		BitMap:           rawCkpt.Bitmap,
		SubmitterAddress: GenRandomByteArray(btctxformatter.AddressLength),
		BlsSig:           rawCkpt.BlsMultiSig.Bytes(),
	}
}

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

// GenRandomSequenceRawCheckpointsWithMeta generates random checkpoints from epoch 0 to a random epoch
func GenRandomSequenceRawCheckpointsWithMeta() []*types.RawCheckpointWithMeta {
	var topEpoch, finalEpoch uint64
	epoch1 := GenRandomEpochNum()
	epoch2 := GenRandomEpochNum()
	if epoch1 > epoch2 {
		topEpoch = epoch1
		finalEpoch = epoch2
	} else {
		topEpoch = epoch2
		finalEpoch = epoch1
	}
	var checkpoints []*types.RawCheckpointWithMeta
	for e := uint64(0); e <= topEpoch; e++ {
		ckpt := GenRandomRawCheckpointWithMeta()
		ckpt.Ckpt.EpochNum = e
		if e <= finalEpoch {
			ckpt.Status = types.Finalized
		}
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

// GenRandomStatus generates random status except for Finalized
func GenRandomStatus() types.CheckpointStatus {
	return types.CheckpointStatus(rand.Int31n(int32(len(types.CheckpointStatus_name) - 1)))
}
