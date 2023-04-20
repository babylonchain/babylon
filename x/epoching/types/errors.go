package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrUnwrappedMsgType = errorsmod.Register(ModuleName, 1, `
										invalid message type in {MsgCreateValidator, MsgDelegate, MsgUndelegate, MsgBeginRedelegate}
										messages. For creating a validator use the wrapped version under 'tx checkpointing create-validator'
										and for the other messages use the wrapped versions under 'tx epoching {delegate,undelegate,redelegate}'`)
	ErrInvalidQueuedMessageType  = errorsmod.Register(ModuleName, 2, "invalid message type of a QueuedMessage")
	ErrUnknownEpochNumber        = errorsmod.Register(ModuleName, 3, "the epoch number is not known in DB")
	ErrUnknownSlashedVotingPower = errorsmod.Register(ModuleName, 5, "the slashed voting power is not known in DB. Maybe the epoch has been checkpointed?")
	ErrUnknownValidator          = errorsmod.Register(ModuleName, 6, "the validator is not known in the validator set.")
	ErrUnknownTotalVotingPower   = errorsmod.Register(ModuleName, 7, "the total voting power is not known in DB.")
	ErrMarshal                   = errorsmod.Register(ModuleName, 8, "marshal error.")
	ErrUnmarshal                 = errorsmod.Register(ModuleName, 9, "unmarshal error.")
	ErrNoWrappedMsg              = errorsmod.Register(ModuleName, 10, "the wrapped msg contains no msg inside.")
	ErrZeroEpochMsg              = errorsmod.Register(ModuleName, 11, "the 0-th epoch does not handle messages")
	ErrInvalidEpoch              = errorsmod.Register(ModuleName, 12, "the epoch is invalid")
	ErrInvalidHeight             = errorsmod.Register(ModuleName, 13, "the height is invalid")
	ErrInsufficientBalance       = errorsmod.Register(ModuleName, 14, "the delegator has insufficient balance to perform delegate")
)
