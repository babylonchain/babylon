package keeper

import (
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
func (k Keeper) TallyBlocks(ctx sdk.Context) {
	// blocksToFinalize is the set of blocks to finalise within this tallying attempt
	blocksToFinalize := []*types.IndexedBlock{}
	// valSets is the BTC validator set at each height with a non-finalised block
	valSets := map[uint64]map[string]uint64{}

	activatedHeight, err := k.BTCStakingKeeper.GetBTCStakingActivatedHeight(ctx)
	if err != nil {
		// invoking TallyBlocks when BTC staking protocol is not activated is a programming error
		panic(fmt.Errorf("cannot tally a block when the BTC staking protocol hasn't been activated yet, current height: %v, activated height: %v",
			ctx.BlockHeight(), activatedHeight))
	}

	// find all blocks that are non-finalised AND have validator set, from latest to the earliest activated height
	// There are 4 different scenarios as follows
	// - has validators, non-finalised: can finalise, add to blocksToFinalize
	// - does not have validators, non-finalised: non-finalisable, skip
	// - has validators, finalised, break here
	// - does not have validators, finalised: impossible to happen, panic
	// After this for loop, the blocks since earliest activated height are either finalised or non-finalisable
	blockRevIter := k.blockStore(ctx).ReverseIterator(sdk.Uint64ToBigEndian(uint64(activatedHeight)), nil)
	for ; blockRevIter.Valid(); blockRevIter.Next() {
		// get the indexed block
		ibBytes := blockRevIter.Value()
		var ib types.IndexedBlock
		k.cdc.MustUnmarshal(ibBytes, &ib)
		// get the validator set of this block
		valSet := k.BTCStakingKeeper.GetVotingPowerTable(ctx, ib.Height)

		if valSet != nil && !ib.Finalized {
			// has validators, non-finalised: can finalise, add block and valset
			blocksToFinalize = append(blocksToFinalize, &ib)
			valSets[ib.Height] = valSet
		} else if valSet == nil && !ib.Finalized {
			// does not have validators, non-finalised: not finalisable, skip
			continue
		} else if valSet != nil && ib.Finalized {
			// has validators and the block has finalised
			// this means that the entire prefix has been finalised, break here
			break
		} else if valSet == nil && ib.Finalized {
			// does not have validators, finalised: impossible to happen, panic
			panic(fmt.Errorf("block %d is finalized, but does not have a validator set", ib.Height))
		}
	}
	// closing the iterator right now before finalising the finalisable blocks
	// this is to follow the contract at https://github.com/cosmos/cosmos-sdk/blob/v0.47.4/store/types/store.go#L239-L240
	blockRevIter.Close()

	// for each of these blocks from earliest to latest, tally the block w.r.t. existing votes
	for i := len(blocksToFinalize) - 1; i >= 0; i-- {
		blockToFinalize := blocksToFinalize[i]
		voterBTCPKs := k.GetVoters(ctx, blockToFinalize.Height)
		valSet := valSets[blockToFinalize.Height]
		if tally(valSet, voterBTCPKs) {
			// if this block gets >2/3 votes, finalise it
			k.finalizeBlock(ctx, blockToFinalize, voterBTCPKs)
		} else {
			// if not, then this block and all subsequent blocks should not be finalised
			// thus, we need to break here
			break
		}
	}
}

// finalizeBlock sets a block to be finalised in KVStore and distributes rewards to
// BTC validators and delegations
func (k Keeper) finalizeBlock(ctx sdk.Context, block *types.IndexedBlock, voterBTCPKs map[string]struct{}) {
	// set block to be finalised in KVStore
	block.Finalized = true
	k.SetBlock(ctx, block)
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
	return votedPower > totalPower*2/3
}
