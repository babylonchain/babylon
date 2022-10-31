package keeper

import (
	"context"

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

	newSubmissionOldestHeaderDepth, err := m.k.GetSubmissionHighestDepth(sdkCtx, submissionKey)

	if err != nil {
		return nil, types.ErrInvalidHeader.Wrap(err.Error())
	}

	rawCheckpointBytes := rawSubmission.GetRawCheckPointBytes()
	// At this point:
	// - every proof of inclusion is valid i.e every transaction is proved to be
	// part of provided block and contains some OP_RETURN data
	// - header is proved to be part of the chain we know about through BTCLightClient
	// - this is new checkpoint submission
	// Get info about this checkpoints epoch
	epochNum, err := m.k.GetCheckpointEpoch(sdkCtx, rawCheckpointBytes)

	if err != nil {
		return nil, err
	}

	err = m.k.checkAncestors(sdkCtx, epochNum, newSubmissionOldestHeaderDepth)

	if err != nil {
		return nil, err
	}

	// Everything is fine, save new checkpoint and update Epoch data
	err = m.k.addEpochSubmission(
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
