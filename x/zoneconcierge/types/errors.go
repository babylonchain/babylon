package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/zoneconcierge module sentinel errors
var (
	ErrSample               = sdkerrors.Register(ModuleName, 1100, "sample error")
	ErrInvalidPacketTimeout = sdkerrors.Register(ModuleName, 1101, "invalid packet timeout")
	ErrInvalidVersion       = sdkerrors.Register(ModuleName, 1102, "invalid version")
	ErrNoChainInfo          = sdkerrors.Register(ModuleName, 1103, "chain info does not exist")
	ErrInvalidHeight        = sdkerrors.Register(ModuleName, 1104, "the indexed header has an invalid height")
)
