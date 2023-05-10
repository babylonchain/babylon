package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/zoneconcierge module sentinel errors
var (
	ErrSample                  = errorsmod.Register(ModuleName, 1100, "sample error")
	ErrInvalidPacketTimeout    = errorsmod.Register(ModuleName, 1101, "invalid packet timeout")
	ErrInvalidVersion          = errorsmod.Register(ModuleName, 1102, "invalid version")
	ErrHeaderNotFound          = errorsmod.Register(ModuleName, 1103, "no header exists at this height")
	ErrInvalidHeader           = errorsmod.Register(ModuleName, 1104, "input header is invalid")
	ErrNoValidAncestorHeader   = errorsmod.Register(ModuleName, 1105, "no valid ancestor for this header")
	ErrForkNotFound            = errorsmod.Register(ModuleName, 1106, "cannot find fork")
	ErrInvalidForks            = errorsmod.Register(ModuleName, 1107, "input forks is invalid")
	ErrChainInfoNotFound       = errorsmod.Register(ModuleName, 1108, "no chain info exists")
	ErrEpochChainInfoNotFound  = errorsmod.Register(ModuleName, 1109, "no chain info exists at this epoch")
	ErrEpochHeadersNotFound    = errorsmod.Register(ModuleName, 1110, "no timestamped header exists at this epoch")
	ErrFinalizedEpochNotFound  = errorsmod.Register(ModuleName, 1111, "cannot find a finalized epoch")
	ErrInvalidProofEpochSealed = errorsmod.Register(ModuleName, 1112, "invalid ProofEpochSealed")
	ErrInvalidMerkleProof      = errorsmod.Register(ModuleName, 1113, "invalid Merkle inclusion proof")
	ErrInvalidChainInfo        = errorsmod.Register(ModuleName, 1114, "invalid chain info")
	ErrInvalidChainIDs         = errorsmod.Register(ModuleName, 1115, "chain ids contain duplicates or empty strings")
)
