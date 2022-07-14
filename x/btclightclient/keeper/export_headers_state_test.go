package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k *Keeper) HeadersState(ctx sdk.Context) headersState {
	return k.headersState(ctx)
}
