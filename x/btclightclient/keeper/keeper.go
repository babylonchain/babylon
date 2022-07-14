package keeper

import (
	"fmt"
	bbl "github.com/babylonchain/babylon/types"
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
func (k Keeper) InsertHeader(ctx sdk.Context, header *bbl.BTCHeaderBytes) error {
	headerHash := header.Hash()
	parentHash := header.ParentHash()

	// Check whether the header already exists, if yes reject
	headerExists := k.headersState(ctx).HeaderExists(headerHash)
	if headerExists {
		return types.ErrDuplicateHeader.Wrap("header with provided hash already exists")
	}

	// Check whether the parent exists, if not reject
	parentExists := k.headersState(ctx).HeaderExists(parentHash)
	if !parentExists {
		return types.ErrHeaderParentDoesNotExist.Wrap("parent for provided hash is not maintained")
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
func (k Keeper) BlockHeight(ctx sdk.Context, header *bbl.BTCHeaderBytes) (uint64, error) {
	headerHash := header.Hash()
	return k.headersState(ctx).GetHeaderHeight(headerHash)
}

// MainChainDepth returns the depth of the header in the main chain or -1 if it does not exist in it
func (k Keeper) MainChainDepth(ctx sdk.Context, headerBytes *bbl.BTCHeaderBytes) (int64, error) {
	// Retrieve the header. If it does not exist, return an error
	headerInfo, err := k.headersState(ctx).GetHeaderByHash(headerBytes.Hash())
	if err != nil {
		return -1, err
	}

	// Retrieve the tip
	tipInfo := k.headersState(ctx).GetTip()

	// If the height of the requested header is larger than the tip, return an error
	if tipInfo.Height < headerInfo.Height {
		return -1, types.ErrHeaderHigherThanTip.Wrap("header higher than tip")
	}

	headerDepth := tipInfo.Height - headerInfo.Height + 1
	mainchain := k.headersState(ctx).GetMainChainUpTo(headerDepth)

	// If we got an empty mainchain or the header does not equal the last element of the mainchain
	// then the header is not maintained inside the mainchain.
	if len(mainchain) == 0 || !headerInfo.Eq(mainchain[len(mainchain)-1]) {
		return -1, nil
	}
	return int64(headerDepth), nil
}

// IsHeaderKDeep returns true if a header is at least k-deep on the main chain
func (k Keeper) IsHeaderKDeep(ctx sdk.Context, headerBytes *bbl.BTCHeaderBytes, depth uint64) bool {
	mainchainDepth, err := k.MainChainDepth(ctx, headerBytes)
	if err != nil || mainchainDepth < 0 {
		return false
	}
	return uint64(mainchainDepth) >= depth
}
