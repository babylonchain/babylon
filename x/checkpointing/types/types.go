package types

type BlsSigHash []byte

type RawCkptHash []byte

type CkptStatus uint32

const (
	// UNCHECKPOINTED indicates the checkpoint has not appeared on BTC
	UNCHECKPOINTED uint32 = 0
	// CHECKPOINTED_NOT_CONFIRMED indicates the checkpoint has been checkpointed on BTC but has insufficent confirmation
	CHECKPOINTED_NOT_CONFIRMED = 1
	// CONFIRMED indicates the checkpoint has sufficient confirmation depth on BTC
	CONFIRMED = 2
)

var RAW_CKPT_STATUS = [3]uint32{UNCHECKPOINTED, CHECKPOINTED_NOT_CONFIRMED, CONFIRMED}
