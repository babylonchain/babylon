package keeper

import (
	"fmt"

	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

// AfterHeaderWithValidCommit is triggered upon each CZ header with a valid QC
func (h Hooks) AfterHeaderWithValidCommit(ctx sdk.Context, txHash []byte, header *ibctmtypes.Header, isOnFork bool) {
	babylonHeader := ctx.BlockHeader()
	indexedHeader := types.IndexedHeader{
		ChainId:       header.Header.ChainID,
		Hash:          header.Header.LastCommitHash,
		Height:        uint64(header.Header.Height),
		BabylonHeader: &babylonHeader,
		BabylonEpoch:  h.k.GetEpoch(ctx).EpochNumber,
		BabylonTxHash: txHash,
	}

	// initialise chain info if not exist
	chainInfo, err := h.k.GetChainInfo(ctx, indexedHeader.ChainId)
	if err != nil {
		if sdkerrors.IsOf(err, types.ErrEpochChainInfoNotFound) {
			// chain info does not exist yet, initialise chain info for this chain
			chainInfo, err = h.k.InitChainInfo(ctx, indexedHeader.ChainId)
			if err != nil {
				panic(fmt.Errorf("failed to initialize chain info of %s: %w", indexedHeader.ChainId, err))
			}
		} else {
			panic(fmt.Errorf("failed to get chain info of %s: %w", indexedHeader.ChainId, err))
		}
	}

	if isOnFork {
		// insert header to fork index
		if err := h.k.insertForkHeader(ctx, indexedHeader.ChainId, &indexedHeader); err != nil {
			panic(err)
		}
		// update the latest fork in chain info
		if err := h.k.tryToUpdateLatestForkHeader(ctx, indexedHeader.ChainId, &indexedHeader); err != nil {
			panic(err)
		}
	} else {
		// ensure the header is the latest one, otherwise ignore it
		// NOTE: while an old header is considered acceptable in IBC-Go (see Case_valid_past_update), but
		// ZoneConcierge should not checkpoint it since Babylon requires monotonic checkpointing
		if !chainInfo.IsLatestHeader(&indexedHeader) {
			return
		}

		// insert header to canonical chain index
		if err := h.k.insertHeader(ctx, indexedHeader.ChainId, &indexedHeader); err != nil {
			panic(err)
		}
		// update the latest canonical header in chain info
		if err := h.k.updateLatestHeader(ctx, indexedHeader.ChainId, &indexedHeader); err != nil {
			panic(err)
		}
	}
}

// AfterEpochEnds is triggered upon an epoch has ended
func (h Hooks) AfterEpochEnds(ctx sdk.Context, epoch uint64) {
	// upon an epoch has ended, index the current chain info for each CZ
	for _, chainID := range h.k.GetAllChainIDs(ctx) {
		h.k.recordEpochChainInfo(ctx, chainID, epoch)
	}
}

// AfterRawCheckpointFinalized is triggered upon an epoch has been finalised
func (h Hooks) AfterRawCheckpointFinalized(ctx sdk.Context, epoch uint64) error {
	// upon an epoch has been finalised, update the last finalised epoch
	h.k.setFinalizedEpoch(ctx, epoch)
	return nil
}

// Other unused hooks

func (h Hooks) AfterBlsKeyRegistered(ctx sdk.Context, valAddr sdk.ValAddress) error { return nil }
func (h Hooks) AfterRawCheckpointConfirmed(ctx sdk.Context, epoch uint64) error     { return nil }

func (h Hooks) AfterRawCheckpointForgotten(ctx sdk.Context, ckpt *checkpointingtypes.RawCheckpoint) error {
	return nil
}
func (h Hooks) AfterRawCheckpointBlsSigVerified(ctx sdk.Context, ckpt *checkpointingtypes.RawCheckpoint) error {
	return nil
}
func (h Hooks) AfterEpochBegins(ctx sdk.Context, epoch uint64)                          {}
func (h Hooks) BeforeSlashThreshold(ctx sdk.Context, valSet epochingtypes.ValidatorSet) {}
