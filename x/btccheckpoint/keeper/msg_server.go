package keeper

import (
	"context"

	btypes "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btccheckpoint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type msgServer struct {
	k Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper}
}

// sameAncestry checks whether the nodes are equal or are contained in the same chain
func (m msgServer) sameAncestry(ctx sdk.Context, fh *btypes.BTCHeaderHashBytes, sh *btypes.BTCHeaderHashBytes) (bool, error) {
	if fh.Eq(sh) {
		return true, nil
	}

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

// Gets proof height in context of btclightclilent, also if proof is composed
// from two different blocks checks that they are on the same fork.
func (m msgServer) checkAllHeadersAreKnown(ctx sdk.Context, rawSub *types.RawCheckpointSubmission) error {
	fh := rawSub.GetFirstBlockHash()
	sh := rawSub.GetSecondBlockHash()

	if fh.Eq(&sh) {
		// both hashes are the same which means, two transactions with their respective
		// proofs were provided in the same block. We only need to check one block for
		// for height
		if m.k.CheckHeaderIsKnown(ctx, &fh) {
			return nil
		} else {
			return types.ErrUnknownHeader
		}
	}

	// at this point we know that both transactions were in different blocks.
	// we need to check two things:
	// - if both blocks are known to header oracle
	// - if both blocks are on the same fork i.e if second block is descendant of the
	// first block
	for _, hash := range []btypes.BTCHeaderHashBytes{fh, sh} {
		if !m.k.CheckHeaderIsKnown(ctx, &hash) {
			return types.ErrUnknownHeader
		}
	}

	// we have checked earlier that both blocks are known to header light client,
	// so in case there is an error we should panic.
	sameAncestry, err := m.sameAncestry(ctx, &fh, &sh)

	if err != nil {
		panic("Headers which are should have been known to btclight client")
	}

	if !sameAncestry {
		return types.ErrProvidedHeaderFromDifferentForks
	}

	return nil
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

func (m msgServer) checkHashesAreAncestors(ctx sdk.Context, hs []*btypes.BTCHeaderHashBytes) bool {
	if len(hs) < 2 {
		// cannot have ancestry relations with only 0 or 1 hash
		return false
	}

	for i := 1; i < len(hs); i++ {
		anc, err := m.sameAncestry(ctx, hs[i-1], hs[i])

		if err != nil {
			// TODO: Light client lost knowledge of one of the chekpoint hashes.
			// decide what to do here. For now returning false, as we cannot check ancestry.
			return false
		}

		if !anc {
			// all block hashes are known to light client, but are no longer at the same
			// fork. Checkpoint defacto lost its validity due to some reorg happening.
			return false
		}
	}

	return true
}

func (m msgServer) checkHeaderIsDescentantOfPreviousEpoch(
	ctx sdk.Context,
	previousEpochSubmissions []*types.SubmissionKey,
	rawSub *types.RawCheckpointSubmission) bool {

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
			// of new submission is a descendant of this one hash
			anc, err := m.sameAncestry(ctx, hs[0], &fh)
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
			// at least two block, check that it is still valid
			if !m.checkHashesAreAncestors(ctx, hs) {
				// checkpoint blockhashes no longer form a chain. Cannot check ancestry
				// with new submission. Move to the next checkpoint
				continue
			}

			// There might be a submission for a previous epoch that was made on
			// a later BTC block. We want at least one ancestor of the first hash
			// in the canonical chain.
			fh := rawSub.GetFirstBlockHash()

			for _, h := range hs {
				// Those hashes were all checked in previous validation steps
				// If there is an error, panic
				anc, err := m.sameAncestry(ctx, h, &fh)

				if err != nil {
					panic("Unexpected anecestry error, all blocks should have been known at this point")
				}

				if anc {
					// found ancestry stop checking
					return true
				}
			}

		}
	}

	return false
}

// TODO at some point add proper logging of error
// TODO emit some events for external consumers. Those should be probably emited
// at EndBlockerCallback
func (m msgServer) InsertBTCSpvProof(ctx context.Context, req *types.MsgInsertBTCSpvProof) (*types.MsgInsertBTCSpvProofResponse, error) {

	address, err := sdk.AccAddressFromBech32(req.Submitter)

	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid submitter address: %s", err)
	}

	rawSubmission, e := types.ParseTwoProofs(address, req.Proofs, m.k.GetPowLimit(), m.k.GetExpectedTag())

	if e != nil {
		return nil, types.ErrInvalidCheckpointProof
	}

	// Get the SDK wrapped context
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	submissionKey := rawSubmission.GetSubmissionKey()

	if m.k.SubmissionExists(sdkCtx, submissionKey) {
		return nil, types.ErrDuplicatedSubmission
	}

	err = m.checkAllHeadersAreKnown(sdkCtx, rawSubmission)

	if err != nil {
		return nil, err
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
		rawCheckpointBytes,
	)

	if err != nil {
		return nil, err
	}

	return &types.MsgInsertBTCSpvProofResponse{}, nil
}

var _ types.MsgServer = msgServer{}
