package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	abci "github.com/tendermint/tendermint/abci/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	// Methods imported from bank should be defined here
}

// StakingKeeper defines the staking module interface contract needed by the
// epoching module.
type StakingKeeper interface {
	GetParams(ctx sdk.Context) stakingtypes.Params
	DequeueAllMatureUBDQueue(ctx sdk.Context, currTime time.Time) (matureUnbonds []stakingtypes.DVPair)
	CompleteUnbonding(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error)
	DequeueAllMatureRedelegationQueue(ctx sdk.Context, currTime time.Time) (matureRedelegations []stakingtypes.DVVTriplet)
	CompleteRedelegation(ctx sdk.Context, delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress) (sdk.Coins, error)
	ApplyAndReturnValidatorSetUpdates(ctx sdk.Context) ([]abci.ValidatorUpdate, error)
	IterateLastValidatorPowers(ctx sdk.Context, handler func(operator sdk.ValAddress, power int64) bool)
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (stakingtypes.Validator, bool)
	GetValidatorDelegations(ctx sdk.Context, valAddr sdk.ValAddress) []stakingtypes.Delegation
	HasMaxUnbondingDelegationEntries(ctx sdk.Context, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) bool
	BondDenom(ctx sdk.Context) string
	HasReceivingRedelegation(ctx sdk.Context, delAddr sdk.AccAddress, valDstAddr sdk.ValAddress) bool
	HasMaxRedelegationEntries(ctx sdk.Context, delegatorAddr sdk.AccAddress, validatorSrcAddr, validatorDstAddr sdk.ValAddress) bool
	ValidateUnbondAmount(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amt sdk.Int) (sdk.Dec, error)
	ValidatorQueueIterator(ctx sdk.Context, endTime time.Time, endHeight int64) sdk.Iterator
	UnbondingToUnbonded(ctx sdk.Context, validator stakingtypes.Validator) stakingtypes.Validator
	RemoveValidator(ctx sdk.Context, address sdk.ValAddress)
}

// Event Hooks
// These can be utilized to communicate between an epoching keeper and another
// keeper which must take particular actions when validators/delegators change
// state. The second keeper must implement this interface, which then the
// epoching keeper can call.

// EpochingHooks event hooks for epoching validator object (noalias)
type EpochingHooks interface {
	AfterEpochBegins(ctx sdk.Context, epoch uint64)            // Must be called after an epoch begins
	AfterEpochEnds(ctx sdk.Context, epoch uint64)              // Must be called after an epoch ends
	BeforeSlashThreshold(ctx sdk.Context, valSet ValidatorSet) // Must be called before a certain threshold (1/3 or 2/3) of validators are slashed in a single epoch
}

// StakingHooks event hooks for staking validator object (noalias)
type StakingHooks interface {
	BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec)                     // Must be called right before a validator is slashed
	AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) error                                  // Must be called when a validator is created
	AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error        // Must be called when a validator is deleted
	AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error         // Must be called when a validator is bonded
	AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error // Must be called when a validator begins unbonding
}
