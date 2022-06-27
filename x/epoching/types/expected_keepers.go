package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

type StakingMsgServer interface {
	// CreateValidator defines a method for creating a new validator.
	CreateValidator(context.Context, *stakingtypes.MsgCreateValidator) (*stakingtypes.MsgCreateValidatorResponse, error)
	// Delegate defines a method for performing a delegation of coins
	// from a delegator to a validator.
	Delegate(context.Context, *stakingtypes.MsgDelegate) (*stakingtypes.MsgDelegateResponse, error)
	// BeginRedelegate defines a method for performing a redelegation
	// of coins from a delegator and source validator to a destination validator.
	BeginRedelegate(context.Context, *stakingtypes.MsgBeginRedelegate) (*stakingtypes.MsgBeginRedelegateResponse, error)
	// Undelegate defines a method for performing an undelegation from a
	// delegate and a validator.
	Undelegate(context.Context, *stakingtypes.MsgUndelegate) (*stakingtypes.MsgUndelegateResponse, error)
}

// TODO: add interfaces of staking, slashing and evidence used in epoching

// Event Hooks
// These can be utilized to communicate between a staking keeper and another
// keeper which must take particular actions when validators/delegators change
// state. The second keeper must implement this interface, which then the
// staking keeper can call.

// EpochingHooks event hooks for epoching validator object (noalias)
type EpochingHooks interface {
	AfterEpochBegins(ctx sdk.Context, epoch sdk.Uint) error // Must be called after an epoch begins
	AfterEpochEnds(ctx sdk.Context, epoch sdk.Uint) error   // Must be called after an epoch ends
}
