package keeper

import (
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
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
var _ checkpointingtypes.CheckpointingHooks = Hooks{}

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
		if err := h.k.InsertForkHeader(ctx, indexedHeader.ChainId, &indexedHeader); err != nil {
			panic(err)
		}
		// update the latest fork in chain info
		if err := h.k.UpdateLatestForkHeader(ctx, indexedHeader.ChainId, &indexedHeader); err != nil {
			panic(err)
		}
	} else {
		// insert header to canonical chain index
		if err := h.k.InsertHeader(ctx, indexedHeader.ChainId, &indexedHeader); err != nil {
			panic(err)
		}
		// update the latest canonical header in chain info
		if err := h.k.UpdateLatestHeader(ctx, indexedHeader.ChainId, &indexedHeader); err != nil {
			panic(err)
		}
	}
}

func (h Hooks) AfterRawCheckpointFinalized(ctx sdk.Context, epoch uint64) error {
	// TODO: xxx
	return nil
}

// Other unused hooks

func (h Hooks) AfterBlsKeyRegistered(ctx sdk.Context, valAddr sdk.ValAddress) error {
	return nil
}

func (h Hooks) AfterRawCheckpointConfirmed(ctx sdk.Context, epoch uint64) error {
	return nil
}
