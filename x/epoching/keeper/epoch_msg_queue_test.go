package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func FuzzEpochMsgQueue(f *testing.F) {
	f.Add(int64(11111))
	f.Add(int64(22222))
	f.Add(int64(55555))
	f.Add(int64(12312))

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, ctx, keeper, _, _, _ := SetupTestKeeper(t)
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

// TODO (stateful tests): fuzz HandleQueueMsg. initialise some validators, let them submit some msgs and trigger HandleQueueMsg
// require mocking valid QueueMsgs
