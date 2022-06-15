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
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetCurrentEpoch returns the current epoch number
func (k Keeper) GetCurrentEpoch(ctx sdk.Context) sdk.Uint {
	panic("TODO: unimplemented")
}

// GetEpochMsgs returns the set of messages queued of a given uncheckpointed epoch
func (k Keeper) GetEpochMsgs(ctx sdk.Context, epoch sdk.Uint) []sdk.Msg {
	panic("TODO: unimplemented")
}

// EnqueueMsg enqueues a message to the queue of the current epoch
func (k Keeper) EnqueueMsg(ctx sdk.Context, msg sdk.Msg) error {
	panic("TODO: unimplemented")
}
