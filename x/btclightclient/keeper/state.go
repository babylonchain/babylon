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
	// Get necessary keys according
	headersKey := types.HeadersObjectKey(height, &headerHash)
	heightKey := types.HeadersObjectHeightKey(&headerHash)

	// Convert the block header into bytes
	headerBytes := bbl.NewBTCHeaderBytesFromBlockHeader(header)

	// save concrete object
	s.headers.Set(headersKey, headerBytes)
	// map header to height
	s.hashToHeight.Set(heightKey, sdk.Uint64ToBigEndian(height))

	s.updateLongestChain(header, height)
}

// CreateTip sets the provided header as the tip
func (s HeadersState) CreateTip(header *wire.BlockHeader) {
	// Retrieve the key for the tip storage
	tipKey := types.TipKey()

	// Convert the *wire.BlockHeader object into a BTCHeaderBytes object
	headerBytes := bbl.NewBTCHeaderBytesFromBlockHeader(header)

	// Convert the BTCHeaderBytes object into a bytes array
	rawBytes := headerBytes.MustMarshal()

	s.tip.Set(tipKey, rawBytes)
}

// GetHeader Retrieve a header by its height and hash
func (s HeadersState) GetHeader(height uint64, hash *chainhash.Hash) (*wire.BlockHeader, error) {
	// Keyed by (height, hash)
	headersKey := types.HeadersObjectKey(height, hash)

	// Retrieve the raw bytes
	rawBytes := s.headers.Get(headersKey)
	if rawBytes == nil {
		return nil, types.ErrHeaderDoesNotExist.Wrap("no header with provided height and hash")
	}

	return blockHeaderFromStoredBytes(rawBytes), nil
}

// GetHeaderHeight Retrieve the Height of a header
func (s HeadersState) GetHeaderHeight(hash *chainhash.Hash) (uint64, error) {
	// Keyed by hash
	hashKey := types.HeadersObjectHeightKey(hash)

	// Retrieve the raw bytes for the height
	bz := s.hashToHeight.Get(hashKey)
	if bz == nil {
		return 0, types.ErrHeaderDoesNotExist.Wrap("no header with provided hash")
	}

	// Convert to uint64 form
	height := sdk.BigEndianToUint64(bz)
	return height, nil
}

// GetHeaderByHash Retrieve a header by its hash
func (s HeadersState) GetHeaderByHash(hash *chainhash.Hash) (*wire.BlockHeader, error) {
	// Get the height of the header in order to use it along with the hash
	// as a (height, hash) key for the object storage
	height, err := s.GetHeaderHeight(hash)
	if err != nil {
		return nil, err
	}
	return s.GetHeader(height, hash)
}

// GetBaseBTCHeader retrieves the BTC header with the minimum height
func (s HeadersState) GetBaseBTCHeader() *wire.BlockHeader {
	// Retrieve the canonical chain
	canonicalChain := s.GetMainChain()
	// If the canonical chain is empty, then there is no base header
	if len(canonicalChain) == 0 {
		return nil
	}
	// The base btc header is the oldest one from the canonical chain
	return canonicalChain[len(canonicalChain)-1]
}

// GetTip returns the tip of the canonical chain
func (s HeadersState) GetTip() *wire.BlockHeader {
	if !s.TipExists() {
		return nil
	}

	// Get the key to the tip storage
	tipKey := types.TipKey()
	return blockHeaderFromStoredBytes(s.tip.Get(tipKey))
}

// GetHeadersByHeight Retrieve headers by their height using an accumulator function
func (s HeadersState) GetHeadersByHeight(height uint64, f func(*wire.BlockHeader) bool) {
	// The s.headers store is keyed by (height, hash)
	// By getting the prefix key using the height,
	// we are getting a store of `hash -> header` that contains all hashes
	// with a particular height.
	store := prefix.NewStore(s.headers, sdk.Uint64ToBigEndian(height))

	iter := store.Iterator(nil, nil)
	defer iter.Close()

	// Iterate through the prefix store and retrieve each header object.
	// Using the header object invoke the accumulator function.
	for ; iter.Valid(); iter.Next() {
		header := blockHeaderFromStoredBytes(iter.Value())
		// The accumulator function notifies us whether the iteration should stop.
		stop := f(header)
		if stop {
			break
		}
	}
}

// GetDescendingHeaders returns a collection of descending headers according to their height
func (s HeadersState) GetDescendingHeaders() []*wire.BlockHeader {
	// Get the prefix store for the (height, hash) -> header collection
	store := prefix.NewStore(s.headers, types.HeadersObjectPrefix)
	// Iterate it in reverse in order to get highest heights first
	// TODO: need to verify this assumption
	iter := store.ReverseIterator(nil, nil)
	defer iter.Close()

	var headers []*wire.BlockHeader
	for ; iter.Valid(); iter.Next() {
		headers = append(headers, blockHeaderFromStoredBytes(iter.Value()))
	}
	return headers
}

// GetMainChain returns the current canonical chain as a collection of block headers
func (s HeadersState) GetMainChain() []*wire.BlockHeader {
	// If there is no tip, there is no base header
	if !s.TipExists() {
		return nil
	}
	currentHeader := s.GetTip()

	// Retrieve a collection of headers in descending height order
	headers := s.GetDescendingHeaders()

	var chain []*wire.BlockHeader
	chain = append(chain, currentHeader)
	// Set the current header to be that of the tip
	// Iterate through the collection and:
	// 		- Discard anything with a higher height from the current header
	// 		- Find the parent of the header and set the current header to it
	// Return the current header
	for _, header := range headers {
		if header.BlockHash().String() == currentHeader.PrevBlock.String() {
			currentHeader = header
			chain = append(chain, header)
		}
	}

	return chain
}

// HeaderExists Check whether a hash is maintained in storage
func (s HeadersState) HeaderExists(hash *chainhash.Hash) bool {
	// Get the prefix store for the hash->height collection
	store := prefix.NewStore(s.hashToHeight, types.HashToHeightPrefix)

	// Convert the *chainhash.Hash object into a BTCHeaderHashBytesObject
	hashBytes := bbl.NewBTCHeaderHashBytesFromChainhash(hash)

	// Convert the BTCHeaderHashBytes object into raw bytes
	rawBytes := hashBytes.MustMarshal()

	return store.Has(rawBytes)
}

// TipExists checks whether the tip of the canonical chain has been set
func (s HeadersState) TipExists() bool {
	tipKey := types.TipKey()
	return s.tip.Has(tipKey)
}

// updateLongestChain checks whether the tip should be updated and acts accordingly
func (s HeadersState) updateLongestChain(header *wire.BlockHeader, height uint64) {
	// If there is no existing tip, then the header is set as the tip
	if !s.TipExists() {
		s.CreateTip(header)
		return
	}

	// Currently, the tip is the one with the biggest height
	// TODO: replace this to use accumulative PoW instead
	// Get the current tip header hash
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
