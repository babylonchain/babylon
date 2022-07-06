package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// GetQueueLength fetches the number of queued messages
func (k Keeper) GetQueueLength(ctx sdk.Context) sdk.Uint {
	store := ctx.KVStore(k.storeKey)

	// get queue len in bytes from DB
	bz := store.Get(types.QueueLengthKey)
	if bz == nil {
		panic(types.ErrUnknownQueueLen)
	}
	// unmarshal
	var queueLen sdk.Uint
	if err := queueLen.Unmarshal(bz); err != nil {
		panic(err)
	}

	return queueLen
}

// SetQueueLength sets the msg queue length
func (k Keeper) SetQueueLength(ctx sdk.Context, queueLen sdk.Uint) {
	store := ctx.KVStore(k.storeKey)

	queueLenBytes, err := queueLen.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.QueueLengthKey, queueLenBytes)
}

// incQueueLength adds the queue length by 1
func (k Keeper) incQueueLength(ctx sdk.Context) {
	queueLen := k.GetQueueLength(ctx)
	incrementedQueueLen := queueLen.AddUint64(1)
	k.SetQueueLength(ctx, incrementedQueueLen)
}

// EnqueueMsg enqueues a message to the queue of the current epoch
func (k Keeper) EnqueueMsg(ctx sdk.Context, msg types.QueuedMessage) {
	store := ctx.KVStore(k.storeKey)

	// insert KV pair, where
	// - key: QueuedMsgKey || queueLenBytes
	// - value: msgBytes
	queueLen := k.GetQueueLength(ctx)
	queueLenBytes, err := queueLen.Marshal()
	if err != nil {
		panic(err)
	}
	msgBytes, err := k.cdc.Marshal(&msg)
	if err != nil {
		panic(err)
	}
	store.Set(append(types.QueuedMsgKey, queueLenBytes...), msgBytes)

	// increment queue length
	k.incQueueLength(ctx)
}

// GetEpochMsgs returns the set of messages queued in the current epoch
func (k Keeper) GetEpochMsgs(ctx sdk.Context) []*types.QueuedMessage {
	queuedMsgs := []*types.QueuedMessage{}
	store := ctx.KVStore(k.storeKey)

	// add each queued msg to queuedMsgs
	iterator := sdk.KVStorePrefixIterator(store, types.QueuedMsgKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		queuedMsgBytes := iterator.Value()
		var queuedMsg types.QueuedMessage
		if err := k.cdc.Unmarshal(queuedMsgBytes, &queuedMsg); err != nil {
			panic(err)
		}
		queuedMsgs = append(queuedMsgs, &queuedMsg)
	}

	return queuedMsgs
}

// ClearEpochMsgs removes all messages in the queue
func (k Keeper) ClearEpochMsgs(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	// remove all epoch msgs
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.QueuedMsgKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		store.Delete(key)
	}

	// set queue len to zero
	k.SetQueueLength(ctx, sdk.NewUint(0))
}

// HandleQueuedMsg unwraps a QueuedMessage and forwards it to the staking module
func (k Keeper) HandleQueuedMsg(ctx sdk.Context, msg *types.QueuedMessage) (*sdk.Result, error) {
	var unwrappedMsgWithType sdk.Msg
	// TODO: after we bump to Cosmos SDK v0.46, add MsgCancelUnbondingDelegation
	switch unwrappedMsg := msg.Msg.(type) {
	case *types.QueuedMessage_MsgCreateValidator:
		unwrappedMsgWithType = unwrappedMsg.MsgCreateValidator
	case *types.QueuedMessage_MsgDelegate:
		unwrappedMsgWithType = unwrappedMsg.MsgDelegate
	case *types.QueuedMessage_MsgUndelegate:
		unwrappedMsgWithType = unwrappedMsg.MsgUndelegate
	case *types.QueuedMessage_MsgBeginRedelegate:
		unwrappedMsgWithType = unwrappedMsg.MsgBeginRedelegate
	default:
		panic(sdkerrors.Wrap(types.ErrInvalidQueuedMessageType, msg.String()))
	}

	// get the handler function from router
	handler := k.router.Handler(unwrappedMsgWithType)
	// handle the unwrapped message
	return handler(ctx, unwrappedMsgWithType)
}
