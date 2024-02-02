package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/x/zoneconcierge/types"
)

// HandleHeaderWithValidCommit handles a CZ header with a valid QC
func (k Keeper) HandleHeaderWithValidCommit(ctx context.Context, txHash []byte, header *types.HeaderInfo, isOnFork bool) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	babylonHeader := sdkCtx.BlockHeader()
	indexedHeader := types.IndexedHeader{
		ChainId:             header.ChainId,
		Hash:                header.AppHash,
		Height:              header.Height,
		Time:                &header.Time,
		BabylonHeaderHash:   babylonHeader.AppHash,
		BabylonHeaderHeight: uint64(babylonHeader.Height),
		BabylonEpoch:        k.GetEpoch(ctx).EpochNumber,
		BabylonTxHash:       txHash,
	}

	k.Logger(sdkCtx).Debug("found new IBC header", "header", indexedHeader)

	var (
		chainInfo *types.ChainInfo
		err       error
	)
	if !k.HasChainInfo(ctx, indexedHeader.ChainId) {
		// chain info does not exist yet, initialise chain info for this chain
		chainInfo, err = k.InitChainInfo(ctx, indexedHeader.ChainId)
		if err != nil {
			panic(fmt.Errorf("failed to initialize chain info of %s: %w", indexedHeader.ChainId, err))
		}
	} else {
		// get chain info
		chainInfo, err = k.GetChainInfo(ctx, indexedHeader.ChainId)
		if err != nil {
			panic(fmt.Errorf("failed to get chain info of %s: %w", indexedHeader.ChainId, err))
		}
	}

	if isOnFork {
		// insert header to fork index
		if err := k.insertForkHeader(ctx, indexedHeader.ChainId, &indexedHeader); err != nil {
			panic(err)
		}
		// update the latest fork in chain info
		if err := k.tryToUpdateLatestForkHeader(ctx, indexedHeader.ChainId, &indexedHeader); err != nil {
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
		if err := k.insertHeader(ctx, indexedHeader.ChainId, &indexedHeader); err != nil {
			panic(err)
		}
		// update the latest canonical header in chain info
		if err := k.updateLatestHeader(ctx, indexedHeader.ChainId, &indexedHeader); err != nil {
			panic(err)
		}
	}
}
