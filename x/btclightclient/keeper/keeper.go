package keeper

import (
	"fmt"
	"github.com/btcsuite/btcd/wire"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   sdk.StoreKey
		memKey     sdk.StoreKey
		hooks      types.BTCLightClientHooks
		paramstore paramtypes.Subspace
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,

) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		hooks:      nil,
		paramstore: ps,
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
func (k Keeper) InsertHeader(ctx sdk.Context, header *wire.BlockHeader) error {
	headerHash := header.BlockHash()
	headerExists := k.HeadersState(ctx).HeaderExists(&headerHash)
	if headerExists {
		return types.ErrDuplicateHeader.Wrap("header with provided hash already exists")
	}

	parentExists := k.HeadersState(ctx).HeaderExists(&header.PrevBlock)
	if !parentExists {
		return types.ErrHeaderParentDoesNotExist.Wrap("parent for provided hash is not maintained")
	}

	parentHeight, err := k.HeadersState(ctx).GetHeaderHeight(&header.PrevBlock)
	if err != nil {
		// Height should always exist if the previous checks have passed
		panic("Height for parent is not maintained")
	}

	parentWork, err := k.HeadersState(ctx).GetHeaderWork(&header.PrevBlock)
	if err != nil {
		// Work should always exist if the previous checks have passed
		panic("Work for parent is not maintained")
	}

	headerWork := types.CalcWork(header)
	cumulativeWork := types.CumulativeWork(headerWork, parentWork)

	previousTip := k.HeadersState(ctx).GetTip()
	// Create the header
<<<<<<< HEAD
	tipUpdated := k.HeadersState(ctx).CreateHeader(header, height+1, cumulativeWork)
	if tipUpdated {
		// Trigger TipUpdated hook
		k.AfterTipUpdated(ctx, height+1)
		// Emit TipUpdated event
		ctx.EventManager().EmitTypedEvent(&types.EventChainExtended{Height: height + 1})
	}
=======
	k.HeadersState(ctx).CreateHeader(header, parentHeight+1, cumulativeWork)

	// Get the new tip
	currentTip := k.HeadersState(ctx).GetTip()

	// Variable maintaining the headers that have been added to the main chain
	var addedToMainChain []*wire.BlockHeader

	// The tip has changed, we need to send events
	if !sameBlock(currentTip, previousTip) {
		if !sameBlock(currentTip, header) {
			panic("The tip was updated but with a different header than the one provided")
		}
		tipHeight := parentHeight + 1
		// Get the highest common ancestor between the new tip and the old tip
		// There are two cases:
		// 	 1. The new tip extends the old tip
		//	    - The highest common ancestor is the old tip
		// 		- No need to send a roll-back event
		//   2. There has been a chain re-org
		// 		- Need to send a roll-back event
		var hca *wire.BlockHeader
		var hcaHeight uint64
		if isParent(currentTip, previousTip) {
			hca = previousTip
			hcaHeight = parentHeight
		} else {
			hca := k.HeadersState(ctx).GetHighestCommonAncestor(previousTip, currentTip)
			hcaHash := hca.BlockHash()
			hcaHeight, err = k.HeadersState(ctx).GetHeaderHeight(&hcaHash)
			if err != nil {
				panic("Height for maintained header not available in storage")
			}
			// chain re-org: trigger a roll-back event to the highest common ancestor
			k.triggerRollBack(ctx, hca, hcaHeight)
		}
		// Find the newly added headers to the main chain
		addedToMainChain = k.HeadersState(ctx).GetInOrderAncestorsUntil(currentTip, hca)
		// Iterate through the added headers and trigger a roll-forward event
		for idx, added := range addedToMainChain {
			// tipHeight + 1 - len(addedToMainChain) -> height of the highest common ancestor
			addedHeight := tipHeight - uint64(len(addedToMainChain)) + 1 + uint64(idx)
			k.triggerRollForward(ctx, added, addedHeight)
		}
	}

>>>>>>> e625af45a6bfe769f34d66e848ef597f8f95fa21
	return nil
}

// BlockHeight returns the height of the provided header
func (k Keeper) BlockHeight(ctx sdk.Context, header *wire.BlockHeader) (uint64, error) {
	headerHash := header.BlockHash()
	return k.HeadersState(ctx).GetHeaderHeight(&headerHash)
}
<<<<<<< HEAD
=======

// HeaderKDeep returns true if a header is at least k-deep on the main chain
func (k Keeper) HeaderKDeep(ctx sdk.Context, header *wire.BlockHeader, depth uint64) bool {
	// TODO: optimize to not traverse the entire mainchain by storing the height along with the header
	mainchain := k.HeadersState(ctx).GetMainChain()
	if depth > uint64(len(mainchain)) {
		return false
	}
	// k-deep -> k headers built on top of the BTC header
	// Discard the first `depth` headers
	kDeepMainChain := mainchain[depth:]
	for _, mainChainHeader := range kDeepMainChain {
		if sameBlock(header, mainChainHeader) {
			return true
		}
	}
	return false
}
>>>>>>> e625af45a6bfe769f34d66e848ef597f8f95fa21
