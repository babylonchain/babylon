package keeper

import (
	epochingkeeper "github.com/babylonchain/babylon/x/epoching/keeper"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	tmcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) ProveHeaderInEpoch(ctx sdk.Context, header *tmproto.Header, epoch *epochingtypes.Epoch) (*tmcrypto.Proof, error) {
	return k.epochingKeeper.ProveAppHashInEpoch(ctx, uint64(header.Height), epoch.EpochNumber)
}

func VerifyHeaderInEpoch(header *tmproto.Header, epoch *epochingtypes.Epoch, proof *tmcrypto.Proof) error {
	return epochingkeeper.VerifyAppHashInclusion(header.AppHash, epoch.AppHashRoot, proof)
}
