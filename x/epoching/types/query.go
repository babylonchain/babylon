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
