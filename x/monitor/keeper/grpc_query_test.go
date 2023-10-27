package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/btctxformatter"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/testutil/mocks"
	ckpttypes "github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/babylonchain/babylon/x/epoching/testepoching"
	types2 "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/monitor/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func FuzzQueryEndedEpochBtcHeight(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		// a genesis validator is generated for setup
		helper := testepoching.NewHelper(t)
		lck := helper.App.BTCLightClientKeeper
		mk := helper.App.MonitorKeeper
		ek := helper.EpochingKeeper
		queryHelper := baseapp.NewQueryServerTestHelper(helper.Ctx, helper.App.InterfaceRegistry())
		types.RegisterQueryServer(queryHelper, mk)
		queryClient := types.NewQueryClient(queryHelper)

		// BeginBlock of block 1, and thus entering epoch 1
		ctx := helper.BeginBlock(r)
		epoch := ek.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		root := lck.GetBaseBTCHeader(ctx)
		chain := datagen.GenRandomValidChainStartingFrom(
			r,
			0,
			root.Header.ToBlockHeader(),
			nil,
			10,
		)
		headerBytes := datagen.HeaderToHeaderBytes(chain)
		err := lck.InsertHeaders(ctx, headerBytes)
		require.NoError(t, err)

		// EndBlock of block 1
		ctx = helper.EndBlock()

		// go to BeginBlock of block 11, and thus entering epoch 2
		for i := uint64(0); i < ek.GetParams(ctx).EpochInterval; i++ {
			ctx = helper.GenAndApplyEmptyBlock(r)
		}
		epoch = ek.GetEpoch(ctx)
		require.Equal(t, uint64(2), epoch.EpochNumber)

		// query epoch 0 ended BTC light client height, should return base height
		req := types.QueryEndedEpochBtcHeightRequest{
			EpochNum: 0,
		}
		resp, err := queryClient.EndedEpochBtcHeight(ctx, &req)
		require.NoError(t, err)
		require.Equal(t, lck.GetBaseBTCHeader(ctx).Height, resp.BtcLightClientHeight)

		// query epoch 1 ended BTC light client height, should return tip height
		req = types.QueryEndedEpochBtcHeightRequest{
			EpochNum: 1,
		}
		resp, err = queryClient.EndedEpochBtcHeight(ctx, &req)
		require.NoError(t, err)
		require.Equal(t, lck.GetTipInfo(ctx).Height, resp.BtcLightClientHeight)
	})
}

func FuzzQueryReportedCheckpointBtcHeight(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		// a genesis validator is generated for setup
		helper := testepoching.NewHelper(t)
		ctl := gomock.NewController(t)
		defer ctl.Finish()
		lck := helper.App.BTCLightClientKeeper
		mk := helper.App.MonitorKeeper
		ek := helper.EpochingKeeper
		ck := helper.App.CheckpointingKeeper
		mockEk := mocks.NewMockEpochingKeeper(ctl)
		ck.SetEpochingKeeper(mockEk)
		queryHelper := baseapp.NewQueryServerTestHelper(helper.Ctx, helper.App.InterfaceRegistry())
		types.RegisterQueryServer(queryHelper, mk)
		queryClient := types.NewQueryClient(queryHelper)

		// BeginBlock of block 1, and thus entering epoch 1
		ctx := helper.BeginBlock(r)
		epoch := ek.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		root := lck.GetBaseBTCHeader(ctx)
		chain := datagen.GenRandomValidChainStartingFrom(
			r,
			0,
			root.Header.ToBlockHeader(),
			nil,
			10,
		)
		headerBytes := datagen.HeaderToHeaderBytes(chain)
		err := lck.InsertHeaders(ctx, headerBytes)
		require.NoError(t, err)

		// Add checkpoint
		valBlsSet, privKeys := datagen.GenerateValidatorSetWithBLSPrivKeys(int(datagen.RandomIntOtherThan(r, 0, 10)))
		valSet := make([]types2.Validator, len(valBlsSet.ValSet))
		for i, val := range valBlsSet.ValSet {
			valSet[i] = types2.Validator{
				Addr:  []byte(val.ValidatorAddress),
				Power: int64(val.VotingPower),
			}
			err := ck.CreateRegistration(ctx, val.BlsPubKey, []byte(val.ValidatorAddress))
			require.NoError(t, err)
		}
		mockCkptWithMeta := &ckpttypes.RawCheckpointWithMeta{Ckpt: datagen.GenerateLegitimateRawCheckpoint(r, privKeys)}
		mockEk.EXPECT().GetValidatorSet(gomock.Any(), gomock.Eq(mockCkptWithMeta.Ckpt.EpochNum)).Return(valSet).AnyTimes()
		// make sure voting power is always sufficient
		mockEk.EXPECT().GetTotalVotingPower(gomock.Any(), gomock.Eq(mockCkptWithMeta.Ckpt.EpochNum)).Return(int64(0)).AnyTimes()
		err = ck.AddRawCheckpoint(
			ctx,
			mockCkptWithMeta,
		)
		require.NoError(t, err)

		// Verify checkpoint
		btcCkpt := btctxformatter.RawBtcCheckpoint{
			Epoch:            mockCkptWithMeta.Ckpt.EpochNum,
			LastCommitHash:   *mockCkptWithMeta.Ckpt.LastCommitHash,
			BitMap:           mockCkptWithMeta.Ckpt.Bitmap,
			SubmitterAddress: datagen.GenRandomByteArray(r, btctxformatter.AddressLength),
			BlsSig:           *mockCkptWithMeta.Ckpt.BlsMultiSig,
		}
		err = ck.VerifyCheckpoint(ctx, btcCkpt)
		require.NoError(t, err)

		// query reported checkpoint BTC light client height
		req := types.QueryReportedCheckpointBtcHeightRequest{
			CkptHash: mockCkptWithMeta.Ckpt.HashStr(),
		}
		resp, err := queryClient.ReportedCheckpointBtcHeight(ctx, &req)
		require.NoError(t, err)
		require.Equal(t, lck.GetTipInfo(ctx).Height, resp.BtcLightClientHeight)

		// query not reported checkpoint BTC light client height, should expect an ErrCheckpointNotReported
		req = types.QueryReportedCheckpointBtcHeightRequest{
			CkptHash: datagen.GenRandomHexStr(r, 32),
		}
		_, err = queryClient.ReportedCheckpointBtcHeight(ctx, &req)
		require.ErrorIs(t, err, types.ErrCheckpointNotReported)
	})
}
