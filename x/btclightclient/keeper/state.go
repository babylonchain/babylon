package keeper

import (
	bbl "github.com/babylonchain/babylon/types"
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
	tip          sdk.KVStore
}

func (k Keeper) HeadersState(ctx sdk.Context) HeadersState {
	// Build the HeadersState storage
	store := ctx.KVStore(k.storeKey)
	return HeadersState{
		cdc:          k.cdc,
		headers:      prefix.NewStore(store, types.HeadersObjectPrefix),
		hashToHeight: prefix.NewStore(store, types.HashToHeightPrefix),
		tip:          prefix.NewStore(store, types.TipPrefix),
	}
}

// CreateHeader Insert the header into the hash->height and (height, hash)->header storage
func (s HeadersState) CreateHeader(header *wire.BlockHeader, height uint64) {
	headerHash := header.BlockHash()
	headersKey := types.HeadersObjectKey(height, &headerHash)
	heightKey := types.HeadersObjectHeightKey(&headerHash)

	headerBytes := bbl.BtcdHeaderToHeaderBytes(header)
	// save concrete object
	s.headers.Set(headersKey, headerBytes)
	// map header to height
	s.hashToHeight.Set(heightKey, sdk.Uint64ToBigEndian(height))

	s.UpdateTip(header, height)
}

// GetHeader Retrieve a header by its height and hash
func (s HeadersState) GetHeader(height uint64, hash *chainhash.Hash) (*wire.BlockHeader, error) {
	headersKey := types.HeadersObjectKey(height, hash)
	rawBytes := s.headers.Get(headersKey)
	if rawBytes == nil {
		return nil, types.ErrHeaderDoesNotExist.Wrap("no header with provided height and hash")
	}
	var headerBytes bbl.BTCHeaderBytes
	headerBytes.Unmarshal(rawBytes)

	header, err := headerBytes.ToBtcdHeader()
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
		var headerBytes bbl.BTCHeaderBytes
		headerBytes.Unmarshal(iter.Value())
		header, err := headerBytes.ToBtcdHeader()
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

// GetDescendingHeaders returns a collection of descending headers according to their height
func (s HeadersState) GetDescendingHeaders() ([]*wire.BlockHeader, error) {
	var headers []*wire.BlockHeader

	store := prefix.NewStore(s.headers, types.HeadersObjectPrefix)
	iter := store.ReverseIterator(nil, nil)
	defer iter.Close()

	// TODO: This assumes that the ReverseIterator will return
	// 		 the blocks in a descending height manner.
	//		 Need to verify.
	for ; iter.Valid(); iter.Next() {
		var headerBytes bbl.BTCHeaderBytes
		headerBytes.Unmarshal(iter.Value())
		header, err := headerBytes.ToBtcdHeader()
		if err != nil {
			return nil, err
		}
		headers = append(headers, header)
	}
	return headers, nil
}

// HeaderExists Check whether a hash is maintained in storage
func (s HeadersState) HeaderExists(hash *chainhash.Hash) bool {
	store := prefix.NewStore(s.hashToHeight, types.HashToHeightPrefix)
	hashBytes := bbl.ChainhashToHeaderHashBytes(hash)
	return store.Has(hashBytes)
}

// GetBaseBTCHeader retrieves the BTC header with the minimum height
func (s HeadersState) GetBaseBTCHeader() (*wire.BlockHeader, error) {
	// If there is no tip, there is no base header
	if !s.TipExists() {
		return nil, nil
	}
	currentHeader := s.GetTip()

	// Retrieve a collection of headers along with their heights
	headers, err := s.GetDescendingHeaders()
	if err != nil {
		return nil, err
	}

	// Set the current header to be that of the tip
	// Iterate through the collection and:
	// 		- Discard anything with a higher height from the current header
	// 		- Find the parent of the header and set the current header to it
	// Return the current header
	for _, header := range headers {
		if header.BlockHash().String() == currentHeader.PrevBlock.String() {
			currentHeader = header
		}
	}

	return currentHeader, nil
}

// CreateTip sets the provided header as the tip
func (s HeadersState) CreateTip(header *wire.BlockHeader) {
	headerBytes := bbl.BtcdHeaderToHeaderBytes(header)
	tipKey := types.TipKey()
	s.tip.Set(tipKey, headerBytes)
}

// UpdateTip checks whether the tip should be updated and acts accordingly
func (s HeadersState) UpdateTip(header *wire.BlockHeader, height uint64) {
	if !s.TipExists() {
		s.CreateTip(header)
		return
	}

	// Currently, the tip is the one with the biggest height
	// TODO: replace this to use accumulative PoW instead
	tip := s.GetTip()
	tipHash := tip.BlockHash()
	tipHeight, err := s.GetHeaderHeight(&tipHash)
	if err != nil {
		panic("Existing tip does not have a maintained height")
	}

	if tipHeight < height {
		s.CreateTip(header)
	}
}

// GetTip returns the currently maintained tip
func (s HeadersState) GetTip() *wire.BlockHeader {
	if !s.TipExists() {
		return nil
	}

	tipKey := types.TipKey()
	var tipBytes bbl.BTCHeaderBytes
	tipBytes.Unmarshal(s.tip.Get(tipKey))
	tip, err := tipBytes.ToBtcdHeader()
	if err != nil {
		panic("Stored tip is not a valid btcd header")
	}
	return tip
}

// TipExists checks whether a tip is maintained
func (s HeadersState) TipExists() bool {
	tipKey := types.TipKey()
	return s.tip.Has(tipKey)
}
