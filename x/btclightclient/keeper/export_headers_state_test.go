package keeper

import (
	"context"
)

func (k *Keeper) HeadersState(ctx context.Context) headersState {
	return k.headersState(ctx)
}
