package types

import "github.com/cosmos/cosmos-sdk/types/query"

// NewQueryRawCheckpointRequest creates a new instance of QueryRawCheckpointRequest.
func NewQueryRawCheckpointRequest(epochNum uint64) *QueryRawCheckpointRequest {
	return &QueryRawCheckpointRequest{EpochNum: epochNum}
}

// NewQueryRawCheckpointListRequest creates a new instance of QueryRawCheckpointListRequest.
func NewQueryRawCheckpointListRequest(req *query.PageRequest, status CheckpointStatus) *QueryRawCheckpointListRequest {
	return &QueryRawCheckpointListRequest{
		Status:     status,
		Pagination: req,
	}
}

func NewQueryEpochStatusRequest(epochNum uint64) *QueryEpochStatusRequest {
	return &QueryEpochStatusRequest{EpochNum: epochNum}
}

func NewQueryRecentEpochStatusCountRequest(epochNum uint64) *QueryRecentEpochStatusCountRequest {
	return &QueryRecentEpochStatusCountRequest{EpochCount: epochNum}
}

func NewQueryLastCheckpointWithStatus(status CheckpointStatus) *QueryLastCheckpointWithStatusRequest {
	return &QueryLastCheckpointWithStatusRequest{Status: status}
}
