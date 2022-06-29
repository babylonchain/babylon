package types

// epoching module event types
const (
	EventTypeBeginEpoch     = "begin_epoch"
	EventTypeEndEpoch       = "end_epoch"
	EventTypeSlashThreshold = "slash_threshold"

	AttributeKeyEpoch          = "epoch"
	AttributeKeyNumSlashedVals = "num_slashed_validators"
	AttributeKeyNumMaxVals     = "num_max_validators"
	AttributeValueCategory     = ModuleName
)
