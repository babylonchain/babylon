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

// GetEpochMsgs returns the set of messages queued of the current epoch
func (k Keeper) GetEpochMsgs(ctx sdk.Context) []types.QueuedMessage {
	panic("TODO: unimplemented")
}

// EnqueueMsg enqueues a message to the queue of the current epoch
func (k Keeper) EnqueueMsg(ctx sdk.Context, msg types.QueuedMessage) error {
	panic("TODO: unimplemented")
}
