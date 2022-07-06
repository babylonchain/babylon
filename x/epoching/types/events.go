package types

// epoching module event types
const (
	EventTypeBeginEpoch            = "begin_epoch"
	EventTypeEndEpoch              = "end_epoch"
	EventTypeHandleQueuedMsg       = "handle_queue_msg"
	EventTypeHandleQueuedMsgFailed = "handle_queue_msg_failed"
	EventTypeSlashThreshold        = "slash_threshold"

	AttributeKeyEpoch             = "epoch"
	AttributeKeyNumSlashedVals    = "num_slashed_validators"
	AttributeKeyNumMaxVals        = "num_max_validators"
	AttributeKeyOriginalEventType = "original_event_type"
	AttributeKeyTxId              = "tx_id"
	AttributeKeyMsgId             = "msg_id"
	AttributeKeyErrorMsg          = "error_msg"
	AttributeValueCategory        = ModuleName
)
