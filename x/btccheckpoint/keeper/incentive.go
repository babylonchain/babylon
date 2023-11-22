package keeper

import (
	"context"
	"github.com/babylonchain/babylon/x/btccheckpoint/types"
)

// rewardBTCTimestamping finds the (submitter, reporter) pairs of all submissions at the
// given finalised epoch according to the given epoch data, then distribute rewards to them
// by invoking the incentive module
func (k Keeper) rewardBTCTimestamping(ctx context.Context, epoch uint64, ed *types.EpochData, bestIdx int) {
	var (
		bestSubmissionAddrs  *types.CheckpointAddressPair
		otherSubmissionAddrs []*types.CheckpointAddressPair
	)

	// iterate over all submission keys to find all submission addresses, including the best one
	for i, sk := range ed.Keys {
		// retrieve submission data, including vigilante addresses
		submissionData := k.GetSubmissionData(ctx, *sk)
		if submissionData == nil {
			// ignore nil submission data for whatever reason
			continue
		}

		// get vigilante addresses of this submission
		submissionAddrs, err := types.NewCheckpointAddressPair(submissionData.VigilanteAddresses)
		if err != nil {
			// failing to unmarshal checkpoint address pair in KVStore is a programming error
			panic(err)
		}

		// assign to best submission or append to other submission according to best submission index
		if i == bestIdx {
			bestSubmissionAddrs = submissionAddrs
		} else {
			otherSubmissionAddrs = append(otherSubmissionAddrs, submissionAddrs)
		}
	}

	// construct reward distribution information and invoke incentive module to distribute rewards
	rewardDistInfo := types.NewRewardDistInfo(bestSubmissionAddrs, otherSubmissionAddrs...)
	k.incentiveKeeper.RewardBTCTimestamping(ctx, epoch, rewardDistInfo)
}
