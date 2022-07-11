package types

// epoching module event types
const (
	EventTypeBeginEpoch             = "begin_epoch"
	EventTypeEndEpoch               = "end_epoch"
	EventTypeHandleQueuedMsg        = "handle_queue_msg"
	EventTypeHandleQueuedMsgFailed  = "handle_queue_msg_failed"
	EventTypeSlashThreshold         = "slash_threshold"
	EventTypeWrappedDelegate        = "wrapped_delegate"
	EventTypeWrappedUndelegate      = "wrapped_undelegate"
	EventTypeWrappedBeginRedelegate = "wrapped_begin_redelegate"

	AttributeKeyEpoch              = "epoch"
	AttributeKeySlashedVotingPower = "slashed_voting_power"
	AttributeKeyTotalVotingPower   = "total_voting_power"
	AttributeKeySlashedValidators  = "slashed_validators"
	AttributeKeyOriginalEventType  = "original_event_type"
	AttributeKeyTxId               = "tx_id"
	AttributeKeyMsgId              = "msg_id"
	AttributeKeyErrorMsg           = "error_msg"
	AttributeKeyValidator          = "validator"
	AttributeKeySrcValidator       = "source_validator"
	AttributeKeyDstValidator       = "destination_validator"
	AttributeKeyEpochBoundary      = "epoch_boundary"

	AttributeValueCategory = ModuleName
)
