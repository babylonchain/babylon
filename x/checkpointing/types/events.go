package types

// checkpointing module event types
const (
	EventTypeRawCheckpointGenerated  = "generate_raw_checkpoint"
	EventTypeCheckpointStatusUpdated = "update_checkpoint_status"
	EventTypeBlsKeyRegistered        = "register_bls_key"
	EventTypeLastCommitHashObserved  = "observe_last_commit_hash"

	AttributeKeyValidator        = "validator"
	AttributeKeyCheckpointStatus = "checkpoint_status"
	AttributeKeyEpochNumber      = "epoch_number"
	AttributeKeyBlockHeight      = "block_height"
	AttributeValueCategory       = ModuleName
)
