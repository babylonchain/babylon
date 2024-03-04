package types

// ToResponse parses one BTC Header Info to BTCHeaderInfoResp.
func (b *BTCHeaderInfo) ToResponse() *BTCHeaderInfoResponse {
	return &BTCHeaderInfoResponse{
		HeaderHex: b.Header.MarshalHex(),
		HashHex:   b.Hash.MarshalHex(),
		Height:    b.Height,
		Work:      *b.Work,
	}
}

// ParseBTCHeadersToResponse parses the infos into resposes.
func ParseBTCHeadersToResponse(infos []*BTCHeaderInfo) (resp []*BTCHeaderInfoResponse) {
	resp = make([]*BTCHeaderInfoResponse, len(infos))
	for i, info := range infos {
		resp[i] = info.ToResponse()
	}
	return resp
}

// Eq returns true if the hashes are equal.
func (m *BTCHeaderInfoResponse) Eq(other *BTCHeaderInfo) bool {
	return m.HashHex == other.Hash.MarshalHex()
}
