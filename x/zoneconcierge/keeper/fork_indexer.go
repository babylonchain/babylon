package keeper

import (
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// canonicalChainStore returns the queue of msgs of a given epoch
// prefix: CanonicalChainKey || chainID
// key: height that this fork starts from
// value: a list of IndexedHeader, representing each header in the fork
func (k Keeper) forkStore(ctx sdk.Context, chainID string) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	canonicalChainStore := prefix.NewStore(store, types.CanonicalChainKey)
	chainIDBytes := []byte(chainID)
	return prefix.NewStore(canonicalChainStore, chainIDBytes)
}
