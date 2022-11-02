package keeper

import (
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmtypes "github.com/tendermint/tendermint/proto/tendermint/types"
)

func (k Keeper) setChainInfo(ctx sdk.Context, chainInfo *types.ChainInfo) {
	store := k.chainInfoStore(ctx)
	store.Set([]byte(chainInfo.ChainId), k.cdc.MustMarshal(chainInfo))
}

func (k Keeper) GetChainInfo(ctx sdk.Context, chainID string) *types.ChainInfo {
	store := k.chainInfoStore(ctx)
	chainInfoBytes := store.Get([]byte(chainID))
	if len(chainInfoBytes) == 0 {
		return nil
	}
	var chainInfo types.ChainInfo
	k.cdc.MustUnmarshal(chainInfoBytes, &chainInfo)
	return &chainInfo
}

func (k Keeper) UpdateLatestHeader(ctx sdk.Context, chainID string, header *tmtypes.Header) error {
	chainInfo := k.GetChainInfo(ctx, chainID)
	if chainInfo == nil {
		chainInfo = &types.ChainInfo{
			ChainId:      chainID,
			LatestHeader: header,
			LatestFork:   nil,
		}
	} else {
		chainInfo.LatestHeader = header
	}
	k.setChainInfo(ctx, chainInfo)
	return nil
}

func (k Keeper) UpdateLatestFork(ctx sdk.Context, chainID string, fork *types.Fork) error {
	chainInfo := k.GetChainInfo(ctx, chainID)
	if chainInfo == nil { // fork can only happen after there exists chain info
		return types.ErrNoChainInfo
	}
	chainInfo.LatestFork = fork
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
