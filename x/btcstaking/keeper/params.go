package keeper

import (
	"context"
	"cosmossdk.io/math"
	"github.com/babylonchain/babylon/x/btcstaking/types"
)

// SetParams sets the x/btcstaking module parameters.
func (k Keeper) SetParams(ctx context.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&p)
	return store.Set(types.ParamsKey, bz)
}

// GetParams returns the current x/btcstaking module parameters.
func (k Keeper) GetParams(ctx context.Context) (p types.Params) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if err != nil {
		panic(err)
	}
	if bz == nil {
		return p
	}
	k.cdc.MustUnmarshal(bz, &p)
	return p
}

// MinCommissionRate returns the minimal commission rate of finality providers
func (k Keeper) MinCommissionRate(ctx context.Context) math.LegacyDec {
	return k.GetParams(ctx).MinCommissionRate
}
