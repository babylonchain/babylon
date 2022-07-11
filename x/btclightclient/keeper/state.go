package keeper

import (
	bbl "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
func (s HeadersState) CreateHeader(headerInfo *types.BTCHeaderInfo) {
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
func (s HeadersState) CreateTip(headerInfo *types.BTCHeaderInfo) {
	// Retrieve the key for the tip storage
	tipKey := types.TipKey()
	s.tip.Set(tipKey, s.cdc.MustMarshal(headerInfo))
}

// GetHeader Retrieve a header by its height and hash
func (s HeadersState) GetHeader(height uint64, hash *bbl.BTCHeaderHashBytes) (*types.BTCHeaderInfo, error) {
	// Keyed by (height, hash)
	headersKey := types.HeadersObjectKey(height, hash)

	// Retrieve the raw bytes
	rawBytes := s.headers.Get(headersKey)
	if rawBytes == nil {
		return nil, types.ErrHeaderDoesNotExist.Wrap("no header with provided height and hash")
	}

	return headerInfoFromStoredBytes(s.cdc, rawBytes), nil
}

// GetHeaderHeight Retrieve the Height of a header
func (s HeadersState) GetHeaderHeight(hash *bbl.BTCHeaderHashBytes) (uint64, error) {
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
func (s HeadersState) GetHeaderWork(hash *bbl.BTCHeaderHashBytes) (*sdk.Uint, error) {
	// Keyed by hash
	hashKey := types.HeadersObjectHeightKey(hash)
	// Retrieve the raw bytes for the work
	bz := s.hashToWork.Get(hashKey)
	if bz == nil {
		return nil, types.ErrHeaderDoesNotExist.Wrap("no header with provided hash")
	}

	// Convert to *big.Int form
	work := new(sdk.Uint)
	err := work.Unmarshal(bz)
	if err != nil {
		panic("Stored header cannot be unmarshalled to sdk.Uint")
	}
	return work, nil
}

// GetHeaderByHash Retrieve a header by its hash
func (s HeadersState) GetHeaderByHash(hash *bbl.BTCHeaderHashBytes) (*types.BTCHeaderInfo, error) {
	// Get the height of the header in order to use it along with the hash
	// as a (height, hash) key for the object storage
	height, err := s.GetHeaderHeight(hash)
	if err != nil {
		return nil, err
	}
	return s.GetHeader(height, hash)
}

// GetBaseBTCHeader retrieves the BTC header with the minimum height
func (s HeadersState) GetBaseBTCHeader() *types.BTCHeaderInfo {
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
func (s HeadersState) GetTip() *types.BTCHeaderInfo {
	if !s.TipExists() {
		return nil
	}

	// Get the key to the tip storage
	tipKey := types.TipKey()
	return headerInfoFromStoredBytes(s.cdc, s.tip.Get(tipKey))
}

// GetHeadersByHeight Retrieve headers by their height using an accumulator function
func (s HeadersState) GetHeadersByHeight(height uint64, f func(*types.BTCHeaderInfo) bool) {
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

// GetDescendingHeadersUpTo returns a collection of descending headers according to their height
func (s HeadersState) GetDescendingHeadersUpTo(tipHeight uint64, depth uint64) []*types.BTCHeaderInfo {
	var headers []*types.BTCHeaderInfo
	s.iterateReverseHeaders(func(header *types.BTCHeaderInfo) bool {
		headers = append(headers, header)
		if tipHeight-header.Height == depth {
			return true
		}
		return false
	})
	return headers
}

// GetMainChainUpTo returns the current canonical chain as a collection of block headers
// 				    starting from the tip and ending on the header that has a depth distance from it.
func (s HeadersState) GetMainChainUpTo(depth uint64) []*types.BTCHeaderInfo {
	// If there is no tip, there is no base header
	if !s.TipExists() {
		return nil
	}
	currentHeader := s.GetTip()

	// Retrieve a collection of headers in descending height order
	headers := s.GetDescendingHeadersUpTo(currentHeader.Height, depth)

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

// GetMainChain retrieves the main chain as a collection of block headers starting from the tip
// 				and ending on the base BTC header.
func (s HeadersState) GetMainChain() []*types.BTCHeaderInfo {
	if !s.TipExists() {
		return nil
	}
	tip := s.GetTip()
	// By providing the depth as the tip.Height, we ensure that we will go as deep as possible
	return s.GetMainChainUpTo(tip.Height)
}

// GetHighestCommonAncestor traverses the ancestors of both headers
//  						to identify the common ancestor with the highest height
func (s HeadersState) GetHighestCommonAncestor(header1 *types.BTCHeaderInfo, header2 *types.BTCHeaderInfo) *types.BTCHeaderInfo {
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
		return header2
	}

	ancestor1 := header1.Hash
	ancestor2 := header2.Hash
	var encountered map[string]bool
	encountered[ancestor1.String()] = true
	encountered[ancestor2.String()] = true
	var found *bbl.BTCHeaderHashBytes = nil

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
				ancestor1 = header.Hash
				if encountered[ancestor1.String()] {
					found = ancestor1
				}
				encountered[ancestor1.String()] = true
			}
			if ancestor2.Eq(header.Hash) {
				ancestor2 = header.Hash
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

// GetInOrderAncestorsUntil returns the list of nodes starting from the block *before* the `ancestor` and ending with the child.
func (s HeadersState) GetInOrderAncestorsUntil(child *types.BTCHeaderInfo, ancestor *types.BTCHeaderInfo) []*types.BTCHeaderInfo {
	if ancestor.Height >= child.Height {
		panic("Ancestor has a higher height than descendant")
	}

	currentHeader := child

	var ancestors []*types.BTCHeaderInfo
	ancestors = append(ancestors, child)
	if child.HasParent(ancestor) {
		return ancestors
	}

	found := false
	s.iterateReverseHeaders(func(header *types.BTCHeaderInfo) bool {
		if header.Eq(ancestor) {
			found = true
			return true
		}
		if currentHeader.HasParent(header) {
			currentHeader = header
			ancestors = append(ancestors, header)
		}
		// Abandon the iteration if the height of the current header is lower
		// than the height of the provided ancestor
		if currentHeader.Height < ancestor.Height {
			return true
		}
		return false
	})

	// If the header was not found, discard the ancestors list
	if !found {
		ancestors = []*types.BTCHeaderInfo{}
	}

	// Reverse the array
	for i, j := 0, len(ancestors)-1; i < j; i, j = i+1, j-1 {
		ancestors[i], ancestors[j] = ancestors[j], ancestors[i]
	}

	return ancestors
}

// HeaderExists Check whether a hash is maintained in storage
func (s HeadersState) HeaderExists(hash *bbl.BTCHeaderHashBytes) bool {
	// Get the prefix store for the hash->height collection
	store := prefix.NewStore(s.hashToHeight, types.HashToHeightPrefix)

	// Convert the BTCHeaderHashBytes object into raw bytes
	rawBytes := hash.MustMarshal()

	return store.Has(rawBytes)
}

// TipExists checks whether the tip of the canonical chain has been set
func (s HeadersState) TipExists() bool {
	tipKey := types.TipKey()
	return s.tip.Has(tipKey)
}

// updateLongestChain checks whether the tip should be updated and returns true if it does
func (s HeadersState) updateLongestChain(headerInfo *types.BTCHeaderInfo) {
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

func (s HeadersState) iterateReverseHeaders(fn func(*types.BTCHeaderInfo) bool) {
	// Get the prefix store for the (height, hash) -> header collection
	store := prefix.NewStore(s.headers, types.HeadersObjectPrefix)
	// Iterate it in reverse in order to get highest heights first
	// TODO: need to verify this assumption
	iter := store.ReverseIterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		header := headerInfoFromStoredBytes(s.cdc, iter.Value())
		stop := fn(header)
		if stop {
			break
		}
	}
}
