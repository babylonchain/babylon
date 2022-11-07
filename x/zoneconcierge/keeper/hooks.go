package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclientkeeper "github.com/cosmos/ibc-go/v5/modules/core/02-client/keeper"
	ibctmtypes "github.com/cosmos/ibc-go/v5/modules/light-clients/07-tendermint/types"
)

type Hooks struct {
	k Keeper
}

// ensures Hooks implements ClientHooks interfaces
var _ ibcclientkeeper.ClientHooks = Hooks{}

func (k Keeper) Hooks() Hooks { return Hooks{k} }

func (h Hooks) AfterHeaderWithValidCommit(ctx sdk.Context, txHash []byte, header *ibctmtypes.Header, isOnFork bool) {
	indexedHeader := types.IndexedHeader{
		ChainId:            header.Header.ChainID,
		Hash:               header.Header.LastCommitHash,
		Height:             uint64(header.Header.Height),
		BabylonBlockHeight: uint64(ctx.BlockHeight()),
		BabylonTxHash:      txHash,
	}
	if isOnFork {
		// insert header to fork index
		h.k.InsertForkHeader(ctx, indexedHeader.ChainId, &indexedHeader)
		// update the latest fork in chain info
		fork := h.k.GetForks(ctx, indexedHeader.ChainId, indexedHeader.Height)
		if fork == nil {
			err := sdkerrors.Wrapf(types.ErrForkNotFound, "fork at height %d should at least contain header %s", indexedHeader.Height, &indexedHeader.Hash)
			panic(err)
		}
		if err := h.k.UpdateLatestForks(ctx, indexedHeader.ChainId, fork); err != nil {
			panic(err)
		}
	} else {
		// insert header to canonical chain index
		h.k.InsertHeader(ctx, indexedHeader.ChainId, &indexedHeader)
		// update the latest canonical header in chain info
		if err := h.k.UpdateLatestHeader(ctx, indexedHeader.ChainId, &indexedHeader); err != nil {
			panic(err)
		}
	}
}
