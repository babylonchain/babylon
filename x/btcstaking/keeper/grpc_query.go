package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

// BTCValidators returns a paginated list of all Babylon maintained validators
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

// PendingBTCDelegations returns all pending BTC delegations
// TODO: find a good way to support pagination of this query
func (k Keeper) PendingBTCDelegations(ctx context.Context, req *types.QueryPendingBTCDelegationsRequest) (*types.QueryPendingBTCDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	btcDels := []*types.BTCDelegation{}

	// get current BTC height
	btcTipHeight, err := k.GetCurrentBTCHeight(sdkCtx)
	if err != nil {
		return nil, err
	}
	// get value of w
	wValue := k.btccKeeper.GetParams(sdkCtx).CheckpointFinalizationTimeout

	// iterate over each BTC validator
	valStore := k.btcValidatorStore(sdkCtx)
	valIter := valStore.Iterator(nil, nil)
	defer valIter.Close()

	for ; valIter.Valid(); valIter.Next() {
		valBTCPKBytes := valIter.Key()
		valBTCPK, err := bbn.NewBIP340PubKey(valBTCPKBytes)
		if err != nil {
			// this can only be programming error
			panic("failed to unmarshal validator BTC PK in KVstore")
		}
		delStore := k.btcDelegationStore(sdkCtx, valBTCPK)
		delIter := delStore.Iterator(nil, nil)

		// iterate over each BTC delegation under this BTC validator
		for ; delIter.Valid(); delIter.Next() {
			var curBTCDels types.BTCDelegatorDelegations
			btcDelsBytes := delIter.Value()
			k.cdc.MustUnmarshal(btcDelsBytes, &curBTCDels)
			for i, btcDel := range curBTCDels.Dels {
				if btcDel.GetStatus(btcTipHeight, wValue) == types.BTCDelegationStatus_PENDING {
					// avoid using btcDel which changes over the iterations
					btcDels = append(btcDels, curBTCDels.Dels[i])
				}
			}
		}

		delIter.Close()
	}

	return &types.QueryPendingBTCDelegationsResponse{BtcDelegations: btcDels}, nil
}

// BTCValidatorPowerAtHeight returns the voting power of the specified validator
// at the provided Babylon height
func (k Keeper) BTCValidatorPowerAtHeight(ctx context.Context, req *types.QueryBTCValidatorPowerAtHeightRequest) (*types.QueryBTCValidatorPowerAtHeightResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	valBTCPK, err := bbn.NewBIP340PubKeyFromHex(req.ValBtcPkHex)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to unmarshal validator BTC PK hex: %v", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	power := k.GetVotingPower(sdkCtx, valBTCPK.MustMarshal(), req.Height)

	return &types.QueryBTCValidatorPowerAtHeightResponse{VotingPower: power}, nil
}

// ActiveBTCValidatorsAtHeight returns the active BTC validators at the provided height
func (k Keeper) ActiveBTCValidatorsAtHeight(ctx context.Context, req *types.QueryActiveBTCValidatorsAtHeightRequest) (*types.QueryActiveBTCValidatorsAtHeightResponse, error) {
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
				BtcPk:                btcValidator.BtcPk,
				Height:               req.Height,
				VotingPower:          votingPower,
				SlashedBabylonHeight: btcValidator.SlashedBabylonHeight,
				SlashedBtcHeight:     btcValidator.SlashedBtcHeight,
			}
			btcValidatorsWithMeta = append(btcValidatorsWithMeta, &btcValidatorWithMeta)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryActiveBTCValidatorsAtHeightResponse{BtcValidators: btcValidatorsWithMeta, Pagination: pageRes}, nil
}

// ActivatedHeight returns the Babylon height in which the BTC Staking protocol was enabled
// TODO: Requires investigation on whether we can enable the BTC staking protocol at genesis
func (k Keeper) ActivatedHeight(ctx context.Context, req *types.QueryActivatedHeightRequest) (*types.QueryActivatedHeightResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	activatedHeight, err := k.GetBTCStakingActivatedHeight(sdkCtx)
	if err != nil {
		return nil, err
	}
	return &types.QueryActivatedHeightResponse{Height: activatedHeight}, nil
}

// BTCValidatorDelegations returns all the delegations of the provided validator filtered by the provided status.
func (k Keeper) BTCValidatorDelegations(ctx context.Context, req *types.QueryBTCValidatorDelegationsRequest) (*types.QueryBTCValidatorDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if len(req.ValBtcPkHex) == 0 {
		return nil, errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest, "validator BTC public key cannot be empty")
	}

	valPK, err := bbn.NewBIP340PubKeyFromHex(req.ValBtcPkHex)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	btcDelStore := k.btcDelegationStore(sdkCtx, valPK)

	btcDels := []*types.BTCDelegatorDelegations{}
	pageRes, err := query.Paginate(btcDelStore, req.Pagination, func(key, value []byte) error {
		var curBTCDels types.BTCDelegatorDelegations
		k.cdc.MustUnmarshal(value, &curBTCDels)
		btcDels = append(btcDels, &curBTCDels)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryBTCValidatorDelegationsResponse{BtcDelegatorDelegations: btcDels, Pagination: pageRes}, nil
}
