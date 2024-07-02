package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/babylonchain/babylon/types"
	finalitytypes "github.com/babylonchain/babylon/x/finality/types"
)

// HandleLiveness handles liveness of each active finality provider for a given height
// including identifying inactive finality providers and applying punishment
func (k Keeper) HandleLiveness(ctx context.Context, height int64) {
	// get all the active finality providers for the height
	fpSet := k.BTCStakingKeeper.GetVotingPowerTable(ctx, uint64(height))
	// get all the voters for the height
	voterBTCPKs := k.GetVoters(ctx, uint64(height))

	// Iterate over all the finality providers which *should* have signed this block
	// store whether or not they have actually signed it, identify inactive
	// ones, and apply punishment
	for fpPkHex := range fpSet {
		fpPk, err := types.NewBIP340PubKeyFromHex(fpPkHex)
		if err != nil {
			panic(fmt.Errorf("invalid finality provider public key %s: %w", fpPkHex, err))
		}

		_, ok := voterBTCPKs[fpPkHex]
		missed := !ok

		err = k.HandleFinalityProviderLiveness(ctx, fpPk, missed, height)
		if err != nil {
			panic(fmt.Errorf("failed to handle liveness of finality provider %s: %w", fpPkHex, err))
		}
	}
}

// HandleFinalityProviderLiveness updates the voting history of the given finality provider and
// jails the finality provider if the number of missed block is reached to the threshold in a
// sliding window
func (k Keeper) HandleFinalityProviderLiveness(ctx context.Context, fpPk *types.BIP340PubKey, missed bool, height int64) error {
	params := k.GetParams(ctx)
	// don't update missed blocks when finality provider is already jailed
	fp, err := k.BTCStakingKeeper.GetFinalityProvider(ctx, fpPk.MustMarshal())
	if err != nil {
		return err
	}

	if fp.IsJailed() || fp.IsSlashed() {
		return nil
	}

	updated, signInfo, err := k.updateSigningInfo(ctx, fpPk, missed, height)
	if err != nil {
		return err
	}

	signedBlocksWindow := params.SignedBlocksWindow
	minSignedPerWindow := params.MinSignedPerWindowInt()

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if missed {
		// TODO emit event

		k.Logger(sdkCtx).Debug(
			"absent finality provider",
			"height", height,
			"public_key", fpPk.MarshalHex(),
			"missed", signInfo.MissedBlocksCounter,
			"threshold", minSignedPerWindow,
		)
	}

	minHeight := signInfo.StartHeight + signedBlocksWindow
	maxMissed := signedBlocksWindow - minSignedPerWindow

	// if we are past the minimum height and the finality provider has missed too many blocks, punish them
	if height > minHeight && signInfo.MissedBlocksCounter > maxMissed {
		updated = true
		// TODO emit event

		// Inactivity detected
		err = k.hooks.AfterInactiveFinalityProviderDetected(ctx, fpPk)
		if err != nil {
			return err
		}

		signInfo.JailedUntil = sdkCtx.HeaderInfo().Time.Add(params.JailDuration)

		// We need to reset the counter & bitmap so that the finality provider won't be
		// immediately jailed upon re-bonding.
		// We don't set the start height as this will get correctly set
		// once they bond again in the AfterFinalityProviderActivated hook
		signInfo.MissedBlocksCounter = 0
		err = k.DeleteMissedBlockBitmap(ctx, fpPk)
		if err != nil {
			return err
		}

		k.Logger(sdkCtx).Info(
			"jailing validator due to liveness fault",
			"height", height,
			"public_key", fpPk.MarshalHex(),
			"min_height", minHeight,
			"threshold", minSignedPerWindow,
			"jailed_until", signInfo.JailedUntil,
		)
	}

	// Set the updated signing info
	if updated {
		return k.FinalityProviderSigningTracker.Set(ctx, fpPk.MustMarshal(), *signInfo)
	}

	return nil
}

func (k Keeper) updateSigningInfo(
	ctx context.Context,
	fpPk *types.BIP340PubKey,
	missed bool,
	height int64,
) (bool, *finalitytypes.FinalityProviderSigningInfo, error) {
	params := k.GetParams(ctx)
	// fetch signing info
	signInfo, err := k.FinalityProviderSigningTracker.Get(ctx, fpPk.MustMarshal())
	if err != nil {
		return false, nil, err
	}

	signedBlocksWindow := params.SignedBlocksWindow

	// Compute the relative index, so we count the blocks the finality provider *should*
	// have signed. We will also use the 0-value default signing info if not present.
	// The index is in the range [0, SignedBlocksWindow)
	// and is used to see if a finality provider signed a block at the given height, which
	// is represented by a bit in the bitmap.
	// The finality provider start height should get mapped to index 0, so we computed index as:
	// (height - startHeight) % signedBlocksWindow
	//
	// NOTE: There is subtle different behavior between genesis finality provider and non-genesis
	// finality providers.
	// A genesis finality provider will start at index 0, whereas a non-genesis finality provider's
	// startHeight will be the block they become active for, but the first block they vote on will be
	// one later. (And thus their first vote is at index 1)
	if signInfo.StartHeight > height {
		return false, nil, fmt.Errorf("invalid state, the finality provider signing info has start height %d, which is greater than the current height %d",
			signInfo.StartHeight, height)
	}
	index := (height - signInfo.StartHeight) % signedBlocksWindow

	// determine if the finality provider signed the previous block
	previous, err := k.GetMissedBlockBitmapValue(ctx, fpPk, index)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get the finality provider's bitmap value: %w", err)
	}

	modifiedSignInfo := false
	switch {
	case !previous && missed:
		// Bitmap value has changed from not missed to missed, so we flip the bit
		// and increment the counter.
		if err := k.SetMissedBlockBitmapValue(ctx, fpPk, index, true); err != nil {
			return false, nil, err
		}

		signInfo.MissedBlocksCounter++
		modifiedSignInfo = true

	case previous && !missed:
		// Bitmap value has changed from missed to not missed, so we flip the bit
		// and decrement the counter.
		if err := k.SetMissedBlockBitmapValue(ctx, fpPk, index, false); err != nil {
			return false, nil, err
		}

		signInfo.MissedBlocksCounter--
		modifiedSignInfo = true

	default:
		// bitmap value at this index has not changed, no need to update counter
	}

	return modifiedSignInfo, &signInfo, nil
}
