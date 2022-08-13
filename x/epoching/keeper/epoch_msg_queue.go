package keeper

import (
	"fmt"
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// InitQueueLength initialises the msg queue length to 0
func (k Keeper) InitQueueLength(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	queueLenBytes := sdk.Uint64ToBigEndian(0)
	store.Set(types.QueueLengthKey, queueLenBytes)
}

// GetQueueLength fetches the number of queued messages
func (k Keeper) GetQueueLength(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	// get queue len in bytes from DB
	bz := store.Get(types.QueueLengthKey)
	if bz == nil {
		panic(types.ErrUnknownQueueLen)
	}
	// unmarshal
	return sdk.BigEndianToUint64(bz)
}

// setQueueLength sets the msg queue length
func (k Keeper) setQueueLength(ctx sdk.Context, queueLen uint64) {
	store := ctx.KVStore(k.storeKey)

	queueLenBytes := sdk.Uint64ToBigEndian(queueLen)
	store.Set(types.QueueLengthKey, queueLenBytes)
}

// incQueueLength adds the queue length by 1
func (k Keeper) incQueueLength(ctx sdk.Context) {
	queueLen := k.GetQueueLength(ctx)
	incrementedQueueLen := queueLen + 1
	k.setQueueLength(ctx, incrementedQueueLen)
}

// EnqueueMsg enqueues a message to the queue of the current epoch
func (k Keeper) EnqueueMsg(ctx sdk.Context, msg types.QueuedMessage) {
	// prefix: QueuedMsgKey
	store := ctx.KVStore(k.storeKey)
	queuedMsgStore := prefix.NewStore(store, types.QueuedMsgKey)

	// key: queueLenBytes
	queueLen := k.GetQueueLength(ctx)
	queueLenBytes := sdk.Uint64ToBigEndian(queueLen)
	// value: msgBytes
	msgBytes, err := k.cdc.MarshalInterface(&msg)
	if err != nil {
		panic(sdkerrors.Wrap(types.ErrMarshal, err.Error()))
	}
	queuedMsgStore.Set(queueLenBytes, msgBytes)

	// increment queue length
	k.incQueueLength(ctx)
}

func (k Keeper) EnqueueGenMsg(ctx sdk.Context, msg types.QueuedMessage) {
	k.EnqueueMsg(ctx, msg)

	epoch := k.GetEpoch(ctx)
	if epoch.EpochNumber == 0 {
		k.HandleQueuedMsgs(ctx, epoch)
	}
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
		var queuedMsg sdk.Msg
		if err := k.cdc.UnmarshalInterface(queuedMsgBytes, &queuedMsg); err != nil {
			panic(sdkerrors.Wrap(types.ErrUnmarshal, err.Error()))
		}
		queuedMsgs = append(queuedMsgs, queuedMsg.(*types.QueuedMessage))
	}

	return queuedMsgs
}

// ClearEpochMsgs removes all messages in the queue
func (k Keeper) ClearEpochMsgs(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	// remove all epoch msgs
	iterator := sdk.KVStorePrefixIterator(store, types.QueuedMsgKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		store.Delete(key)
	}

	// set queue len to zero
	k.setQueueLength(ctx, 0)
}

func (k Keeper) HandleQueuedMsgs(ctx sdk.Context, epoch types.Epoch) {
	// get all msgs in the msg queue
	queuedMsgs := k.GetEpochMsgs(ctx)
	// forward each msg in the msg queue to the right keeper
	for _, msg := range queuedMsgs {
		res, err := k.HandleQueuedMsg(ctx, msg)
		// skip this failed msg and emit and event signalling it
		// we do not panic here as some users may wrap an invalid message
		// (e.g., self-delegate coins more than its balance, wrong coding of addresses, ...)
		// honest validators will have consistent execution results on the queued messages
		if err != nil {
			// emit an event signalling the failed execution
			err := ctx.EventManager().EmitTypedEvent(
				&types.EventHandleQueuedMsg{
					EpochNumber: epoch.EpochNumber,
					TxId:        msg.TxId,
					MsgId:       msg.MsgId,
					Error:       err.Error(),
				},
			)
			if err != nil {
				panic(err)
			}
			// skip this failed msg
			continue
		}
		// for each event, emit an wrapped event EventTypeHandleQueuedMsg, which attaches the original attributes plus the original event type, the epoch number, txid and msgid to the event here
		for _, event := range res.Events {
			err := ctx.EventManager().EmitTypedEvent(
				&types.EventHandleQueuedMsg{
					OriginalEventType:  event.Type,
					EpochNumber:        epoch.EpochNumber,
					TxId:               msg.TxId,
					MsgId:              msg.MsgId,
					OriginalAttributes: event.Attributes,
				},
			)
			if err != nil {
				panic(err)
			}
		}
	}

	// clear the current msg queue
	k.ClearEpochMsgs(ctx)
	// trigger AfterEpochEnds hook
	k.AfterEpochEnds(ctx, epoch.EpochNumber)
	// emit EndEpoch event
	err := ctx.EventManager().EmitTypedEvent(
		&types.EventEndEpoch{
			EpochNumber: epoch.EpochNumber,
		},
	)
	if err != nil {
		panic(err)
	}
}

// HandleQueuedMsg unwraps a QueuedMessage and forwards it to the staking module
func (k Keeper) HandleQueuedMsg(ctx sdk.Context, msg *types.QueuedMessage) (*sdk.Result, error) {
	unwrappedMsgWithType := msg.WithType()

	// get the handler function from router
	handler := k.router.Handler(unwrappedMsgWithType)

	// Create a new Context based off of the existing Context with a MultiStore branch
	// in case message processing fails. At this point, the MultiStore is a branch of a branch.
	handlerCtx, msCache := cacheTxContext(ctx, msg.TxId, msg.MsgId)

	// handle the unwrapped message
	result, err := handler(handlerCtx, unwrappedMsgWithType)

	if err == nil {
		msCache.Write()
	}

	return result, err
}

// based on a function with the same name in `baseapp.go``
func cacheTxContext(ctx sdk.Context, txid []byte, msgid []byte) (sdk.Context, sdk.CacheMultiStore) {
	ms := ctx.MultiStore()
	// TODO: https://github.com/cosmos/cosmos-sdk/issues/2824
	msCache := ms.CacheMultiStore()
	if msCache.TracingEnabled() {
		msCache = msCache.SetTracingContext(
			sdk.TraceContext(
				map[string]interface{}{
					"txHash":  fmt.Sprintf("%X", txid),
					"msgHash": fmt.Sprintf("%X", msgid),
				},
			),
		).(sdk.CacheMultiStore)
	}

	return ctx.WithMultiStore(msCache), msCache
}
