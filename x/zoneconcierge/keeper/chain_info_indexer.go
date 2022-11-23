package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) setChainInfo(ctx sdk.Context, chainInfo *types.ChainInfo) {
	store := k.chainInfoStore(ctx)
	store.Set([]byte(chainInfo.ChainId), k.cdc.MustMarshal(chainInfo))
}

// GetChainInfo returns the ChainInfo struct for a chain with a given ID
// Since IBC does not provide API that allows to initialise chain info right before creating an IBC connection,
// we can only check its existence every time, and return an empty one if it's not initialised yet.
func (k Keeper) GetChainInfo(ctx sdk.Context, chainID string) *types.ChainInfo {
	store := k.chainInfoStore(ctx)

	if !store.Has([]byte(chainID)) {
		return &types.ChainInfo{
			ChainId:      chainID,
			LatestHeader: nil,
			LatestForks: &types.Forks{
				Headers: []*types.IndexedHeader{},
			},
		}
	}
	chainInfoBytes := store.Get([]byte(chainID))
	var chainInfo types.ChainInfo
	k.cdc.MustUnmarshal(chainInfoBytes, &chainInfo)
	return &chainInfo
}

func (k Keeper) tryToUpdateLatestHeader(ctx sdk.Context, chainID string, header *types.IndexedHeader) error {
	if header == nil {
		return sdkerrors.Wrapf(types.ErrInvalidHeader, "header is nil")
	}
	// NOTE: we can accept header without ancestor since IBC connection can be established at any height
	chainInfo := k.GetChainInfo(ctx, chainID)
	if chainInfo.LatestHeader != nil {
		// ensure the header is the latest one
		// NOTE: submitting an old header is considered acceptable in IBC-Go (see Case_valid_past_update),
		// but the chain info indexer will not record such old header since it's not the latest one
		if chainInfo.LatestHeader.Height > header.Height {
			return nil
		}
	}
	chainInfo.LatestHeader = header
	k.setChainInfo(ctx, chainInfo)
	return nil
}

func (k Keeper) trpToUpdateLatestForkHeader(ctx sdk.Context, chainID string, header *types.IndexedHeader) error {
	if header == nil {
		return sdkerrors.Wrapf(types.ErrInvalidHeader, "header is nil")
	}

	chainInfo := k.GetChainInfo(ctx, chainID)

	if len(chainInfo.LatestForks.Headers) == 0 {
		// no fork at the moment, add this fork header as the latest one
		chainInfo.LatestForks.Headers = append(chainInfo.LatestForks.Headers, header)
	} else if chainInfo.LatestForks.Headers[0].Height == header.Height {
		// there exists fork headers at the same height, add this fork header to the set of latest fork headers
		chainInfo.LatestForks.Headers = append(chainInfo.LatestForks.Headers, header)
	} else if chainInfo.LatestForks.Headers[0].Height < header.Height {
		// this fork header is newer than the previous one, add this fork header as the latest one
		chainInfo.LatestForks = &types.Forks{
			Headers: []*types.IndexedHeader{header},
		}
	} else {
		// this fork header is older than the current latest fork, don't record this fork header in chain info
		return nil
	}

	k.setChainInfo(ctx, chainInfo)
	return nil
}

// GetAllChainIDs gets all chain IDs that integrate Babylon
func (k Keeper) GetAllChainIDs(ctx sdk.Context) []string {
	chainIDs := []string{}
	iter := k.chainInfoStore(ctx).Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		chainIDBytes := iter.Key()
		chainID := string(chainIDBytes)
		chainIDs = append(chainIDs, chainID)
	}
	return chainIDs
}

// msgChainInfoStore stores the information of canonical chains and forks for CZs
// prefix: ChainInfoKey
// key: chainID
// value: ChainInfo
func (k Keeper) chainInfoStore(ctx sdk.Context) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	return prefix.NewStore(store, types.ChainInfoKey)
}
