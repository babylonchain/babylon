package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
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
func (k *Keeper) SetHooks(sh types.EpochingHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set validator hooks twice")
	}

	k.hooks = sh

	return k
}

// GetCurrentEpoch returns the current epoch number
func (k Keeper) GetCurrentEpoch(ctx sdk.Context) sdk.Uint {
	panic("TODO: unimplemented")
}

// GetEpochMsgs returns the set of messages queued of the current epoch
func (k Keeper) GetEpochMsgs(ctx sdk.Context) []types.QueuedMessage {
	panic("TODO: unimplemented")
}

// EnqueueMsg enqueues a message to the queue of the current epoch
func (k Keeper) EnqueueMsg(ctx sdk.Context, msg types.QueuedMessage) error {
	panic("TODO: unimplemented")
}
