//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// IBCTransferTestSuite tests IBC transfer end-to-end
func TestIBCTranferTestSuite(t *testing.T) {
	suite.Run(t, new(IBCTransferTestSuite))
}

// TestBTCTimestampingTestSuite tests BTC timestamping protocol end-to-end
func TestBTCTimestampingTestSuite(t *testing.T) {
	suite.Run(t, new(BTCTimestampingTestSuite))
}

// TestBTCTimestampingPhase2HermesTestSuite tests BTC timestamping phase 2 protocol end-to-end,
// with the Hermes relayer
// TODO: Uncomment once we have fix broadcasting of timestamps
// func TestBTCTimestampingPhase2HermesTestSuite(t *testing.T) {
// 	suite.Run(t, new(BTCTimestampingPhase2HermesTestSuite))
// }

// TestBTCTimestampingPhase2RlyTestSuite tests BTC timestamping phase 2 protocol end-to-end,
// with the Go relayer
// TODO: Uncomment once we have fix broadcasting of timestamps
// func TestBTCTimestampingPhase2RlyTestSuite(t *testing.T) {
// 	suite.Run(t, new(BTCTimestampingPhase2RlyTestSuite))
// }

// TestBTCStakingTestSuite tests BTC staking protocol end-to-end
func TestBTCStakingTestSuite(t *testing.T) {
	suite.Run(t, new(BTCStakingTestSuite))
}
