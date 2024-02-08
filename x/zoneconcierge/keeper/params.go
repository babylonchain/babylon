package keeper

import (
	"context"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
)

// SetParams sets the x/zoneconcierge module parameters.
func (k Keeper) SetParams(ctx context.Context, p types.Params) error {
	if err := p.Validate(); err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&p)
	if err := store.Set(types.ParamsKey, bz); err != nil {
		panic(err)
	}

	return nil
}

// GetParams returns the current x/zoneconcierge module parameters.
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
