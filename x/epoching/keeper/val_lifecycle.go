package keeper

import (
	"fmt"

	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: test the lifecycle functionality

// called upon receiving MsgWrappedCreateValidator
func (k Keeper) InitValState(ctx sdk.Context, valAddr sdk.ValAddress) {

	lc := types.ValidatorLifecycle{
		ValAddr:             valAddr.String(), // bech32-encoded string
		CreateRequestHeight: uint64(ctx.BlockHeight()),
	}
	k.setValLifecycle(ctx, valAddr, &lc)
}

func (k Keeper) UpdateValState(ctx sdk.Context, valAddr sdk.ValAddress, state types.ValState) {
	lc := k.getValLifecycle(ctx, valAddr)
	switch state {
	case types.ValStateCreateRequestSubmitted:
		panic(fmt.Errorf("call InitValState instead"))
	case types.ValStateBonded:
		lc.BondedHeight = uint64(ctx.BlockHeight())
	case types.ValStateUnbondingRequestSubmitted:
		lc.UnbondingRequestHeight = uint64(ctx.BlockHeight())
	case types.ValStateUnbonding:
		lc.UnbondingHeight = uint64(ctx.BlockHeight())
	case types.ValStateUnbonded:
		lc.UnbondedHeight = uint64(ctx.BlockHeight())
	}
	k.setValLifecycle(ctx, valAddr, lc)
}

func (k Keeper) setValLifecycle(ctx sdk.Context, valAddr sdk.ValAddress, lc *types.ValidatorLifecycle) {
	store := k.valLifecycleStore(ctx)
	lcBytes := k.cdc.MustMarshal(lc)
	store.Set([]byte(valAddr), lcBytes)
}

func (k Keeper) getValLifecycle(ctx sdk.Context, valAddr sdk.ValAddress) *types.ValidatorLifecycle {
	store := k.valLifecycleStore(ctx)
	lcBytes := store.Get([]byte(valAddr))
	var lc types.ValidatorLifecycle
	k.cdc.MustUnmarshal(lcBytes, &lc)
	return &lc
}

// valLifecycleStore returns the total voting power of the validator set of a given epoch
// prefix: ValidatorLifecycleKey
// key: val_addr
// value: ValidatorLifecycle object
func (k Keeper) valLifecycleStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.ValidatorLifecycleKey)
}
