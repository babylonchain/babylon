package keeper

import (
	"context"
	"fmt"

	"github.com/babylonchain/babylon/x/finality/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TallyBlocks tries to finalise all blocks that are non-finalised AND have a non-nil
// finality provider set, from earliest to the latest.
//
// This function is invoked upon each `EndBlock` *after* the BTC staking protocol is activated
// It ensures that at height `h`, the ancestor chain `[activated_height, h-1]` contains either
// - finalised blocks (i.e., block with finality provider set AND QC of this finality provider set)
// - non-finalisable blocks (i.e., block with no active finality providers)
// but without block that has finality providers set AND does not receive QC
func (k Keeper) TallyBlocks(ctx context.Context) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	activatedHeight, err := k.BTCStakingKeeper.GetBTCStakingActivatedHeight(ctx)
	if err != nil {
		// invoking TallyBlocks when BTC staking protocol is not activated is a programming error
		panic(fmt.Errorf("cannot tally a block when the BTC staking protocol hasn't been activated yet, current height: %v, activated height: %v",
			sdkCtx.HeaderInfo().Height, activatedHeight))
	}

	// start finalising blocks since max(activatedHeight, nextHeightToFinalize)
	startHeight := k.getNextHeightToFinalize(ctx)
	if startHeight < activatedHeight {
		startHeight = activatedHeight
	}

	// find all blocks that are non-finalised AND have finality provider set since max(activatedHeight, lastFinalizedHeight+1)
	// There are 4 different scenarios as follows
	// - has finality providers, non-finalised: tally and try to finalise
	// - does not have finality providers, non-finalised: non-finalisable, continue
	// - has finality providers, finalised: impossible to happen, panic
	// - does not have finality providers, finalised: impossible to happen, panic
	// After this for loop, the blocks since earliest activated height are either finalised or non-finalisable
	for i := startHeight; i <= uint64(sdkCtx.HeaderInfo().Height); i++ {
		ib, err := k.GetBlock(ctx, i)
		if err != nil {
			panic(err) // failing to get an existing block is a programming error
		}

		// get the finality provider set of this block
		fpSet := k.BTCStakingKeeper.GetVotingPowerTable(ctx, ib.Height)

		if fpSet != nil && !ib.Finalized {
			// has finality providers, non-finalised: tally and try to finalise the block
			voterBTCPKs := k.GetVoters(ctx, ib.Height)
			if tally(fpSet, voterBTCPKs) {
				// if this block gets >2/3 votes, finalise it
				k.finalizeBlock(ctx, ib, voterBTCPKs)
			} else {
				// if not, then this block and all subsequent blocks should not be finalised
				// thus, we need to break here
				break
			}
		} else if fpSet == nil && !ib.Finalized {
			// does not have finality providers, non-finalised: not finalisable,
			// increment the next height to finalise and continue
			k.setNextHeightToFinalize(ctx, ib.Height+1)
			continue
		} else if fpSet != nil && ib.Finalized {
			// has finality providers and the block has finalised
			// this can only be a programming error
			panic(fmt.Errorf("block %d is finalized, but last finalized height in DB does not reach here", ib.Height))
		} else if fpSet == nil && ib.Finalized {
			// does not have finality providers, finalised: impossible to happen, panic
			panic(fmt.Errorf("block %d is finalized, but does not have a finality provider set", ib.Height))
		}
	}
}

// finalizeBlock sets a block to be finalised in KVStore and distributes rewards to
// finality providers and delegations
func (k Keeper) finalizeBlock(ctx context.Context, block *types.IndexedBlock, voterBTCPKs map[string]struct{}) {
	// set block to be finalised in KVStore
	block.Finalized = true
	k.SetBlock(ctx, block)
	// set next height to finalise as height+1
	k.setNextHeightToFinalize(ctx, block.Height+1)
	// distribute rewards to BTC staking stakeholders w.r.t. the reward distribution cache
	rdc, err := k.BTCStakingKeeper.GetRewardDistCache(ctx, block.Height)
	if err != nil {
		// failing to get a reward distribution cache before distributing reward is a programming error
		panic(err)
	}
	// filter out voted finality providers
	rdc.FilterVotedFinalityProviders(voterBTCPKs)
	// reward voted finality providers
	k.IncentiveKeeper.RewardBTCStaking(ctx, block.Height, rdc)
	// remove reward distribution cache afterwards
	k.BTCStakingKeeper.RemoveRewardDistCache(ctx, block.Height)
	// record the last finalized height metric
	types.RecordLastFinalizedHeight(block.Height)
}

// tally checks whether a block with the given finality provider set and votes reaches a quorum or not
func tally(fpSet map[string]uint64, voterBTCPKs map[string]struct{}) bool {
	totalPower := uint64(0)
	votedPower := uint64(0)
	for pkStr, power := range fpSet {
		totalPower += power
		if _, ok := voterBTCPKs[pkStr]; ok {
			votedPower += power
		}
	}
	return votedPower*3 > totalPower*2
}

// setNextHeightToFinalize sets the next height to finalise as the given height
func (k Keeper) setNextHeightToFinalize(ctx context.Context, height uint64) {
	store := k.storeService.OpenKVStore(ctx)
	heightBytes := sdk.Uint64ToBigEndian(height)
	if err := store.Set(types.NextHeightToFinalizeKey, heightBytes); err != nil {
		panic(err)
	}
}

// getNextHeightToFinalize gets the next height to finalise
func (k Keeper) getNextHeightToFinalize(ctx context.Context) uint64 {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.NextHeightToFinalizeKey)
	if err != nil {
		panic(err)
	}
	if bz == nil {
		return 0
	}
	height := sdk.BigEndianToUint64(bz)
	return height
}
