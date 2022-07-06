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
func (s HeadersState) CreateHeader(header *wire.BlockHeader, height uint64) error {
	headerHash := header.BlockHash()
	// Get necessary keys according
	headersKey, err := types.HeadersObjectKey(height, &headerHash)
	if err != nil {
		return err
	}
	heightKey, err := types.HeadersObjectHeightKey(&headerHash)
	if err != nil {
		return err
	}

	// Convert the block header into bytes
	headerBytes, err := bbl.NewBTCHeaderBytesFromBlockHeader(header)
	if err != nil {
		return err
	}

	// save concrete object
	s.headers.Set(headersKey, headerBytes)
	// map header to height
	s.hashToHeight.Set(heightKey, sdk.Uint64ToBigEndian(height))

	return s.UpdateTip(header, height)
}

// GetHeader Retrieve a header by its height and hash
func (s HeadersState) GetHeader(height uint64, hash *chainhash.Hash) (*wire.BlockHeader, error) {
	// Keyed by (height, hash)
	headersKey, err := types.HeadersObjectKey(height, hash)
	if err != nil {
		return nil, err
	}
	// Retrieve the raw bytes
	rawBytes := s.headers.Get(headersKey)
	if rawBytes == nil {
		return nil, types.ErrHeaderDoesNotExist.Wrap("no header with provided height and hash")
	}

	// Get the BTCHeaderBytes object
	headerBytes, err := bbl.NewBTCHeaderBytesFromBytes(rawBytes)
	if err != nil {
		return nil, err
	}

	// Convert it into a btcd header
	header, err := headerBytes.ToBlockHeader()
	if err != nil {
		return nil, err
	}

	return header, nil
}

// GetHeaderHeight Retrieve the Height of a header
func (s HeadersState) GetHeaderHeight(hash *chainhash.Hash) (uint64, error) {
	// Keyed by hash
	hashKey, err := types.HeadersObjectHeightKey(hash)
	if err != nil {
		return 0, err
	}

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

// GetHeadersByHeight Retrieve headers by their height using an accumulator function
func (s HeadersState) GetHeadersByHeight(height uint64, f func(*wire.BlockHeader) bool) error {
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
		// Convert the bytes value into a BTCHeaderBytes object
		headerBytes, err := bbl.NewBTCHeaderBytesFromBytes(iter.Value())
		if err != nil {
			return err
		}
		// Convert the BTCHeaderBytes object into a *wire.BlockHeader object
		header, err := headerBytes.ToBlockHeader()
		if err != nil {
			return err
		}
		// The accumulator function notifies us whether the iteration should stop.
		stop := f(header)
		if stop {
			return nil
		}
	}
	return nil
}

// GetDescendingHeaders returns a collection of descending headers according to their height
func (s HeadersState) GetDescendingHeaders() ([]*wire.BlockHeader, error) {
	// Get the prefix store for the (height, hash) -> header collection
	store := prefix.NewStore(s.headers, types.HeadersObjectPrefix)
	// Iterate it in reverse in order to get highest heights first
	// TODO: need to verify this assumption
	iter := store.ReverseIterator(nil, nil)
	defer iter.Close()

	var headers []*wire.BlockHeader
	for ; iter.Valid(); iter.Next() {
		// Convert the bytes value into a BTCHeaderBytes object
		headerBytes, err := bbl.NewBTCHeaderBytesFromBytes(iter.Value())
		if err != nil {
			return nil, err
		}
		// Convert the BTCHeaderBytes object into a *wire.BlockHeader object
		header, err := headerBytes.ToBlockHeader()
		if err != nil {
			return nil, err
		}
		headers = append(headers, header)
	}
	return headers, nil
}

// HeaderExists Check whether a hash is maintained in storage
func (s HeadersState) HeaderExists(hash *chainhash.Hash) (bool, error) {
	// Get the prefix store for the hash->height collection
	store := prefix.NewStore(s.hashToHeight, types.HashToHeightPrefix)

	// Convert the *chainhash.Hash object into a BTCHeaderHashBytesObject
	hashBytes, err := bbl.NewBTCHeaderHashBytesFromChainhash(hash)
	if err != nil {
		return false, err
	}

	// Convert the BTCHeaderHashBytes object into raw bytes
	rawBytes, err := hashBytes.Marshal()
	if err != nil {
		return false, err
	}

	return store.Has(rawBytes), nil
}

// GetMainChain returns the current canonical chain as a collection of block headers
func (s HeadersState) GetMainChain() ([]*wire.BlockHeader, error) {
	// If there is no tip, there is no base header
	if !s.TipExists() {
		return nil, nil
	}
	currentHeader, err := s.GetTip()
	if err != nil {
		return nil, err
	}

	// Retrieve a collection of headers in descending height order
	headers, err := s.GetDescendingHeaders()
	if err != nil {
		return nil, err
	}

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

	return chain, nil
}

// GetBaseBTCHeader retrieves the BTC header with the minimum height
func (s HeadersState) GetBaseBTCHeader() (*wire.BlockHeader, error) {
	// Retrieve the canonical chain
	canonicalChain, err := s.GetMainChain()
	if err != nil {
		return nil, err
	}
	// If the canonical chain is empty, then there is no base header
	if len(canonicalChain) == 0 {
		return nil, nil
	}
	// The base btc header is the oldest one from the canonical chain
	return canonicalChain[len(canonicalChain)-1], nil
}

// CreateTip sets the provided header as the tip
func (s HeadersState) CreateTip(header *wire.BlockHeader) error {
	// Retrieve the key for the tip storage
	tipKey := types.TipKey()

	// Convert the *wire.BlockHeader object into a BTCHeaderBytes object
	headerBytes, err := bbl.NewBTCHeaderBytesFromBlockHeader(header)
	if err != nil {
		return err
	}

	// Convert the BTCHeaderBytes object into a bytes array
	rawBytes, err := headerBytes.Marshal()
	if err != nil {
		return err
	}

	s.tip.Set(tipKey, rawBytes)
	return nil
}

// UpdateTip checks whether the tip should be updated and acts accordingly
func (s HeadersState) UpdateTip(header *wire.BlockHeader, height uint64) error {
	// If there is no existing tip, then the header is set as the tip
	if !s.TipExists() {
		return s.CreateTip(header)
	}

	// Currently, the tip is the one with the biggest height
	// TODO: replace this to use accumulative PoW instead
	// Get the current tip header hash
	tip, err := s.GetTip()
	if err != nil {
		return err
	}
	tipHash := tip.BlockHash()
	tipHeight, err := s.GetHeaderHeight(&tipHash)
	if err != nil {
		panic("Existing tip does not have a maintained height")
	}

	if tipHeight < height {
		return s.CreateTip(header)
	}
	return nil
}

// GetTip returns the tip of the canonical chain
func (s HeadersState) GetTip() (*wire.BlockHeader, error) {
	if !s.TipExists() {
		return nil, nil
	}

	// Get the key to the tip storage
	tipKey := types.TipKey()
	// Convert the tip raw bytes into a BTCHeaderBytes object
	tipBytes, err := bbl.NewBTCHeaderBytesFromBytes(s.tip.Get(tipKey))
	if err != nil {
		return nil, err
	}
	// Convert the BTCHeaderBytes object into a *wire.BlockHeader object
	tip, err := tipBytes.ToBlockHeader()
	if err != nil {
		panic("Stored tip is not a valid btcd header")
	}
	return tip, nil
}

// TipExists checks whether the tip of the canonical chain has been set
func (s HeadersState) TipExists() bool {
	tipKey := types.TipKey()
	return s.tip.Has(tipKey)
}
