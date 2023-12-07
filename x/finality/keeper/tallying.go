package keeper

import (
	"context"
	"fmt"

	"github.com/babylonchain/babylon/x/finality/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TallyBlocks tries to finalise all blocks that are non-finalised AND have a non-nil
// BTC validator set, from earliest to the latest.
//
// This function is invoked upon each `EndBlock` *after* the BTC staking protocol is activated
// It ensures that at height `h`, the ancestor chain `[activated_height, h-1]` contains either
// - finalised blocks (i.e., block with validator set AND QC of this validator set)
// - non-finalisable blocks (i.e., block with no active validator)
// but without block that has validator set AND does not receive QC
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

	// find all blocks that are non-finalised AND have validator set since max(activatedHeight, lastFinalizedHeight+1)
	// There are 4 different scenarios as follows
	// - has validators, non-finalised: tally and try to finalise
	// - does not have validators, non-finalised: non-finalisable, continue
	// - has validators, finalised: impossible to happen, panic
	// - does not have validators, finalised: impossible to happen, panic
	// After this for loop, the blocks since earliest activated height are either finalised or non-finalisable
	for i := startHeight; i <= uint64(sdkCtx.HeaderInfo().Height); i++ {
		ib, err := k.GetBlock(ctx, i)
		if err != nil {
			panic(err) // failing to get an existing block is a programming error
		}

		// get the validator set of this block
		valSet := k.BTCStakingKeeper.GetVotingPowerTable(ctx, ib.Height)

		if valSet != nil && !ib.Finalized {
			// has validators, non-finalised: tally and try to finalise the block
			voterBTCPKs := k.GetVoters(ctx, ib.Height)
			if tally(valSet, voterBTCPKs) {
				// if this block gets >2/3 votes, finalise it
				k.finalizeBlock(ctx, ib, voterBTCPKs)
			} else {
				// if not, then this block and all subsequent blocks should not be finalised
				// thus, we need to break here
				break
			}
		} else if valSet == nil && !ib.Finalized {
			// does not have validators, non-finalised: not finalisable,
			// increment the next height to finalise and continue
			k.setNextHeightToFinalize(ctx, ib.Height+1)
			continue
		} else if valSet != nil && ib.Finalized {
			// has validators and the block has finalised
			// this can only be a programming error
			panic(fmt.Errorf("block %d is finalized, but last finalized height in DB does not reach here", ib.Height))
		} else if valSet == nil && ib.Finalized {
			// does not have validators, finalised: impossible to happen, panic
			panic(fmt.Errorf("block %d is finalized, but does not have a validator set", ib.Height))
		}
	}
}

// finalizeBlock sets a block to be finalised in KVStore and distributes rewards to
// BTC validators and delegations
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
	// filter out voted BTC validators
	rdc.FilterVotedBTCVals(voterBTCPKs)
	// reward voted BTC validators
	k.IncentiveKeeper.RewardBTCStaking(ctx, block.Height, rdc)
	// remove reward distribution cache afterwards
	k.BTCStakingKeeper.RemoveRewardDistCache(ctx, block.Height)
}

// tally checks whether a block with the given validator set and votes reaches a quorum or not
func tally(valSet map[string]uint64, voterBTCPKs map[string]struct{}) bool {
	totalPower := uint64(0)
	votedPower := uint64(0)
	for pkStr, power := range valSet {
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
