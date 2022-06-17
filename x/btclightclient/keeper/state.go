package keeper

import (
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
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
func (s HeadersState) Create(header *wire.BlockHeader) {
	height, err := s.GetHeaderHeight(&header.PrevBlock)
	if err != nil {
		// Parent should always exist
		panic("Parent does not exist.")
	}

	headerHash := header.BlockHash()
	headersKey := types.HeadersObjectKey(height+1, &headerHash)
	heightKey := types.HeadersObjectHeightKey(&headerHash)

	headerRawBytes := types.BtcdHeaderToBytes(header)
	// save concrete object
	s.headers.Set(headersKey, headerRawBytes.HeaderBytes)
	// map header to height
	s.hashToHeight.Set(heightKey, sdk.Uint64ToBigEndian(height))
}

// GetHeader Retrieve a header by its height and hash
func (s HeadersState) GetHeader(height uint64, hash *chainhash.Hash) (*wire.BlockHeader, error) {
	headersKey := types.HeadersObjectKey(height, hash)
	rawBytes := s.headers.Get(headersKey)
	if rawBytes == nil {
		return nil, types.ErrHeaderDoesNotExist.Wrap("no header with provided height and hash")
	}

	headerBytes := &types.BTCHeaderBytes{HeaderBytes: rawBytes}
	header, err := types.BytesToBtcdHeader(headerBytes)
	if err != nil {
		return nil, err
	}
	return header, nil
}

// GetHeaderHeight Retrieve the Height of a header
func (s HeadersState) GetHeaderHeight(hash *chainhash.Hash) (uint64, error) {
	hashKey := types.HeadersObjectHeightKey(hash)
	bz := s.hashToHeight.Get(hashKey)
	if bz == nil {
		return 0, types.ErrHeaderDoesNotExist.Wrap("no header with provided hash")
	}
	height := sdk.BigEndianToUint64(bz)
	return height, nil
}

// GetHeaderByHash Retrieve a header by its hash
func (s HeadersState) GetHeaderByHash(hash *chainhash.Hash) (*wire.BlockHeader, error) {
	height, err := s.GetHeaderHeight(hash)
	if err != nil {
		return nil, err
	}
	return s.GetHeader(height, hash)
}

// GetHeadersByHeight Retrieve headers by their height
func (s HeadersState) GetHeadersByHeight(height uint64, f func(*wire.BlockHeader) bool) error {
	store := prefix.NewStore(s.headers, sdk.Uint64ToBigEndian(height))
	iter := store.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		rawBytes := iter.Value()
		headerBytes := &types.BTCHeaderBytes{HeaderBytes: rawBytes}

		header, err := types.BytesToBtcdHeader(headerBytes)
		if err != nil {
			return err
		}
		stop := f(header)
		if stop {
			break
		}
	}
	return nil
}

// Exists Check whether a hash is maintained in storage
func (s HeadersState) Exists(hash *chainhash.Hash) bool {
	_, err := s.GetHeaderHeight(hash)
	return err == nil
}

func (k Keeper) TipState(ctx sdk.Context) TipState {
	panic("implement me")
}

type TipState struct {
	// TODO
}
