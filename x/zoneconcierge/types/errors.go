package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/zoneconcierge module sentinel errors
var (
	ErrInvalidVersion          = errorsmod.Register(ModuleName, 1101, "invalid version")
	ErrHeaderNotFound          = errorsmod.Register(ModuleName, 1102, "no header exists at this height")
	ErrInvalidHeader           = errorsmod.Register(ModuleName, 1103, "input header is invalid")
	ErrChainInfoNotFound       = errorsmod.Register(ModuleName, 1104, "no chain info exists")
	ErrEpochChainInfoNotFound  = errorsmod.Register(ModuleName, 1105, "no chain info exists at this epoch")
	ErrEpochHeadersNotFound    = errorsmod.Register(ModuleName, 1106, "no timestamped header exists at this epoch")
	ErrFinalizedEpochNotFound  = errorsmod.Register(ModuleName, 1107, "cannot find a finalized epoch")
	ErrInvalidProofEpochSealed = errorsmod.Register(ModuleName, 1108, "invalid ProofEpochSealed")
	ErrInvalidMerkleProof      = errorsmod.Register(ModuleName, 1109, "invalid Merkle inclusion proof")
	ErrInvalidChainInfo        = errorsmod.Register(ModuleName, 1110, "invalid chain info")
	ErrInvalidChainIDs         = errorsmod.Register(ModuleName, 1111, "chain ids contain duplicates or empty strings")
)
