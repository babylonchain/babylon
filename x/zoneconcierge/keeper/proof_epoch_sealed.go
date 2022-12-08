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

	// get the validator set of the sealed epoch
	proof.ValidatorSet, err = k.checkpointingKeeper.GetBLSPubKeySet(ctx, epochNumber)
	if err != nil {
		return nil, err
	}

	// TODO: proof of inclusion for epoch metadata in sealer header

	// TODO: proof of inclusion for validator set in sealer header

	return proof, nil
}

// VerifyEpochSealed verifies that the given `epoch` is sealed by the `rawCkpt` by using the given `proof`
// The verification rules include:
// - The raw checkpoint's last_commit_hash is same as in the header of the sealer epoch
// - More than 1/3 (in voting power) validators in the validator set of this epoch have signed last_commit_hash of the sealer epoch
// - The epoch medatata is committed to the app_hash of the sealer epoch
// - The validator set is committed to the app_hash of the sealer epoch
func VerifyEpochSealed(ctx sdk.Context, epoch *epochingtypes.Epoch, rawCkpt *checkpointingtypes.RawCheckpoint, proof *types.ProofEpochSealed) error {
	// TODO: Ensure The raw checkpoint's last_commit_hash is same as in the header of the sealer epoch
	// TODO: Ensure More than 1/3 (in voting power) validators in the validator set of this epoch have signed last_commit_hash of the sealer epoch
	// TODO: Ensure The epoch medatata is committed to the app_hash of the sealer epoch
	// TODO: Ensure The validator set is committed to the app_hash of the sealer epoch
	panic("TODO: implement me")
}
