package keeper

import (
	sdkmath "cosmossdk.io/math"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type headersState struct {
	cdc          codec.BinaryCodec
	headers      sdk.KVStore
	hashToHeight sdk.KVStore
	hashToWork   sdk.KVStore
	tip          sdk.KVStore
}

func (k Keeper) headersState(ctx sdk.Context) headersState {
	// Build the headersState storage
	store := ctx.KVStore(k.storeKey)
	return headersState{
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
func (s headersState) CreateHeader(headerInfo *types.BTCHeaderInfo) {
	headerHash := headerInfo.Hash
	height := headerInfo.Height
	cumulativeWork := headerInfo.Work

	// Get necessary keys according
	headersKey := types.HeadersObjectKey(height, headerHash)
	heightKey := types.HeadersObjectHeightKey(headerHash)
	workKey := types.HeadersObjectWorkKey(headerHash)

	// save concrete object
	s.headers.Set(headersKey, s.cdc.MustMarshal(headerInfo))
	// map header to height
	s.hashToHeight.Set(heightKey, sdk.Uint64ToBigEndian(height))
	// map header to work
	workBytes, err := cumulativeWork.Marshal()
	if err != nil {
		panic("Work cannot be marshalled")
	}
	s.hashToWork.Set(workKey, workBytes)

	s.updateLongestChain(headerInfo)
}

// CreateTip sets the provided header as the tip
func (s headersState) CreateTip(headerInfo *types.BTCHeaderInfo) {
	// Retrieve the key for the tip storage
	tipKey := types.TipKey()
	s.tip.Set(tipKey, s.cdc.MustMarshal(headerInfo))
}

// GetHeader Retrieve a header by its height and hash
func (s headersState) GetHeader(height uint64, hash *bbn.BTCHeaderHashBytes) (*types.BTCHeaderInfo, error) {
	// Keyed by (height, hash)
	headersKey := types.HeadersObjectKey(height, hash)

	if !s.headers.Has(headersKey) {
		return nil, types.ErrHeaderDoesNotExist.Wrap("no header with provided height and hash")
	}
	// Retrieve the raw bytes
	rawBytes := s.headers.Get(headersKey)

	return headerInfoFromStoredBytes(s.cdc, rawBytes), nil
}

// GetHeaderHeight Retrieve the Height of a header
func (s headersState) GetHeaderHeight(hash *bbn.BTCHeaderHashBytes) (uint64, error) {
	// Keyed by hash
	hashKey := types.HeadersObjectHeightKey(hash)

	// Retrieve the raw bytes for the height
	if !s.hashToHeight.Has(hashKey) {
		return 0, types.ErrHeaderDoesNotExist.Wrap("no header with provided hash")
	}

	bz := s.hashToHeight.Get(hashKey)
	// Convert to uint64 form
	height := sdk.BigEndianToUint64(bz)
	return height, nil
}

// GetHeaderWork Retrieve the work of a header
func (s headersState) GetHeaderWork(hash *bbn.BTCHeaderHashBytes) (*sdkmath.Uint, error) {
	// Keyed by hash
	hashKey := types.HeadersObjectHeightKey(hash)
	// Retrieve the raw bytes for the work
	bz := s.hashToWork.Get(hashKey)
	if bz == nil {
		return nil, types.ErrHeaderDoesNotExist.Wrap("no header with provided hash")
	}

	// Convert to *big.Int form
	work := new(sdkmath.Uint)
	err := work.Unmarshal(bz)
	if err != nil {
		panic("Stored header cannot be unmarshalled to sdk.Uint")
	}
	return work, nil
}

// GetHeaderByHash Retrieve a header by its hash
func (s headersState) GetHeaderByHash(hash *bbn.BTCHeaderHashBytes) (*types.BTCHeaderInfo, error) {
	// Get the height of the header in order to use it along with the hash
	// as a (height, hash) key for the object storage
	height, err := s.GetHeaderHeight(hash)
	if err != nil {
		return nil, err
	}
	return s.GetHeader(height, hash)
}

// GetBaseBTCHeader retrieves the BTC header with the minimum height
func (s headersState) GetBaseBTCHeader() *types.BTCHeaderInfo {
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
func (s headersState) GetTip() *types.BTCHeaderInfo {
	if !s.TipExists() {
		return nil
	}
	// Get the key to the tip storage
	tipKey := types.TipKey()
	return headerInfoFromStoredBytes(s.cdc, s.tip.Get(tipKey))
}

// HeadersByHeight Retrieve headers by their height using an accumulator function
func (s headersState) HeadersByHeight(height uint64, f func(*types.BTCHeaderInfo) bool) {
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
		header := headerInfoFromStoredBytes(s.cdc, iter.Value())
		// The accumulator function notifies us whether the iteration should stop.
		stop := f(header)
		if stop {
			break
		}
	}
}

// getDescendingHeadersUpTo returns a collection of descending headers according to their height
func (s headersState) getDescendingHeadersUpTo(startHeight uint64, depth uint64) []*types.BTCHeaderInfo {
	var headers []*types.BTCHeaderInfo
	s.iterateReverseHeaders(func(header *types.BTCHeaderInfo) bool {
		// Use `depth+1` because we want to first gather all the headers
		// with a depth of `depth`.
		if startHeight-header.Height == depth+1 {
			return true
		}
		headers = append(headers, header)
		return false
	})
	return headers
}

// GetHeaderAncestryUpTo returns a list of headers starting from the header parameter and leading to
//
//	the header that has a `depth` distance from it.
func (s headersState) GetHeaderAncestryUpTo(currentHeader *types.BTCHeaderInfo, depth uint64) []*types.BTCHeaderInfo {
	// Retrieve a collection of headers in descending height order
	// Use depth+1 since we want all headers at the depth height.
	headers := s.getDescendingHeadersUpTo(currentHeader.Height, depth)

	var chain []*types.BTCHeaderInfo
	chain = append(chain, currentHeader)
	// Set the current header to be that of the tip
	// Iterate through the collection and:
	// 		- Discard anything with a higher height from the current header
	// 		- Find the parent of the header and set the current header to it
	// Return the current header
	for _, header := range headers {
		if currentHeader.HasParent(header) {
			currentHeader = header
			chain = append(chain, header)
		}
	}

	return chain
}

// GetMainChainUpTo returns the current canonical chain as a collection of block headers
// starting from the tip and ending on the header that has `depth` distance from it.
func (s headersState) GetMainChainUpTo(depth uint64) []*types.BTCHeaderInfo {
	// If there is no tip, there is no base header
	if !s.TipExists() {
		return nil
	}
	return s.GetHeaderAncestryUpTo(s.GetTip(), depth)
}

// GetMainChain retrieves the main chain as a collection of block headers starting from the tip
//
//	and ending on the base BTC header.
func (s headersState) GetMainChain() []*types.BTCHeaderInfo {
	if !s.TipExists() {
		return nil
	}
	tip := s.GetTip()
	// By providing the depth as the tip.Height, we ensure that we will go as deep as possible
	return s.GetMainChainUpTo(tip.Height)
}

// GetHighestCommonAncestor traverses the ancestors of both headers
// to identify the common ancestor with the highest height
func (s headersState) GetHighestCommonAncestor(header1 *types.BTCHeaderInfo, header2 *types.BTCHeaderInfo) *types.BTCHeaderInfo {
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
	if header1.HasParent(header2) {
		return header2
	}
	if header2.HasParent(header1) {
		return header1
	}

	ancestor1 := header1.Hash
	ancestor2 := header2.Hash
	encountered := make(map[string]bool, 0)
	encountered[ancestor1.String()] = true
	encountered[ancestor2.String()] = true
	var found *bbn.BTCHeaderHashBytes = nil

	var resHeader *types.BTCHeaderInfo = nil

	s.iterateReverseHeaders(func(header *types.BTCHeaderInfo) bool {
		// During iteration, we will encounter an ancestor for which its header hash
		// has been set on the hash map.
		// However, we do not have the entry yet, so we set the found flag to that hash
		// and when we encounter it during iteration we return it.
		if found != nil && header.Hash.Eq(found) {
			resHeader = header
			return true
		} else {
			if ancestor1.Eq(header.Hash) {
				ancestor1 = header.Header.ParentHash()
				if encountered[ancestor1.String()] {
					found = ancestor1
				}
				encountered[ancestor1.String()] = true
			}
			if ancestor2.Eq(header.Hash) {
				ancestor2 = header.Header.ParentHash()
				if encountered[ancestor2.String()] {
					found = ancestor2
				}
				encountered[ancestor2.String()] = true
			}
		}
		return false
	})
	return resHeader
}

// GetInOrderAncestorsUntil returns the list of nodes starting from the block *after* the `ancestor` and ending with the `descendant`.
func (s headersState) GetInOrderAncestorsUntil(descendant *types.BTCHeaderInfo, ancestor *types.BTCHeaderInfo) []*types.BTCHeaderInfo {
	if ancestor.Height > descendant.Height {
		panic("Ancestor has a higher height than descendant")
	}
	if ancestor.Height == descendant.Height {
		// return an empty list
		return []*types.BTCHeaderInfo{}
	}

	if descendant.HasParent(ancestor) {
		return []*types.BTCHeaderInfo{descendant}
	}

	ancestors := s.GetHeaderAncestryUpTo(descendant, descendant.Height-ancestor.Height)
	if !ancestors[len(ancestors)-1].Eq(ancestor) {
		// `ancestor` is not an ancestor of `descendant`, return an empty list
		return []*types.BTCHeaderInfo{}
	}

	// Discard the last element of the ancestry which corresponds to `ancestor`
	ancestors = ancestors[:len(ancestors)-1]

	// Reverse the ancestry
	for i, j := 0, len(ancestors)-1; i < j; i, j = i+1, j-1 {
		ancestors[i], ancestors[j] = ancestors[j], ancestors[i]
	}

	return ancestors
}

// HeaderExists Check whether a hash is maintained in storage
func (s headersState) HeaderExists(hash *bbn.BTCHeaderHashBytes) bool {
	if hash == nil {
		return false
	}
	return s.hashToHeight.Has(hash.MustMarshal())
}

// TipExists checks whether the tip of the canonical chain has been set
func (s headersState) TipExists() bool {
	tipKey := types.TipKey()
	return s.tip.Has(tipKey)
}

// updateLongestChain checks whether the tip should be updated and returns true if it does
func (s headersState) updateLongestChain(headerInfo *types.BTCHeaderInfo) {
	// If there is no existing tip, then the header is set as the tip
	if !s.TipExists() {
		s.CreateTip(headerInfo)
		return
	}

	// Get the current tip header hash
	tip := s.GetTip()

	// If the work of the current tip is less than the work of the provided header,
	// the provided header is set as the tip.
	if headerInfo.Work.GT(*tip.Work) {
		s.CreateTip(headerInfo)
	}
}

func (s headersState) iterateReverseHeaders(fn func(*types.BTCHeaderInfo) bool) {
	// Iterate it in reverse in order to get highest heights first
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
