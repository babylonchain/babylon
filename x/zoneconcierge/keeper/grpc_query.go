package keeper

import (
	"context"

	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	epochingtypes "github.com/babylonchain/babylon/x/epoching/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
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

// Header returns the header and fork headers at a given height
func (k Keeper) Header(c context.Context, req *types.QueryHeaderRequest) (*types.QueryHeaderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if len(req.ChainId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "chain ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(c)

	header, err := k.GetHeader(ctx, req.ChainId, req.Height)
	if err != nil {
		return nil, err
	}
	forks := k.GetForks(ctx, req.ChainId, req.Height)
	resp := &types.QueryHeaderResponse{
		Header:      header,
		ForkHeaders: forks,
	}

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

// ListEpochHeaders returns all headers of a chain with given ID
// TODO: support pagination in this RPC
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

	// find the last finalised chain info and the earliest epoch that snapshots this chain info
	finalizedEpoch, chainInfo, err := k.GetLastFinalizedChainInfo(ctx, req.ChainId)
	if err != nil {
		return nil, err
	}

	// find the epoch metadata of the finalised epoch
	epochInfo, err := k.epochingKeeper.GetHistoricalEpoch(ctx, finalizedEpoch)
	if err != nil {
		return nil, err
	}

	// find the raw checkpoint and the best submission key for the finalised epoch
	_, rawCheckpoint, bestSubmissionKey, err := k.btccKeeper.GetEpochDataWithBestSubmission(ctx, finalizedEpoch)
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

	// generate all proofs
	resp.ProofTxInBlock, resp.ProofHeaderInEpoch, resp.ProofEpochSealed, resp.ProofEpochSubmitted, err = k.proveFinalizedChainInfo(ctx, chainInfo, epochInfo, bestSubmissionKey)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (k Keeper) FinalizedChainInfoUntilHeight(c context.Context, req *types.QueryFinalizedChainInfoUntilHeightRequest) (*types.QueryFinalizedChainInfoUntilHeightResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if len(req.ChainId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "chain ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(c)
	resp := &types.QueryFinalizedChainInfoUntilHeightResponse{}

	// find and assign the last finalised chain info and the earliest epoch that snapshots this chain info
	finalizedEpoch, chainInfo, err := k.GetLastFinalizedChainInfo(ctx, req.ChainId)
	if err != nil {
		return nil, err
	}
	resp.FinalizedChainInfo = chainInfo

	if chainInfo.LatestHeader.Height <= req.Height { // the requested height is after the last finalised chain info
		// find and assign the epoch metadata of the finalised epoch
		resp.EpochInfo, err = k.epochingKeeper.GetHistoricalEpoch(ctx, finalizedEpoch)
		if err != nil {
			return nil, err
		}

		// find and assign the raw checkpoint and the best submission key for the finalised epoch
		_, resp.RawCheckpoint, resp.BtcSubmissionKey, err = k.btccKeeper.GetEpochDataWithBestSubmission(ctx, finalizedEpoch)
		if err != nil {
			return nil, err
		}
	} else { // the requested height is before the last finalised chain info
		// starting from the requested height, iterate backward until a timestamped header
		closestHeader, err := k.FindClosestHeader(ctx, req.ChainId, req.Height)
		if err != nil {
			return nil, err
		}
		// assign the finalizedEpoch, and retrieve epoch info, raw ckpt and submission key
		finalizedEpoch = closestHeader.BabylonEpoch
		resp.FinalizedChainInfo, err = k.GetEpochChainInfo(ctx, req.ChainId, finalizedEpoch)
		if err != nil {
			return nil, err
		}
		resp.EpochInfo, err = k.epochingKeeper.GetHistoricalEpoch(ctx, finalizedEpoch)
		if err != nil {
			return nil, err
		}
		_, resp.RawCheckpoint, resp.BtcSubmissionKey, err = k.btccKeeper.GetEpochDataWithBestSubmission(ctx, finalizedEpoch)
		if err != nil {
			return nil, err
		}
	}

	// if the query does not want the proofs, return here
	if !req.Prove {
		return resp, nil
	}

	// generate all proofs
	resp.ProofTxInBlock, resp.ProofHeaderInEpoch, resp.ProofEpochSealed, resp.ProofEpochSubmitted, err = k.proveFinalizedChainInfo(ctx, resp.FinalizedChainInfo, resp.EpochInfo, resp.BtcSubmissionKey)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// proveFinalizedChainInfo generates proofs that a chainInfo has been finalised by the given epoch with epochInfo
// It includes proofTxInBlock, proofHeaderInEpoch, proofEpochSealed and proofEpochSubmitted
// The proofs can be verified by a verifier with access to a BTC and Babylon light client
// CONTRACT: this is only a private helper function for simplifying the implementation of RPC calls
func (k Keeper) proveFinalizedChainInfo(
	ctx sdk.Context,
	chainInfo *types.ChainInfo,
	epochInfo *epochingtypes.Epoch,
	bestSubmissionKey *btcctypes.SubmissionKey,
) (*tmproto.TxProof, *tmcrypto.Proof, *types.ProofEpochSealed, []*btcctypes.TransactionInfo, error) {
	// Proof that the Babylon tx is in block
	proofTxInBlock, err := k.ProveTxInBlock(ctx, chainInfo.LatestHeader.BabylonTxHash)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// proof that the block is in this epoch
	proofHeaderInEpoch, err := k.ProveHeaderInEpoch(ctx, chainInfo.LatestHeader.BabylonHeader, epochInfo)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// proof that the epoch is sealed
	proofEpochSealed, err := k.ProveEpochSealed(ctx, epochInfo.EpochNumber)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// proof that the epoch's checkpoint is submitted to BTC
	// i.e., the two `TransactionInfo`s for the checkpoint
	proofEpochSubmitted, err := k.ProveEpochSubmitted(ctx, bestSubmissionKey)
	if err != nil {
		// The only error in ProveEpochSubmitted is the nil bestSubmission.
		// Since the epoch w.r.t. the bestSubmissionKey is finalised, this
		// can only be a programming error, so we should panic here.
		panic(err)
	}
	return proofTxInBlock, proofHeaderInEpoch, proofEpochSealed, proofEpochSubmitted, nil
}
