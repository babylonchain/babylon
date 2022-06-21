package keeper

import (
	"fmt"

	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"
)

const (
	DefaultEpochNumber = 0
	DefaultQueueLength = 0
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   sdk.StoreKey
		memKey     sdk.StoreKey
		hooks      types.EpochingHooks
		paramstore paramtypes.Subspace
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
		hooks:      nil,
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
		return sdk.NewUint(uint64(DefaultEpochNumber)), nil
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

func (k Keeper) IncrementEpochNumber(ctx sdk.Context) error {
	epochNumber, err := k.GetEpochNumber(ctx)
	if err != nil {
		return err
	}
	incrementedEpochNumber := epochNumber.AddUint64(1)
	return k.SetQueueLength(ctx, incrementedEpochNumber)
}

func (k Keeper) GetQueueLength(ctx sdk.Context) (sdk.Uint, error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.QueueLengthKey)
	if bz == nil {
		return sdk.NewUint(uint64(DefaultQueueLength)), nil
	}
	var queueLen sdk.Uint
	err := queueLen.Unmarshal(bz)

	return queueLen, err
}

func (k Keeper) SetQueueLength(ctx sdk.Context, queueLen sdk.Uint) error {
	store := ctx.KVStore(k.storeKey)

	queueLenBytes, err := queueLen.Marshal()
	if err != nil {
		return err
	}

	store.Set(types.EpochNumberKey, queueLenBytes)

	return nil
}

func (k Keeper) IncrementQueueLength(ctx sdk.Context) error {
	queueLen, err := k.GetQueueLength(ctx)
	if err != nil {
		return err
	}
	incrementedQueueLen := queueLen.AddUint64(1)
	return k.SetQueueLength(ctx, incrementedQueueLen)
}

// GetEpochMsgs returns the set of messages queued of the current epoch
func (k Keeper) GetEpochMsgs(ctx sdk.Context) ([]types.QueuedMessage, error) {
	queuedMsgs := []types.QueuedMessage{}

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.QueuedMsgKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		queuedMsgBytes := store.Get(key)
		var queuedMsg types.QueuedMessage
		if err := k.cdc.Unmarshal(queuedMsgBytes, &queuedMsg); err != nil {
			return nil, err
		}
		queuedMsgs = append(queuedMsgs, queuedMsg)
	}

	return queuedMsgs, nil
}

// EnqueueMsg enqueues a message to the queue of the current epoch
func (k Keeper) EnqueueMsg(ctx sdk.Context, msg types.QueuedMessage) error {
	store := ctx.KVStore(k.storeKey)

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

	return k.IncrementQueueLength(ctx)
}
