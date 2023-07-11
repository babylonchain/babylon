//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// TestBTCTimestampingTestSuite tests BTC timestamping protocol end-to-end
func TestBTCTimestampingTestSuite(t *testing.T) {
	suite.Run(t, new(BTCTimestampingTestSuite))
}

// TestBTCStakingTestSuite tests BTC staking protocol end-to-end
func TestBTCStakingTestSuite(t *testing.T) {
	suite.Run(t, new(BTCStakingTestSuite))
}
