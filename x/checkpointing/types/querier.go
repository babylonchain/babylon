package types

import "github.com/cosmos/cosmos-sdk/types/query"

// NewQueryRawCheckpointRequest creates a new instance of QueryRawCheckpointRequest.
func NewQueryRawCheckpointRequest(epoch_num uint64) *QueryRawCheckpointRequest {
	return &QueryRawCheckpointRequest{EpochNum: epoch_num}
}

// NewQueryRawCheckpointListRequest creates a new instance of QueryRawCheckpointListRequest.
func NewQueryRawCheckpointListRequest(req *query.PageRequest, status CheckpointStatus) *QueryRawCheckpointListRequest {
	return &QueryRawCheckpointListRequest{
		Status:     status,
		Pagination: req,
	}
}
