package keeper

import (
	"fmt"

	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
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
	msgBytes, err := k.cdc.Marshal(&msg)
	if err != nil {
		panic(sdkerrors.Wrap(types.ErrMarshal, err.Error()))
	}
	queuedMsgStore.Set(queueLenBytes, msgBytes)

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
			panic(sdkerrors.Wrap(types.ErrUnmarshal, err.Error()))
		}
		queuedMsgs = append(queuedMsgs, &queuedMsg)
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

// HandleQueuedMsg unwraps a QueuedMessage and forwards it to the staking module
func (k Keeper) HandleQueuedMsg(ctx sdk.Context, msg *types.QueuedMessage) (*sdk.Result, error) {
	var unwrappedMsgWithType sdk.Msg
	// TODO (non-urgent): after we bump to Cosmos SDK v0.46, add MsgCancelUnbondingDelegation
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

	fmt.Println("DISPATCHING", sdktypes.MsgTypeURL(unwrappedMsgWithType), unwrappedMsgWithType)

	switch unwrappedMsg := msg.Msg.(type) {
	case *types.QueuedMessage_MsgDelegate:
		k.printDelegations(ctx, unwrappedMsg.MsgDelegate.ValidatorAddress)
		k.checkHasDelegatorStartingInfo("PRE DELEG", ctx, unwrappedMsg.MsgDelegate.ValidatorAddress, unwrappedMsg.MsgDelegate.DelegatorAddress)
	case *types.QueuedMessage_MsgUndelegate:
		k.checkHasDelegatorStartingInfo("PRE UNDELEG", ctx, unwrappedMsg.MsgUndelegate.ValidatorAddress, unwrappedMsg.MsgUndelegate.DelegatorAddress)
	case *types.QueuedMessage_MsgBeginRedelegate:
		k.checkHasDelegatorStartingInfo("PRE REDELEG", ctx, unwrappedMsg.MsgBeginRedelegate.ValidatorSrcAddress, unwrappedMsg.MsgBeginRedelegate.DelegatorAddress)
	default:

	}

	// get the handler function from router
	handler := k.router.Handler(unwrappedMsgWithType)
	// handle the unwrapped message
	result, err := handler(ctx, unwrappedMsgWithType)

	if err != nil {
		fmt.Println("DISPATCH FAILED", sdktypes.MsgTypeURL(unwrappedMsgWithType), err)
	}
	switch unwrappedMsg := msg.Msg.(type) {
	case *types.QueuedMessage_MsgDelegate:
		k.checkHasDelegatorStartingInfo("POST DELEG", ctx, unwrappedMsg.MsgDelegate.ValidatorAddress, unwrappedMsg.MsgDelegate.DelegatorAddress)
		k.printDelegations(ctx, unwrappedMsg.MsgDelegate.ValidatorAddress)

	case *types.QueuedMessage_MsgBeginRedelegate:
		k.checkHasDelegatorStartingInfo("POST REDELEG", ctx, unwrappedMsg.MsgBeginRedelegate.ValidatorDstAddress, unwrappedMsg.MsgBeginRedelegate.DelegatorAddress)
		k.printDelegations(ctx, unwrappedMsg.MsgBeginRedelegate.ValidatorSrcAddress)
	default:
	}

	return result, err
}

func (k Keeper) checkHasDelegatorStartingInfo(lbl string, ctx sdk.Context, val string, del string) {
	valAddr, _ := sdk.ValAddressFromBech32(val)
	delAddr, _ := sdk.AccAddressFromBech32(del)
	if !k.distr.HasDelegatorStartingInfo(ctx, valAddr, delAddr) {
		fmt.Println(lbl, ": ", "NO DELEGATION START INFO", " del: ", del, " val: ", val)
	} else {
		fmt.Println(lbl, ": ", "HAS DELEGATION START INFO ", " del: ", del, " val: ", val)
	}
}

func (k Keeper) printDelegations(ctx sdk.Context, val string) {
	valAddr, _ := sdk.ValAddressFromBech32(val)
	srcVal, _ := k.stk.GetValidator(ctx, valAddr)
	srcAddr := srcVal.GetOperator()
	delegations := k.stk.GetValidatorDelegations(ctx, srcAddr)
	fmt.Println("DELEGATING TO ", val, ": ", len(delegations))
	for i := range delegations {
		fmt.Println("DELEGATION: ", delegations[i])
	}
}
