package types

import (
	bbl "github.com/babylonchain/babylon/types"
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
	var hashBytes bbl.BTCHeaderHashBytes
	err := hashBytes.UnmarshalHex(hash)
	if err != nil {
		return nil, err
	}
	res := &QueryContainsRequest{Hash: &hashBytes}
	return res, nil
}
