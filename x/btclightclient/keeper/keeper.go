package keeper

import (
	"fmt"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/cometbft/cometbft/libs/log"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		cdc       codec.BinaryCodec
		storeKey  storetypes.StoreKey
		memKey    storetypes.StoreKey
		hooks     types.BTCLightClientHooks
		btcConfig bbn.BtcConfig
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	btcConfig bbn.BtcConfig,
) *Keeper {

	return &Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		memKey:    memKey,
		hooks:     nil,
		btcConfig: btcConfig,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetHooks sets the btclightclient hooks
func (k *Keeper) SetHooks(bh types.BTCLightClientHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set btclightclient hooks twice")
	}
	k.hooks = bh

	return k
}

// InsertHeader inserts a btcd header into the header state
func (k Keeper) InsertHeader(ctx sdk.Context, header *bbn.BTCHeaderBytes) error {
	if header == nil {
		return types.ErrEmptyMessage
	}
	headerHash := header.Hash()
	parentHash := header.ParentHash()

	// Check whether the header already exists, if yes reject
	headerExists := k.headersState(ctx).HeaderExists(headerHash)
	if headerExists {
		return types.ErrDuplicateHeader
	}

	// Check whether the parent exists, if not reject
	parentExists := k.headersState(ctx).HeaderExists(parentHash)
	if !parentExists {
		return types.ErrHeaderParentDoesNotExist
	}

	// Retrieve the height of the parent to calculate the current height
	parentHeight, err := k.headersState(ctx).GetHeaderHeight(parentHash)
	if err != nil {
		// Height should always exist if the previous checks have passed
		panic("Height for parent is not maintained")
	}

	// Retrieve the work of the parent to calculate the cumulative work
	parentWork, err := k.headersState(ctx).GetHeaderWork(parentHash)
	if err != nil {
		// Work should always exist if the previous checks have passed
		panic("Work for parent is not maintained")
	}

	// Calculate the cumulative work
	headerWork := types.CalcWork(header)
	cumulativeWork := types.CumulativeWork(headerWork, *parentWork)

	// Construct the BTCHeaderInfo object
	headerInfo := types.NewBTCHeaderInfo(header, headerHash, parentHeight+1, &cumulativeWork)

	// Retrieve the previous tip for future usage
	previousTip := k.headersState(ctx).GetTip()

	// Create the header
	k.headersState(ctx).CreateHeader(headerInfo)

	// Get the new tip
	currentTip := k.headersState(ctx).GetTip()

	// Variable maintaining the headers that have been added to the main chain
	var addedToMainChain []*types.BTCHeaderInfo

	k.triggerHeaderInserted(ctx, headerInfo)
	// The tip has changed, we need to send events
	if !currentTip.Eq(previousTip) {
		if !currentTip.Eq(headerInfo) {
			panic("The tip was updated but with a different header than the one provided")
		}
		// Get the highest common ancestor between the new tip and the old tip
		// There are two cases:
		// 	 1. The new tip extends the old tip
		//	    - The highest common ancestor is the old tip
		// 		- No need to send a roll-back event
		//   2. There has been a chain re-org
		// 		- Need to send a roll-back event
		var hca *types.BTCHeaderInfo
		if currentTip.HasParent(previousTip) {
			hca = previousTip
		} else {
			hca = k.headersState(ctx).GetHighestCommonAncestor(previousTip, currentTip)
			// chain re-org: trigger a roll-back event to the highest common ancestor
			k.triggerRollBack(ctx, hca)
		}
		// Find the newly added headers to the main chain
		addedToMainChain = k.headersState(ctx).GetInOrderAncestorsUntil(currentTip, hca)
		// Iterate through the added headers and trigger a roll-forward event
		for _, added := range addedToMainChain {
			// tipHeight + 1 - len(addedToMainChain) -> height of the highest common ancestor
			k.triggerRollForward(ctx, added)
		}
	}

	return nil
}

// BlockHeight returns the height of the provided header
func (k Keeper) BlockHeight(ctx sdk.Context, headerHash *bbn.BTCHeaderHashBytes) (uint64, error) {
	if headerHash == nil {
		return 0, types.ErrEmptyMessage
	}
	return k.headersState(ctx).GetHeaderHeight(headerHash)
}

// MainChainDepth returns the depth of the header in the main chain or -1 if it does not exist in it
func (k Keeper) MainChainDepth(ctx sdk.Context, headerHashBytes *bbn.BTCHeaderHashBytes) (int64, error) {
	if headerHashBytes == nil {
		return -1, types.ErrEmptyMessage
	}
	// Retrieve the header. If it does not exist, return an error
	headerInfo, err := k.headersState(ctx).GetHeaderByHash(headerHashBytes)
	if err != nil {
		return -1, err
	}

	// Retrieve the tip
	tipInfo := k.headersState(ctx).GetTip()

	// If the height of the requested header is larger than the tip, return -1
	if tipInfo.Height < headerInfo.Height {
		return -1, nil
	}

	// The depth is the number of blocks that have been build on top of the header
	// For example:
	// 		Tip: 0-deep
	// 		Tip height is 10, headerInfo height is 5: 5-deep etc.
	headerDepth := tipInfo.Height - headerInfo.Height
	mainchain := k.headersState(ctx).GetMainChainUpTo(headerDepth)

	// If we got an empty mainchain or the header does not equal the last element of the mainchain
	// then the header is not maintained inside the mainchain.
	if len(mainchain) == 0 || !headerInfo.Eq(mainchain[len(mainchain)-1]) {
		return -1, nil
	}
	return int64(headerDepth), nil
}

// IsHeaderKDeep returns true if a header is at least k-deep on the main chain
func (k Keeper) IsHeaderKDeep(ctx sdk.Context, headerHashBytes *bbn.BTCHeaderHashBytes, depth uint64) (bool, error) {
	if headerHashBytes == nil {
		return false, types.ErrEmptyMessage
	}
	mainchainDepth, err := k.MainChainDepth(ctx, headerHashBytes)
	if err != nil {
		return false, err
	}
	// If MainChainDepth returned a negative depth, then the header is not on the mainchain
	if mainchainDepth < 0 {
		return false, nil
	}
	// return true if the provided depth is more than equal the mainchain depth
	return depth >= uint64(mainchainDepth), nil
}

// IsAncestor returns true/false depending on whether `parent` is an ancestor of `child`.
// Returns false if the parent and the child are the same header.
func (k Keeper) IsAncestor(ctx sdk.Context, parentHashBytes *bbn.BTCHeaderHashBytes, childHashBytes *bbn.BTCHeaderHashBytes) (bool, error) {
	// nil checks
	if parentHashBytes == nil || childHashBytes == nil {
		return false, types.ErrEmptyMessage
	}
	// Retrieve parent and child header
	parentHeader, err := k.headersState(ctx).GetHeaderByHash(parentHashBytes)
	if err != nil {
		return false, types.ErrHeaderDoesNotExist.Wrapf("parent does not exist")
	}
	childHeader, err := k.headersState(ctx).GetHeaderByHash(childHashBytes)
	if err != nil {
		return false, types.ErrHeaderDoesNotExist.Wrapf("child does not exist")
	}

	// If the height of the child is equal or less than the parent, then the input is invalid
	if childHeader.Height <= parentHeader.Height {
		return false, nil
	}

	// Retrieve the ancestry
	ancestry := k.headersState(ctx).GetHeaderAncestryUpTo(childHeader, childHeader.Height-parentHeader.Height)
	// If it is empty, return false
	if len(ancestry) == 0 {
		return false, nil
	}
	// Return whether the last element of the ancestry is equal to the parent
	return ancestry[len(ancestry)-1].Eq(parentHeader), nil
}

func (k Keeper) GetTipInfo(ctx sdk.Context) *types.BTCHeaderInfo {
	return k.headersState(ctx).GetTip()
}

// TODO: The following functions, GetHeaderByHash and GetHeaderByHeight, are super inefficient
// and should be replaced with a better implementation. This requires changing the
// underlying data model for the whole btclightclient module.
// GetHeaderByHash returns header with given hash from main chain or returns nil if such header is not found
// or is not on main chain
func (k Keeper) GetHeaderByHash(ctx sdk.Context, hash *bbn.BTCHeaderHashBytes) *types.BTCHeaderInfo {
	depth, err := k.MainChainDepth(ctx, hash)

	if depth < 0 || err != nil {
		return nil
	}

	info, err := k.headersState(ctx).GetHeaderByHash(hash)

	if err != nil {
		return nil
	}

	return info
}

// GetHeaderByHeight returns header with given height from main chain, returns nil if such header is not found
func (k Keeper) GetHeaderByHeight(ctx sdk.Context, height uint64) *types.BTCHeaderInfo {
	var info *types.BTCHeaderInfo

	k.headersState(ctx).HeadersByHeight(height, func(hi *types.BTCHeaderInfo) bool {
		depth, err := k.MainChainDepth(ctx, hi.Hash)

		if depth < 0 || err != nil {
			return false
		}

		info = hi
		return true
	})

	return info
}

// GetHighestCommonAncestor traverses the ancestors of both headers
// to identify the common ancestor with the highest height
func (k Keeper) GetHighestCommonAncestor(ctx sdk.Context, header1 *types.BTCHeaderInfo, header2 *types.BTCHeaderInfo) *types.BTCHeaderInfo {
	return k.headersState(ctx).GetHighestCommonAncestor(header1, header2)
}

// GetInOrderAncestorsUntil returns the list of nodes starting from the block *after* the `ancestor` and ending with the `descendant`.
func (k Keeper) GetInOrderAncestorsUntil(ctx sdk.Context, descendant *types.BTCHeaderInfo, ancestor *types.BTCHeaderInfo) []*types.BTCHeaderInfo {
	return k.headersState(ctx).GetInOrderAncestorsUntil(descendant, ancestor)
}
