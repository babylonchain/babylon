package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/epoching module sentinel errors
var (
	ErrUnwrappedMsgType         = sdkerrors.Register(ModuleName, 1, "invalid message type in {MsgCreateValidator, MsgDelegate, MsgUndelegate, MsgBeginRedelegate} messages. use wrapped versions instead")
	ErrInvalidQueuedMessageType = sdkerrors.Register(ModuleName, 2, "invalid message type of a QueuedMessage")
	ErrUnknownEpochNumber       = sdkerrors.Register(ModuleName, 3, "the epoch number is not known in DB")
	ErrUnknownQueueLen          = sdkerrors.Register(ModuleName, 4, "the msg queue length is not known in DB")
	ErrUnknownSlashedValSetSize = sdkerrors.Register(ModuleName, 5, "the slashed validator set size is not known in DB")
)
