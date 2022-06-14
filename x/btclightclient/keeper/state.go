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

func (s HeadersState) Create(header *types.BTCBlockHeader) {
	// Insert a header into storage

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

func (s HeadersState) GetHeader(height uint64, hash []byte) (*types.BTCBlockHeader, error) {
	// Retrieve a header by its height and hash

	headersKey := types.HeadersObjectKey(height, hash)
	store := prefix.NewStore(s.headers, types.HeadersObjectPrefix)
	bz := store.Get(headersKey)
	if bz == nil {
		return nil, types.ErrHeaderDoesNotExist.Wrap("no header with provided height and hash")
	}

	header := new(types.BTCBlockHeader)
	s.cdc.MustUnmarshal(bz, header)
	return header, nil
}

func (s HeadersState) GetHeaderHeight(hash []byte) (uint64, error) {
	// Retrieve the Height of a header

	hashKey := types.HeadersObjectHeightKey(hash)
	store := prefix.NewStore(s.headers, types.HashToHeightPrefix)
	bz := store.Get(hashKey)
	if bz == nil {
		return 0, types.ErrHeaderDoesNotExist.Wrap("no header with provided hash")
	}
	height := sdk.BigEndianToUint64(bz)
	return height, nil
}

func (s HeadersState) GetHeaderByHash(hash []byte) (*types.BTCBlockHeader, error) {
	// Retrieve a header by its hash

	height, err := s.GetHeaderHeight(hash)
	if err != nil {
		return nil, err
	}
	return s.GetHeader(height, hash)
}

func (s HeadersState) GetHeadersByHeight(height uint64, f func(*types.BTCBlockHeader) bool) {
	// Retrieve headers by their height
	// func parameter is used for pagination
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

func (s HeadersState) GetHeaders(f func([]byte) bool) {
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

func (s HeadersState) Exists(hash []byte) bool {
	hashKey := types.HeadersObjectHeightKey(hash)
	store := prefix.NewStore(s.headers, types.HashToHeightPrefix)
	bz := store.Get(hashKey)
	return bz == nil
}

func (k Keeper) TipState(ctx sdk.Context) TipState {
	panic("implement me")
}

type TipState struct {
	// TODO
}
