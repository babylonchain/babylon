package keeper

import (
	"context"
	"errors"
	"fmt"

	"github.com/babylonchain/babylon/x/btccheckpoint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) getCheckpointInfo(ctx context.Context, epochNum uint64, epochData *types.EpochData) (*types.BTCCheckpointInfo, error) {
	bestSubmission := k.GetEpochBestSubmissionBtcInfo(ctx, epochData)

	if bestSubmission == nil {
		return nil, errors.New("checkpoint for given epoch not yet submitted")
	}

	bestSubmissionHeight, err := k.GetBlockHeight(ctx, &bestSubmission.YoungestBlockHash)

	if err != nil {
		return nil, fmt.Errorf("error getting best submission height: %w", err)
	}

	bestSubmissionData := k.GetSubmissionData(ctx, bestSubmission.SubmissionKey)

	return &types.BTCCheckpointInfo{
		EpochNumber:                        epochNum,
		BestSubmissionBtcBlockHeight:       bestSubmissionHeight,
		BestSubmissionBtcBlockHash:         &bestSubmission.YoungestBlockHash,
		BestSubmissionTransactions:         bestSubmissionData.TxsInfo,
		BestSubmissionVigilanteAddressList: []*types.CheckpointAddresses{bestSubmissionData.VigilanteAddresses},
	}, nil
}

func (k Keeper) BtcCheckpointInfo(c context.Context, req *types.QueryBtcCheckpointInfoRequest) (*types.QueryBtcCheckpointInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	checkpointEpoch := req.GetEpochNum()

	ed := k.GetEpochData(ctx, checkpointEpoch)
	ckptInfo, err := k.getCheckpointInfo(ctx, checkpointEpoch, ed)
	if err != nil {
		return nil, fmt.Errorf("failed to get lowest BTC height and hash in keys of epoch %d: %w", req.EpochNum, err)
	}

	return &types.QueryBtcCheckpointInfoResponse{
		Info: ckptInfo.ToResponse(),
	}, nil
}

func (k Keeper) BtcCheckpointsInfo(ctx context.Context, req *types.QueryBtcCheckpointsInfoRequest) (*types.QueryBtcCheckpointsInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ckptInfoList := []*types.BTCCheckpointInfoResponse{}
	epochDataStore := k.epochDataStore(ctx)
	// iterate over epochDataStore, where key is the epoch number and value is the epoch data
	pageRes, err := query.Paginate(epochDataStore, req.Pagination, func(key, value []byte) error {
		var epochData types.EpochData
		if err := k.cdc.Unmarshal(value, &epochData); err != nil {
			return fmt.Errorf("failed to decode epoch data %+v: %w", value, err)
		}
		epochNum := sdk.BigEndianToUint64(key)

		ckptInfo, err := k.getCheckpointInfo(ctx, epochNum, &epochData)
		if err != nil {
			return fmt.Errorf("failed to get lowest BTC height and hash in keys of epoch %d: %w", epochNum, err)
		}

		// append ckpt info
		ckptInfoList = append(ckptInfoList, ckptInfo.ToResponse())
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryBtcCheckpointsInfoResponse{
		InfoList:   ckptInfoList,
		Pagination: pageRes,
	}, nil
}

func getOffset(pageReq *query.PageRequest) uint64 {
	if pageReq == nil {
		return 0
	} else {
		return pageReq.Offset
	}
}

func buildPageResponse(numOfKeys uint64, pageReq *query.PageRequest) *query.PageResponse {
	if pageReq == nil {
		return &query.PageResponse{}
	}

	if !pageReq.CountTotal {
		return &query.PageResponse{}
	}

	return &query.PageResponse{Total: numOfKeys}
}

func (k Keeper) EpochSubmissions(c context.Context, req *types.QueryEpochSubmissionsRequest) (*types.QueryEpochSubmissionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	checkpointEpoch := req.GetEpochNum()

	_, limit, err := query.ParsePagination(req.Pagination)

	offset := getOffset(req.Pagination)

	if err != nil {
		return nil, err
	}

	epochData := k.GetEpochData(ctx, checkpointEpoch)

	if epochData == nil || len(epochData.Keys) == 0 {

		return &types.QueryEpochSubmissionsResponse{
			Keys:       []*types.SubmissionKey{},
			Pagination: buildPageResponse(0, req.Pagination),
		}, nil
	}

	numberOfKeys := uint64(len((epochData.Keys)))

	if offset >= numberOfKeys {
		// offset larger than number of keys return empty response
		return &types.QueryEpochSubmissionsResponse{
			Keys:       []*types.SubmissionKey{},
			Pagination: buildPageResponse(numberOfKeys, req.Pagination),
		}, nil
	}

	var responseKeys []*types.SubmissionKey

	for i := offset; i < numberOfKeys; i++ {
		if len(responseKeys) == limit {
			break
		}

		responseKeys = append(responseKeys, epochData.Keys[i])
	}

	return &types.QueryEpochSubmissionsResponse{
		Keys:       responseKeys,
		Pagination: buildPageResponse(numberOfKeys, req.Pagination),
	}, nil
}
