package types

import (
	"context"
	"time"

	storetypes "cosmossdk.io/store/types"
	cmtprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	// Methods imported from bank should be defined here
}

// StakingKeeper defines the staking module interface contract needed by the
// epoching module.
type StakingKeeper interface {
	GetParams(ctx context.Context) (stakingtypes.Params, error)
	DequeueAllMatureUBDQueue(ctx context.Context, currTime time.Time) ([]stakingtypes.DVPair, error)
	CompleteUnbonding(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (sdk.Coins, error)
	DequeueAllMatureRedelegationQueue(ctx context.Context, currTime time.Time) ([]stakingtypes.DVVTriplet, error)
	CompleteRedelegation(ctx context.Context, delAddr sdk.AccAddress, valSrcAddr, valDstAddr sdk.ValAddress) (sdk.Coins, error)
	ApplyAndReturnValidatorSetUpdates(ctx context.Context) ([]abci.ValidatorUpdate, error)
	IterateLastValidatorPowers(ctx context.Context, handler func(operator sdk.ValAddress, power int64) bool) error
	GetValidator(ctx context.Context, addr sdk.ValAddress) (stakingtypes.Validator, error)
	GetValidatorDelegations(ctx context.Context, valAddr sdk.ValAddress) ([]stakingtypes.Delegation, error)
	HasMaxUnbondingDelegationEntries(ctx context.Context, delegatorAddr sdk.AccAddress, validatorAddr sdk.ValAddress) (bool, error)
	BondDenom(ctx context.Context) (string, error)
	HasReceivingRedelegation(ctx context.Context, delAddr sdk.AccAddress, valDstAddr sdk.ValAddress) (bool, error)
	HasMaxRedelegationEntries(ctx context.Context, delegatorAddr sdk.AccAddress, validatorSrcAddr, validatorDstAddr sdk.ValAddress) (bool, error)
	ValidateUnbondAmount(ctx context.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amt math.Int) (math.LegacyDec, error)
	ValidatorQueueIterator(ctx context.Context, endTime time.Time, endHeight int64) (storetypes.Iterator, error)
	UnbondAllMatureValidators(ctx context.Context) error
	GetValidatorByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (stakingtypes.Validator, error)
	GetPubKeyByConsAddr(ctx context.Context, consAddr sdk.ConsAddress) (cmtprotocrypto.PublicKey, error)
}

// Event Hooks
// These can be utilized to communicate between an epoching keeper and another
// keeper which must take particular actions when validators/delegators change
// state. The second keeper must implement this interface, which then the
// epoching keeper can call.

// EpochingHooks event hooks for epoching validator object (noalias)
type EpochingHooks interface {
	AfterEpochBegins(ctx context.Context, epoch uint64)            // Must be called after an epoch begins
	AfterEpochEnds(ctx context.Context, epoch uint64)              // Must be called after an epoch ends
	BeforeSlashThreshold(ctx context.Context, valSet ValidatorSet) // Must be called before a certain threshold (1/3 or 2/3) of validators are slashed in a single epoch
}

// StakingHooks event hooks for staking validator object (noalias)
type StakingHooks interface {
	BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction math.LegacyDec)              // Must be called right before a validator is slashed
	AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress) error                                  // Must be called when a validator is created
	AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error        // Must be called when a validator is deleted
	AfterValidatorBonded(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error         // Must be called when a validator is bonded
	AfterValidatorBeginUnbonding(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) error // Must be called when a validator begins unbonding
}
