package keeper

import (
	"context"
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"

	"cosmossdk.io/store/prefix"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type headersState struct {
	cdc          codec.BinaryCodec
	storeAdapter storetypes.KVStore
	headers      storetypes.KVStore
	hashToHeight storetypes.KVStore
}

func (k Keeper) headersState(ctx context.Context) headersState {
	// Build the headersState storage
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return headersState{
		cdc:          k.cdc,
		storeAdapter: storeAdapter,
		headers:      prefix.NewStore(storeAdapter, types.HeadersObjectPrefix),
		hashToHeight: prefix.NewStore(storeAdapter, types.HashToHeightPrefix),
	}
}

// insertHeader Insert the header into the following storages:
// - hash->height
// - height -> HeaderInfo
func (s headersState) insertHeader(h *types.BTCHeaderInfo) {
	// Get necessary keys according
	headersKey := types.HeadersObjectKey(h.Height)
	heightKey := types.HeadersObjectHeightKey(h.Hash)

	// save concrete object
	s.headers.Set(headersKey, s.cdc.MustMarshal(h))
	s.hashToHeight.Set(heightKey, sdk.Uint64ToBigEndian(h.Height))
}

func (s headersState) deleteHeader(h *types.BTCHeaderInfo) {
	// Get necessary keys
	headersKey := types.HeadersObjectKey(h.Height)
	heightKey := types.HeadersObjectHeightKey(h.Hash)

	// save concrete object
	s.headers.Delete(headersKey)
	s.hashToHeight.Delete(heightKey)
}

func (s headersState) setLastRollbackPoint(reorgPoint *types.BTCHeaderInfo) {
	s.storeAdapter.Set(types.LastRollbackPointKey, s.cdc.MustMarshal(reorgPoint))
}

func (s headersState) GetLastRollbackPoint() *types.BTCHeaderInfo {
	reorgPointBytes := s.storeAdapter.Get(types.LastRollbackPointKey)
	if len(reorgPointBytes) == 0 {
		// reorg never happened
		return nil
	}
	var reorgPoint types.BTCHeaderInfo
	s.cdc.MustUnmarshal(reorgPointBytes, &reorgPoint)
	return &reorgPoint
}

func (s headersState) rollBackHeadersUpTo(height uint64) {
	headersToDelete := make([]*types.BTCHeaderInfo, 0)

	handleInfoFn := func(header *types.BTCHeaderInfo) bool {
		if len(headersToDelete) == 0 && height >= header.Height {
			// first header in iteration i.e the one with highest height and rollback to block
			// higher than current tip has been requested. stop the iteration
			return true
		}

		if header.Height == height {
			return true
		}

		headersToDelete = append(headersToDelete, header)
		return false
	}

	s.IterateReverseHeaders(handleInfoFn)

	// delete rollbacked headers from storage and set up new tip
	for _, header := range headersToDelete {
		s.deleteHeader(header)
	}
}

// GetHeaderByHeight Retrieve a header by its height and hash
func (s headersState) GetHeaderByHeight(height uint64) (*types.BTCHeaderInfo, error) {
	headersKey := types.HeadersObjectKey(height)

	// Retrieve the raw bytes
	rawBytes := s.headers.Get(headersKey)

	if rawBytes == nil {
		return nil, types.ErrHeaderDoesNotExist.Wrap("no header with provided height")
	}

	return headerInfoFromStoredBytes(s.cdc, rawBytes), nil
}

// GetHeaderByHash Retrieve a header by its hash
func (s headersState) GetHeaderByHash(hash *bbn.BTCHeaderHashBytes) (*types.BTCHeaderInfo, error) {
	// Keyed by hash
	hashKey := types.HeadersObjectHeightKey(hash)

	heightBytes := s.hashToHeight.Get(hashKey)

	if heightBytes == nil {
		return nil, types.ErrHeaderDoesNotExist.Wrap("no header with provided hash")
	}

	// Retrieve the raw bytes
	headerBytes := s.headers.Get(heightBytes)

	if headerBytes == nil {
		height := sdk.BigEndianToUint64(heightBytes)
		// panic here, as it means we got mapping hash->height but no mapping height->header
		// and those should always be in sync
		errMsg := fmt.Sprintf("header height exists but header does not. HeaderHash: %s, HeaderHeight: %d", hash.String(), height)
		panic(errMsg)
	}

	return headerInfoFromStoredBytes(s.cdc, headerBytes), nil
}

// GetTip returns the tip of the canonical chain
func (s headersState) GetTip() *types.BTCHeaderInfo {
	var tip *types.BTCHeaderInfo
	handleTipFn := func(header *types.BTCHeaderInfo) bool {
		// first retrieved header is tip
		tip = header
		return true
	}
	s.IterateReverseHeaders(handleTipFn)
	return tip
}

// HeaderExists Check whether a hash is maintained in storage
func (s headersState) HeaderExists(hash *bbn.BTCHeaderHashBytes) bool {
	if hash == nil {
		return false
	}

	_, err := s.GetHeaderByHash(hash)

	return err == nil
}

// TipExists checks whether the tip of the canonical chain has been set
func (s headersState) TipExists() bool {
	return s.GetTip() != nil
}

func (s headersState) IterateReverseHeaders(fn func(*types.BTCHeaderInfo) bool) {
	// Iterate it in reverse in order to get the highest heights first
	iter := s.headers.ReverseIterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		header := headerInfoFromStoredBytes(s.cdc, iter.Value())
		stop := fn(header)
		if stop {
			break
		}
	}
}

// IterateForwardHeaders iterates over all headers in store in increasing order
// - if startPoint is 0, it will start from the lowest height
// - if startPoint is lower that the lowest height, it will start from the lowest height
// - if startPoint is higher than the highest height, it will not iterate at all i.e provided
// callback will not be called
func (s headersState) IterateForwardHeaders(startPoint uint64, fn func(*types.BTCHeaderInfo) bool) {
	// Iterate it in increasing order to get lowest heights first
	var startKey []byte = nil
	if startPoint != 0 {
		startKey = types.HeadersObjectKey(startPoint)
	}

	iter := s.headers.Iterator(startKey, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		header := headerInfoFromStoredBytes(s.cdc, iter.Value())
		stop := fn(header)
		if stop {
			break
		}
	}
}

func (s headersState) BaseHeader() *types.BTCHeaderInfo {
	var baseHeader *types.BTCHeaderInfo
	handleBaseHeaderFn := func(header *types.BTCHeaderInfo) bool {
		// first retrieved header is base header
		baseHeader = header
		return true
	}
	s.IterateForwardHeaders(0, handleBaseHeaderFn)
	return baseHeader
}
