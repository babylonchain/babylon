package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
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

// TODO: add interfaces of staking, slashing and evidence used in epoching

// Event Hooks
// These can be utilized to communicate between a staking keeper and another
// keeper which must take particular actions when validators/delegators change
// state. The second keeper must implement this interface, which then the
// staking keeper can call.

// EpochingHooks event hooks for epoching validator object (noalias)
type EpochingHooks interface {
	BeginEpoch(ctx sdk.Context, epoch sdk.Uint) error // Must be called when an epoch begins
	EndEpoch(ctx sdk.Context, epoch sdk.Uint) error   // Must be called when an epoch ends
}
