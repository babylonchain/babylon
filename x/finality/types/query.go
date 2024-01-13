package types

import (
	"fmt"
)

// NewQueriedBlockStatus takes a human-readable queried block status format and returns our custom enum.
// Options: NonFinalized | Finalized | Any
func NewQueriedBlockStatus(status string) (QueriedBlockStatus, error) {
	if status == "NonFinalized" {
		return QueriedBlockStatus_NON_FINALIZED, nil
	}
	if status == "Finalized" {
		return QueriedBlockStatus_FINALIZED, nil
	}
	if status == "Any" {
		return QueriedBlockStatus_ANY, nil
	}
	return QueriedBlockStatus_NON_FINALIZED, fmt.Errorf("invalid queried block status %s", status)
}
