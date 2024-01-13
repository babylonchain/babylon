package types

import (
	"github.com/babylonchain/babylon/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

// NewQueryHashesRequest creates a new instance of QueryHashesRequest.
func NewQueryHashesRequest(req *query.PageRequest) *QueryHashesRequest {
	return &QueryHashesRequest{Pagination: req}
}

// NewQueryContainsRequest creates a new instance of QueryContainsRequest.
func NewQueryContainsRequest(hash string) (*QueryContainsRequest, error) {
	hashBytes, err := types.NewBTCHeaderHashBytesFromHex(hash)
	if err != nil {
		return nil, err
	}
	res := &QueryContainsRequest{Hash: &hashBytes}
	return res, nil
}

func NewQueryHeaderDepthRequest(hash string) (*QueryHeaderDepthRequest, error) {
	_, err := types.NewBTCHeaderHashBytesFromHex(hash)
	if err != nil {
		return nil, err
	}
	res := &QueryHeaderDepthRequest{Hash: hash}
	return res, nil
}

func NewQueryMainChainRequest(req *query.PageRequest) *QueryMainChainRequest {
	return &QueryMainChainRequest{Pagination: req}
}

func NewQueryTipRequest() *QueryTipRequest {
	return &QueryTipRequest{}
}

func NewQueryBaseHeaderRequest() *QueryBaseHeaderRequest {
	return &QueryBaseHeaderRequest{}
}
