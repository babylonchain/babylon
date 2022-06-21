package types

// NewBaseBTCHeader creates a new Params instance
func NewBaseBTCHeader(headerBytes []byte, height uint64) BaseBTCHeader {
	btcHeaderBytes := &BTCHeaderBytes{HeaderBytes: headerBytes}
	return BaseBTCHeader{Header: btcHeaderBytes, Height: height}
}

// DefaultBaseBTCHeader returns a default set of parameters
func DefaultBaseBTCHeader(headerBytes []byte, height uint64) BaseBTCHeader {
	return NewBaseBTCHeader(headerBytes, height)
}

// Validate validates the base BTC header
func (p BaseBTCHeader) Validate() error {
	return nil
}
