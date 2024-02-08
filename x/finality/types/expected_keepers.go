package types

import (
	"context"

	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
)

type BTCStakingKeeper interface {
	GetFinalityProvider(ctx context.Context, fpBTCPK []byte) (*bstypes.FinalityProvider, error)
	HasFinalityProvider(ctx context.Context, fpBTCPK []byte) bool
	SlashFinalityProvider(ctx context.Context, fpBTCPK []byte) error
	GetVotingPower(ctx context.Context, fpBTCPK []byte, height uint64) uint64
	GetVotingPowerTable(ctx context.Context, height uint64) map[string]uint64
	GetBTCStakingActivatedHeight(ctx context.Context) (uint64, error)
	GetRewardDistCache(ctx context.Context, height uint64) (*bstypes.RewardDistCache, error)
	RemoveRewardDistCache(ctx context.Context, height uint64)
}

// IncentiveKeeper defines the expected interface needed to distribute rewards.
type IncentiveKeeper interface {
	RewardBTCStaking(ctx context.Context, height uint64, rdc *bstypes.RewardDistCache)
}
