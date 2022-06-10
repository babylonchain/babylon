package keeper

import (
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	headersStateNamespace  = []byte{0x0}                        // reserve this namespace for headers
	headersObjectNamespace = append(headersStateNamespace, 0x0) // where we save the concrete header bytes
	hashToHeightNamespace  = append(headersStateNamespace, 0x1) // where we map hash to height

	tipStateNamespace = []byte{0x1}                             // reserve this namespace for the tip
)

func (k Keeper) HeadersState(ctx sdk.Context) HeadersState {
    // Build the HeadersState storage
	store := ctx.KVStore(k.storeKey)
	return HeadersState{
		cdc:          k.cdc,
		headers:      prefix.NewStore(store, headersObjectNamespace),
		hashToHeight: prefix.NewStore(store, hashToHeightNamespace),
	}
}

func (k Keeper) TipState(ctx sdk.Context) TipState {
	panic("implement me")
}

type HeadersState struct {
	cdc          codec.BinaryCodec
	headers      sdk.KVStore
	hashToHeight sdk.KVStore
}

func (s HeadersState) Create(height uint64, header *types.BitcoinHeader) {
    // Method for inserting headers into the store

	// TODO: get the hash of the bitcoin header
	pk, headerHash := s.getPrimaryKey(height, header)

	// save concrete object
	s.headers.Set(pk, s.cdc.MustMarshal(header))
	// map header to height
	s.hashToHeight.Set(headerHash, sdk.Uint64ToBigEndian(height))
}

// TODO: Implement a getter function

func (s HeadersState) HeadersByHeight(height uint64, f func(header *types.BitcoinHeader) (stop bool)) {
    // Method  for retrieving headers by their height
	store := prefix.NewStore(s.headers, sdk.Uint64ToBigEndian(height))
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		headerRawBytes := iter.Value()
		header := new(types.BitcoinHeader)
		s.cdc.MustUnmarshal(headerRawBytes, header)
		stop := f(header)
		if stop {
			break
		}
	}
}

func (s HeadersState) getPrimaryKey(height uint64, header *types.BitcoinHeader) (primaryKey []byte, headerHash []byte) {
	// heightPartKey := sdk.Uint64ToBigEndian(height)
	// headerHashPartKey := hashHeader(header)
	// primaryKey := append(heightPartKey, headerHashPartKey...)
	// return primaryKey, headerHashPartKey
	return []byte("temp"), []byte("hash")
}

type TipState struct {
}
