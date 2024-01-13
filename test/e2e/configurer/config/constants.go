package config

const (
	// PropDepositBlocks estimated number of blocks it takes to deposit for a proposal
	PropDepositBlocks float32 = 10
	// PropVoteBlocks number of blocks it takes to vote for a single validator to vote for a proposal
	PropVoteBlocks float32 = 1.2
	// PropBufferBlocks number of blocks used as a calculation buffer
	PropBufferBlocks float32 = 6
)
