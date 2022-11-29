package keeper_test

import (
	"github.com/babylonchain/babylon/btctxformatter"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/boljen/go-bitmap"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/testutil/mocks"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

// FuzzKeeperAddRawCheckpoint checks
// 1. if the RawCheckpointWithMeta object is nil, an error is returned
// 2. if the RawCheckpointWithMeta object does not exist when GetRawCheckpoint is called, an error is returned
// 3. if a RawCheckpointWithMeta object with the same epoch number already exists, an error is returned
func FuzzKeeperAddRawCheckpoint(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 1)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		ckptKeeper, ctx, _ := testkeeper.CheckpointingKeeper(t, nil, nil, client.Context{})

		// test nil raw checkpoint
		err := ckptKeeper.AddRawCheckpoint(ctx, nil)
		require.Errorf(t, err, "add a nil raw checkpoint")

		// test random raw checkpoint
		mockCkptWithMeta := datagen.GenRandomRawCheckpointWithMeta()
		ckpt, err := ckptKeeper.GetRawCheckpoint(ctx, mockCkptWithMeta.Ckpt.EpochNum)
		require.Nil(t, ckpt)
		require.Errorf(t, err, "raw checkpoint does not exist")
		err = ckptKeeper.AddRawCheckpoint(
			ctx,
			mockCkptWithMeta,
		)
		require.NoError(t, err)
		ckpt, err = ckptKeeper.GetRawCheckpoint(ctx, mockCkptWithMeta.Ckpt.EpochNum)
		require.NoError(t, err)
		t.Logf("mocked ckpt: %v\n got ckpt: %v\n", mockCkptWithMeta, ckpt)
		require.True(t, ckpt.Equal(mockCkptWithMeta))

		// test existing raw checkpoint by epoch number
		_, err = ckptKeeper.BuildRawCheckpoint(
			ctx,
			mockCkptWithMeta.Ckpt.EpochNum,
			datagen.GenRandomLastCommitHash(),
		)
		require.Errorf(t, err, "raw checkpoint with the same epoch already exists")
	})
}

// FuzzKeeperSetCheckpointStatus checks if the checkpoint status
// is not correct, the status will not be changed
func FuzzKeeperSetCheckpointStatus(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 1)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ek := mocks.NewMockEpochingKeeper(ctrl)
		ckptKeeper, ctx, _ := testkeeper.CheckpointingKeeper(t, ek, nil, client.Context{})

		mockCkptWithMeta := datagen.GenRandomRawCheckpointWithMeta()
		mockCkptWithMeta.Status = types.Accumulating
		epoch := mockCkptWithMeta.Ckpt.EpochNum

		_ = ckptKeeper.AddRawCheckpoint(
			ctx,
			mockCkptWithMeta,
		)
		ckptKeeper.SetCheckpointSubmitted(ctx, epoch)
		status, err := ckptKeeper.GetStatus(ctx, epoch)
		require.NoError(t, err)
		require.Equal(t, types.Accumulating, status)
		mockCkptWithMeta.Status = types.Sealed
		err = ckptKeeper.UpdateCheckpoint(ctx, mockCkptWithMeta)
		require.NoError(t, err)
		ckptKeeper.SetCheckpointSubmitted(ctx, epoch)
		status, err = ckptKeeper.GetStatus(ctx, epoch)
		require.NoError(t, err)
		require.Equal(t, types.Submitted, status)
		ckptKeeper.SetCheckpointConfirmed(ctx, epoch)
		status, err = ckptKeeper.GetStatus(ctx, epoch)
		require.NoError(t, err)
		require.Equal(t, types.Confirmed, status)
		ckptKeeper.SetCheckpointConfirmed(ctx, epoch)
		status, err = ckptKeeper.GetStatus(ctx, epoch)
		require.NoError(t, err)
		require.Equal(t, types.Confirmed, status)
		ckptKeeper.SetCheckpointFinalized(ctx, epoch)
		status, err = ckptKeeper.GetStatus(ctx, epoch)
		require.NoError(t, err)
		require.Equal(t, types.Finalized, status)
	})
}

// FuzzKeeperCheckpointEpoch checks the following scenarios
// 1. given a valid slice of checkpoint bytes, should return its epoch number
// 2. given a dummy checkpoint, should return ErrInvalidRawCheckpoint
// 3. given a conflicting checkpoint, should panic
func FuzzKeeperCheckpointEpoch(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 1)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		ek := mocks.NewMockEpochingKeeper(ctrl)
		ek.EXPECT().GetValidatorSet(gomock.Any(), gomock.Any()).Return(valSet).AnyTimes()
		ek.EXPECT().GetTotalVotingPower(gomock.Any(), gomock.Any()).Return(int64(20)).AnyTimes()
		ckptKeeper, ctx, _ := testkeeper.CheckpointingKeeper(t, ek, nil, client.Context{})
		for i, val := range valSet {
			err := ckptKeeper.CreateRegistration(ctx, pubkeys[i], val.Addr)
			require.NoError(t, err)
		}

		// add local checkpoint, signed by the first validator
		bm := bitmap.New(104)
		bm.Set(0, true)
		localCkptWithMeta := datagen.GenRandomRawCheckpointWithMeta()
		localCkptWithMeta.Status = types.Sealed
		localCkptWithMeta.PowerSum = 10
		localCkptWithMeta.Ckpt.Bitmap = bm
		msgBytes := append(sdk.Uint64ToBigEndian(localCkptWithMeta.Ckpt.EpochNum), *localCkptWithMeta.Ckpt.LastCommitHash...)
		sig := bls12381.Sign(blsPrivKey1, msgBytes)
		localCkptWithMeta.Ckpt.BlsMultiSig = &sig
		_ = ckptKeeper.AddRawCheckpoint(
			ctx,
			localCkptWithMeta,
		)

		// 1. check valid checkpoint
		btcCkptBytes := generateBtcCkptBytes(
			localCkptWithMeta.Ckpt.EpochNum,
			localCkptWithMeta.Ckpt.LastCommitHash.MustMarshal(),
			localCkptWithMeta.Ckpt.Bitmap,
			localCkptWithMeta.Ckpt.BlsMultiSig.Bytes(),
			t,
		)
		epoch, err := ckptKeeper.CheckpointEpoch(ctx, btcCkptBytes)
		require.NoError(t, err)
		require.Equal(t, localCkptWithMeta.Ckpt.EpochNum, epoch)

		// 2. check a checkpoint with invalid sig
		btcCkptBytes = generateBtcCkptBytes(
			localCkptWithMeta.Ckpt.EpochNum,
			localCkptWithMeta.Ckpt.LastCommitHash.MustMarshal(),
			localCkptWithMeta.Ckpt.Bitmap,
			randNBytes(btctxformatter.BlsSigLength),
			t,
		)
		epoch, err = ckptKeeper.CheckpointEpoch(ctx, btcCkptBytes)
		require.ErrorIs(t, err, types.ErrInvalidRawCheckpoint)

		// 3. check a conflicting checkpoint; signed on a random lastcommithash
		conflictLastCommitHash := randNBytes(btctxformatter.LastCommitHashLength)
		msgBytes = append(sdk.Uint64ToBigEndian(localCkptWithMeta.Ckpt.EpochNum), conflictLastCommitHash...)
		btcCkptBytes = generateBtcCkptBytes(
			localCkptWithMeta.Ckpt.EpochNum,
			conflictLastCommitHash,
			localCkptWithMeta.Ckpt.Bitmap,
			bls12381.Sign(blsPrivKey1, msgBytes),
			t,
		)
		require.Panics(t, func() {
			_, _ = ckptKeeper.CheckpointEpoch(ctx, btcCkptBytes)
		})
	})
}

func randNBytes(n int) []byte {
	bytes := make([]byte, n)
	rand.Read(bytes)
	return bytes
}

func generateBtcCkptBytes(epoch uint64, lch []byte, bitmap []byte, blsSig []byte, t *testing.T) []byte {
	tag := randNBytes(btctxformatter.TagLength)
	babylonTag := btctxformatter.BabylonTag(tag[:btctxformatter.TagLength])
	address := randNBytes(btctxformatter.AddressLength)

	rawBTCCkpt := &btctxformatter.RawBtcCheckpoint{
		Epoch:            epoch,
		LastCommitHash:   lch,
		BitMap:           bitmap,
		SubmitterAddress: address,
		BlsSig:           blsSig,
	}
	firstHalf, secondHalf, err := btctxformatter.EncodeCheckpointData(
		babylonTag,
		btctxformatter.CurrentVersion,
		rawBTCCkpt,
	)
	require.NoError(t, err)
	decodedFirst, err := btctxformatter.IsBabylonCheckpointData(babylonTag, btctxformatter.CurrentVersion, firstHalf)
	require.NoError(t, err)
	decodedSecond, err := btctxformatter.IsBabylonCheckpointData(babylonTag, btctxformatter.CurrentVersion, secondHalf)
	require.NoError(t, err)
	ckptData, err := btctxformatter.ConnectParts(btctxformatter.CurrentVersion, decodedFirst.Data, decodedSecond.Data)
	require.NoError(t, err)

	return ckptData
}
