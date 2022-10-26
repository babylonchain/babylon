package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/btccheckpoint module sentinel errors
var (
	ErrInvalidCheckpointProof            = sdkerrors.Register(ModuleName, 1100, "Invalid checkpoint proof")
	ErrDuplicatedSubmission              = sdkerrors.Register(ModuleName, 1101, "Duplicated submission")
	ErrNoCheckpointsForPreviousEpoch     = sdkerrors.Register(ModuleName, 1102, "No checkpoints for previous epoch")
	ErrUnknownHeader                     = sdkerrors.Register(ModuleName, 1103, "Proof headers are not known in BTCLightclient")
	ErrProvidedHeaderFromDifferentForks  = sdkerrors.Register(ModuleName, 1104, "Proof header from different forks")
	ErrProvidedHeaderDoesNotHaveAncestor = sdkerrors.Register(ModuleName, 1105, "Proof header does not have ancestor in previous epoch")
	ErrEpochAlreadyFinalized             = sdkerrors.Register(ModuleName, 1106, "Submission denied. Epoch already finalized")
)
