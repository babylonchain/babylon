package keeper

import (
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
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
var _ epochingtypes.EpochingHooks = Hooks{}

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

func (h Hooks) AfterEpochEnds(ctx sdk.Context, epoch uint64) {
	// upon an epoch has ended, index the current chain info for each CZ
	for _, chainID := range h.k.GetAllChainIDs(ctx) {
		if err := h.k.RecordEpochChainInfo(ctx, chainID, epoch); err != nil {
			panic(err) // this happens only when the chain info does not exist, which is a programming error
		}
	}
}

func (h Hooks) AfterRawCheckpointFinalized(ctx sdk.Context, epoch uint64) error {
	// upon an epoch has been finalised, update the last finalised epoch
	h.k.setFinalizedEpoch(ctx, epoch)
	return nil
}

// Other unused hooks

func (h Hooks) AfterBlsKeyRegistered(ctx sdk.Context, valAddr sdk.ValAddress) error     { return nil }
func (h Hooks) AfterRawCheckpointConfirmed(ctx sdk.Context, epoch uint64) error         { return nil }
func (h Hooks) AfterEpochBegins(ctx sdk.Context, epoch uint64)                          {}
func (h Hooks) BeforeSlashThreshold(ctx sdk.Context, valSet epochingtypes.ValidatorSet) {}
