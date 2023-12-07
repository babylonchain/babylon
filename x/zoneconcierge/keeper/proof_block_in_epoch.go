package keeper

import (
	"context"

	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	tmcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
)

func (k Keeper) ProveHeaderInEpoch(ctx context.Context, headerHeight uint64, epoch *epochingtypes.Epoch) (*tmcrypto.Proof, error) {
	return k.epochingKeeper.ProveAppHashInEpoch(ctx, headerHeight, epoch.EpochNumber)
}
