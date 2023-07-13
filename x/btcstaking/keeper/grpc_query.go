package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/babylonchain/babylon/x/btcstaking/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) BTCValidators(ctx context.Context, req *types.QueryBTCValidatorsRequest) (*types.QueryBTCValidatorsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.btcValidatorStore(sdkCtx)

	var btcValidators []*types.BTCValidator
	pageRes, err := query.Paginate(store, req.Pagination, func(key, value []byte) error {
		var btcValidator types.BTCValidator
		k.cdc.MustUnmarshal(value, &btcValidator)
		btcValidators = append(btcValidators, &btcValidator)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryBTCValidatorsResponse{BtcValidators: btcValidators, Pagination: pageRes}, nil
}

func (k Keeper) BTCValidatorsAtHeight(ctx context.Context, req *types.QueryBTCValidatorsAtHeightRequest) (*types.QueryBTCValidatorsAtHeightResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := k.votingPowerStore(sdkCtx, req.Height)

	var btcValidatorsWithMeta []*types.BTCValidatorWithMeta
	pageRes, err := query.Paginate(store, req.Pagination, func(key, value []byte) error {
		btcValidator, err := k.GetBTCValidator(sdkCtx, key)
		if err != nil {
			return err
		}

		votingPower := k.GetVotingPower(sdkCtx, key, req.Height)
		if votingPower > 0 {
			btcValidatorWithMeta := types.BTCValidatorWithMeta{
				BtcPk:       btcValidator.BtcPk,
				Height:      req.Height,
				VotingPower: votingPower,
			}
			btcValidatorsWithMeta = append(btcValidatorsWithMeta, &btcValidatorWithMeta)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryBTCValidatorsAtHeightResponse{BtcValidators: btcValidatorsWithMeta, Pagination: pageRes}, nil
}

func (k Keeper) BTCValidatorDelegationsAtHeight(ctx context.Context, req *types.QueryBTCValidatorDelegationsAtHeightRequest) (*types.QueryBTCValidatorDelegationsAtHeightResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// convert pk to bytes
	valPkBytes, err := req.ValBtcPk.Marshal()
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	btcHeight, err := k.GetBTCHeightAtBabylonHeight(sdkCtx, req.Height)
	if err != nil {
		return nil, err
	}

	// get value of w
	wValue := k.btccKeeper.GetParams(sdkCtx).CheckpointFinalizationTimeout
	store := k.btcDelegationStore(sdkCtx, valPkBytes)

	var btcDelsWithMeta []*types.BTCDelegationWithMeta
	pageRes, err := query.Paginate(store, req.Pagination, func(key, value []byte) error {
		btcDel, err := k.GetBTCDelegation(sdkCtx, valPkBytes, key)
		if err != nil {
			return err
		}

		delPower := btcDel.VotingPower(btcHeight, wValue)
		if delPower > 0 {
			btcDelMeta := types.BTCDelegationWithMeta{
				BtcPk:       btcDel.BtcPk,
				StartHeight: btcDel.StartHeight,
				EndHeight:   btcDel.EndHeight,
				VotingPower: delPower,
			}
			btcDelsWithMeta = append(btcDelsWithMeta, &btcDelMeta)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryBTCValidatorDelegationsAtHeightResponse{BtcDelegations: btcDelsWithMeta, Pagination: pageRes}, nil
}
