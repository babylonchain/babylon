package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: add more tests on the lifecycle record

// RecordNewValState adds a state for a validator lifecycle, including bonded, unbonding and unbonded
func (k Keeper) RecordNewValState(ctx sdk.Context, valAddr sdk.ValAddress, state types.BondState) error {
	lc := k.GetValLifecycle(ctx, valAddr)
	if lc == nil {
		lc = &types.ValidatorLifecycle{
			ValAddr: valAddr.String(), // bech32-encoded string
			ValLife: []*types.ValStateUpdate{},
		}
	}
	height, time := ctx.BlockHeight(), ctx.BlockTime()
	valStateUpdate := types.ValStateUpdate{
		State:       state,
		BlockHeight: uint64(height),
		BlockTime:   &time,
	}
	lc.ValLife = append(lc.ValLife, &valStateUpdate)
	k.SetValLifecycle(ctx, valAddr, lc)
	return nil
}

func (k Keeper) SetValLifecycle(ctx sdk.Context, valAddr sdk.ValAddress, lc *types.ValidatorLifecycle) {
	store := k.valLifecycleStore(ctx)
	lcBytes := k.cdc.MustMarshal(lc)
	store.Set([]byte(valAddr), lcBytes)
}

func (k Keeper) GetValLifecycle(ctx sdk.Context, valAddr sdk.ValAddress) *types.ValidatorLifecycle {
	store := k.valLifecycleStore(ctx)
	lcBytes := store.Get([]byte(valAddr))
	if len(lcBytes) == 0 {
		return nil
	}
	var lc types.ValidatorLifecycle
	k.cdc.MustUnmarshal(lcBytes, &lc)
	return &lc
}

// valLifecycleStore returns the store of the validator lifecycle
// prefix: ValidatorLifecycleKey
// key: val_addr
// value: ValidatorLifecycle object
func (k Keeper) valLifecycleStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.ValidatorLifecycleKey)
}
