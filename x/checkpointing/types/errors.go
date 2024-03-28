package types

import errorsmod "cosmossdk.io/errors"

// x/checkpointing module sentinel errors
var (
	ErrCkptDoesNotExist        = errorsmod.Register(ModuleName, 1201, "raw checkpoint does not exist")
	ErrCkptAlreadyExist        = errorsmod.Register(ModuleName, 1202, "raw checkpoint already exists")
	ErrCkptHashNotEqual        = errorsmod.Register(ModuleName, 1203, "hash does not equal to raw checkpoint")
	ErrCkptNotAccumulating     = errorsmod.Register(ModuleName, 1204, "raw checkpoint is no longer accumulating BLS sigs")
	ErrCkptAlreadyVoted        = errorsmod.Register(ModuleName, 1205, "raw checkpoint already accumulated the validator")
	ErrInvalidRawCheckpoint    = errorsmod.Register(ModuleName, 1206, "raw checkpoint is invalid")
	ErrInvalidCkptStatus       = errorsmod.Register(ModuleName, 1207, "raw checkpoint's status is invalid")
	ErrInvalidPoP              = errorsmod.Register(ModuleName, 1208, "proof-of-possession is invalid")
	ErrBlsKeyDoesNotExist      = errorsmod.Register(ModuleName, 1209, "BLS public key does not exist")
	ErrBlsKeyAlreadyExist      = errorsmod.Register(ModuleName, 1210, "BLS public key already exists")
	ErrBlsPrivKeyDoesNotExist  = errorsmod.Register(ModuleName, 1211, "BLS private key does not exist")
	ErrInvalidBlsSignature     = errorsmod.Register(ModuleName, 1212, "BLS signature is invalid")
	ErrConflictingCheckpoint   = errorsmod.Register(ModuleName, 1213, "Conflicting checkpoint is found")
	ErrInvalidAppHash          = errorsmod.Register(ModuleName, 1214, "Provided app hash is Invalid")
	ErrInsufficientVotingPower = errorsmod.Register(ModuleName, 1215, "Accumulated voting power is not greater than 2/3 of total power")
	ErrFinalizedEpochNotFound  = errorsmod.Register(ModuleName, 1216, "cannot find a finalized epoch")
)
