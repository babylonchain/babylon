package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetQueueLength fetches the number of queued messages
func (k Keeper) GetQueueLength(ctx sdk.Context) (sdk.Uint, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.QueueLengthKey)
	if bz == nil {
		return sdk.NewUint(0), nil
	}
	var queueLen sdk.Uint
	err := queueLen.Unmarshal(bz)

	return queueLen, err
}

// setQueueLength sets the msg queue length
func (k Keeper) setQueueLength(ctx sdk.Context, queueLen sdk.Uint) error {
	store := ctx.KVStore(k.storeKey)

	queueLenBytes, err := queueLen.Marshal()
	if err != nil {
		return err
	}

	store.Set(types.QueueLengthKey, queueLenBytes)

	return nil
}

// incQueueLength adds the queue length by 1
func (k Keeper) incQueueLength(ctx sdk.Context) error {
	queueLen, err := k.GetQueueLength(ctx)
	if err != nil {
		return err
	}
	incrementedQueueLen := queueLen.AddUint64(1)
	return k.setQueueLength(ctx, incrementedQueueLen)
}

// EnqueueMsg enqueues a message to the queue of the current epoch
func (k Keeper) EnqueueMsg(ctx sdk.Context, msg types.QueuedMessage) error {
	store := ctx.KVStore(k.storeKey)

	// insert KV pair, where
	// - key: QueuedMsgKey || queueLenBytes
	// - value: msgBytes
	queueLen, err := k.GetQueueLength(ctx)
	if err != nil {
		return err
	}
	queueLenBytes, err := queueLen.Marshal()
	if err != nil {
		return err
	}
	msgBytes, err := k.cdc.Marshal(&msg)
	if err != nil {
		return err
	}
	store.Set(append(types.QueuedMsgKey, queueLenBytes...), msgBytes)

	// increment queue length
	return k.incQueueLength(ctx)
}

// GetEpochMsgs returns the set of messages queued in the current epoch
func (k Keeper) GetEpochMsgs(ctx sdk.Context) ([]*types.QueuedMessage, error) {
	queuedMsgs := []*types.QueuedMessage{}
	store := ctx.KVStore(k.storeKey)

	// add each queued msg to queuedMsgs
	iterator := sdk.KVStorePrefixIterator(store, types.QueuedMsgKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		queuedMsgBytes := iterator.Value()
		var queuedMsg types.QueuedMessage
		if err := k.cdc.Unmarshal(queuedMsgBytes, &queuedMsg); err != nil {
			return nil, err
		}
		queuedMsgs = append(queuedMsgs, &queuedMsg)
	}

	return queuedMsgs, nil
}

// ClearEpochMsgs removes all messages in the queue
func (k Keeper) ClearEpochMsgs(ctx sdk.Context) error {
	store := ctx.KVStore(k.storeKey)

	// remove all epoch msgs
	iterator := sdk.KVStorePrefixIterator(store, types.QueuedMsgKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		store.Delete(key)
	}

	// set queue len to zero
	return k.setQueueLength(ctx, sdk.NewUint(0))
}
