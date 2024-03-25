package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/btcstaking module sentinel errors
var (
	ErrFpNotFound                   = errorsmod.Register(ModuleName, 1100, "the finality provider is not found")
	ErrBTCDelegatorNotFound         = errorsmod.Register(ModuleName, 1101, "the BTC delegator is not found")
	ErrBTCDelegationNotFound        = errorsmod.Register(ModuleName, 1102, "the BTC delegation is not found")
	ErrFpRegistered                 = errorsmod.Register(ModuleName, 1103, "the finality provider has already been registered")
	ErrFpAlreadySlashed             = errorsmod.Register(ModuleName, 1104, "the finality provider has already been slashed")
	ErrBTCStakingNotActivated       = errorsmod.Register(ModuleName, 1105, "the BTC staking protocol is not activated yet")
	ErrBTCHeightNotFound            = errorsmod.Register(ModuleName, 1106, "the BTC height is not found")
	ErrReusedStakingTx              = errorsmod.Register(ModuleName, 1107, "the BTC staking tx is already used")
	ErrInvalidCovenantPK            = errorsmod.Register(ModuleName, 1108, "the BTC staking tx specifies a wrong covenant PK")
	ErrInvalidStakingTx             = errorsmod.Register(ModuleName, 1109, "the BTC staking tx is not valid")
	ErrInvalidSlashingTx            = errorsmod.Register(ModuleName, 1110, "the BTC slashing tx is not valid")
	ErrInvalidCovenantSig           = errorsmod.Register(ModuleName, 1111, "the covenant signature is not valid")
	ErrCommissionLTMinRate          = errorsmod.Register(ModuleName, 1112, "commission cannot be less than min rate")
	ErrCommissionGTMaxRate          = errorsmod.Register(ModuleName, 1113, "commission cannot be more than one")
	ErrInvalidDelegationState       = errorsmod.Register(ModuleName, 1114, "Unexpected delegation state")
	ErrInvalidUnbondingTx           = errorsmod.Register(ModuleName, 1115, "the BTC unbonding tx is not valid")
	ErrRewardDistCacheNotFound      = errorsmod.Register(ModuleName, 1116, "the reward distribution cache is not found")
	ErrEmptyFpList                  = errorsmod.Register(ModuleName, 1117, "the finality provider list is empty")
	ErrInvalidProofOfPossession     = errorsmod.Register(ModuleName, 1118, "the proof of possession is not valid")
	ErrDuplicatedFp                 = errorsmod.Register(ModuleName, 1119, "the staking request contains duplicated finality provider public key")
	ErrInvalidBTCUndelegateReq      = errorsmod.Register(ModuleName, 1120, "invalid undelegation request")
	ErrVotingPowerTableNotUpdated   = errorsmod.Register(ModuleName, 1121, "voting power table has not been updated")
	ErrVotingPowerDistCacheNotFound = errorsmod.Register(ModuleName, 1122, "the voting power distribution cache is not found")
	ErrParamsNotFound               = errorsmod.Register(ModuleName, 1123, "the parameters are not found")
)
