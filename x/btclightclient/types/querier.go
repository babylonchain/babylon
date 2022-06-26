package types

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
)

// NewQueryParamsRequest creates a new instance of QueryParamsRequest.
func NewQueryParamsRequest() *QueryParamsRequest {
	return &QueryParamsRequest{}
}

// NewQueryHashesRequest creates a new instance of QueryHashesRequest.
func NewQueryHashesRequest() *QueryHashesRequest {
	return &QueryHashesRequest{}
}

// NewQueryContainsRequest creates a new instance of QueryContainsRequest.
func NewQueryContainsRequest(hash string) (*QueryContainsRequest, error) {
	// Convert the hex hash into the
	// reverse bytes representation that we expect
	chHash, err := chainhash.NewHashFromStr(hash)
	if err != nil {
		return nil, err
	}
	chBytes := ChainhashToBytes(chHash)
	return &QueryContainsRequest{Hash: chBytes}, nil
}
