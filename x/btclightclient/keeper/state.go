package keeper

import (
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type HeadersState struct {
	cdc          codec.BinaryCodec
	headers      sdk.KVStore
	hashToHeight sdk.KVStore
}

func (k Keeper) HeadersState(ctx sdk.Context) HeadersState {
	// Build the HeadersState storage
	store := ctx.KVStore(k.storeKey)
	return HeadersState{
		cdc:          k.cdc,
		headers:      prefix.NewStore(store, types.HeadersObjectPrefix),
		hashToHeight: prefix.NewStore(store, types.HashToHeightPrefix),
	}
}

// Create Insert a header into storage
func (s HeadersState) Create(header *types.BTCBlockHeader) {
	height, err := s.GetHeaderHeight(header.PrevBlock)
	if err != nil {
		// Parent should always exist
		panic("Parent does not exist.")
	}

	headersKey := types.HeadersObjectKey(height+1, header.Hash)
	heightKey := types.HeadersObjectHeightKey(header.Hash)

	// save concrete object
	s.headers.Set(headersKey, s.cdc.MustMarshal(header))
	// map header to height
	s.hashToHeight.Set(heightKey, sdk.Uint64ToBigEndian(height))
}

// GetHeader Retrieve a header by its height and hash
func (s HeadersState) GetHeader(height uint64, hash types.BlockHash) (*types.BTCBlockHeader, error) {
	headersKey := types.HeadersObjectKey(height, hash)
	bz := s.headers.Get(headersKey)
	if bz == nil {
		return nil, types.ErrHeaderDoesNotExist.Wrap("no header with provided height and hash")
	}

	header := new(types.BTCBlockHeader)
	s.cdc.MustUnmarshal(bz, header)
	return header, nil
}

// GetHeaderHeight Retrieve the Height of a header
func (s HeadersState) GetHeaderHeight(hash types.BlockHash) (uint64, error) {
	hashKey := types.HeadersObjectHeightKey(hash)
	bz := s.hashToHeight.Get(hashKey)
	if bz == nil {
		return 0, types.ErrHeaderDoesNotExist.Wrap("no header with provided hash")
	}
	height := sdk.BigEndianToUint64(bz)
	return height, nil
}

// GetHeaderByHash Retrieve a header by its hash
func (s HeadersState) GetHeaderByHash(hash types.BlockHash) (*types.BTCBlockHeader, error) {
	height, err := s.GetHeaderHeight(hash)
	if err != nil {
		return nil, err
	}
	return s.GetHeader(height, hash)
}

// GetHeadersByHeight Retrieve headers by their height
func (s HeadersState) GetHeadersByHeight(height uint64, f func(*types.BTCBlockHeader) bool) {
	store := prefix.NewStore(s.headers, sdk.Uint64ToBigEndian(height))
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		headerRawBytes := iter.Value()
		header := new(types.BTCBlockHeader)
		s.cdc.MustUnmarshal(headerRawBytes, header)
		stop := f(header)
		if stop {
			break
		}
	}
}

// GetBlockHashes Retrieve all block hashes
func (s HeadersState) GetBlockHashes(f func(types.BlockHash) bool) {
	iter := s.hashToHeight.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		// The header is the key of the KV store
		header := iter.Key()
		stop := f(header)
		if stop {
			break
		}
	}
}

// Exists Check whether a hash is maintained in storage
func (s HeadersState) Exists(hash types.BlockHash) bool {
	_, err := s.GetHeaderHeight(hash)
	return err != nil
}

func (k Keeper) TipState(ctx sdk.Context) TipState {
	panic("implement me")
}

type TipState struct {
	// TODO
}
