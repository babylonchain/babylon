package bindings

type BabylonQuery struct {
	Epoch                *struct{}          `json:"epoch,omitempty"`
	LatestFinalizedEpoch *struct{}          `json:"latest_finalized_epoch,omitempty"`
	BtcTip               *struct{}          `json:"btc_tip,omitempty"`
	BtcBaseHeader        *struct{}          `json:"btc_base_header,omitempty"`
	BtcHeaderByHash      *BtcHeaderByHash   `json:"btc_header_by_hash,omitempty"`
	BtcHeaderByNumber    *BtcHeaderByNumber `json:"btc_header_by_number,omitempty"`
}

type BtcHeaderByHash struct {
	Hash string `json:"hash,omitempty"`
}

type BtcHeaderByNumber struct {
	Height uint64 `json:"height,omitempty"`
}

type CurrentEpochResponse struct {
	Epoch uint64 `json:"epoch"`
}

type LatestFinalizedEpochResponse struct {
	Epoch uint64 `json:"epoch"`
}

type BtcBlockHeader struct {
	Version       int32  `json:"version,omitempty"`
	PrevBlockhash string `json:"prev_blockhash,omitempty"`
	MerkleRoot    string `json:"merkle_root,omitempty"`
	Time          uint32 `json:"time,omitempty"`
	Bits          uint32 `json:"bits,omitempty"`
	Nonce         uint32 `json:"nonce,omitempty"`
}

type BtcBlockHeaderInfo struct {
	Header *BtcBlockHeader `json:"header,omitempty"`
	Height uint64          `json:"height,omitempty"`
}

type BtcTipResponse struct {
	HeaderInfo *BtcBlockHeaderInfo `json:"header_info,omitempty"`
}

type BtcBaseHeaderResponse struct {
	HeaderInfo *BtcBlockHeaderInfo `json:"header_info,omitempty"`
}

type BtcHeaderByQueryResponse struct {
	HeaderInfo *BtcBlockHeaderInfo `json:"header_info,omitempty"`
}
