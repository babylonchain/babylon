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
