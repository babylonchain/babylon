package keeper

import (
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: init

func (k Keeper) setChainInfo(ctx sdk.Context, chainInfo *types.ChainInfo) {
	store := k.chainInfoStore(ctx)
	store.Set([]byte(chainInfo.ChainId), k.cdc.MustMarshal(chainInfo))
}

func (k Keeper) GetChainInfo(ctx sdk.Context, chainID string) (*types.ChainInfo, error) {
	store := k.chainInfoStore(ctx)
	chainInfoBytes := store.Get([]byte(chainID))
	if len(chainInfoBytes) == 0 {
		return nil, types.ErrNoChainInfo
	}
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
		return types.ErrNoChainInfo
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
