package bindings

type BabylonQuery struct {
	Epoch                *struct{} `json:"epoch,omitempty"`
	LatestFinalizedEpoch *struct{} `json:"latest_finalized_epoch,omitempty"`
}

type CurrentEpochResponse struct {
	Epoch uint64 `json:"epoch"`
}

type LatestFinalizedEpochResponse struct {
	Epoch uint64 `json:"epoch"`
}
