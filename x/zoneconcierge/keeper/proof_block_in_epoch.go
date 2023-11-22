package keeper

import (
	"context"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	tmcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

func (k Keeper) ProveHeaderInEpoch(ctx context.Context, header *tmproto.Header, epoch *epochingtypes.Epoch) (*tmcrypto.Proof, error) {
	return k.epochingKeeper.ProveAppHashInEpoch(ctx, uint64(header.Height), epoch.EpochNumber)
}
