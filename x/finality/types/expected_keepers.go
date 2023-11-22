package types

import (
	"context"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
)

type BTCStakingKeeper interface {
	GetBTCValidator(ctx context.Context, valBTCPK []byte) (*bstypes.BTCValidator, error)
	HasBTCValidator(ctx context.Context, valBTCPK []byte) bool
	SlashBTCValidator(ctx context.Context, valBTCPK []byte) error
	GetVotingPower(ctx context.Context, valBTCPK []byte, height uint64) uint64
	GetVotingPowerTable(ctx context.Context, height uint64) map[string]uint64
	GetBTCStakingActivatedHeight(ctx context.Context) (uint64, error)
	RecordRewardDistCache(ctx context.Context)
	GetRewardDistCache(ctx context.Context, height uint64) (*bstypes.RewardDistCache, error)
	RemoveRewardDistCache(ctx context.Context, height uint64)
}

// IncentiveKeeper defines the expected interface needed to distribute rewards.
type IncentiveKeeper interface {
	RewardBTCStaking(ctx context.Context, height uint64, rdc *bstypes.RewardDistCache)
}
