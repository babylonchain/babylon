package keeper

import (
	"context"
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
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

// BTCValidator returns the validator with the specified validator BTC PK
func (k Keeper) BTCValidator(ctx context.Context, req *types.QueryBTCValidatorRequest) (*types.QueryBTCValidatorResponse, error) {
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

	val, err := k.GetBTCValidator(sdkCtx, valPK.MustMarshal())

	if err != nil {
		return nil, err
	}

	return &types.QueryBTCValidatorResponse{BtcValidator: val}, nil
}

// BTCDelegations returns all BTC delegations under a given status
func (k Keeper) BTCDelegations(ctx context.Context, req *types.QueryBTCDelegationsRequest) (*types.QueryBTCDelegationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// get current BTC height
	btcTipHeight, err := k.GetCurrentBTCHeight(sdkCtx)
	if err != nil {
		return nil, err
	}
	// get value of w
	wValue := k.btccKeeper.GetParams(sdkCtx).CheckpointFinalizationTimeout

	store := k.btcDelegationStore(sdkCtx)
	var btcDels []*types.BTCDelegation
	pageRes, err := query.FilteredPaginate(store, req.Pagination, func(_ []byte, value []byte, accumulate bool) (bool, error) {
		var btcDel types.BTCDelegation
		k.cdc.MustUnmarshal(value, &btcDel)

		// hit if the queried status is ANY or matches the BTC delegation status
		if req.Status == types.BTCDelegationStatus_ANY || btcDel.GetStatus(btcTipHeight, wValue) == req.Status {
			if accumulate {
				btcDels = append(btcDels, &btcDel)
			}
			return true, nil
		}

		return false, nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryBTCDelegationsResponse{
		BtcDelegations: btcDels,
		Pagination:     pageRes,
	}, nil
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

// BTCValidatorCurrentPower returns the voting power of the specified validator
// at the current height
func (k Keeper) BTCValidatorCurrentPower(ctx context.Context, req *types.QueryBTCValidatorCurrentPowerRequest) (*types.QueryBTCValidatorCurrentPowerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	valBTCPK, err := bbn.NewBIP340PubKeyFromHex(req.ValBtcPkHex)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to unmarshal validator BTC PK hex: %v", err)
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	power := uint64(0)
	curHeight := uint64(sdkCtx.BlockHeight())

	// if voting power table is recorded at the current height, use this voting power
	if k.HasVotingPowerTable(sdkCtx, curHeight) {
		power = k.GetVotingPower(sdkCtx, valBTCPK.MustMarshal(), curHeight)
	} else {
		// NOTE: it's possible that the voting power is not recorded at the current height,
		// e.g., `EndBlock` is not reached yet
		// in this case, we use the last height
		curHeight -= 1
		power = k.GetVotingPower(sdkCtx, valBTCPK.MustMarshal(), curHeight)
	}

	return &types.QueryBTCValidatorCurrentPowerResponse{Height: curHeight, VotingPower: power}, nil
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
	btcDelStore := k.btcDelegatorStore(sdkCtx, valPK)

	btcDels := []*types.BTCDelegatorDelegations{}
	pageRes, err := query.Paginate(btcDelStore, req.Pagination, func(key, value []byte) error {
		delBTCPK, err := bbn.NewBIP340PubKey(key)
		if err != nil {
			return err
		}

		curBTCDels, err := k.getBTCDelegatorDelegations(sdkCtx, valPK, delBTCPK)
		if err != nil {
			return err
		}

		btcDels = append(btcDels, curBTCDels)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryBTCValidatorDelegationsResponse{BtcDelegatorDelegations: btcDels, Pagination: pageRes}, nil
}

// BTCDelegation returns existing btc delegation by staking tx hash
func (k Keeper) BTCDelegation(ctx context.Context, req *types.QueryBTCDelegationRequest) (*types.QueryBTCDelegationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// decode staking tx hash
	stakingTxHash, err := chainhash.NewHashFromStr(req.StakingTxHashHex)
	if err != nil {
		return nil, err
	}

	// find BTC delegation
	btcDel := k.getBTCDelegation(sdkCtx, *stakingTxHash)
	if btcDel == nil {
		return nil, types.ErrBTCDelegationNotFound
	}

	// check whether it's active
	currentTip := k.btclcKeeper.GetTipInfo(sdkCtx)
	currentWValue := k.btccKeeper.GetParams(sdkCtx).CheckpointFinalizationTimeout
	isActive := btcDel.GetStatus(
		currentTip.Height,
		currentWValue,
	) == types.BTCDelegationStatus_ACTIVE

	// get its undelegation info
	var undelegationInfo *types.BTCUndelegationInfo
	if btcDel.BtcUndelegation != nil {
		undelegationInfo = &types.BTCUndelegationInfo{
			UnbondingTx:           btcDel.BtcUndelegation.UnbondingTx,
			ValidatorUnbondingSig: btcDel.BtcUndelegation.ValidatorUnbondingSig,
			CovenantUnbondingSig:  btcDel.BtcUndelegation.CovenantUnbondingSig,
		}
	}

	return &types.QueryBTCDelegationResponse{
		BtcPk:            btcDel.BtcPk,
		ValBtcPk:         btcDel.ValBtcPk,
		StartHeight:      btcDel.StartHeight,
		EndHeight:        btcDel.EndHeight,
		TotalSat:         btcDel.TotalSat,
		StakingTx:        hex.EncodeToString(btcDel.StakingTx.Tx),
		StakingScript:    hex.EncodeToString(btcDel.StakingTx.Script),
		Active:           isActive,
		UndelegationInfo: undelegationInfo,
	}, nil
}
