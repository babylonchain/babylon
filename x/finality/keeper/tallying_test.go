package keeper_test

import (
	"encoding/hex"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	bbn "github.com/babylonchain/babylon/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/babylonchain/babylon/x/finality/keeper"
	"github.com/babylonchain/babylon/x/finality/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func FuzzTallying_PanicCases(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		bsKeeper := types.NewMockBTCStakingKeeper(ctrl)
		iKeeper := types.NewMockIncentiveKeeper(ctrl)
		fKeeper, ctx := keepertest.FinalityKeeper(t, bsKeeper, iKeeper)

		// Case 1: expect to panic if tallying upon BTC staking protocol is not activated
		bsKeeper.EXPECT().GetBTCStakingActivatedHeight(gomock.Any()).Return(uint64(0), bstypes.ErrBTCStakingNotActivated).Times(1)
		require.Panics(t, func() { fKeeper.TallyBlocks(ctx) })

		// Case 2: expect to panic if finalised block with nil validator set
		fKeeper.SetBlock(ctx, &types.IndexedBlock{
			Height:         1,
			AppHash: datagen.GenRandomByteArray(r, 32),
			Finalized:      true,
		})
		// activate BTC staking protocol at height 1
		ctx = ctx.WithBlockHeight(1)
		bsKeeper.EXPECT().GetBTCStakingActivatedHeight(gomock.Any()).Return(uint64(1), nil).Times(1)
		bsKeeper.EXPECT().GetVotingPowerTable(gomock.Any(), gomock.Eq(uint64(1))).Return(nil).Times(1)
		require.Panics(t, func() { fKeeper.TallyBlocks(ctx) })
	})
}

func FuzzTallying_FinalizingNoBlock(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		bsKeeper := types.NewMockBTCStakingKeeper(ctrl)
		iKeeper := types.NewMockIncentiveKeeper(ctrl)
		fKeeper, ctx := keepertest.FinalityKeeper(t, bsKeeper, iKeeper)

		// activate BTC staking protocol at a random height
		activatedHeight := datagen.RandomInt(r, 10) + 1

		// index a list of blocks, don't give them QCs, and tally them
		// Expect they are not finalised
		for i := activatedHeight; i < activatedHeight+10; i++ {
			// index blocks
			fKeeper.SetBlock(ctx, &types.IndexedBlock{
				Height:         i,
				AppHash: datagen.GenRandomByteArray(r, 32),
				Finalized:      false,
			})
			// this block does not have QC
			err := giveNoQCToHeight(r, ctx, bsKeeper, fKeeper, i)
			require.NoError(t, err)
		}
		// add mock queries to GetBTCStakingActivatedHeight
		ctx = ctx.WithBlockHeight(int64(activatedHeight) + 10 - 1)
		bsKeeper.EXPECT().GetBTCStakingActivatedHeight(gomock.Any()).Return(activatedHeight, nil).Times(1)
		// tally blocks and none of them should be finalised
		fKeeper.TallyBlocks(ctx)
		for i := activatedHeight; i < activatedHeight+10; i++ {
			ib, err := fKeeper.GetBlock(ctx, i)
			require.NoError(t, err)
			require.False(t, ib.Finalized)
		}
	})

}

func FuzzTallying_FinalizingSomeBlocks(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		bsKeeper := types.NewMockBTCStakingKeeper(ctrl)
		iKeeper := types.NewMockIncentiveKeeper(ctrl)
		fKeeper, ctx := keepertest.FinalityKeeper(t, bsKeeper, iKeeper)

		// activate BTC staking protocol at a random height
		activatedHeight := datagen.RandomInt(r, 10) + 1

		// index a list of blocks, give some of them QCs, and tally them.
		// Expect they are all finalised
		numWithQCs := datagen.RandomInt(r, 5) + 1
		for i := activatedHeight; i < activatedHeight+10; i++ {
			// index blocks
			fKeeper.SetBlock(ctx, &types.IndexedBlock{
				Height:         i,
				AppHash: datagen.GenRandomByteArray(r, 32),
				Finalized:      false,
			})
			if i < activatedHeight+numWithQCs {
				// this block has QC
				err := giveQCToHeight(r, ctx, bsKeeper, fKeeper, i)
				require.NoError(t, err)
			} else {
				// this block does not have QC
				err := giveNoQCToHeight(r, ctx, bsKeeper, fKeeper, i)
				require.NoError(t, err)
			}
		}
		// we don't test incentive in this function
		bsKeeper.EXPECT().GetRewardDistCache(gomock.Any(), gomock.Any()).Return(bstypes.NewRewardDistCache(), nil).Times(int(numWithQCs))
		iKeeper.EXPECT().RewardBTCStaking(gomock.Any(), gomock.Any(), gomock.Any()).Return().Times(int(numWithQCs))
		bsKeeper.EXPECT().RemoveRewardDistCache(gomock.Any(), gomock.Any()).Return().Times(int(numWithQCs))
		// add mock queries to GetBTCStakingActivatedHeight
		ctx = ctx.WithBlockHeight(int64(activatedHeight) + 10 - 1)
		bsKeeper.EXPECT().GetBTCStakingActivatedHeight(gomock.Any()).Return(activatedHeight, nil).Times(1)
		// tally blocks and none of them should be finalised
		fKeeper.TallyBlocks(ctx)
		for i := activatedHeight; i < activatedHeight+10; i++ {
			ib, err := fKeeper.GetBlock(ctx, i)
			require.NoError(t, err)
			if i < activatedHeight+numWithQCs {
				require.True(t, ib.Finalized)
			} else {
				require.False(t, ib.Finalized)
			}
		}
	})

}

func giveQCToHeight(r *rand.Rand, ctx sdk.Context, bsKeeper *types.MockBTCStakingKeeper, fKeeper *keeper.Keeper, height uint64) error {
	// 4 BTC vals
	valSet := map[string]uint64{}
	// 3 votes
	for i := 0; i < 3; i++ {
		votedValPK, err := datagen.GenRandomBIP340PubKey(r)
		if err != nil {
			return err
		}
		votedSig, err := bbn.NewSchnorrEOTSSig(datagen.GenRandomByteArray(r, 32))
		if err != nil {
			return err
		}
		fKeeper.SetSig(ctx, height, votedValPK, votedSig)
		// add val
		valSet[votedValPK.MarshalHex()] = 1
	}
	// the rest val that does not vote
	valSet[hex.EncodeToString(datagen.GenRandomByteArray(r, 32))] = 1
	bsKeeper.EXPECT().GetVotingPowerTable(gomock.Any(), gomock.Eq(height)).Return(valSet).Times(1)

	return nil
}

func giveNoQCToHeight(r *rand.Rand, ctx sdk.Context, bsKeeper *types.MockBTCStakingKeeper, fKeeper *keeper.Keeper, height uint64) error {
	// 1 vote
	votedValPK, err := datagen.GenRandomBIP340PubKey(r)
	if err != nil {
		return err
	}
	votedSig, err := bbn.NewSchnorrEOTSSig(datagen.GenRandomByteArray(r, 32))
	if err != nil {
		return err
	}
	fKeeper.SetSig(ctx, height, votedValPK, votedSig)
	// 4 BTC vals
	valSet := map[string]uint64{
		votedValPK.MarshalHex():                               1,
		hex.EncodeToString(datagen.GenRandomByteArray(r, 32)): 1,
		hex.EncodeToString(datagen.GenRandomByteArray(r, 32)): 1,
		hex.EncodeToString(datagen.GenRandomByteArray(r, 32)): 1,
	}
	bsKeeper.EXPECT().GetVotingPowerTable(gomock.Any(), gomock.Eq(height)).Return(valSet).MaxTimes(1)

	return nil
}
