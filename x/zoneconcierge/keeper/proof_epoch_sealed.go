package keeper

import (
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) ProveEpochSealed(ctx sdk.Context, epochNumber uint64) (*types.ProofEpochSealed, error) {
	var (
		proof *types.ProofEpochSealed = &types.ProofEpochSealed{}
		err   error                   = nil
	)

	proof.ValidatorSet, err = k.checkpointingKeeper.GetBLSPubKeySet(ctx, epochNumber)
	if err != nil {
		return nil, err
	}

	// TODO: record sealer header in epoch metadata

	// TODO: proof of inclusion for epoch metadata in sealer header

	// TODO: proof of inclusion for validator set in sealer header

	return proof, nil
}

func VerifyEpochSealed(ctx sdk.Context, epoch *epochingtypes.Epoch, rawCkpt *checkpointingtypes.RawCheckpoint, proof *types.ProofEpochSealed) error {
	panic("TODO: implement me")
}
