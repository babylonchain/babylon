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

	height, err := k.HeadersState(ctx).GetHeaderHeight(&header.PrevBlock)
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
	k.HeadersState(ctx).CreateHeader(header, height+1, cumulativeWork)

	// Get the new tip
	currentTip := k.HeadersState(ctx).GetTip()

	// Variable maintaining the headers that have been added to the main chain
	var addedToMainChain []*wire.BlockHeader

	// The tip has changed, we need to send events
	if currentTip.BlockHash().String() != previousTip.BlockHash().String() {
		// The new tip extends the old tip
		if isParent(currentTip, previousTip) {
			// If the new header extended
			addedToMainChain = append(addedToMainChain, currentTip)
		} else {
			// There has been a chain re-org
			// Get the highest common ancestor between
			hca, hcaHeight := k.HeadersState(ctx).GetHighestCommonAncestor(previousTip, header)
			// Trigger a roll-back event to that ancestor
			k.triggerRollBack(ctx, hca, hcaHeight)

			// Find the newly added headers to the main chain
			addedToMainChain = k.HeadersState(ctx).GetInOrderAncestorsUntil(header, hca)
		}
	}
	// Iterate through the headers that were added to the main chain
	// and trigger a roll-forward event
	for idx, added := range addedToMainChain {
		// height + 1 -> height of the tip
		// height + 1 - len(addedToMainChain) -> height of highest common ancestor
		k.triggerRollForward(ctx, added, height+1-uint64(len(addedToMainChain))+uint64(idx))
	}

	return nil
}

// BlockHeight returns the height of the provided header
func (k Keeper) BlockHeight(ctx sdk.Context, header *wire.BlockHeader) (uint64, error) {
	headerHash := header.BlockHash()
	return k.HeadersState(ctx).GetHeaderHeight(&headerHash)
}

func isParent(child *wire.BlockHeader, parent *wire.BlockHeader) bool {
	return child.PrevBlock.String() == parent.BlockHash().String()
}
