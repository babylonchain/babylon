package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/epoching module sentinel errors
var (
	ErrInvalidMsgType     = sdkerrors.Register(ModuleName, 1, "invalid message type in {MsgCreateValidator, MsgDelegate, MsgUndelegate, MsgBeginRedelegate} messages. use wrapped versions instead")
	ErrUnknownEpochNumber = sdkerrors.Register(ModuleName, 1, "the epoch number is not known in DB")
)
