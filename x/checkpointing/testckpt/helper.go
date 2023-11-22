package testckpt

import (
	"context"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/babylonchain/babylon/app"
	appparams "github.com/babylonchain/babylon/app/params"
	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/x/checkpointing/keeper"
	"github.com/babylonchain/babylon/x/checkpointing/types"
	epochingkeeper "github.com/babylonchain/babylon/x/epoching/keeper"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
)

// Helper is a structure which wraps the entire app and exposes functionalities for testing the epoching module
type Helper struct {
	t *testing.T

	Ctx                 context.Context
	App                 *app.BabylonApp
	CheckpointingKeeper *keeper.Keeper
	MsgSrvr             types.MsgServer
	QueryClient         types.QueryClient
	StakingKeeper       *stakingkeeper.Keeper
	EpochingKeeper      *epochingkeeper.Keeper

	GenAccs []authtypes.GenesisAccount
}

// CreateValidator calls handler to create a new staking validator
func (h *Helper) CreateValidator(addr sdk.ValAddress, pk cryptotypes.PubKey, blsPK *bls12381.PublicKey, pop *types.ProofOfPossession, stakeAmount sdkmath.Int) {
	coin := sdk.NewCoin(appparams.DefaultBondDenom, stakeAmount)
	h.createValidator(addr, pk, blsPK, pop, coin)
}

// CreateValidatorWithValPower calls handler to create a new staking validator with zero commission
func (h *Helper) CreateValidatorWithValPower(addr sdk.ValAddress, pk cryptotypes.PubKey,
	blsPK *bls12381.PublicKey, pop *types.ProofOfPossession, valPower int64) sdkmath.Int {
	amount := h.StakingKeeper.TokensFromConsensusPower(h.Ctx, valPower)
	coin := sdk.NewCoin(appparams.DefaultBondDenom, amount)
	h.createValidator(addr, pk, blsPK, pop, coin)
	return amount
}

// CreateValidatorMsg returns a message used to create validator in this service.
func (h *Helper) CreateValidatorMsg(addr sdk.ValAddress, pk cryptotypes.PubKey, blsPK *bls12381.PublicKey, pop *types.ProofOfPossession, stakeAmount sdkmath.Int) *types.MsgWrappedCreateValidator {
	coin := sdk.NewCoin(appparams.DefaultBondDenom, stakeAmount)
	msg, err := stakingtypes.NewMsgCreateValidator(addr.String(), pk, coin, stakingtypes.Description{}, ZeroCommission(), sdkmath.OneInt())
	require.NoError(h.t, err)
	wmsg, err := types.NewMsgWrappedCreateValidator(msg, blsPK, pop)
	require.NoError(h.t, err)
	return wmsg
}

func (h *Helper) createValidator(addr sdk.ValAddress, pk cryptotypes.PubKey, blsPK *bls12381.PublicKey, pop *types.ProofOfPossession, coin sdk.Coin) {
	h.Handle(func(ctx context.Context) (proto.Message, error) {
		return h.CreateValidatorMsg(addr, pk, blsPK, pop, coin.Amount), nil
	})
}

// Handle executes an action function with the Helper's context, wraps the result into an SDK service result, and performs two assertions before returning it
func (h *Helper) Handle(action func(context.Context) (proto.Message, error)) *sdk.Result {
	res, err := action(h.Ctx)
	r, _ := sdk.WrapServiceResult(sdk.UnwrapSDKContext(h.Ctx), res, err)
	require.NotNil(h.t, r)
	require.NoError(h.t, err)
	return r
}

// ZeroCommission constructs a commission rates with all zeros.
func ZeroCommission() stakingtypes.CommissionRates {
	return stakingtypes.NewCommissionRates(sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec(), sdkmath.LegacyZeroDec())
}
