package bindings

type BabylonQuery struct {
	Epoch *struct{} `json:"epoch,omitempty"`
}

type CurrentEpochResponse struct {
	Epoch uint64 `json:"epoch"`
}
