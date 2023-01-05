package keeper

import (
	"context"
	"fmt"

	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) ChainList(c context.Context, req *types.QueryChainListRequest) (*types.QueryChainListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	chainIDs := k.GetAllChainIDs(ctx)
	// TODO: pagination for this API
	resp := &types.QueryChainListResponse{ChainIds: chainIDs}
	return resp, nil
}

// ChainInfo returns the latest info of a chain with given ID
func (k Keeper) ChainInfo(c context.Context, req *types.QueryChainInfoRequest) (*types.QueryChainInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if len(req.ChainId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "chain ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// find the chain info of this epoch
	chainInfo := k.GetChainInfo(ctx, req.ChainId)
	resp := &types.QueryChainInfoResponse{ChainInfo: chainInfo}
	return resp, nil
}

// EpochChainInfo returns the info of a chain with given ID in a given epoch
func (k Keeper) EpochChainInfo(c context.Context, req *types.QueryEpochChainInfoRequest) (*types.QueryEpochChainInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if len(req.ChainId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "chain ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// find the chain info of the given epoch
	chainInfo, err := k.GetEpochChainInfo(ctx, req.ChainId, req.EpochNum)
	if err != nil {
		return nil, err
	}
	resp := &types.QueryEpochChainInfoResponse{ChainInfo: chainInfo}
	return resp, nil
}

// ListHeaders returns all headers of a chain with given ID, with pagination support
func (k Keeper) ListHeaders(c context.Context, req *types.QueryListHeadersRequest) (*types.QueryListHeadersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if len(req.ChainId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "chain ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(c)

	headers := []*types.IndexedHeader{}
	store := k.canonicalChainStore(ctx, req.ChainId)
	pageRes, err := query.Paginate(store, req.Pagination, func(key, value []byte) error {
		var header types.IndexedHeader
		k.cdc.MustUnmarshal(value, &header)
		headers = append(headers, &header)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &types.QueryListHeadersResponse{
		Headers:    headers,
		Pagination: pageRes,
	}
	return resp, nil
}

// ListEpochHeaders returns all headers of a chain with given ID, with pagination support
func (k Keeper) ListEpochHeaders(c context.Context, req *types.QueryListEpochHeadersRequest) (*types.QueryListEpochHeadersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if len(req.ChainId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "chain ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(c)

	headers, err := k.GetEpochHeaders(ctx, req.ChainId, req.EpochNum)
	if err != nil {
		return nil, err
	}

	resp := &types.QueryListEpochHeadersResponse{
		Headers: headers,
	}
	return resp, nil
}

func (k Keeper) FinalizedChainInfo(c context.Context, req *types.QueryFinalizedChainInfoRequest) (*types.QueryFinalizedChainInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if len(req.ChainId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "chain ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// find the last finalised epoch
	finalizedEpoch, err := k.GetFinalizedEpoch(ctx)
	if err != nil {
		return nil, err
	}

	// find the chain info of this epoch
	chainInfo, err := k.GetEpochChainInfo(ctx, req.ChainId, finalizedEpoch)
	if err != nil {
		return nil, err
	}

	// It's possible that the chain info's epoch is way before the last finalised epoch
	// e.g., when there is no relayer for many epochs
	// NOTE: if an epoch is finalised then all of its previous epochs are also finalised
	if chainInfo.LatestHeader.BabylonEpoch < finalizedEpoch {
		finalizedEpoch = chainInfo.LatestHeader.BabylonEpoch
	}

	// find the epoch metadata
	epochInfo, err := k.epochingKeeper.GetHistoricalEpoch(ctx, finalizedEpoch)
	if err != nil {
		return nil, err
	}

	// find the btc checkpoint tx index of this epoch
	ed := k.btccKeeper.GetEpochData(ctx, finalizedEpoch)
	if ed.Status != btcctypes.Finalized {
		err := fmt.Errorf("epoch %d should have been finalized, but is in status %s", finalizedEpoch, ed.Status.String())
		panic(err) // this can only be a programming error
	}
	if len(ed.Key) == 0 {
		err := fmt.Errorf("finalized epoch %d should have at least 1 checkpoint submission", finalizedEpoch)
		panic(err) // this can only be a programming error
	}
	bestSubmissionKey := ed.Key[0] // index of checkpoint tx on BTC

	// get raw checkpoint of this epoch
	rawCheckpointBytes := ed.RawCheckpoint
	rawCheckpoint, err := checkpointingtypes.FromBTCCkptBytesToRawCkpt(rawCheckpointBytes)
	if err != nil {
		return nil, err
	}

	resp := &types.QueryFinalizedChainInfoResponse{
		FinalizedChainInfo: chainInfo,
		// metadata related to this chain info, including the epoch, the raw checkpoint of this epoch, and the BTC tx index of the raw checkpoint
		EpochInfo:        epochInfo,
		RawCheckpoint:    rawCheckpoint,
		BtcSubmissionKey: bestSubmissionKey,
	}

	// if the query does not want the proofs, return here
	if !req.Prove {
		return resp, nil
	}

	// Proof that the Babylon tx is in block
	resp.ProofTxInBlock, err = k.ProveTxInBlock(ctx, chainInfo.LatestHeader.BabylonTxHash)
	if err != nil {
		return nil, err
	}

	// proof that the block is in this epoch
	resp.ProofHeaderInEpoch, err = k.ProveHeaderInEpoch(ctx, chainInfo.LatestHeader.BabylonHeader, epochInfo)
	if err != nil {
		return nil, err
	}

	// proof that the epoch is sealed
	resp.ProofEpochSealed, err = k.ProveEpochSealed(ctx, finalizedEpoch)
	if err != nil {
		return nil, err
	}

	// proof that the epoch's checkpoint is submitted to BTC
	// i.e., the two `TransactionInfo`s for the checkpoint
	resp.ProofEpochSubmitted, err = k.ProveEpochSubmitted(ctx, *bestSubmissionKey)
	if err != nil {
		// The only error in ProveEpochSubmitted is the nil bestSubmission.
		// Since the epoch w.r.t. the bestSubmissionKey is finalised, this
		// can only be a programming error, so we should panic here.
		panic(err)
	}

	return resp, nil
}
