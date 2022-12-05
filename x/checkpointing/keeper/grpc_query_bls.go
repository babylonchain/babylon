package keeper

import (
	"context"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/jinzhu/copier"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) BlsPublicKeyList(c context.Context, req *types.QueryBlsPublicKeyListRequest) (*types.QueryBlsPublicKeyListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	sdkCtx := sdk.UnwrapSDKContext(c)
	valBLSKeys, err := k.GetBLSPubKeySet(sdkCtx, req.EpochNum)
	if err != nil {
		return nil, err
	}

	if req.Pagination == nil {
		return &types.QueryBlsPublicKeyListResponse{
			ValidatorWithBlsKeys: valBLSKeys,
		}, nil
	}

	total := uint64(len(valBLSKeys))
	start := req.Pagination.Offset
	if start > total-1 {
		return nil, status.Error(codes.InvalidArgument, "pagination offset out of range")
	}
	var end uint64
	if req.Pagination.Limit == 0 {
		end = total
	} else {
		end = req.Pagination.Limit + start
	}
	if end > total {
		end = total
	}
	var copiedValBLSKeys []*types.ValidatorWithBlsKey
	err = copier.Copy(&copiedValBLSKeys, valBLSKeys[start:end])
	if err != nil {
		return nil, err
	}

	return &types.QueryBlsPublicKeyListResponse{
		ValidatorWithBlsKeys: copiedValBLSKeys,
	}, nil
}
