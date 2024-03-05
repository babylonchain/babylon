package types

// ToResponse generates a RawCheckpointResponse struct from RawCheckpoint.
func (r *RawCheckpoint) ToResponse() *RawCheckpointResponse {
	return &RawCheckpointResponse{
		EpochNum:     r.EpochNum,
		BlockHashHex: r.BlockHash.String(),
		Bitmap:       r.Bitmap,
		BlsMultiSig:  r.BlsMultiSig,
	}
}

// ToResponse generates a RawCheckpointWithMetaResponse struct from RawCheckpointWithMeta.
func (r *RawCheckpointWithMeta) ToResponse() *RawCheckpointWithMetaResponse {
	return &RawCheckpointWithMetaResponse{
		Ckpt:       r.Ckpt.ToResponse(),
		Status:     r.Status,
		StatusDesc: r.Status.String(),
		BlsAggrPk:  r.BlsAggrPk,
		PowerSum:   r.PowerSum,
		Lifecycle:  NewCheckpointStateUpdateResponse(r.Lifecycle),
	}
}

// ToResponse generates a CheckpointStateUpdateResponse struct from CheckpointStateUpdate.
func (c *CheckpointStateUpdate) ToResponse() *CheckpointStateUpdateResponse {
	return &CheckpointStateUpdateResponse{
		State:       c.State,
		StatusDesc:  c.State.String(),
		BlockHeight: c.BlockHeight,
		BlockTime:   c.BlockTime,
	}
}

// NewCheckpointStateUpdateResponse creates a new CheckpointStateUpdateResponse slice from []*CheckpointStateUpdate.
func NewCheckpointStateUpdateResponse(cs []*CheckpointStateUpdate) (resp []*CheckpointStateUpdateResponse) {
	resp = make([]*CheckpointStateUpdateResponse, len(cs))
	for i, c := range cs {
		resp[i] = c.ToResponse()
	}
	return resp
}
