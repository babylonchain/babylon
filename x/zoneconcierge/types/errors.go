package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/zoneconcierge module sentinel errors
var (
	ErrSample                = sdkerrors.Register(ModuleName, 1100, "sample error")
	ErrInvalidPacketTimeout  = sdkerrors.Register(ModuleName, 1101, "invalid packet timeout")
	ErrInvalidVersion        = sdkerrors.Register(ModuleName, 1102, "invalid version")
	ErrNoChainInfo           = sdkerrors.Register(ModuleName, 1103, "chain does not exist")
	ErrHeaderNotExist        = sdkerrors.Register(ModuleName, 1104, "no header exists at this height")
	ErrNoValidAncestorHeader = sdkerrors.Register(ModuleName, 1105, "no valid ancestor for this header")
)
