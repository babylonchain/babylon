package keeper

import (
	"fmt"

	"github.com/babylonchain/babylon/crypto/bls12381"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
)

func (k Keeper) ProveEpochSealed(ctx sdk.Context, epochNumber uint64) (*types.ProofEpochSealed, error) {
	var (
		proof *types.ProofEpochSealed = &types.ProofEpochSealed{}
		err   error
	)

	// get the validator set of the sealed epoch
	proof.ValidatorSet, err = k.checkpointingKeeper.GetBLSPubKeySet(ctx, epochNumber)
	if err != nil {
		return nil, err
	}

	// TODO: proof of inclusion for epoch metadata in sealer header
	proof.ProofEpochInfo = &tmcrypto.ProofOps{}

	// TODO: proof of inclusion for validator set in sealer header
	proof.ProofEpochValSet = &tmcrypto.ProofOps{}

	return proof, nil
}

// VerifyEpochSealed verifies that the given `epoch` is sealed by the `rawCkpt` by using the given `proof`
// The verification rules include:
// - basic sanity checks
// - The raw checkpoint's last_commit_hash is same as in the header of the sealer epoch
// - More than 1/3 (in voting power) validators in the validator set of this epoch have signed last_commit_hash of the sealer epoch
// - The epoch medatata is committed to the app_hash of the sealer epoch
// - The validator set is committed to the app_hash of the sealer epoch
func VerifyEpochSealed(epoch *epochingtypes.Epoch, rawCkpt *checkpointingtypes.RawCheckpoint, proof *types.ProofEpochSealed) error {
	// nil check
	if epoch == nil {
		return fmt.Errorf("epoch is nil")
	} else if rawCkpt == nil {
		return fmt.Errorf("rawCkpt is nil")
	} else if proof == nil {
		return fmt.Errorf("proof is nil")
	}

	// sanity check
	if err := epoch.ValidateBasic(); err != nil {
		return err
	} else if err := rawCkpt.ValidateBasic(); err != nil {
		return err
	} else if err = proof.ValidateBasic(); err != nil {
		return err
	}

	// TODO: Ensure The epoch medatata is committed to the app_hash of the sealer header
	// TODO: Ensure The validator set is committed to the app_hash of the sealer header

	// ensure epoch number is same in epoch and rawCkpt
	if epoch.EpochNumber != rawCkpt.EpochNum {
		return fmt.Errorf("epoch.EpochNumber (%d) is not equal to rawCkpt.EpochNum (%d)", epoch.EpochNumber, rawCkpt.EpochNum)
	}

	// ensure the raw checkpoint's last_commit_hash is same as in the header of the sealer header
	// NOTE: since this proof is assembled by a Babylon node who has verified the checkpoint,
	// the two lch values should always be the same, otherwise this Babylon node is malicious.
	// This is different from the checkpoint verification rules in checkpointing,
	// where a checkpoint with valid BLS multisig but different lch signals a dishonest majority equivocation.
	lchInCkpt := rawCkpt.LastCommitHash
	lchInSealerHeader := checkpointingtypes.LastCommitHash(epoch.SealerHeader.LastCommitHash)
	if !lchInCkpt.Equal(lchInSealerHeader) {
		return fmt.Errorf("LastCommitHash is not same in rawCkpt (%s) and epoch's SealerHeader (%s)", lchInCkpt.String(), lchInSealerHeader.String())
	}

	/*
		Ensure more than 1/3 (in voting power) validators of this epoch have signed (epoch_num || last_commit_hash) in the raw checkpoint
	*/
	valSet := checkpointingtypes.ValidatorWithBlsKeySet{ValSet: proof.ValidatorSet}
	// filter validator set that contributes to the signature
	signerSet, signerSetPower, err := valSet.FindSubsetWithPowerSum(rawCkpt.Bitmap)
	if err != nil {
		return err
	}
	// ensure the signerSet has > 1/3 voting power
	if signerSetPower <= valSet.GetTotalPower()*1/3 {
		return fmt.Errorf("the BLS signature involves insufficient voting power")
	}
	// verify BLS multisig
	signedMsgBytes := rawCkpt.SignedMsg()
	ok, err := bls12381.VerifyMultiSig(*rawCkpt.BlsMultiSig, signerSet.GetBLSKeySet(), signedMsgBytes)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("BLS signature does not match the public key")
	}

	return nil
}
