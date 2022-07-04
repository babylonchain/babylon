package keeper

import (
	"fmt"

	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		cdc              codec.BinaryCodec
		storeKey         sdk.StoreKey
		memKey           sdk.StoreKey
		hooks            types.EpochingHooks
		paramstore       paramtypes.Subspace
		stk              types.StakingKeeper
		stakingMsgServer types.StakingMsgServer
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	stk types.StakingKeeper,
	stakingMsgServer types.StakingMsgServer,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		memKey:           memKey,
		paramstore:       ps,
		hooks:            nil,
		stk:              stk,
		stakingMsgServer: stakingMsgServer,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Set the validator hooks
func (k *Keeper) SetHooks(eh types.EpochingHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set validator hooks twice")
	}

	k.hooks = eh

	return k
}

// GetEpochNumber fetches epoch number
func (k Keeper) GetEpochNumber(ctx sdk.Context) sdk.Uint {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.EpochNumberKey)
	if bz == nil {
		panic(types.ErrUnknownEpochNumber)
	}
	var epochNumber sdk.Uint
	if err := epochNumber.Unmarshal(bz); err != nil {
		panic(err)
	}

	return epochNumber
}

// SetEpochNumber sets epoch number
func (k Keeper) SetEpochNumber(ctx sdk.Context, epochNumber sdk.Uint) {
	store := ctx.KVStore(k.storeKey)

	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.EpochNumberKey, epochNumberBytes)
}

// IncEpochNumber adds epoch number by 1
func (k Keeper) IncEpochNumber(ctx sdk.Context) sdk.Uint {
	epochNumber := k.GetEpochNumber(ctx)
	incrementedEpochNumber := epochNumber.AddUint64(1)
	k.SetEpochNumber(ctx, incrementedEpochNumber)
	return incrementedEpochNumber
}

// GetEpochBoundary gets the epoch boundary, i.e., the height of the block that ends this epoch
// example: in epoch 1, epoch interval is 5 blocks, boundary will be 1*5=5
// 0 | 1 2 3 4 5 | 6 7 8 9 10 |
// 0 |     1     |     2      |
func (k Keeper) GetEpochBoundary(ctx sdk.Context) sdk.Uint {
	epochNumber := k.GetEpochNumber(ctx)
	// epoch number is 0 at the 0-th block, i.e., genesis
	if epochNumber.IsZero() {
		return sdk.NewUint(0)
	}
	// case when epoch number > 0
	epochInterval := sdk.NewUint(k.GetParams(ctx).EpochInterval)
	return epochNumber.Mul(epochInterval)
}

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
// TODO: after we bump to Cosmos SDK v0.46, add MsgCancelUnbondingDelegation
func (k Keeper) HandleQueuedMsg(ctx sdk.Context, msg *types.QueuedMessage) {
	switch unwrappedMsg := msg.Msg.(type) {
	case *types.QueuedMessage_MsgCreateValidator:
		unwrappedMsgWithType := unwrappedMsg.MsgCreateValidator
		if _, err := k.stakingMsgServer.CreateValidator(sdk.WrapSDKContext(ctx), unwrappedMsgWithType); err != nil {
			panic(err)
		}
	case *types.QueuedMessage_MsgDelegate:
		unwrappedMsgWithType := unwrappedMsg.MsgDelegate
		if _, err := k.stakingMsgServer.Delegate(sdk.WrapSDKContext(ctx), unwrappedMsgWithType); err != nil {
			panic(err)
		}
	case *types.QueuedMessage_MsgUndelegate:
		unwrappedMsgWithType := unwrappedMsg.MsgUndelegate
		if _, err := k.stakingMsgServer.Undelegate(sdk.WrapSDKContext(ctx), unwrappedMsgWithType); err != nil {
			panic(err)
		}
	case *types.QueuedMessage_MsgBeginRedelegate:
		unwrappedMsgWithType := unwrappedMsg.MsgBeginRedelegate
		if _, err := k.stakingMsgServer.BeginRedelegate(sdk.WrapSDKContext(ctx), unwrappedMsgWithType); err != nil {
			panic(err)
		}
	default:
		panic(sdkerrors.Wrap(types.ErrInvalidQueuedMessageType, msg.String()))
	}
}
