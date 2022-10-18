package keeper

import (
	"context"

	btypes "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btccheckpoint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type msgServer struct {
	k Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper}
}

// onTheSameFork checks if fh is ancestor of sh, or if sh is ancestor of fh.
// This way we can be sure both blocks are on the same fork.
func (m msgServer) onTheSameFork(ctx sdk.Context, fh *btypes.BTCHeaderHashBytes, sh *btypes.BTCHeaderHashBytes) (bool, error) {
	isFirstAncestor, err := m.k.IsAncestor(ctx, fh, sh)

	if err != nil {
		return false, err
	}

	isSecondAncestor, err := m.k.IsAncestor(ctx, sh, fh)

	if err != nil {
		return false, err
	}

	return isFirstAncestor || isSecondAncestor, nil
}

// checkHashesFromOneBlock checks if all hashes are from the same block i.e
// if all hashes are equal
func checkHashesFromOneBlock(hs []*btypes.BTCHeaderHashBytes) bool {
	if len(hs) < 2 {
		return true
	}

	for i := 1; i < len(hs); i++ {
		if !hs[i-1].Eq(hs[i]) {
			return false
		}
	}

	return true
}

// checkHashesAreOnTheSameFork checks if provided hases are on the same fork, if
// one of the hashes is not known to light client it returns error
func (m msgServer) checkHashesAreOnTheSameFork(ctx sdk.Context, hs []*btypes.BTCHeaderHashBytes) (bool, error) {

	if len(hs) == 0 {
		// with empty hashes, cannot check for ancestry
		return false, nil
	}

	if len(hs) == 1 {
		// there is only one hash, it is by defintion on one fork.
		return true, nil
	}

	for i := 1; i < len(hs); i++ {
		onTheSameFork, err := m.onTheSameFork(ctx, hs[i-1], hs[i])

		if err != nil {
			return false, err
		}

		if !onTheSameFork {
			// all block hashes are known to light client, but are no longer at the same
			// fork. Checkpoint defacto lost its validity due to some reorg happening.
			return false, nil
		}
	}

	return true, nil
}

func (m msgServer) submissionKeyOnOneFork(ctx sdk.Context, key *types.SubmissionKey) (bool, error) {
	keyHashes := key.GetKeyBlockHashes()
	return m.checkHashesAreOnTheSameFork(ctx, keyHashes)
}

func (m msgServer) checkHeaderIsDescentantOfPreviousEpoch(
	ctx sdk.Context,
	previousEpochSubmissions []*types.SubmissionKey,
	rawSub *types.RawCheckpointSubmission) bool {
	// At this point we already checkeed that bloc hashes in rawSub are on the same
	// fork

	for _, sub := range previousEpochSubmissions {
		// This should always be true, if we have some submission key composed from
		// less than 2 transaction keys in previous epoch, something went really wrong
		if len(sub.Key) < 2 {
			panic("Submission key composed of less than 2 transactions keys in database")
		}

		hs := sub.GetKeyBlockHashes()

		// All this functionality could be implemented in checkHashesAreAncestors
		// and appending first hash of new subbmision to old checkpoint hashes, but there
		// different error conditions here which require different loging.
		if checkHashesFromOneBlock(hs) {
			fh := rawSub.GetFirstBlockHash()
			// all the hashes are from the same block, we only need to check if firstHash
			// of new submission is ancestor of this one hash
			anc, err := m.k.IsAncestor(ctx, hs[0], &fh)
			if err != nil {
				// TODO: light client lost knowledge of blockhash from previous epoch
				// (we know that this is not rawSub as we checked that earlier)
				// It means either some bug / or fork had happened. For now just move
				// forward as we are not able to establish ancestry here
				continue
			}

			if anc {
				// found ancestry stop checking
				return true
			}
		} else {
			// hashes are not from the same block i.e this checkpoint was split between
			// at least two blocks, check if those blocks are still on the same fork
			onSameFork, err := m.checkHashesAreOnTheSameFork(ctx, hs)

			if err != nil {
				// Submission is no longer known to light client
				// TODO it could probably be delted.
				continue
			}

			if !onSameFork {
				// checkpoint blockhashes no longer form a chain. Cannot check ancestry
				// with new submission. Move to the next checkpoint
				continue
			}

			lastHashFromSavedCheckpoint := hs[len(hs)-1]

			fh := rawSub.GetFirstBlockHash()

			// do not check err as all those hashes were checked in previous validation steps
			anc, err := m.k.IsAncestor(ctx, lastHashFromSavedCheckpoint, &fh)

			if err != nil {
				panic("Unexpected anecestry error, all blocks should have been known at this point")
			}

			if anc {
				// found ancestry stop checking
				return true
			}
		}
	}

	return false
}

// TODO at some point add proper logging of error
// TODO emit some events for external consumers. Those should be probably emited
// at EndBlockerCallback
func (m msgServer) InsertBTCSpvProof(ctx context.Context, req *types.MsgInsertBTCSpvProof) (*types.MsgInsertBTCSpvProofResponse, error) {
	rawSubmission, err := types.ParseSubmission(req, m.k.GetPowLimit(), m.k.GetExpectedTag())

	if err != nil {
		return nil, types.ErrInvalidCheckpointProof.Wrap(err.Error())
	}

	// Get the SDK wrapped context
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	submissionKey := rawSubmission.GetSubmissionKey()

	if m.k.SubmissionExists(sdkCtx, submissionKey) {
		return nil, types.ErrDuplicatedSubmission
	}

	submissionState, err := m.k.GetSubmissionBtcState(sdkCtx, submissionKey)

	if err != nil {
		return nil, types.ErrUnknownHeader
	}

	// If we have submission splitted between two blocks and at least one of them
	// is on fork, we need to check if those blocks are on the same chain.
	// In case of submission splitted between two blocks but on main chain, ancestry
	// is implicit.
	// TODO: Either get rid of accepting subbmisions on forks, or design some better
	// mechanism to deal with that case
	if submissionState == OnFork {
		onTheSameFork, err := m.submissionKeyOnOneFork(
			sdkCtx,
			&submissionKey,
		)

		if err != nil {
			panic("Headers which shoud have been known to btc light client")
		}

		if !onTheSameFork {
			return nil, types.ErrProvidedHeaderFromDifferentForks
		}
	}

	rawCheckpointBytes := rawSubmission.GetRawCheckPointBytes()
	// At this point:
	// - every proof of inclusion is valid i.e every transaction is proved to be
	// part of provided block and contains some OP_RETURN data
	// - header is proved to be part of the chain we know about through BTCLightClient
	// - this is new checkpoint submission
	// Inform checkpointing module about it.
	epochNum, err := m.k.GetCheckpointEpoch(sdkCtx, rawCheckpointBytes)

	if err != nil {
		return nil, err
	}

	// This seems to be valid babylon checkpoint, check ancestors.
	// If this submission is not for initial epoch there must already exsits checkpoints
	// for previous epoch which are ancestors of provided submissions
	if epochNum > 1 {
		// this is valid checkpoint for not initial epoch, we need to check previous epoch
		// checkpoints
		previousEpochData := m.k.GetEpochData(sdkCtx, epochNum-1)

		// First check if there are any checkpoints for previous epoch at all.
		if previousEpochData == nil {
			return nil, types.ErrNoCheckpointsForPreviousEpoch
		}

		if len(previousEpochData.Key) == 0 {
			return nil, types.ErrNoCheckpointsForPreviousEpoch
		}

		isDescendant := m.checkHeaderIsDescentantOfPreviousEpoch(sdkCtx, previousEpochData.Key, rawSubmission)

		if !isDescendant {
			return nil, types.ErrProvidedHeaderDoesNotHaveAncestor
		}
	}

	// Everything is fine, save new checkpoint and update Epoch data
	err = m.k.AddEpochSubmission(
		sdkCtx,
		epochNum,
		submissionKey,
		rawSubmission.GetSubmissionData(epochNum),
		submissionState,
		rawCheckpointBytes,
	)

	if err != nil {
		return nil, err
	}

	return &types.MsgInsertBTCSpvProofResponse{}, nil
}

var _ types.MsgServer = msgServer{}
