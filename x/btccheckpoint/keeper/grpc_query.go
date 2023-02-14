package keeper

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) lowestBtcHeightAndHash(ctx sdk.Context, subKey *types.SubmissionKey) (uint64, []byte, error) {
	// initializing to max, as then every header height will be smaller
	var lowestHeaderNumber uint64 = math.MaxUint64
	var lowestHeaderHash []byte

	for _, tk := range subKey.Key {

		if !k.CheckHeaderIsOnMainChain(ctx, tk.Hash) {
			return 0, nil, errors.New("one of submission headers not on main chain")
		}

		headerNumber, err := k.GetBlockHeight(ctx, tk.Hash)

		if err != nil {
			// CheckHeaderIsOnMainChain  (which uses main chain depth) returned true which
			// means header should be saved and we should know its heigh but GetBlockHeight
			// returned error. Something is really bad, panic.
			panic("Inconsistent data model in btc light client")
		}

		if headerNumber < lowestHeaderNumber {
			lowestHeaderNumber = headerNumber
			lowestHeaderHash = *tk.Hash
		}
	}

	return lowestHeaderNumber, lowestHeaderHash, nil
}

func (k Keeper) getCheckpointInfo(ctx sdk.Context, epochNum uint64, subKeys []*types.SubmissionKey) (*types.BTCCheckpointInfo, error) {
	if len(subKeys) == 0 {
		return nil, errors.New("empty subKeys")
	}

	info := types.BTCCheckpointInfo{
		EpochNumber:            epochNum,
		EarliestBtcBlockNumber: math.MaxUint64, // initializing to max, as then every header height will be smaller
		VigilanteAddressList:   []*types.CheckpointAddresses{},
	}

	for _, subKey := range subKeys {
		headerNumber, headerHash, err := k.lowestBtcHeightAndHash(ctx, subKey)
		if err != nil {
			// submission is not valid for some reason, ignore it
			continue
		}

		// get vigilante address
		sd := k.GetSubmissionData(ctx, *subKey)
		if sd == nil {
			// submission is not valid for some reason, ignore it
			continue
		}

		// ensure lowest header number and hash
		if headerNumber < info.EarliestBtcBlockNumber {
			info.EarliestBtcBlockNumber = headerNumber
			info.EarliestBtcBlockHash = headerHash
		}
		// append vigilante addresses
		vAddrs := *sd.VigilanteAddresses // make a new copy
		info.VigilanteAddressList = append(info.VigilanteAddressList, &vAddrs)
	}

	if info.EarliestBtcBlockNumber == math.MaxUint64 {
		return nil, errors.New("there is no valid submission for given raw checkpoint")
	}

	return &info, nil
}

func (k Keeper) BtcCheckpointInfo(c context.Context, req *types.QueryBtcCheckpointInfoRequest) (*types.QueryBtcCheckpointInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	checkpointEpoch := req.GetEpochNum()

	epochData := k.GetEpochData(ctx, checkpointEpoch)

	// Check if we have any submission for given epoch
	if epochData == nil || len(epochData.Key) == 0 {
		return nil, errors.New("checkpoint for given epoch not yet submitted")
	}

	ckptInfo, err := k.getCheckpointInfo(ctx, checkpointEpoch, epochData.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to get lowest BTC height and hash in keys of epoch %d: %w", req.EpochNum, err)
	}

	resp := &types.QueryBtcCheckpointInfoResponse{
		Info: ckptInfo,
	}
	return resp, nil
}

func (k Keeper) BtcCheckpointsInfo(c context.Context, req *types.QueryBtcCheckpointsInfoRequest) (*types.QueryBtcCheckpointsInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	epochDataStore := prefix.NewStore(store, types.EpochDataPrefix)

	ckptInfoList := []*types.BTCCheckpointInfo{}
	// iterate over epochDataStore, where key is the epoch number and value is the epoch data
	pageRes, err := query.Paginate(epochDataStore, req.Pagination, func(key, value []byte) error {
		epochNum := sdk.BigEndianToUint64(key)
		var epochData types.EpochData
		k.cdc.MustUnmarshal(value, &epochData)

		// Check if we have any submission for given epoch
		if len(epochData.Key) == 0 {
			return errors.New("checkpoint for given epoch not yet submitted")
		}

		ckptInfo, err := k.getCheckpointInfo(ctx, epochNum, epochData.Key)
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

	if epochData == nil || len(epochData.Key) == 0 {

		return &types.QueryEpochSubmissionsResponse{
			Keys:       []*types.SubmissionKey{},
			Pagination: buildPageResponse(0, req.Pagination),
		}, nil
	}

	numberOfKeys := uint64(len((epochData.Key)))

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

		responseKeys = append(responseKeys, epochData.Key[i])
	}

	return &types.QueryEpochSubmissionsResponse{
		Keys:       responseKeys,
		Pagination: buildPageResponse(numberOfKeys, req.Pagination),
	}, nil
}
