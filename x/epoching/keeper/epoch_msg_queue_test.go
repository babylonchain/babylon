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
	coin100 = sdk.NewInt64Coin(sdk.DefaultBondDenom, 100)
	coin50  = sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)
	// val1    = sdk.ValAddress("_____validator1_____")
	// val2    = sdk.ValAddress("_____validator2_____")
	// val3    = sdk.ValAddress("_____validator3_____")
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

// FuzzHandleQueueMsg tests HandleQueueMsg over MsgWrappedDelegate.
// It enqueues some MsgWrappedDelegate, enters a new epoch (which triggers HandleQueueMsg), and check if the message queue is consistent or not, and if the queue msgs are processed or not.
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

		// generate a random number of validators
		numNewVals := rand.Intn(1000)
		for i := 0; i < numNewVals; i++ {
			helper.WrappedDelegate(genAddr, val, coin50.Amount)
		}
		// ensure the msgs are queued
		epochMsgs := keeper.GetEpochMsgs(ctx)
		require.Equal(t, numNewVals, len(epochMsgs))

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
		addedPoWer := helper.StakingKeeper.TokensToConsensusPower(ctx, sdk.NewInt(int64(numNewVals*50)))
		require.Equal(t, valPower+addedPoWer, valPower2)
	})
}
