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

	epochDataStore := k.epochDataStore(ctx)

	ckptInfoList := []*types.BTCCheckpointInfo{}
	// iterate over epochDataStore, where key is the epoch number and value is the epoch data
	pageRes, err := query.Paginate(epochDataStore, req.Pagination, func(key, value []byte) error {
		epochNum := sdk.BigEndianToUint64(key)
		var epochData types.EpochData
		k.cdc.MustUnmarshal(value, &epochData)

		ckptInfo, err := k.getCheckpointInfo(ctx, epochNum, &epochData)

		if err != nil {
			return fmt.Errorf("failed to get lowest BTC height and hash in keys of epoch %d: %w", epochNum, err)
		}
		// append ckpt info
		ckptInfoList = append(ckptInfoList, ckptInfo)

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &types.QueryBtcCheckpointsInfoResponse{
		InfoList:   ckptInfoList,
		Pagination: pageRes,
	}
	return resp, nil
}

func (k Keeper) EpochSubmissions(c context.Context, req *types.QueryEpochSubmissionsRequest) (*types.QueryEpochSubmissionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	epoch := req.GetEpochNum()
	epochData := k.GetEpochData(ctx, epoch)
	if epochData == nil || len(epochData.Keys) == 0 {
		return &types.QueryEpochSubmissionsResponse{
			Keys: []*types.SubmissionKeyResponse{},
		}, nil
	}

	submKeysResp := make([]*types.SubmissionKeyResponse, len(epochData.Keys))
	for i, submKey := range epochData.Keys {
		skr, err := submKey.ToResponse()
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		submKeysResp[i] = skr
	}

	return &types.QueryEpochSubmissionsResponse{
		Keys: submKeysResp,
	}, nil
}
