package keeper

import (
	"github.com/babylonchain/babylon/x/epoching/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: add more tests on the lifecycle record

// RecordNewDelegationState adds a state for a delegation lifecycle, including created, bonded, unbonding and unbonded
func (k Keeper) RecordNewDelegationState(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, state types.BondState) error {
	lc := k.GetDelegationLifecycle(ctx, delAddr)
	if lc == nil {
		lc = &types.DelegationLifecycle{
			DelAddr: delAddr.String(), // bech32-encoded string
			DelLife: []*types.DelegationStateUpdate{},
		}
	}
	height, time := ctx.BlockHeight(), ctx.BlockTime()
	DelegationStateUpdate := types.DelegationStateUpdate{
		State:       state,
		ValAddr:     valAddr.String(),
		BlockHeight: uint64(height),
		BlockTime:   &time,
	}
	lc.DelLife = append(lc.DelLife, &DelegationStateUpdate)
	k.SetDelegationLifecycle(ctx, delAddr, lc)
	return nil
}

func (k Keeper) SetDelegationLifecycle(ctx sdk.Context, delAddr sdk.AccAddress, lc *types.DelegationLifecycle) {
	store := k.delegationLifecycleStore(ctx)
	lcBytes := k.cdc.MustMarshal(lc)
	store.Set([]byte(delAddr), lcBytes)
}

func (k Keeper) GetDelegationLifecycle(ctx sdk.Context, delAddr sdk.AccAddress) *types.DelegationLifecycle {
	store := k.delegationLifecycleStore(ctx)
	lcBytes := store.Get([]byte(delAddr))
	if len(lcBytes) == 0 {
		return nil
	}
	var lc types.DelegationLifecycle
	k.cdc.MustUnmarshal(lcBytes, &lc)
	return &lc
}

// delegationLifecycleStore returns the store of the delegation lifecycle
// prefix: DelegationLifecycleKey
// key: del_addr
// value: DelegationLifecycle object
func (k Keeper) delegationLifecycleStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.DelegationLifecycleKey)
}
