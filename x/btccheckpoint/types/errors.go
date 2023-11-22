package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/btccheckpoint module sentinel errors
var (
	ErrInvalidCheckpointProof            = errorsmod.Register(ModuleName, 1100, "Invalid checkpoint proof")
	ErrDuplicatedSubmission              = errorsmod.Register(ModuleName, 1101, "Duplicated submission")
	ErrNoCheckpointsForPreviousEpoch     = errorsmod.Register(ModuleName, 1102, "No checkpoints for previous epoch")
	ErrInvalidHeader                     = errorsmod.Register(ModuleName, 1103, "Proof headers are invalid")
	ErrProvidedHeaderDoesNotHaveAncestor = errorsmod.Register(ModuleName, 1104, "Proof header does not have ancestor in previous epoch")
	ErrEpochAlreadyFinalized             = errorsmod.Register(ModuleName, 1105, "Submission denied. Epoch already finalized")
)
