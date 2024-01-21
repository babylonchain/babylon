package keeper

import (
	"context"
	"fmt"

	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	tmcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
)

func (k Keeper) ProveCZHeaderInEpoch(_ context.Context, header *types.IndexedHeader, epoch *epochingtypes.Epoch) (*tmcrypto.ProofOps, error) {
	czHeaderKey := types.GetCZHeaderKey(header.ChainId, header.Height)
	_, _, proof, err := k.QueryStore(types.StoreKey, czHeaderKey, int64(epoch.GetSealerBlockHeight()))
	if err != nil {
		return nil, err
	}

	return proof, nil
}

func (k Keeper) ProveEpochInfo(epoch *epochingtypes.Epoch) (*tmcrypto.ProofOps, error) {
	epochInfoKey := types.GetEpochInfoKey(epoch.EpochNumber)
	_, _, proof, err := k.QueryStore(epochingtypes.StoreKey, epochInfoKey, int64(epoch.GetSealerBlockHeight()))
	if err != nil {
		return nil, err
	}

	return proof, nil
}

func (k Keeper) ProveValSet(epoch *epochingtypes.Epoch) (*tmcrypto.ProofOps, error) {
	valSetKey := types.GetValSetKey(epoch.EpochNumber)
	_, _, proof, err := k.QueryStore(checkpointingtypes.StoreKey, valSetKey, int64(epoch.GetSealerBlockHeight()))
	if err != nil {
		return nil, err
	}
	return proof, nil
}

// ProveEpochSealed proves an epoch has been sealed, i.e.,
// - the epoch's validator set has a valid multisig over the sealer header
// - the epoch's validator set is committed to the sealer header's app_hash
// - the epoch's metadata is committed to the sealer header's app_hash
func (k Keeper) ProveEpochSealed(ctx context.Context, epochNumber uint64) (*types.ProofEpochSealed, error) {
	var (
		proof = &types.ProofEpochSealed{}
		err   error
	)

	// get the validator set of the sealed epoch
	proof.ValidatorSet, err = k.checkpointingKeeper.GetBLSPubKeySet(ctx, epochNumber)
	if err != nil {
		return nil, err
	}

	// get sealer header and the query height
	epoch, err := k.epochingKeeper.GetHistoricalEpoch(ctx, epochNumber)
	if err != nil {
		return nil, err
	}

	// proof of inclusion for epoch metadata in sealer header
	proof.ProofEpochInfo, err = k.ProveEpochInfo(epoch)
	if err != nil {
		return nil, err
	}

	// proof of inclusion for validator set in sealer header
	proof.ProofEpochValSet, err = k.ProveValSet(epoch)
	if err != nil {
		return nil, err
	}

	return proof, nil
}

// ProveEpochSubmitted generates proof that the epoch's checkpoint is submitted to BTC
// i.e., the two `TransactionInfo`s for the checkpoint
func (k Keeper) ProveEpochSubmitted(ctx context.Context, sk *btcctypes.SubmissionKey) ([]*btcctypes.TransactionInfo, error) {
	bestSubmissionData := k.btccKeeper.GetSubmissionData(ctx, *sk)
	if bestSubmissionData == nil {
		return nil, fmt.Errorf("the best submission key for epoch %d has no submission data", bestSubmissionData.Epoch)
	}
	return bestSubmissionData.TxsInfo, nil
}

// proveFinalizedChainInfo generates proofs that a chainInfo has been finalised by the given epoch with epochInfo
// It includes proofTxInBlock, proofHeaderInEpoch, proofEpochSealed and proofEpochSubmitted
// The proofs can be verified by a verifier with access to a BTC and Babylon light client
// CONTRACT: this is only a private helper function for simplifying the implementation of RPC calls
func (k Keeper) proveFinalizedChainInfo(
	ctx context.Context,
	chainInfo *types.ChainInfo,
	epochInfo *epochingtypes.Epoch,
	bestSubmissionKey *btcctypes.SubmissionKey,
) (*types.ProofFinalizedChainInfo, error) {
	var (
		err   error
		proof = &types.ProofFinalizedChainInfo{}
	)

	// Proof that the CZ header is timestamped in epoch
	proof.ProofCzHeaderInEpoch, err = k.ProveCZHeaderInEpoch(ctx, chainInfo.LatestHeader, epochInfo)
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
