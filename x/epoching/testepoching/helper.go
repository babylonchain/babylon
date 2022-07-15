package testepoching

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Helper is a structure which wraps the staking handler
// and provides methods useful in tests
type Helper struct {
	t *testing.T

	k   stakingkeeper.Keeper
	ctx sdk.Context
}

// NewHelper creates staking Handler wrapper for tests
func NewHelper(t *testing.T, ctx sdk.Context, k stakingkeeper.Keeper) *Helper {
	return &Helper{t, k, ctx}
}

// CreateValidator calls handler to create a new staking validator
func (h *Helper) CreateValidator(addr sdk.ValAddress, pk cryptotypes.PubKey, stakeAmount sdk.Int, ok bool) {
	coin := sdk.NewCoin(sdk.DefaultBondDenom, stakeAmount)
	h.createValidator(addr, pk, coin, ok)
}

// CreateValidatorWithValPower calls handler to create a new staking validator with zero
// commission
func (h *Helper) CreateValidatorWithValPower(addr sdk.ValAddress, pk cryptotypes.PubKey, valPower int64, ok bool) sdk.Int {
	amount := h.k.TokensFromConsensusPower(h.ctx, valPower)
	coin := sdk.NewCoin(sdk.DefaultBondDenom, amount)
	h.createValidator(addr, pk, coin, ok)
	return amount
}

// CreateValidatorMsg returns a message used to create validator in this service.
func (h *Helper) CreateValidatorMsg(addr sdk.ValAddress, pk cryptotypes.PubKey, stakeAmount sdk.Int) *stakingtypes.MsgCreateValidator {
	coin := sdk.NewCoin(sdk.DefaultBondDenom, stakeAmount)
	msg, err := stakingtypes.NewMsgCreateValidator(addr, pk, coin, stakingtypes.Description{}, ZeroCommission(), sdk.OneInt())
	require.NoError(h.t, err)
	return msg
}

func (h *Helper) createValidator(addr sdk.ValAddress, pk cryptotypes.PubKey, coin sdk.Coin, ok bool) {
	msg, err := stakingtypes.NewMsgCreateValidator(addr, pk, coin, stakingtypes.Description{}, ZeroCommission(), sdk.OneInt())
	require.NoError(h.t, err)
	h.Handle(msg, ok)
}

// Delegate calls handler to delegate stake for a validator
func (h *Helper) Delegate(delegator sdk.AccAddress, val sdk.ValAddress, amount sdk.Int) {
	coin := sdk.NewCoin(sdk.DefaultBondDenom, amount)
	msg := stakingtypes.NewMsgDelegate(delegator, val, coin)
	h.Handle(msg, true)
}

// DelegateWithPower calls handler to delegate stake for a validator
func (h *Helper) DelegateWithPower(delegator sdk.AccAddress, val sdk.ValAddress, power int64) {
	coin := sdk.NewCoin(sdk.DefaultBondDenom, h.k.TokensFromConsensusPower(h.ctx, power))
	msg := stakingtypes.NewMsgDelegate(delegator, val, coin)
	h.Handle(msg, true)
}

// Undelegate calls handler to unbound some stake from a validator.
func (h *Helper) Undelegate(delegator sdk.AccAddress, val sdk.ValAddress, amount sdk.Int, ok bool) *sdk.Result {
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, amount)
	msg := stakingtypes.NewMsgUndelegate(delegator, val, unbondAmt)
	return h.Handle(msg, ok)
}

// Handle calls staking handler on a given message
func (h *Helper) Handle(msg sdk.Msg, ok bool) *sdk.Result {
	handler := staking.NewHandler(h.k)
	res, err := handler(h.ctx, msg)
	if ok {
		require.NoError(h.t, err)
		require.NotNil(h.t, res)
	} else {
		require.Error(h.t, err)
		require.Nil(h.t, res)
	}
	return res
}

// CheckValidator asserts that a validor exists and has a given status (if status!="")
// and if has a right jailed flag.
func (h *Helper) CheckValidator(addr sdk.ValAddress, status stakingtypes.BondStatus, jailed bool) stakingtypes.Validator {
	v, ok := h.k.GetValidator(h.ctx, addr)
	require.True(h.t, ok)
	require.Equal(h.t, jailed, v.Jailed, "wrong Jalied status")
	if status >= 0 {
		require.Equal(h.t, status, v.Status)
	}
	return v
}

// CheckDelegator asserts that a delegator exists
func (h *Helper) CheckDelegator(delegator sdk.AccAddress, val sdk.ValAddress, found bool) {
	_, ok := h.k.GetDelegation(h.ctx, delegator, val)
	require.Equal(h.t, ok, found)
}

// TurnBlock calls EndBlocker and updates the block time
func (h *Helper) TurnBlock(newTime time.Time) sdk.Context {
	h.ctx = h.ctx.WithBlockTime(newTime)
	staking.EndBlocker(h.ctx, h.k)
	return h.ctx
}

// TurnBlockTimeDiff calls EndBlocker and updates the block time by adding the
// duration to the current block time
func (h *Helper) TurnBlockTimeDiff(diff time.Duration) sdk.Context {
	h.ctx = h.ctx.WithBlockTime(h.ctx.BlockHeader().Time.Add(diff))
	staking.EndBlocker(h.ctx, h.k)
	return h.ctx
}

// ZeroCommission constructs a commission rates with all zeros.
func ZeroCommission() stakingtypes.CommissionRates {
	return stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
}
