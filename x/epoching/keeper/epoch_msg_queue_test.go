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
		require.Empty(t, keeper.GetEpochMsgs(ctx))
		require.Equal(t, uint64(0), keeper.GetQueueLength(ctx))

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
		epochMsgs := keeper.GetEpochMsgs(ctx)
		for i, msg := range epochMsgs {
			require.Equal(t, sdk.Uint64ToBigEndian(uint64(i)), msg.TxId)
			require.Equal(t, sdk.Uint64ToBigEndian(uint64(i)), msg.MsgId)
			require.Nil(t, msg.Msg)
		}

		// after clearing the msg queue, ensure that the epoch msg queue is empty
		keeper.ClearEpochMsgs(ctx)
		require.Empty(t, keeper.GetEpochMsgs(ctx))
		require.Equal(t, uint64(0), keeper.GetQueueLength(ctx))
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
		ctx, keeper, genAccs := helper.Ctx, helper.EpochingKeeper, helper.GenAccs
		valSet0 := helper.EpochingKeeper.GetValidatorSet(helper.Ctx, 0)

		// validator to be delegated
		val := valSet0[0].Addr
		valPower, err := helper.EpochingKeeper.GetValidatorVotingPower(ctx, 0, val)
		require.NoError(t, err)

		// get genesis account's address, whose holder will be the delegator
		require.NotNil(t, genAccs)
		require.NotEmpty(t, genAccs)
		genAddr := genAccs[0].GetAddress()

		// delegate a random amount of tokens to the validator
		numNewVals := rand.Int63n(1000) + 1
		for i := int64(0); i < numNewVals; i++ {
			helper.WrappedDelegate(genAddr, val, coinWithOnePower.Amount)
		}
		// ensure the msgs are queued
		epochMsgs := keeper.GetEpochMsgs(ctx)
		require.Equal(t, numNewVals, int64(len(epochMsgs)))

		// enter the 1st block and thus epoch 1
		// Note that we missed epoch 0's BeginBlock/EndBlock and thus EpochMsgs are not handled
		ctx = helper.GenAndApplyEmptyBlock()
		// enter epoch 2
		for i := uint64(0); i < keeper.GetParams(ctx).EpochInterval; i++ {
			ctx = helper.GenAndApplyEmptyBlock()
		}

		// ensure queued msgs have been handled
		queueLen := keeper.GetQueueLength(ctx)
		require.Equal(t, uint64(0), queueLen)
		epochMsgs = keeper.GetEpochMsgs(ctx)
		require.Equal(t, 0, len(epochMsgs))

		// ensure the voting power has been added w.r.t. the newly delegated tokens
		valPower2, err := helper.EpochingKeeper.GetValidatorVotingPower(ctx, 2, val)
		require.NoError(t, err)
		addedPower := helper.StakingKeeper.TokensToConsensusPower(ctx, coinWithOnePower.Amount.MulRaw(numNewVals))
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
		valSet0 := helper.EpochingKeeper.GetValidatorSet(helper.Ctx, 0)

		// validator to be undelegated
		val := valSet0[0].Addr
		valPower, err := helper.EpochingKeeper.GetValidatorVotingPower(ctx, 0, val)
		require.NoError(t, err)

		// get genesis account's address, whose holder will be the delegator
		require.NotNil(t, genAccs)
		require.NotEmpty(t, genAccs)
		genAddr := genAccs[0].GetAddress()

		// unbond a random amount of tokens from the validator
		numNewVals := rand.Int63n(7) + 1 // numNewVals \in [1, 7] since UBD queue contains at most DefaultMaxEntries=7 validators
		for i := int64(0); i < numNewVals; i++ {
			helper.WrappedUndelegate(genAddr, val, coinWithOnePower.Amount)
		}
		// ensure the msgs are queued
		epochMsgs := keeper.GetEpochMsgs(ctx)
		require.Equal(t, numNewVals, int64(len(epochMsgs)))

		// enter the 1st block and thus epoch 1
		// Note that we missed epoch 0's BeginBlock/EndBlock and thus EpochMsgs are not handled
		ctx = helper.GenAndApplyEmptyBlock()
		// enter epoch 2
		for i := uint64(0); i < keeper.GetParams(ctx).EpochInterval; i++ {
			ctx = helper.GenAndApplyEmptyBlock()
		}

		// ensure queued msgs have been handled
		queueLen := keeper.GetQueueLength(ctx)
		require.Equal(t, uint64(0), queueLen)
		epochMsgs = keeper.GetEpochMsgs(ctx)
		require.Equal(t, 0, len(epochMsgs))

		// ensure the voting power has been reduced w.r.t. the unbonding tokens
		valPower2, err := helper.EpochingKeeper.GetValidatorVotingPower(ctx, 2, val)
		require.NoError(t, err)
		reducedPower := helper.StakingKeeper.TokensToConsensusPower(ctx, coinWithOnePower.Amount.MulRaw(numNewVals))
		require.Equal(t, valPower-reducedPower, valPower2)

		// ensure the genesis account has these unbonding tokens
		unbondingDels := helper.StakingKeeper.GetAllUnbondingDelegations(ctx, genAddr)
		require.Equal(t, 1, len(unbondingDels))                            // there is only 1 type of tokens
		require.Equal(t, numNewVals, int64(len(unbondingDels[0].Entries))) // there are numNewVals entries
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
		valSet0 := helper.EpochingKeeper.GetValidatorSet(helper.Ctx, 0)

		// 2 validators, where the operator will redelegate some token from val1 to val2
		val1 := valSet0[0].Addr
		val1Power, err := helper.EpochingKeeper.GetValidatorVotingPower(ctx, 0, val1)
		require.NoError(t, err)
		val2 := valSet0[1].Addr
		val2Power, err := helper.EpochingKeeper.GetValidatorVotingPower(ctx, 0, val2)
		require.NoError(t, err)
		require.Equal(t, val1Power, val2Power)

		// get genesis account's address, whose holder will be the delegator
		require.NotNil(t, genAccs)
		require.NotEmpty(t, genAccs)
		genAddr := genAccs[0].GetAddress()

		// redelegate a random amount of tokens from val1 to val2
		numNewVals := rand.Int63n(7) + 1 // numNewVals \in [1, 7] since UBD queue contains at most DefaultMaxEntries=7 validators
		for i := int64(0); i < numNewVals; i++ {
			helper.WrappedBeginRedelegate(genAddr, val1, val2, coinWithOnePower.Amount)
		}
		// ensure the msgs are queued
		epochMsgs := keeper.GetEpochMsgs(ctx)
		require.Equal(t, numNewVals, int64(len(epochMsgs)))

		// enter the 1st block and thus epoch 1
		// Note that we missed epoch 0's BeginBlock/EndBlock and thus EpochMsgs are not handled
		ctx = helper.GenAndApplyEmptyBlock()
		// enter epoch 2
		for i := uint64(0); i < keeper.GetParams(ctx).EpochInterval; i++ {
			ctx = helper.GenAndApplyEmptyBlock()
		}

		// ensure queued msgs have been handled
		queueLen := keeper.GetQueueLength(ctx)
		require.Equal(t, uint64(0), queueLen)
		epochMsgs = keeper.GetEpochMsgs(ctx)
		require.Equal(t, 0, len(epochMsgs))

		// ensure the voting power has been redelegated from val1 to val2
		// Note that in Babylon, redelegation happens unconditionally upon `EndEpoch`, rather than upon checkpointed. Meanwhile in Cosmos SDK, redelegation happens upon `EndBlock`.
		// This is because slashable security only requires `unbonding` -> `unbonded` to depend on checkpoints, and redelegation does not unbond any stake from the system.
		val1Power2, err := helper.EpochingKeeper.GetValidatorVotingPower(ctx, 2, val1)
		require.NoError(t, err)
		val2Power2, err := helper.EpochingKeeper.GetValidatorVotingPower(ctx, 2, val2)
		require.NoError(t, err)
		redelegatedPower := helper.StakingKeeper.TokensToConsensusPower(ctx, coinWithOnePower.Amount.MulRaw(numNewVals))
		// ensure the voting power of val1 has reduced
		require.Equal(t, val1Power-redelegatedPower, val1Power2)
		// ensure the voting power of val2 has increased
		require.Equal(t, val2Power+redelegatedPower, val2Power2)

		// ensure the genesis account has these redelegating tokens
		redelegations := helper.StakingKeeper.GetAllRedelegations(ctx, genAddr, val1, val2)
		require.Equal(t, 1, len(redelegations))                            // there is only 1 type of tokens
		require.Equal(t, numNewVals, int64(len(redelegations[0].Entries))) // there are numNewVals entries
		for _, entry := range redelegations[0].Entries {
			require.Equal(t, coinWithOnePower.Amount, entry.InitialBalance) // each redelegating entry has tokens of 1 voting power
		}
	})
}
