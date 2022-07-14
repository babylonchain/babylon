package keeper

import (
	"context"

	btypes "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btccheckpoint/types"
	btcchaincfg "github.com/btcsuite/btcd/chaincfg"
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

// Gets proof height in context of btclightclilent, also if proof is composed
// from two different blocks checks that they are on the same fork.
func (m msgServer) getProofHeight(rawSub *types.RawCheckpointSubmission) (uint64, error) {
	var latestblock = uint64(0)

	fh := rawSub.GetFirstBlockHash()
	sh := rawSub.GetSecondBlockHash()

	if fh.Eq(&sh) {
		// both hashes are the same which means, two transactions with their respective
		// proofs were provided in the same block. We only need to check one block for
		// for height
		num, err := m.k.GetBlockHeight(fh)

		if err != nil {
			return 0, err
		}

		return num, nil
	}

	// at this point we know that both transactions were in different blocks.
	// we need to check two things:
	// - if both blocks are known to header oracle
	// - if both blocks are on the same fork i.e if second block is descendant of the
	// first block
	for _, hash := range []btypes.BTCHeaderHashBytes{fh, sh} {
		num, err := m.k.GetBlockHeight(hash)
		if err != nil {
			return 0, err
		}
		// the highest block number with a checkpoint being stable implies that all blocks are stable
		if num > latestblock {
			latestblock = num
		}
	}

	// we have checked earlier that both blocks are known to header light client,
	// so no need to check err.
	isAncestor, err := m.k.IsAncestor(fh, sh)

	if err != nil {
		panic("Headers which are should have been known to btclight client")
	}

	if !isAncestor {
		return 0, types.ErrProvidedHeaderFromDifferentForks
	}

	return latestblock, nil
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

func (m msgServer) checkHashesAreAncestors(hs []*btypes.BTCHeaderHashBytes) bool {
	if len(hs) < 2 {
		// cannot have ancestry relations with only 0 or 1 hash
		return false
	}

	for i := 1; i < len(hs); i++ {
		anc, err := m.k.IsAncestor(*hs[i-1], *hs[i])

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
			// all the hashes are from the same block, we only need to check if firstHash
			// of new submission is ancestor of this one hash
			anc, err := m.k.IsAncestor(*hs[0], rawSub.GetFirstBlockHash())

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
			if !m.checkHashesAreAncestors(hs) {
				// checkpoint blockhashes no longer form a chain. Cannot check ancestry
				// with new submission. Move to the next checkpoint
				continue
			}

			lastHashFromSavedCheckpoint := hs[len(hs)-1]

			// do not check err as all those hashes were checked in previous validation steps
			anc, err := m.k.IsAncestor(*lastHashFromSavedCheckpoint, rawSub.GetFirstBlockHash())

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

// TODO maybe move it to keeper
func (m msgServer) checkAndSaveEpochData(
	sdkCtx sdk.Context,
	epochNum uint64,
	subKey types.SubmissionKey,
	subData types.SubmissionData) {
	ed := m.k.GetEpochData(sdkCtx, epochNum)

	// TODO: SaveEpochData and SaveSubmission should be done in one transaction.
	// Not sure cosmos-sdk has facialities to do it.
	// Otherwise it is possible to end up with node which updated submission list
	// but did not save submission itself.

	if ed == nil {
		// we do not have any data saved yet
		newEd := types.NewEpochData(subKey)
		ed = &newEd
	} else {
		// epoch data already existis for given epoch, append new submission, and save
		// submission key and data
		ed.AppendKey(subKey)
	}

	m.k.SaveEpochData(sdkCtx, epochNum, ed)
	m.k.SaveSubmission(sdkCtx, subKey, subData)
}

// TODO at some point add proper logging of error
// TODO emit some events for external consumers. Those should be probably emited
// at EndBlockerCallback
func (m msgServer) InsertBTCSpvProof(ctx context.Context, req *types.MsgInsertBTCSpvProof) (*types.MsgInsertBTCSpvProofResponse, error) {

	address, err := sdk.AccAddressFromBech32(req.Submitter)

	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid submitter address: %s", err)
	}

	// TODO get PowLimit from config
	rawSubmission, e := types.ParseTwoProofs(address, req.Proofs, btcchaincfg.MainNetParams.PowLimit)

	if e != nil {
		return nil, types.ErrInvalidCheckpointProof
	}

	// Get the SDK wrapped context
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	submissionKey := rawSubmission.GetSubmissionKey()

	if m.k.SubmissionExists(sdkCtx, submissionKey) {
		return nil, types.ErrDuplicatedSubmission
	}

	// TODO for now we do nothing with processed blockHeight but ultimatly it should
	// be a part of timestamp
	_, err = m.getProofHeight(rawSubmission)

	if err != nil {
		return nil, err
	}

	// At this point:
	// - every proof of inclusion is valid i.e every transaction is proved to be
	// part of provided block and contains some OP_RETURN data
	// - header is proved to be part of the chain we know about through BTCLightClient
	// - this is new checkpoint submission
	// Inform checkpointing module about it.
	epochNum, err := m.k.GetCheckpointEpoch(rawSubmission.GetRawCheckPointBytes())

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

		isDescendant := m.checkHeaderIsDescentantOfPreviousEpoch(previousEpochData.Key, rawSubmission)

		if !isDescendant {
			return nil, types.ErrProvidedHeaderDoesNotHaveAncestor
		}
	}

	// Everything is fine, save new checkpoint and update Epoch data
	m.checkAndSaveEpochData(sdkCtx, epochNum, submissionKey, rawSubmission.GetSubmissionData())

	return &types.MsgInsertBTCSpvProofResponse{}, nil
}

var _ types.MsgServer = msgServer{}
