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

// StakingKeeper defines the expected interface needed to retrieve validator staking status
type StakingKeeper interface {
	GetLastTotalPower(ctx sdk.Context) sdk.Uint
}

// EpochingKeeper defines the expected interface needed to retrieve epoch info
type EpochingKeeper interface {
	GetCurrentEpoch(ctx sdk.Context) sdk.Uint
}

// Event Hooks
// These can be utilized to communicate between a checkpointing keeper and another
// keeper which must take particular actions when raw checkpoints change
// state. The second keeper must implement this interface, which then the
// checkpointing keeper can call.

// CheckpointingHooks event hooks for raw checkpoint object (noalias)
type CheckpointingHooks interface {
	AfterBlsKeyRegistered(ctx sdk.Context, valAddr sdk.ValAddress) error // Must be called when a BLS key is registered
	AfterRawCheckpointConfirmed(ctx sdk.Context, epoch sdk.Uint) error   // Must be called when a raw checkpoint is CONFIRMED
}
