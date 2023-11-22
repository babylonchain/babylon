package keeper_test

import (
	"math/rand"
	"testing"

	"cosmossdk.io/math"

	"github.com/babylonchain/babylon/app"
	"github.com/babylonchain/babylon/testutil/datagen"
	checkpointingkeeper "github.com/babylonchain/babylon/x/checkpointing/keeper"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/babylonchain/babylon/x/epoching/testepoching"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/stretchr/testify/require"
)

func FuzzGetValidatorBlsKeySet(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		// a genesis validator is generated for setup
		helper := testepoching.NewHelper(t)
		ctx := helper.Ctx
		ek := helper.EpochingKeeper
		ck := helper.App.CheckpointingKeeper
		queryHelper := baseapp.NewQueryServerTestHelper(helper.Ctx, helper.App.InterfaceRegistry())
		types.RegisterQueryServer(queryHelper, ck)
		msgServer := checkpointingkeeper.NewMsgServerImpl(ck)
		genesisVal := ek.GetValidatorSet(helper.Ctx, 0)[0]
		genesisBLSPubkey, err := ck.GetBlsPubKey(helper.Ctx, genesisVal.Addr)
		require.NoError(t, err)

		// epoch 1 right now
		epoch := ek.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		// 1. get validator BLS set when there's only a genesis validator
		valBlsSet := ck.GetValidatorBlsKeySet(ctx, epoch.EpochNumber)
		require.Equal(t, genesisVal.GetValAddressStr(), valBlsSet.ValSet[0].ValidatorAddress)
		require.True(t, genesisBLSPubkey.Equal(valBlsSet.ValSet[0].BlsPubKey))
		require.Equal(t, uint64(genesisVal.Power), valBlsSet.ValSet[0].VotingPower)

		// add n new validators via MsgWrappedCreateValidator
		n := r.Intn(10) + 1
		addrs, err := app.AddTestAddrs(helper.App, helper.Ctx, n, math.NewInt(100000000))
		require.NoError(t, err)

		wcvMsgs := make([]*types.MsgWrappedCreateValidator, n)
		for i := 0; i < n; i++ {
			msg, err := buildMsgWrappedCreateValidator(addrs[i])
			require.NoError(t, err)
			wcvMsgs[i] = msg
			_, err = msgServer.WrappedCreateValidator(ctx, msg)
			require.NoError(t, err)
		}

		// go to block 11, and thus entering epoch 2
		for i := uint64(0); i < ek.GetParams(ctx).EpochInterval; i++ {
			ctx, err = helper.GenAndApplyEmptyBlock(r)
			require.NoError(t, err)
		}
		epoch = ek.GetEpoch(ctx)
		require.Equal(t, uint64(2), epoch.EpochNumber)

		// 2. get validator BLS set when there are n+1 validators
		epochNum := uint64(2)
		valBlsSet2 := ck.GetValidatorBlsKeySet(ctx, epochNum)
		expectedValSet := ek.GetValidatorSet(ctx, 2)
		for i, expectedVal := range expectedValSet {
			expectedBlsPubkey, err := ck.GetBlsPubKey(ctx, expectedVal.Addr)
			require.NoError(t, err)
			require.Equal(t, expectedVal.GetValAddressStr(), valBlsSet2.ValSet[i].ValidatorAddress)
			require.True(t, expectedBlsPubkey.Equal(valBlsSet2.ValSet[i].BlsPubKey))
			require.Equal(t, uint64(expectedVal.Power), valBlsSet2.ValSet[i].VotingPower)
		}
	})
}
