package bindings

type BabylonQuery struct {
	Epoch                    *struct{}          `json:"epoch,omitempty"`
	LatestFinalizedEpochInfo *struct{}          `json:"latest_finalized_epoch_info,omitempty"`
	BtcTip                   *struct{}          `json:"btc_tip,omitempty"`
	BtcBaseHeader            *struct{}          `json:"btc_base_header,omitempty"`
	BtcHeaderByHash          *BtcHeaderByHash   `json:"btc_header_by_hash,omitempty"`
	BtcHeaderByHeight        *BtcHeaderByHeight `json:"btc_header_by_height,omitempty"`
}

type BtcHeaderByHash struct {
	Hash string `json:"hash,omitempty"`
}

type BtcHeaderByHeight struct {
	Height uint64 `json:"height"`
}

type CurrentEpochResponse struct {
	Epoch uint64 `json:"epoch"`
}

type LatestFinalizedEpochInfoResponse struct {
	EpochInfo *FinalizedEpochInfo `json:"epoch_info,omitempty"`
}

type FinalizedEpochInfo struct {
	EpochNumber     uint64 `json:"epoch_number"`
	LastBlockHeight uint64 `json:"last_block_height"`
}

type BtcBlockHeader struct {
	Version       int32  `json:"version"`
	PrevBlockhash string `json:"prev_blockhash,omitempty"`
	MerkleRoot    string `json:"merkle_root,omitempty"`
	Time          uint32 `json:"time"`
	Bits          uint32 `json:"bits"`
	Nonce         uint32 `json:"nonce"`
}

type BtcBlockHeaderInfo struct {
	Header *BtcBlockHeader `json:"header,omitempty"`
	Height uint64          `json:"height"`
}

type BtcTipResponse struct {
	HeaderInfo *BtcBlockHeaderInfo `json:"header_info,omitempty"`
}

type BtcBaseHeaderResponse struct {
	HeaderInfo *BtcBlockHeaderInfo `json:"header_info,omitempty"`
}

type BtcHeaderQueryResponse struct {
	HeaderInfo *BtcBlockHeaderInfo `json:"header_info,omitempty"`
}
