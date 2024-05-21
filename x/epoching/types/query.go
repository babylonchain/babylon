package types

import (
	"encoding/hex"
)

// ToResponse parses a Epoch into a query response epoch struct.
func (e *Epoch) ToResponse() *EpochResponse {
	return &EpochResponse{
		EpochNumber:          e.EpochNumber,
		CurrentEpochInterval: e.CurrentEpochInterval,
		FirstBlockHeight:     e.FirstBlockHeight,
		LastBlockTime:        e.LastBlockTime,
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

// ToResponse parses a ValStateUpdate into a query response valset update struct.
func (v *ValStateUpdate) ToResponse() *ValStateUpdateResponse {
	return &ValStateUpdateResponse{
		StateDesc:   v.State.String(),
		BlockHeight: v.BlockHeight,
		BlockTime:   v.BlockTime,
	}
}

// NewValsetUpdateResponses parses all the valset updates as response.
func NewValsetUpdateResponses(vs []*ValStateUpdate) []*ValStateUpdateResponse {
	resp := make([]*ValStateUpdateResponse, len(vs))
	for i, v := range vs {
		resp[i] = v.ToResponse()
	}
	return resp
}

// NewQueuedMessagesResponse parses all the queued messages as response.
func NewQueuedMessagesResponse(msgs []*QueuedMessage) []*QueuedMessageResponse {
	resp := make([]*QueuedMessageResponse, len(msgs))
	for i, m := range msgs {
		resp[i] = m.ToResponse()
	}
	return resp
}
