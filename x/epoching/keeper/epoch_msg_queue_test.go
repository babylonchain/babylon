package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/x/epoching/testepoching"
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

var (
	coinWithOnePower = sdk.NewInt64Coin(sdk.DefaultBondDenom, sdk.DefaultPowerReduction.Int64())
)

// FuzzEnqueueMsg tests EnqueueMsg. It enqueues some wrapped msgs, and check if the message queue includes the enqueued msgs or not
func FuzzEnqueueMsg(f *testing.F) {
	f.Add(int64(11111))
	f.Add(int64(22222))
	f.Add(int64(55555))
	f.Add(int64(12312))

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		helper := testepoching.NewHelper(t)
		ctx, keeper := helper.Ctx, helper.EpochingKeeper
		// ensure that the epoch msg queue is correct at the genesis
		require.Empty(t, keeper.GetCurrentEpochMsgs(ctx))
		require.Equal(t, uint64(0), keeper.GetCurrentQueueLength(ctx))

		// enter the 1st block and thus epoch 1
		// Note that the genesis block does not trigger BeginBlock or EndBlock
		ctx = helper.GenAndApplyEmptyBlock()
		epoch := keeper.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)
		// ensure that the epoch msg queue is correct at epoch 1
		require.Empty(t, keeper.GetCurrentEpochMsgs(ctx))
		require.Equal(t, uint64(0), keeper.GetCurrentQueueLength(ctx))

		// Enqueue a random number of msgs
		numQueuedMsgs := rand.Uint64() % 100
		for i := uint64(0); i < numQueuedMsgs; i++ {
			msg := types.QueuedMessage{
				TxId:  sdk.Uint64ToBigEndian(i),
				MsgId: sdk.Uint64ToBigEndian(i),
			}
			keeper.EnqueueMsg(ctx, msg)
		}

		// ensure that each msg in the queue is correct
		epochMsgs := keeper.GetCurrentEpochMsgs(ctx)
		for i, msg := range epochMsgs {
			require.Equal(t, sdk.Uint64ToBigEndian(uint64(i)), msg.TxId)
			require.Equal(t, sdk.Uint64ToBigEndian(uint64(i)), msg.MsgId)
			require.Nil(t, msg.Msg)
		}
	})
}

// FuzzHandleQueuedMsg_MsgWrappedDelegate tests HandleQueueMsg over MsgWrappedDelegate.
// It enqueues some MsgWrappedDelegate, enters a new epoch (which triggers HandleQueueMsg), and check if the newly delegated tokens take effect or not
func FuzzHandleQueuedMsg_MsgWrappedDelegate(f *testing.F) {
	f.Add(int64(11111))
	f.Add(int64(22222))
	f.Add(int64(55555))
	f.Add(int64(12312))

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		helper := testepoching.NewHelperWithValSet(t)
		keeper, genAccs := helper.EpochingKeeper, helper.GenAccs

		// get genesis account's address, whose holder will be the delegator
		require.NotNil(t, genAccs)
		require.NotEmpty(t, genAccs)
		genAddr := genAccs[0].GetAddress()

		// enter the 1st block and thus epoch 1
		// Note that the genesis block does not trigger BeginBlock or EndBlock
		ctx := helper.GenAndApplyEmptyBlock()
		epoch := keeper.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		// get validator to be undelegated
		valSet := keeper.GetCurrentValidatorSet(helper.Ctx)
		val := valSet[0].Addr
		valPower, err := keeper.GetCurrentValidatorVotingPower(ctx, val)
		require.NoError(t, err)

		// delegate a random amount of tokens to the validator
		numNewDels := rand.Int63n(1000) + 1
		for i := int64(0); i < numNewDels; i++ {
			helper.WrappedDelegate(genAddr, val, coinWithOnePower.Amount)
		}
		// ensure the msgs are queued
		epochMsgs := keeper.GetCurrentEpochMsgs(ctx)
		require.Equal(t, numNewDels, int64(len(epochMsgs)))

		// enter epoch 2
		params := keeper.GetParams(ctx)
		for i := uint64(0); i < params.EpochInterval; i++ {
			ctx = helper.GenAndApplyEmptyBlock()
		}
		epoch = keeper.GetEpoch(ctx)
		require.Equal(t, uint64(2), epoch.EpochNumber)

		// ensure epoch 2 has initialised an empty msg queue
		require.Empty(t, keeper.GetCurrentEpochMsgs(ctx))

		// ensure the voting power has been added w.r.t. the newly delegated tokens
		valPower2, err := keeper.GetCurrentValidatorVotingPower(ctx, val)
		require.NoError(t, err)
		addedPower := helper.StakingKeeper.TokensToConsensusPower(ctx, coinWithOnePower.Amount.MulRaw(numNewDels))
		require.Equal(t, valPower+addedPower, valPower2)
	})
}

// FuzzHandleQueuedMsg_MsgWrappedUndelegate tests HandleQueueMsg over MsgWrappedUndelegate.
// It enqueues some MsgWrappedUndelegate, enters a new epoch (which triggers HandleQueueMsg), and check if the tokens become unbonding or not
func FuzzHandleQueuedMsg_MsgWrappedUndelegate(f *testing.F) {
	f.Add(int64(11111))
	f.Add(int64(22222))
	f.Add(int64(55555))
	f.Add(int64(12312))

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		helper := testepoching.NewHelperWithValSet(t)
		ctx, keeper, genAccs := helper.Ctx, helper.EpochingKeeper, helper.GenAccs

		// get genesis account's address, whose holder will be the delegator
		require.NotNil(t, genAccs)
		require.NotEmpty(t, genAccs)
		genAddr := genAccs[0].GetAddress()

		// enter the 1st block and thus epoch 1
		ctx = helper.GenAndApplyEmptyBlock()
		epoch := keeper.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		valSet1 := helper.EpochingKeeper.GetCurrentValidatorSet(helper.Ctx)
		val := valSet1[0].Addr // validator to be undelegated
		valPower, err := helper.EpochingKeeper.GetCurrentValidatorVotingPower(ctx, val)
		require.NoError(t, err)

		// unbond a random amount of tokens from the validator
		// Note that for any pair of delegator and validator, there can be `<=DefaultMaxEntries=7` concurrent undelegations at any time slot
		// Otherwise, only `DefaultMaxEntries` undelegations will be processed at this height and the rest will be rejected
		// See https://github.com/cosmos/cosmos-sdk/blob/v0.45.5/x/staking/keeper/delegation.go#L814-L816
		numNewUndels := rand.Int63n(7) + 1
		for i := int64(0); i < numNewUndels; i++ {
			helper.WrappedUndelegate(genAddr, val, coinWithOnePower.Amount)
		}
		// ensure the msgs are queued
		epochMsgs := keeper.GetCurrentEpochMsgs(ctx)
		require.Equal(t, numNewUndels, int64(len(epochMsgs)))

		// enter epoch 2
		for i := uint64(0); i < keeper.GetParams(ctx).EpochInterval; i++ {
			ctx = helper.GenAndApplyEmptyBlock()
		}
		epoch = keeper.GetEpoch(ctx)
		require.Equal(t, uint64(2), epoch.EpochNumber)

		// ensure epoch 2 has initialised an empty msg queue
		require.Empty(t, keeper.GetCurrentEpochMsgs(ctx))

		// ensure the voting power has been reduced w.r.t. the unbonding tokens
		valPower2, err := helper.EpochingKeeper.GetCurrentValidatorVotingPower(ctx, val)
		require.NoError(t, err)
		reducedPower := helper.StakingKeeper.TokensToConsensusPower(ctx, coinWithOnePower.Amount.MulRaw(numNewUndels))
		require.Equal(t, valPower-reducedPower, valPower2)

		// ensure the genesis account has these unbonding tokens
		unbondingDels := helper.StakingKeeper.GetAllUnbondingDelegations(ctx, genAddr)
		require.Equal(t, 1, len(unbondingDels))                              // there is only 1 type of tokens
		require.Equal(t, numNewUndels, int64(len(unbondingDels[0].Entries))) // there are numNewUndels entries
		for _, entry := range unbondingDels[0].Entries {
			require.Equal(t, coinWithOnePower.Amount, entry.Balance) // each unbonding delegation entry has tokens of 1 voting power
		}
	})
}

// FuzzHandleQueuedMsg_MsgWrappedBeginRedelegate tests HandleQueueMsg over MsgWrappedBeginRedelegate.
// It enqueues some MsgWrappedBeginRedelegate, enters a new epoch (which triggers HandleQueueMsg), and check if the tokens become unbonding or not
func FuzzHandleQueuedMsg_MsgWrappedBeginRedelegate(f *testing.F) {
	f.Add(int64(11111))
	f.Add(int64(22222))
	f.Add(int64(55555))
	f.Add(int64(12312))

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		helper := testepoching.NewHelperWithValSet(t)
		ctx, keeper, genAccs := helper.Ctx, helper.EpochingKeeper, helper.GenAccs

		// get genesis account's address, whose holder will be the delegator
		require.NotNil(t, genAccs)
		require.NotEmpty(t, genAccs)
		genAddr := genAccs[0].GetAddress()

		// enter the 1st block and thus epoch 1
		ctx = helper.GenAndApplyEmptyBlock()
		epoch := keeper.GetEpoch(ctx)
		require.Equal(t, uint64(1), epoch.EpochNumber)

		valSet1 := helper.EpochingKeeper.GetCurrentValidatorSet(ctx)

		// 2 validators, where the operator will redelegate some token from val1 to val2
		val1 := valSet1[0].Addr
		val1Power, err := helper.EpochingKeeper.GetCurrentValidatorVotingPower(ctx, val1)
		require.NoError(t, err)
		val2 := valSet1[1].Addr
		val2Power, err := helper.EpochingKeeper.GetCurrentValidatorVotingPower(ctx, val2)
		require.NoError(t, err)
		require.Equal(t, val1Power, val2Power)

		// redelegate a random amount of tokens from val1 to val2
		// same as undelegation, there can be `<=DefaultMaxEntries=7` concurrent redelegation requests for any tuple (delegatorAddr, srcValidatorAddr, dstValidatorAddr)
		// See https://github.com/cosmos/cosmos-sdk/blob/v0.45.5/x/staking/keeper/delegation.go#L908-L910
		numNewRedels := rand.Int63n(7) + 1
		for i := int64(0); i < numNewRedels; i++ {
			helper.WrappedBeginRedelegate(genAddr, val1, val2, coinWithOnePower.Amount)
		}
		// ensure the msgs are queued
		epochMsgs := keeper.GetCurrentEpochMsgs(ctx)
		require.Equal(t, numNewRedels, int64(len(epochMsgs)))

		// enter epoch 2
		for i := uint64(0); i < keeper.GetParams(ctx).EpochInterval; i++ {
			ctx = helper.GenAndApplyEmptyBlock()
		}
		epoch = keeper.GetEpoch(ctx)
		require.Equal(t, uint64(2), epoch.EpochNumber)

		// ensure epoch 2 has initialised an empty msg queue
		require.Empty(t, keeper.GetCurrentEpochMsgs(ctx))

		// ensure the voting power has been redelegated from val1 to val2
		// Note that in Cosmos SDK, redelegation happens upon the next `EndBlock`, rather than waiting for 14 days.
		// This is because redelegation does not affect PoS security: upon redelegation requests, no token is leaving the system.
		// SImilarly, in Babylon, redelegation happens unconditionally upon `EndEpoch`, rather than depending on checkpoint status.
		val1Power2, err := helper.EpochingKeeper.GetCurrentValidatorVotingPower(ctx, val1)
		require.NoError(t, err)
		val2Power2, err := helper.EpochingKeeper.GetCurrentValidatorVotingPower(ctx, val2)
		require.NoError(t, err)
		redelegatedPower := helper.StakingKeeper.TokensToConsensusPower(ctx, coinWithOnePower.Amount.MulRaw(numNewRedels))
		// ensure the voting power of val1 has reduced
		require.Equal(t, val1Power-redelegatedPower, val1Power2)
		// ensure the voting power of val2 has increased
		require.Equal(t, val2Power+redelegatedPower, val2Power2)

		// ensure the genesis account has these redelegating tokens
		redelegations := helper.StakingKeeper.GetAllRedelegations(ctx, genAddr, val1, val2)
		require.Equal(t, 1, len(redelegations))                              // there is only 1 type of tokens
		require.Equal(t, numNewRedels, int64(len(redelegations[0].Entries))) // there are numNewRedels entries
		for _, entry := range redelegations[0].Entries {
			require.Equal(t, coinWithOnePower.Amount, entry.InitialBalance) // each redelegating entry has tokens of 1 voting power
		}
	})
}
