package keeper

import (
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) InitChainInfo(ctx sdk.Context, chainID string, latestHeader *types.IndexedHeader) error {
	store := k.chainInfoStore(ctx)
	// the chain info should not exist at this point
	if store.Has([]byte(chainID)) {
		return types.ErrReInitChainInfo
	}
	chainInfo := &types.ChainInfo{
		ChainId:      chainID,
		LatestHeader: latestHeader,
		LatestForks:  nil,
	}
	k.setChainInfo(ctx, chainInfo)
	return nil
}

func (k Keeper) setChainInfo(ctx sdk.Context, chainInfo *types.ChainInfo) {
	store := k.chainInfoStore(ctx)
	store.Set([]byte(chainInfo.ChainId), k.cdc.MustMarshal(chainInfo))
}

func (k Keeper) GetChainInfo(ctx sdk.Context, chainID string) (*types.ChainInfo, error) {
	store := k.chainInfoStore(ctx)
	// GetChainInfo can be invoked only after the chain info is initialised
	if !store.Has([]byte(chainID)) {
		return nil, types.ErrChainInfoNotFound
	}
	chainInfoBytes := store.Get([]byte(chainID))
	var chainInfo types.ChainInfo
	k.cdc.MustUnmarshal(chainInfoBytes, &chainInfo)
	return &chainInfo, nil
}

func (k Keeper) UpdateLatestHeader(ctx sdk.Context, chainID string, header *types.IndexedHeader) error {
	chainInfo, err := k.GetChainInfo(ctx, chainID)
	// new header can only be received after there exists chain info
	if err != nil {
		return err
	}
	chainInfo.LatestHeader = header
	k.setChainInfo(ctx, chainInfo)
	return nil
}

func (k Keeper) UpdateLatestForks(ctx sdk.Context, chainID string, forks *types.Forks) error {
	chainInfo, err := k.GetChainInfo(ctx, chainID)
	// fork can only happen after there exists chain info
	if err != nil {
		return types.ErrChainInfoNotFound
	}
	chainInfo.LatestForks = forks
	k.setChainInfo(ctx, chainInfo)
	return nil
}

// msgChainInfoStore stores the information of canonical chains and forks for CZs
// prefix: ChainInfoKey
// key: chainID
// value: ChainInfo
func (k Keeper) chainInfoStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.ChainInfoKey)
}
