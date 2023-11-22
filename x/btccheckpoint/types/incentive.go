package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CheckpointAddressPair is a pair of (submitter, reporter) addresses of a checkpoint
// submission
type CheckpointAddressPair struct {
	Submitter sdk.AccAddress
	Reporter  sdk.AccAddress
}

func NewCheckpointAddressPair(addrs *CheckpointAddresses) (*CheckpointAddressPair, error) {
	var (
		submitter sdk.AccAddress
		reporter  sdk.AccAddress
	)
	if err := submitter.Unmarshal(addrs.Submitter); err != nil {
		return nil, fmt.Errorf("failed to unmarshal submitter address in bytes: %w", err)
	}
	if err := reporter.Unmarshal(addrs.Reporter); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reporter address in bytes: %w", err)
	}
	return &CheckpointAddressPair{
		Submitter: submitter,
		Reporter:  reporter,
	}, nil
}

// RewardDistInfo includes information necessary for incentive module to distribute rewards to
// a given finalised epoch
type RewardDistInfo struct {
	// Best is the address pair of the best checkpoint submission
	Best *CheckpointAddressPair
	// Others is a list of other address pairs
	Others []*CheckpointAddressPair
}

func NewRewardDistInfo(best *CheckpointAddressPair, others ...*CheckpointAddressPair) *RewardDistInfo {
	return &RewardDistInfo{
		Best:   best,
		Others: others,
	}
}
