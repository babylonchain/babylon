package keeper

import (
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/big"
)

type HeadersState struct {
	cdc          codec.BinaryCodec
	headers      sdk.KVStore
	hashToHeight sdk.KVStore
	hashToWork   sdk.KVStore
	tip          sdk.KVStore
}

func (k Keeper) HeadersState(ctx sdk.Context) HeadersState {
	// Build the HeadersState storage
	store := ctx.KVStore(k.storeKey)
	return HeadersState{
		cdc:          k.cdc,
		headers:      prefix.NewStore(store, types.HeadersObjectPrefix),
		hashToHeight: prefix.NewStore(store, types.HashToHeightPrefix),
		hashToWork:   prefix.NewStore(store, types.HashToWorkPrefix),
		tip:          prefix.NewStore(store, types.TipPrefix),
	}
}

// CreateHeader Insert the header into the following storages:
// - hash->height
// - hash->work
// - (height, hash)->header storage
// Returns a boolean value indicating whether there is a new tip
func (s HeadersState) CreateHeader(header *wire.BlockHeader, height uint64, cumulativeWork *big.Int) {
	headerHash := header.BlockHash()
	// Get necessary keys according
	headersKey := types.HeadersObjectKey(height, &headerHash)
	heightKey := types.HeadersObjectHeightKey(&headerHash)
	workKey := types.HeadersObjectWorkKey(&headerHash)

	// Convert the block header into bytes
	headerBytes := bbl.NewBTCHeaderBytesFromBlockHeader(header)

	// save concrete object
	s.headers.Set(headersKey, headerBytes)
	// map header to height
	s.hashToHeight.Set(heightKey, sdk.Uint64ToBigEndian(height))
	// map header to work
	s.hashToWork.Set(workKey, cumulativeWork.Bytes())

	s.updateLongestChain(header, cumulativeWork)
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

// GetHeaderWork Retrieve the work of a header
func (s HeadersState) GetHeaderWork(hash *chainhash.Hash) (*big.Int, error) {
	// Keyed by hash
	hashKey := types.HeadersObjectHeightKey(hash)
	// Retrieve the raw bytes for the work
	bz := s.hashToWork.Get(hashKey)
	if bz == nil {
		return nil, types.ErrHeaderDoesNotExist.Wrap("no header with provided hash")
	}

	// Convert to *big.Int form
	work := new(big.Int).SetBytes(bz)
	return work, nil
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
	var headers []*wire.BlockHeader
	s.iterateReverseHeaders(func(header *wire.BlockHeader) bool {
		headers = append(headers, header)
		return false
	})
	return headers
}

// GetMainChain returns the current canonical chain as a collection of block headers
// 				starting from the tip and ending on the base header
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

// GetHighestCommonAncestor traverses the ancestors of both headers
//  						to identify the common ancestor with the highest height
func (s HeadersState) GetHighestCommonAncestor(header1 *wire.BlockHeader, header2 *wire.BlockHeader) *wire.BlockHeader {
	// The algorithm works as follows:
	// 1. Initialize a hashmap hash -> bool denoting whether the hash
	//    of an ancestor of either header1 or header2 has been encountered
	// 2. Maintain ancestor1 and ancestor2 as variables that point
	//	  to the current ancestor hash of the header1 and header2 parameters
	// 3. Whenever a node is encountered with a hash that is equal to ancestor{1,2},
	//    update the ancestor{1,2} variables.
	// 4. If ancestor1 or ancestor2 is set to the hash table,
	//    then that's the hash of the earliest ancestor
	// 5. Using the hash of the heighest ancestor wait until we get the header bytes
	// 	  in order to avoid an extra access.
	if isParent(header1, header2) {
		return header2
	}
	if isParent(header2, header1) {
		return header1
	}
	ancestor1 := header1.BlockHash()
	ancestor2 := header2.BlockHash()
	var encountered map[string]bool
	encountered[ancestor1.String()] = true
	encountered[ancestor2.String()] = true
	var found *chainhash.Hash = nil

	var resHeader *wire.BlockHeader = nil

	s.iterateReverseHeaders(func(btcdHeader *wire.BlockHeader) bool {
		// During iteration, we will encounter an ancestor for which its header hash
		// has been set on the hash map.
		// However, we do not have the entry yet, so we set the found flag to that hash
		// and when we encounter it during iteration we return it.
		if found != nil && sameHash(*found, btcdHeader.BlockHash()) {
			resHeader = btcdHeader
			return true
		} else {
			if ancestor1 == btcdHeader.BlockHash() {
				ancestor1 = btcdHeader.PrevBlock
				if encountered[ancestor1.String()] {
					found = &ancestor1
				}
				encountered[ancestor1.String()] = true
			}
			if ancestor2 == btcdHeader.BlockHash() {
				ancestor2 = btcdHeader.PrevBlock
				if encountered[ancestor2.String()] {
					found = &ancestor2
				}
				encountered[ancestor2.String()] = true
			}
		}
		return false
	})
	return resHeader
}

// GetInOrderAncestorsUntil returns the list of nodes starting from the child and ending with the block *before* the `ancestor`.
func (s HeadersState) GetInOrderAncestorsUntil(child *wire.BlockHeader, ancestor *wire.BlockHeader) []*wire.BlockHeader {
	currentHeader := child

	var ancestors []*wire.BlockHeader
	ancestors = append(ancestors, child)
	if isParent(child, ancestor) {
		return ancestors
	}
	s.iterateReverseHeaders(func(header *wire.BlockHeader) bool {
		if header.BlockHash() == ancestor.BlockHash() {
			return true
		}
		if header.BlockHash().String() == currentHeader.PrevBlock.String() {
			currentHeader = header
			ancestors = append(ancestors, header)
		}
		return false
	})

	return ancestors
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

// updateLongestChain checks whether the tip should be updated and returns true if it does
func (s HeadersState) updateLongestChain(header *wire.BlockHeader, cumulativeWork *big.Int) {
	// If there is no existing tip, then the header is set as the tip
	if !s.TipExists() {
		s.CreateTip(header)
		return
	}

	// Get the current tip header hash
	tip := s.GetTip()

	tipHash := tip.BlockHash()
	// Retrieve the tip's work from storage
	tipWork, err := s.GetHeaderWork(&tipHash)
	if err != nil {
		panic("Existing tip does not have a maintained work")
	}

	// If the work of the current tip is less than the work of the provided header,
	// the provided header is set as the tip.
	if tipWork.Cmp(cumulativeWork) < 0 {
		s.CreateTip(header)
	}
}

func (s HeadersState) iterateReverseHeaders(fn func(*wire.BlockHeader) bool) {
	// Get the prefix store for the (height, hash) -> header collection
	store := prefix.NewStore(s.headers, types.HeadersObjectPrefix)
	// Iterate it in reverse in order to get highest heights first
	// TODO: need to verify this assumption
	iter := store.ReverseIterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		btcdHeader := blockHeaderFromStoredBytes(iter.Value())
		stop := fn(btcdHeader)
		if stop {
			break
		}
	}
}
