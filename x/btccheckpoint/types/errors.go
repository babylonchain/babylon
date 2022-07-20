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
	ErrProvidedHeaderFromDifferentForks  = sdkerrors.Register(ModuleName, 1103, "Proof header from different forks")
	ErrProvidedHeaderDoesNotHaveAncestor = sdkerrors.Register(ModuleName, 1104, "Proof header does not have ancestor in previous epoch")
	ErrEpochAlreadyConfirmedOrFinalized  = sdkerrors.Register(ModuleName, 1105, "Submission denied. Epoch already confirmed/finalized")
)
