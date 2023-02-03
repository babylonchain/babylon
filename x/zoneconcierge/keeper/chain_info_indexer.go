package keeper

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) setChainInfo(ctx sdk.Context, chainInfo *types.ChainInfo) {
	store := k.chainInfoStore(ctx)
	store.Set([]byte(chainInfo.ChainId), k.cdc.MustMarshal(chainInfo))
}

func (k Keeper) InitChainInfo(ctx sdk.Context, chainID string) (*types.ChainInfo, error) {
	if len(chainID) == 0 {
		return nil, fmt.Errorf("chainID is empty")
	}
	// ensure chain info has not been initialised yet
	if k.HasChainInfo(ctx, chainID) {
		return nil, sdkerrors.Wrapf(types.ErrInvalidChainInfo, "chain info has already initialized")
	}

	chainInfo := &types.ChainInfo{
		ChainId:      chainID,
		LatestHeader: nil,
		LatestForks: &types.Forks{
			Headers: []*types.IndexedHeader{},
		},
		TimestampedHeadersCount: 0,
	}

	k.setChainInfo(ctx, chainInfo)
	return chainInfo, nil
}

// HasChainInfo returns whether the chain info exists for a given ID
// Since IBC does not provide API that allows to initialise chain info right before creating an IBC connection,
// we can only check its existence every time, and return an empty one if it's not initialised yet.
func (k Keeper) HasChainInfo(ctx sdk.Context, chainID string) bool {
	store := k.chainInfoStore(ctx)
	return store.Has([]byte(chainID))
}

// GetChainInfo returns the ChainInfo struct for a chain with a given ID
// Since IBC does not provide API that allows to initialise chain info right before creating an IBC connection,
// we can only check its existence every time, and return an empty one if it's not initialised yet.
func (k Keeper) GetChainInfo(ctx sdk.Context, chainID string) (*types.ChainInfo, error) {
	store := k.chainInfoStore(ctx)

	if !store.Has([]byte(chainID)) {
		return nil, types.ErrEpochChainInfoNotFound
	}
	chainInfoBytes := store.Get([]byte(chainID))
	var chainInfo types.ChainInfo
	k.cdc.MustUnmarshal(chainInfoBytes, &chainInfo)
	return &chainInfo, nil
}

// updateLatestHeader updates the chainInfo w.r.t. the given header, including
// - replace the old latest header with the given one
// - increment the number of timestamped headers
// Note that this function is triggered only upon receiving headers from the relayer,
// and only a subset of headers in CZ are relayed. Thus TimestampedHeadersCount is not
// equal to the total number of headers in CZ.
func (k Keeper) updateLatestHeader(ctx sdk.Context, chainID string, header *types.IndexedHeader) error {
	if header == nil {
		return sdkerrors.Wrapf(types.ErrInvalidHeader, "header is nil")
	}
	chainInfo, err := k.GetChainInfo(ctx, chainID)
	if err != nil {
		// chain info has not been initialised yet
		return fmt.Errorf("failed to get chain info of %s: %w", chainID, err)
	}
	chainInfo.LatestHeader = header     // replace the old latest header with the given one
	chainInfo.TimestampedHeadersCount++ // increment the number of timestamped headers

	k.setChainInfo(ctx, chainInfo)
	return nil
}

// tryToUpdateLatestForkHeader tries to update the chainInfo w.r.t. the given fork header
// - If no fork exists, add this fork header as the latest one
// - If there is a fork header at the same height, add this fork to the set of latest fork headers
// - If this fork header is newer than the previous one, replace the old fork headers with this fork header
// - If this fork header is older than the current latest fork, ignore
func (k Keeper) tryToUpdateLatestForkHeader(ctx sdk.Context, chainID string, header *types.IndexedHeader) error {
	if header == nil {
		return sdkerrors.Wrapf(types.ErrInvalidHeader, "header is nil")
	}

	chainInfo, err := k.GetChainInfo(ctx, chainID)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrChainInfoNotFound, "cannot insert fork header when chain info is not initialized")
	}

	if len(chainInfo.LatestForks.Headers) == 0 {
		// no fork at the moment, add this fork header as the latest one
		chainInfo.LatestForks.Headers = append(chainInfo.LatestForks.Headers, header)
	} else if chainInfo.LatestForks.Headers[0].Height == header.Height {
		// there exists fork headers at the same height, add this fork header to the set of latest fork headers
		chainInfo.LatestForks.Headers = append(chainInfo.LatestForks.Headers, header)
	} else if chainInfo.LatestForks.Headers[0].Height < header.Height {
		// this fork header is newer than the previous one, replace the old fork headers with this fork header
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
