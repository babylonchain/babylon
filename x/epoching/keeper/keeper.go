package keeper

import (
	"fmt"

	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		StakingMsgServer types.StakingMsgServer
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
		StakingMsgServer: stakingMsgServer,
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
func (k Keeper) GetEpochNumber(ctx sdk.Context) (sdk.Uint, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.EpochNumberKey)
	if bz == nil {
		return sdk.NewUint(0), types.ErrUnknownEpochNumber
	}
	var epochNumber sdk.Uint
	err := epochNumber.Unmarshal(bz)

	return epochNumber, err
}

// SetEpochNumber sets epoch number
func (k Keeper) SetEpochNumber(ctx sdk.Context, epochNumber sdk.Uint) error {
	store := ctx.KVStore(k.storeKey)

	epochNumberBytes, err := epochNumber.Marshal()
	if err != nil {
		return err
	}

	store.Set(types.EpochNumberKey, epochNumberBytes)

	return nil
}

// IncEpochNumber adds epoch number by 1
func (k Keeper) IncEpochNumber(ctx sdk.Context) (sdk.Uint, error) {
	epochNumber, err := k.GetEpochNumber(ctx)
	if err != nil {
		return sdk.NewUint(0), err
	}
	incrementedEpochNumber := epochNumber.AddUint64(1)
	return incrementedEpochNumber, k.SetEpochNumber(ctx, incrementedEpochNumber)
}

// GetEpochBoundary gets the epoch boundary, i.e., the height of the block that ends this epoch
func (k Keeper) GetEpochBoundary(ctx sdk.Context) (sdk.Uint, error) {
	epochNumber, err := k.GetEpochNumber(ctx)
	if err != nil {
		return sdk.NewUint(0), err
	}
	// epoch number is 0 at the 0-th block, i.e., genesis
	if epochNumber.IsZero() {
		return sdk.NewUint(0), nil
	}
	epochInterval := sdk.NewUint(k.GetParams(ctx).EpochInterval)
	// example: in epoch 1, epoch interval is 5 blocks, boundary will be 1*5=5
	// 0 | 1 2 3 4 5 | 6 7 8 9 10 |
	// 0 |     1     |     2      |
	return epochNumber.Mul(epochInterval), nil
}

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
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.QueuedMsgKey)
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		store.Delete(key)
	}

	// set queue len to zero
	return k.setQueueLength(ctx, sdk.NewUint(0))
}
