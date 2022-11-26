package keeper

import (
	"fmt"

	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// InitMsgQueue initialises the msg queue length of the current epoch to 0
func (k Keeper) InitMsgQueue(ctx sdk.Context) {
	store := k.msgQueueLengthStore(ctx)

	epochNumber := k.GetEpoch(ctx).EpochNumber
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	queueLenBytes := sdk.Uint64ToBigEndian(0)
	store.Set(epochNumberBytes, queueLenBytes)
}

// GetQueueLength fetches the number of queued messages of a given epoch
func (k Keeper) GetQueueLength(ctx sdk.Context, epochNumber uint64) uint64 {
	store := k.msgQueueLengthStore(ctx)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)

	// get queue len in bytes from DB
	bz := store.Get(epochNumberBytes)
	if bz == nil {
		return 0 // BBN has not reached this epoch yet
	}
	// unmarshal
	return sdk.BigEndianToUint64(bz)
}

// GetQueueLength fetches the number of queued messages of the current epoch
func (k Keeper) GetCurrentQueueLength(ctx sdk.Context) uint64 {
	epochNumber := k.GetEpoch(ctx).EpochNumber
	return k.GetQueueLength(ctx, epochNumber)
}

// incCurrentQueueLength adds the queue length of the current epoch by 1
func (k Keeper) incCurrentQueueLength(ctx sdk.Context) {
	store := k.msgQueueLengthStore(ctx)

	epochNumber := k.GetEpoch(ctx).EpochNumber
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)

	queueLen := k.GetQueueLength(ctx, epochNumber)
	incrementedQueueLen := queueLen + 1
	incrementedQueueLenBytes := sdk.Uint64ToBigEndian(incrementedQueueLen)

	store.Set(epochNumberBytes, incrementedQueueLenBytes)
}

// EnqueueMsg enqueues a message to the queue of the current epoch
func (k Keeper) EnqueueMsg(ctx sdk.Context, msg types.QueuedMessage) {
	epochNumber := k.GetEpoch(ctx).EpochNumber
	store := k.msgQueueStore(ctx, epochNumber)

	// key: index, in this case = queueLenBytes
	queueLen := k.GetCurrentQueueLength(ctx)
	queueLenBytes := sdk.Uint64ToBigEndian(queueLen)
	// value: msgBytes
	msgBytes, err := k.cdc.MarshalInterface(&msg)
	if err != nil {
		panic(sdkerrors.Wrap(types.ErrMarshal, err.Error()))
	}
	store.Set(queueLenBytes, msgBytes)

	// increment queue length
	k.incCurrentQueueLength(ctx)
}

// GetEpochMsgs returns the set of messages queued in a given epoch
func (k Keeper) GetEpochMsgs(ctx sdk.Context, epochNumber uint64) []*types.QueuedMessage {
	queuedMsgs := []*types.QueuedMessage{}
	store := k.msgQueueStore(ctx, epochNumber)

	// add each queued msg to queuedMsgs
	iterator := sdk.KVStorePrefixIterator(store, nil)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		queuedMsgBytes := iterator.Value()
		var sdkMsg sdk.Msg
		if err := k.cdc.UnmarshalInterface(queuedMsgBytes, &sdkMsg); err != nil {
			panic(sdkerrors.Wrap(types.ErrUnmarshal, err.Error()))
		}
		queuedMsg, ok := sdkMsg.(*types.QueuedMessage)
		if !ok {
			panic("invalid queued message")
		}
		queuedMsgs = append(queuedMsgs, queuedMsg)
	}

	return queuedMsgs
}

// GetCurrentEpochMsgs returns the set of messages queued in the current epoch
func (k Keeper) GetCurrentEpochMsgs(ctx sdk.Context) []*types.QueuedMessage {
	epochNumber := k.GetEpoch(ctx).EpochNumber
	return k.GetEpochMsgs(ctx, epochNumber)
}

// HandleQueuedMsg unwraps a QueuedMessage and forwards it to the staking module
func (k Keeper) HandleQueuedMsg(ctx sdk.Context, msg *types.QueuedMessage) (*sdk.Result, error) {
	var (
		unwrappedMsgWithType sdk.Msg
		err                  error
	)
	unwrappedMsgWithType = msg.UnwrapToSdkMsg()

	// failed to decode validator address
	if err != nil {
		panic(err)
	}

	// get the handler function from router
	handler := k.router.Handler(unwrappedMsgWithType)

	// Create a new Context based off of the existing Context with a MultiStore branch
	// in case message processing fails. At this point, the MultiStore is a branch of a branch.
	handlerCtx, msCache := cacheTxContext(ctx, msg.TxId, msg.MsgId, msg.BlockHeight)

	// handle the unwrapped message
	result, err := handler(handlerCtx, unwrappedMsgWithType)
	if err != nil {
		return result, err
	}

	// release the cache
	msCache.Write()

	// record lifecycle for delegation
	switch unwrappedMsg := msg.Msg.(type) {
	case *types.QueuedMessage_MsgCreateValidator:
		// handle self-delegation
		delAddr, err := sdk.AccAddressFromBech32(unwrappedMsg.MsgCreateValidator.DelegatorAddress)
		if err != nil {
			return nil, err
		}
		valAddr, err := sdk.ValAddressFromBech32(unwrappedMsg.MsgCreateValidator.ValidatorAddress)
		if err != nil {
			return nil, err
		}
		// self-bonded to the created validator
		k.RecordNewDelegationState(ctx, delAddr, valAddr, types.BondState_CREATED)
		if err != nil {
			return nil, err
		}
		k.RecordNewDelegationState(ctx, delAddr, valAddr, types.BondState_BONDED)
	case *types.QueuedMessage_MsgDelegate:
		delAddr, err := sdk.AccAddressFromBech32(unwrappedMsg.MsgDelegate.DelegatorAddress)
		if err != nil {
			return nil, err
		}
		valAddr, err := sdk.ValAddressFromBech32(unwrappedMsg.MsgDelegate.ValidatorAddress)
		if err != nil {
			return nil, err
		}
		// created and bonded to the validator
		k.RecordNewDelegationState(ctx, delAddr, valAddr, types.BondState_CREATED)
		k.RecordNewDelegationState(ctx, delAddr, valAddr, types.BondState_BONDED)
	case *types.QueuedMessage_MsgUndelegate:
		delAddr, err := sdk.AccAddressFromBech32(unwrappedMsg.MsgUndelegate.DelegatorAddress)
		if err != nil {
			return nil, err
		}
		valAddr, err := sdk.ValAddressFromBech32(unwrappedMsg.MsgUndelegate.ValidatorAddress)
		if err != nil {
			return nil, err
		}
		// unbonding from the validator
		// (in `ApplyMatureUnbonding`) AFTER mature, unbonded from the validator
		k.RecordNewDelegationState(ctx, delAddr, valAddr, types.BondState_UNBONDING)
	case *types.QueuedMessage_MsgBeginRedelegate:
		delAddr, err := sdk.AccAddressFromBech32(unwrappedMsg.MsgBeginRedelegate.DelegatorAddress)
		if err != nil {
			return nil, err
		}
		srcValAddr, err := sdk.ValAddressFromBech32(unwrappedMsg.MsgBeginRedelegate.ValidatorSrcAddress)
		if err != nil {
			return nil, err
		}
		// unbonding from the source validator
		// (in `ApplyMatureUnbonding`) AFTER mature, unbonded from the source validator, created/bonded to the destination validator
		k.RecordNewDelegationState(ctx, delAddr, srcValAddr, types.BondState_UNBONDING)
	default:
		panic(sdkerrors.Wrap(types.ErrInvalidQueuedMessageType, msg.String()))
	}

	return result, nil
}

// based on a function with the same name in `baseapp.go“
func cacheTxContext(ctx sdk.Context, txid []byte, msgid []byte, height uint64) (sdk.Context, sdk.CacheMultiStore) {
	ms := ctx.MultiStore()
	// TODO: https://github.com/cosmos/cosmos-sdk/issues/2824
	msCache := ms.CacheMultiStore()
	if msCache.TracingEnabled() {
		msCache = msCache.SetTracingContext(
			sdk.TraceContext(
				map[string]interface{}{
					"txHash":  fmt.Sprintf("%X", txid),
					"msgHash": fmt.Sprintf("%X", msgid),
					"height":  fmt.Sprintf("%d", height),
				},
			),
		).(sdk.CacheMultiStore)
	}

	return ctx.WithMultiStore(msCache), msCache
}

// msgQueueStore returns the queue of msgs of a given epoch
// prefix: MsgQueueKey || epochNumber
// key: index
// value: msg
func (k Keeper) msgQueueStore(ctx sdk.Context, epochNumber uint64) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	msgQueueStore := prefix.NewStore(store, types.MsgQueueKey)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	return prefix.NewStore(msgQueueStore, epochNumberBytes)
}

// msgQueueLengthStore returns the length of the msg queue of a given epoch
// prefix: QueueLengthKey
// key: epochNumber
// value: queue length
func (k Keeper) msgQueueLengthStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.QueueLengthKey)
}
