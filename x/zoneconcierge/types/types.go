package types

import "time"

// IsLatestHeader checks if a given header is higher than the latest header in chain info
func (ci *ChainInfo) IsLatestHeader(header *IndexedHeader) bool {
	if ci.LatestHeader != nil && ci.LatestHeader.Height > header.Height {
		return false
	}
	return true
}

type HeaderInfo struct {
	ClientId string
	ChainId  string
	AppHash  []byte
	Height   uint64
	Time     time.Time
}
