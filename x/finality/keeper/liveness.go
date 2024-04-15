package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	"github.com/bits-and-blooms/bitset"
	sdk "github.com/cosmos/cosmos-sdk/types"

	btcstakingtypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/babylonchain/babylon/x/finality/types"
)

// HandleLiveness checks liveness of each finality provider from the active set and
// jail inactive ones
func (k Keeper) HandleLiveness(ctx context.Context) {
	prevHeight := sdk.UnwrapSDKContext(ctx).HeaderInfo().Height - 1
	if prevHeight <= 1 {
		return
	}

	// get finality providers from the active set
	activeFps, err := k.BTCStakingKeeper.GetActiveFinalityProviders(ctx, uint64(prevHeight))
	if err != nil {
		panic(fmt.Errorf("failed to get active finality providers: %w", err))
	}
	for _, fp := range activeFps {
		if err := k.HandleFinalityProviderLiveness(
			ctx,
			k.GetParams(ctx),
			prevHeight,
			fp,
			k.GetVoters(ctx, uint64(prevHeight)),
		); err != nil {
			panic(fmt.Errorf("failed to handle liveness of finality provider %s at height %d: %w",
				fp.BtcPk.MarshalHex(), prevHeight, err))
		}
	}

}

func (k Keeper) HandleFinalityProviderLiveness(
	ctx context.Context, params types.Params, height int64,
	fp *btcstakingtypes.FinalityProvider, voters map[string]struct{}) error {

	// TODO: add `jailed` to finality provider

	// don't update missed blocks when the finality provider is already jailed

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	pkBytes := fp.BtcPk.MustMarshal()

	// fetch signing info
	signInfo, err := k.FinalityProviderSigningTracker.Get(ctx, pkBytes)
	if err != nil {
		return err
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
	// NOTE: There is subtle different behavior between genesis finality providers and non-genesis finality providers.
	// A genesis finality provider will start at index 0, whereas a non-genesis finality provider's startHeight will be the block
	// they bonded on, but the first block they vote on will be one later. (And thus their first vote is at index 1)
	index := (height - signInfo.StartHeight) % signedBlocksWindow
	if signInfo.StartHeight > height {
		return fmt.Errorf("the finality provider has start height %d greater than the checking height %d",
			signInfo.StartHeight, height)
	}

	// determine if the finality provider signed the previous block
	previous, err := k.GetMissedBlockBitmapValue(ctx, pkBytes, index)
	if err != nil {
		return fmt.Errorf("failed to get the finality provider's bitmap value: %w", err)
	}

	modifiedSignInfo := false

	_, signed := voters[fp.BtcPk.MarshalHex()]
	switch {
	case !previous && !signed:
		// Bitmap value has changed from not missed to missed, so we flip the bit
		// and increment the counter.
		if err := k.SetMissedBlockBitmapValue(ctx, pkBytes, index, true); err != nil {
			return err
		}

		signInfo.MissedBlocksCounter++
		modifiedSignInfo = true

	case previous && signed:
		// Bitmap value has changed from missed to not missed, so we flip the bit
		// and decrement the counter.
		if err := k.SetMissedBlockBitmapValue(ctx, pkBytes, index, false); err != nil {
			return err
		}

		signInfo.MissedBlocksCounter--
		modifiedSignInfo = true

	default:
		// bitmap value at this index has not changed, no need to update counter
	}

	minSignedPerWindow := params.MinSignedPerWindowInt()

	if !signed {
		// TODO emit finality provider absent event
	}

	minHeight := signInfo.StartHeight + signedBlocksWindow
	maxMissed := signedBlocksWindow - minSignedPerWindow

	// if we are past the minimum height and the finality provider has missed too many blocks, punish them
	if height > minHeight && signInfo.MissedBlocksCounter > maxMissed {
		modifiedSignInfo = true
		// Downtime confirmed: jail the finality provider
		err = k.BTCStakingKeeper.JailFinalityProvider(ctx, pkBytes)
		if err != nil {
			return err
		}
		signInfo.JailedUntil = sdkCtx.HeaderInfo().Time.Add(params.JailDuration)

		// We need to reset the counter & bitmap so that the finality provider won't be
		// immediately jailed for downtime upon re-bonding.
		// We don't set the start height as this will get correctly set
		// once they bond again in the Afterfinality providerBonded hook!
		signInfo.MissedBlocksCounter = 0
		err = k.DeleteMissedBlockBitmap(ctx, pkBytes)
		if err != nil {
			return err
		}

		// TODO: emit jailing event
		k.Logger(sdkCtx).Info(
			"jailing finality provider due to liveness fault",
			"height", height,
			"finality_provider", fp.BtcPk.MarshalHex(),
			"min_height", minHeight,
			"threshold", minSignedPerWindow,
			"jailed_until", signInfo.JailedUntil,
		)
	}

	// Set the updated signing info
	if modifiedSignInfo {
		return k.FinalityProviderSigningTracker.Set(ctx, pkBytes, signInfo)
	}
	return nil
}

// GetMissedBlockBitmapValue returns true if a finality provider missed signing a block
// at the given index and false otherwise. The index provided is assumed to be
// the index in the range [0, SignedBlocksWindow), which represents the bitmap
// where each bit represents a height, and is determined by the finality provider's
// IndexOffset modulo SignedBlocksWindow. This index is used to fetch the chunk
// in the bitmap and the relative bit in that chunk.
func (k Keeper) GetMissedBlockBitmapValue(ctx context.Context, fpPk []byte, index int64) (bool, error) {
	// get the chunk or "word" in the logical bitmap
	chunkIndex := index / types.MissedBlockBitmapChunkSize

	bs := bitset.New(uint(types.MissedBlockBitmapChunkSize))
	chunk, err := k.getMissedBlockBitmapChunk(ctx, fpPk, chunkIndex)
	if err != nil {
		return false, errorsmod.Wrapf(err, "failed to get bitmap chunk; index: %d", index)
	}

	if chunk != nil {
		if err := bs.UnmarshalBinary(chunk); err != nil {
			return false, errorsmod.Wrapf(err, "failed to decode bitmap chunk; index: %d", index)
		}
	}

	// get the bit position in the chunk of the logical bitmap, where Test()
	// checks if the bit is set.
	bitIndex := index % types.MissedBlockBitmapChunkSize
	return bs.Test(uint(bitIndex)), nil
}

// SetMissedBlockBitmapValue sets, i.e. flips, a bit in the finality provider's missed
// block bitmap. When missed=true, the bit is set, otherwise it set to zero. The
// index provided is assumed to be the index in the range [0, SignedBlocksWindow),
// which represents the bitmap where each bit represents a height, and is
// determined by the finality provider's IndexOffset modulo SignedBlocksWindow. This
// index is used to fetch the chunk in the bitmap and the relative bit in that
// chunk.
func (k Keeper) SetMissedBlockBitmapValue(ctx context.Context, fpPk []byte, index int64, missed bool) error {
	// get the chunk or "word" in the logical bitmap
	chunkIndex := index / types.MissedBlockBitmapChunkSize

	bs := bitset.New(uint(types.MissedBlockBitmapChunkSize))
	chunk, err := k.getMissedBlockBitmapChunk(ctx, fpPk, chunkIndex)
	if err != nil {
		return errorsmod.Wrapf(err, "failed to get bitmap chunk; index: %d", index)
	}

	if chunk != nil {
		if err := bs.UnmarshalBinary(chunk); err != nil {
			return errorsmod.Wrapf(err, "failed to decode bitmap chunk; index: %d", index)
		}
	}

	// get the bit position in the chunk of the logical bitmap
	bitIndex := uint(index % types.MissedBlockBitmapChunkSize)
	if missed {
		bs.Set(bitIndex)
	} else {
		bs.Clear(bitIndex)
	}

	updatedChunk, err := bs.MarshalBinary()
	if err != nil {
		return errorsmod.Wrapf(err, "failed to encode bitmap chunk; index: %d", index)
	}

	return k.SetMissedBlockBitmapChunk(ctx, fpPk, chunkIndex, updatedChunk)
}

// DeleteMissedBlockBitmap removes a finality provider's missed block bitmap from state.
func (k Keeper) DeleteMissedBlockBitmap(ctx context.Context, addr sdk.ConsAddress) error {
	rng := collections.NewPrefixedPairRange[[]byte, uint64](addr.Bytes())
	return k.FinalityProviderMissedBlockBitmap.Clear(ctx, rng)
}

// getMissedBlockBitmapChunk gets the bitmap chunk at the given chunk index for
// a finality provider's missed block signing window.
func (k Keeper) getMissedBlockBitmapChunk(ctx context.Context, fpPk []byte, chunkIndex int64) ([]byte, error) {
	chunk, err := k.FinalityProviderMissedBlockBitmap.Get(ctx, collections.Join(fpPk, uint64(chunkIndex)))
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return nil, err
	}
	return chunk, nil
}

// SetMissedBlockBitmapChunk sets the bitmap chunk at the given chunk index for
// a finality provider's missed block signing window.
func (k Keeper) SetMissedBlockBitmapChunk(ctx context.Context, fpPk []byte, chunkIndex int64, chunk []byte) error {
	return k.FinalityProviderMissedBlockBitmap.Set(ctx, collections.Join(fpPk, uint64(chunkIndex)), chunk)
}
