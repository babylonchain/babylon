package types

import "fmt"

// NewBTCDelegationStatus takes a human-readable delegation status format and returns our custom enum.
// Options: Active | Pending | Expired
func NewBTCDelegationStatus(status string) (BTCDelegationStatus, error) {
	if status == "Active" {
		return BTCDelegationStatus_ACTIVE, nil
	}
	if status == "Pending" {
		return BTCDelegationStatus_PENDING, nil
	}
	if status == "Expired" {
		return BTCDelegationStatus_EXPIRED, nil
	}
	return BTCDelegationStatus_ACTIVE, fmt.Errorf("invalid delegation status %s", status)
}
