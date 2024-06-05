package keeper

import (
	"context"

	bbntypes "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

const maxQueryChainsInfoLimit = 100

func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) ChainList(c context.Context, req *types.QueryChainListRequest) (*types.QueryChainListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	chainIDs := []string{}
	store := k.chainInfoStore(ctx)
	pageRes, err := query.Paginate(store, req.Pagination, func(key, value []byte) error {
		chainID := string(key)
		chainIDs = append(chainIDs, chainID)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &types.QueryChainListResponse{
		ChainIds:   chainIDs,
		Pagination: pageRes,
	}
	return resp, nil
}

// ChainsInfo returns the latest info for a given list of chains
func (k Keeper) ChainsInfo(c context.Context, req *types.QueryChainsInfoRequest) (*types.QueryChainsInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// return if no chain IDs are provided
	if len(req.ChainIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "chain IDs cannot be empty")
	}

	// return if chain IDs exceed the limit
	if len(req.ChainIds) > maxQueryChainsInfoLimit {
		return nil, status.Errorf(codes.InvalidArgument, "cannot query more than %d chains", maxQueryChainsInfoLimit)
	}

	// return if chain IDs contain duplicates or empty strings
	if err := bbntypes.CheckForDuplicatesAndEmptyStrings(req.ChainIds); err != nil {
		return nil, status.Error(codes.InvalidArgument, types.ErrInvalidChainIDs.Wrap(err.Error()).Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	var chainsInfo []*types.ChainInfo
	for _, chainID := range req.ChainIds {
		chainInfo, err := k.GetChainInfo(ctx, chainID)
		if err != nil {
			return nil, err
		}

		chainsInfo = append(chainsInfo, chainInfo)
	}

	resp := &types.QueryChainsInfoResponse{ChainsInfo: chainsInfo}
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

// EpochChainsInfo returns the latest info for list of chains in a given epoch
func (k Keeper) EpochChainsInfo(c context.Context, req *types.QueryEpochChainsInfoRequest) (*types.QueryEpochChainsInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// return if no chain IDs are provided
	if len(req.ChainIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "chain IDs cannot be empty")
	}

	// return if chain IDs exceed the limit
	if len(req.ChainIds) > maxQueryChainsInfoLimit {
		return nil, status.Errorf(codes.InvalidArgument, "cannot query more than %d chains", maxQueryChainsInfoLimit)
	}

	// return if chain IDs contain duplicates or empty strings
	if err := bbntypes.CheckForDuplicatesAndEmptyStrings(req.ChainIds); err != nil {
		return nil, status.Error(codes.InvalidArgument, types.ErrInvalidChainIDs.Wrap(err.Error()).Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	var chainsInfo []*types.ChainInfo
	for _, chainID := range req.ChainIds {
		// check if chain ID is valid
		if !k.HasChainInfo(ctx, chainID) {
			return nil, status.Error(codes.InvalidArgument, types.ErrChainInfoNotFound.Wrapf("chain ID %s", chainID).Error())
		}

		// if the chain info is not found in the given epoch, return with empty fields
		if !k.EpochChainInfoExists(ctx, chainID, req.EpochNum) {
			chainsInfo = append(chainsInfo, &types.ChainInfo{ChainId: chainID})
			continue
		}

		// find the chain info of the given epoch
		chainInfoWithProof, err := k.GetEpochChainInfo(ctx, chainID, req.EpochNum)
		if err != nil {
			return nil, err
		}

		chainsInfo = append(chainsInfo, chainInfoWithProof.ChainInfo)
	}

	resp := &types.QueryEpochChainsInfoResponse{ChainsInfo: chainsInfo}
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

// FinalizedChainsInfo returns the finalized info for a given list of chains
func (k Keeper) FinalizedChainsInfo(c context.Context, req *types.QueryFinalizedChainsInfoRequest) (*types.QueryFinalizedChainsInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// return if no chain IDs are provided
	if len(req.ChainIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "chain ID cannot be empty")
	}

	// return if chain IDs exceed the limit
	if len(req.ChainIds) > maxQueryChainsInfoLimit {
		return nil, status.Errorf(codes.InvalidArgument, "cannot query more than %d chains", maxQueryChainsInfoLimit)
	}

	// return if chain IDs contain duplicates or empty strings
	if err := bbntypes.CheckForDuplicatesAndEmptyStrings(req.ChainIds); err != nil {
		return nil, status.Error(codes.InvalidArgument, types.ErrInvalidChainIDs.Wrap(err.Error()).Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	resp := &types.QueryFinalizedChainsInfoResponse{FinalizedChainsInfo: []*types.FinalizedChainInfo{}}

	// find the last finalised epoch
	lastFinalizedEpoch := k.GetLastFinalizedEpoch(ctx)
	for _, chainID := range req.ChainIds {
		// check if chain ID is valid
		if !k.HasChainInfo(ctx, chainID) {
			return nil, status.Error(codes.InvalidArgument, types.ErrChainInfoNotFound.Wrapf("chain ID %s", chainID).Error())
		}

		data := &types.FinalizedChainInfo{ChainId: chainID}

		// if the chain info is not found in the last finalised epoch, return with empty fields
		if !k.EpochChainInfoExists(ctx, chainID, lastFinalizedEpoch) {
			resp.FinalizedChainsInfo = append(resp.FinalizedChainsInfo, data)
			continue
		}

		// find the chain info in the last finalised epoch
		chainInfoWithProof, err := k.GetEpochChainInfo(ctx, chainID, lastFinalizedEpoch)
		if err != nil {
			return nil, err
		}
		chainInfo := chainInfoWithProof.ChainInfo

		// set finalizedEpoch as the earliest epoch that snapshots this chain info.
		// it's possible that the chain info's epoch is way before the last finalised epoch
		// e.g., when there is no relayer for many epochs
		// NOTE: if an epoch is finalised then all of its previous epochs are also finalised
		finalizedEpoch := lastFinalizedEpoch
		if chainInfo.LatestHeader.BabylonEpoch < finalizedEpoch {
			finalizedEpoch = chainInfo.LatestHeader.BabylonEpoch
		}

		data.FinalizedChainInfo = chainInfo

		// find the epoch metadata of the finalised epoch
		data.EpochInfo, err = k.epochingKeeper.GetHistoricalEpoch(ctx, finalizedEpoch)
		if err != nil {
			return nil, err
		}

		rawCheckpoint, err := k.checkpointingKeeper.GetRawCheckpoint(ctx, finalizedEpoch)
		if err != nil {
			return nil, err
		}

		data.RawCheckpoint = rawCheckpoint.Ckpt

		// find the raw checkpoint and the best submission key for the finalised epoch
		_, data.BtcSubmissionKey, err = k.btccKeeper.GetBestSubmission(ctx, finalizedEpoch)
		if err != nil {
			return nil, err
		}

		// generate all proofs
		if req.Prove {
			data.Proof, err = k.proveFinalizedChainInfo(ctx, chainInfo, data.EpochInfo, data.BtcSubmissionKey)
			if err != nil {
				return nil, err
			}
		}

		resp.FinalizedChainsInfo = append(resp.FinalizedChainsInfo, data)
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

	// find the last finalised epoch
	lastFinalizedEpoch := k.GetLastFinalizedEpoch(ctx)
	// find the chain info in the last finalised epoch
	chainInfoWithProof, err := k.GetEpochChainInfo(ctx, req.ChainId, lastFinalizedEpoch)
	if err != nil {
		return nil, err
	}
	chainInfo := chainInfoWithProof.ChainInfo

	// set finalizedEpoch as the earliest epoch that snapshots this chain info.
	// it's possible that the chain info's epoch is way before the last finalised epoch
	// e.g., when there is no relayer for many epochs
	// NOTE: if an epoch is finalised then all of its previous epochs are also finalised
	finalizedEpoch := lastFinalizedEpoch
	if chainInfo.LatestHeader.BabylonEpoch < finalizedEpoch {
		finalizedEpoch = chainInfo.LatestHeader.BabylonEpoch
	}

	resp.FinalizedChainInfo = chainInfo

	if chainInfo.LatestHeader.Height <= req.Height { // the requested height is after the last finalised chain info
		// find and assign the epoch metadata of the finalised epoch
		resp.EpochInfo, err = k.epochingKeeper.GetHistoricalEpoch(ctx, finalizedEpoch)
		if err != nil {
			return nil, err
		}

		rawCheckpoint, err := k.checkpointingKeeper.GetRawCheckpoint(ctx, finalizedEpoch)

		if err != nil {
			return nil, err
		}

		resp.RawCheckpoint = rawCheckpoint.Ckpt

		// find and assign the raw checkpoint and the best submission key for the finalised epoch
		_, resp.BtcSubmissionKey, err = k.btccKeeper.GetBestSubmission(ctx, finalizedEpoch)
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
		chainInfoWithProof, err := k.GetEpochChainInfo(ctx, req.ChainId, finalizedEpoch)
		if err != nil {
			return nil, err
		}
		resp.FinalizedChainInfo = chainInfoWithProof.ChainInfo
		resp.EpochInfo, err = k.epochingKeeper.GetHistoricalEpoch(ctx, finalizedEpoch)
		if err != nil {
			return nil, err
		}

		rawCheckpoint, err := k.checkpointingKeeper.GetRawCheckpoint(ctx, finalizedEpoch)

		if err != nil {
			return nil, err
		}

		resp.RawCheckpoint = rawCheckpoint.Ckpt

		_, resp.BtcSubmissionKey, err = k.btccKeeper.GetBestSubmission(ctx, finalizedEpoch)
		if err != nil {
			return nil, err
		}
	}

	// if the query does not want the proofs, return here
	if !req.Prove {
		return resp, nil
	}

	// generate all proofs
	resp.Proof, err = k.proveFinalizedChainInfo(ctx, chainInfo, resp.EpochInfo, resp.BtcSubmissionKey)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
