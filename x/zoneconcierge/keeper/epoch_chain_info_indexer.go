package keeper

import (
	"fmt"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetEpochChainInfo gets the latest chain info of a given epoch for a given chain ID
func (k Keeper) GetEpochChainInfo(ctx sdk.Context, chainID string, epochNumber uint64) (*types.ChainInfo, error) {
	store := k.epochChainInfoStore(ctx, chainID)
	epochNumberBytes := sdk.Uint64ToBigEndian(epochNumber)
	if !store.Has(epochNumberBytes) {
		return nil, types.ErrEpochChainInfoNotFound
	}
	epochChainInfoBytes := store.Get(epochNumberBytes)
	var chainInfo types.ChainInfo
	k.cdc.MustUnmarshal(epochChainInfoBytes, &chainInfo)
	return &chainInfo, nil
}

// GetLastFinalizedChainInfo gets the last finalised chain info recorded for a given chain ID
// and the earliest epoch that snapshots this chain info
func (k Keeper) GetLastFinalizedChainInfo(ctx sdk.Context, chainID string) (uint64, *types.ChainInfo, error) {
	// find the last finalised epoch
	finalizedEpoch, err := k.GetFinalizedEpoch(ctx)
	if err != nil {
		return 0, nil, err
	}

	// find the chain info of this epoch
	chainInfo, err := k.GetEpochChainInfo(ctx, chainID, finalizedEpoch)
	if err != nil {
		return finalizedEpoch, nil, err
	}

	// It's possible that the chain info's epoch is way before the last finalised epoch
	// e.g., when there is no relayer for many epochs
	// NOTE: if an epoch is finalised then all of its previous epochs are also finalised
	if chainInfo.LatestHeader.BabylonEpoch < finalizedEpoch {
		// remember the last finalised epoch
		finalizedEpoch = chainInfo.LatestHeader.BabylonEpoch
		// replace the chain info w.r.t. this last finalised epoch
		chainInfo, err = k.GetEpochChainInfo(ctx, chainID, finalizedEpoch)
		if err != nil {
			return finalizedEpoch, nil, err
		}
	}

	return finalizedEpoch, chainInfo, nil
}

// GetEpochHeaders gets the headers timestamped in a given epoch, in the ascending order
func (k Keeper) GetEpochHeaders(ctx sdk.Context, chainID string, epochNumber uint64) ([]*types.IndexedHeader, error) {
	headers := []*types.IndexedHeader{}

	// find the last timestamped header of this chain in the epoch
	epochChainInfo, err := k.GetEpochChainInfo(ctx, chainID, epochNumber)
	if err != nil {
		return nil, err
	}
	// it's possible that this epoch's snapshot is not updated for many epochs
	// this implies that this epoch does not timestamp any header for this chain at all
	if epochChainInfo.LatestHeader.BabylonEpoch < epochNumber {
		return nil, types.ErrEpochHeadersNotFound
	}
	// now we have the last header in this epoch
	headers = append(headers, epochChainInfo.LatestHeader)

	// append all previous headers until reaching the previous epoch
	canonicalChainStore := k.canonicalChainStore(ctx, chainID)
	lastHeaderKey := sdk.Uint64ToBigEndian(epochChainInfo.LatestHeader.Height)
	// NOTE: even in ReverseIterator, start and end should still be specified in ascending order
	canonicalChainIter := canonicalChainStore.ReverseIterator(nil, lastHeaderKey)
	defer canonicalChainIter.Close()
	for ; canonicalChainIter.Valid(); canonicalChainIter.Next() {
		var prevHeader types.IndexedHeader
		k.cdc.MustUnmarshal(canonicalChainIter.Value(), &prevHeader)
		if prevHeader.BabylonEpoch < epochNumber {
			// we have reached the previous epoch, break the loop
			break
		}
		headers = append(headers, &prevHeader)
	}

	// reverse the list so that it remains ascending order
	bbn.Reverse(headers)

	return headers, nil
}

// ChainInfoExists checks whether the given chainInfo exists on Babylon or not
func (k Keeper) ChainInfoExists(ctx sdk.Context, chainInfo *types.ChainInfo) error {
	if chainInfo == nil {
		return fmt.Errorf("nil chainInfo")
	}
	if err := chainInfo.ValidateBasic(); err != nil {
		return err
	}

	expectedChainInfo, err := k.GetEpochChainInfo(ctx, chainInfo.ChainId, chainInfo.LatestHeader.BabylonEpoch)
	if err != nil {
		return err
	}

	if !chainInfo.Equal(expectedChainInfo) {
		return types.ErrInvalidChainInfo.Wrapf("chainInfo does not match: expected chainInfo in KVStore: %v", expectedChainInfo)
	}
	return nil
}

// recordEpochChainInfo records the chain info for a given epoch number of given chain ID
// where the latest chain info is retrieved from the chain info indexer
func (k Keeper) recordEpochChainInfo(ctx sdk.Context, chainID string, epochNumber uint64) {
	// get the latest known chain info
	// NOTE: GetChainInfo returns an empty ChainInfo object when the ChainInfo does not exist
	chainInfo := k.GetChainInfo(ctx, chainID)
	// NOTE: we can record epoch chain info without ancestor since IBC connection can be established at any height
	store := k.epochChainInfoStore(ctx, chainID)
	store.Set(sdk.Uint64ToBigEndian(epochNumber), k.cdc.MustMarshal(chainInfo))
}

// epochChainInfoStore stores each epoch's latest ChainInfo for a CZ
// prefix: EpochChainInfoKey || chainID
// key: epochNumber
// value: ChainInfo
func (k Keeper) epochChainInfoStore(ctx sdk.Context, chainID string) prefix.Store {
	store := ctx.KVStore(k.storeKey)
	epochChainInfoStore := prefix.NewStore(store, types.EpochChainInfoKey)
	chainIDBytes := []byte(chainID)
	return prefix.NewStore(epochChainInfoStore, chainIDBytes)
}
