package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/epoching module sentinel errors
var (
	ErrUnwrappedMsgType          = sdkerrors.Register(ModuleName, 1, "invalid message type in {MsgCreateValidator, MsgDelegate, MsgUndelegate, MsgBeginRedelegate} messages. use wrapped versions instead")
	ErrInvalidQueuedMessageType  = sdkerrors.Register(ModuleName, 2, "invalid message type of a QueuedMessage")
	ErrUnknownEpochNumber        = sdkerrors.Register(ModuleName, 3, "the epoch number is not known in DB")
	ErrUnknownQueueLen           = sdkerrors.Register(ModuleName, 4, "the msg queue length is not known in DB")
	ErrUnknownSlashedVotingPower = sdkerrors.Register(ModuleName, 5, "the slashed voting power is not known in DB. Maybe the epoch has been checkpointed?")
	ErrUnknownValidator          = sdkerrors.Register(ModuleName, 6, "the slashed validator is not in the validator set.")
	ErrUnknownTotalVotingPower   = sdkerrors.Register(ModuleName, 7, "the total voting power is not known in DB.")
	ErrMarshal                   = sdkerrors.Register(ModuleName, 8, "marshal error.")
	ErrUnmarshal                 = sdkerrors.Register(ModuleName, 9, "unmarshal error.")
	ErrNoWrappedMsg              = sdkerrors.Register(ModuleName, 10, "the wrapped msg contains no msg inside.")
)
