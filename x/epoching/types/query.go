package types

import "encoding/hex"

// ToResponse parses a Epoch into a query response epoch struct.
func (e *Epoch) ToResponse() *EpochResponse {
	return &EpochResponse{
		EpochNumber:          e.EpochNumber,
		CurrentEpochInterval: e.CurrentEpochInterval,
		FirstBlockHeight:     e.FirstBlockHeight,
		LastBlockTime:        e.LastBlockTime,
		AppHashRootHex:       hex.EncodeToString(e.AppHashRoot),
		SealerAppHashHex:     hex.EncodeToString(e.SealerAppHash),
		SealerBlockHash:      hex.EncodeToString(e.SealerBlockHash),
	}
}

// ToResponse parses a QueuedMessage into a query response queued message struct.
func (q *QueuedMessage) ToResponse() *QueuedMessageResponse {
	return &QueuedMessageResponse{
		TxId:        hex.EncodeToString(q.TxId),
		MsgId:       hex.EncodeToString(q.MsgId),
		BlockHeight: q.BlockHeight,
		BlockTime:   q.BlockTime,
		Msg:         q.UnwrapToSdkMsg().String(),
	}
}
