package keeper

import (
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// proveFinalizedChainInfo generates proofs that a chainInfo has been finalised by the given epoch with epochInfo
// It includes proofTxInBlock, proofHeaderInEpoch, proofEpochSealed and proofEpochSubmitted
// The proofs can be verified by a verifier with access to a BTC and Babylon light client
// CONTRACT: this is only a private helper function for simplifying the implementation of RPC calls
func (k Keeper) proveFinalizedChainInfo(
	ctx sdk.Context,
	chainInfo *types.ChainInfo,
	epochInfo *epochingtypes.Epoch,
	bestSubmissionKey *btcctypes.SubmissionKey,
) (*types.ProofFinalizedChainInfo, error) {
	var (
		err   error
		proof = &types.ProofFinalizedChainInfo{}
	)

	// Proof that the Babylon tx is in block
	proof.ProofTxInBlock, err = k.ProveTxInBlock(ctx, chainInfo.LatestHeader.BabylonTxHash)
	if err != nil {
		return nil, err
	}

	// proof that the block is in this epoch
	proof.ProofHeaderInEpoch, err = k.ProveHeaderInEpoch(ctx, chainInfo.LatestHeader.BabylonHeader, epochInfo)
	if err != nil {
		return nil, err
	}

	// proof that the epoch is sealed
	proof.ProofEpochSealed, err = k.ProveEpochSealed(ctx, epochInfo.EpochNumber)
	if err != nil {
		return nil, err
	}

	// proof that the epoch's checkpoint is submitted to BTC
	// i.e., the two `TransactionInfo`s for the checkpoint
	proof.ProofEpochSubmitted, err = k.ProveEpochSubmitted(ctx, bestSubmissionKey)
	if err != nil {
		// The only error in ProveEpochSubmitted is the nil bestSubmission.
		// Since the epoch w.r.t. the bestSubmissionKey is finalised, this
		// can only be a programming error, so we should panic here.
		panic(err)
	}

	return proof, nil
}

// TODO: implement a standalone verifier VerifyFinalizedChainInfo that
// verifies whether a chainInfo is finalised or not, with access to
// Bitcoin and Babylon light clients
