//go:build e2e
// +build e2e

package e2e

import (
	"fmt"

	"github.com/babylonchain/babylon/test/e2e/initialization"
	ct "github.com/babylonchain/babylon/x/checkpointing/types"
)

// Most simple test, just checking that two chains are up and connected through
// ibc
func (s *IntegrationTestSuite) TestConnectIbc() {
	chainA := s.configurer.GetChainConfig(0)
	chainB := s.configurer.GetChainConfig(1)
	_, err := chainA.GetDefaultNode()
	s.NoError(err)
	_, err = chainB.GetDefaultNode()
	s.NoError(err)
}

func (s *IntegrationTestSuite) TestIbcCheckpointing() {
	chainA := s.configurer.GetChainConfig(0)

	chainA.WaitUntilHeight(25)

	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	// Finalize epoch 1 and 2, as first headers of opposing chain are in epoch 2
	nonValidatorNode.FinalizeSealedEpochs(1, 2)

	epoch2, err := nonValidatorNode.QueryCheckpointForEpoch(2)
	s.NoError(err)

	if epoch2.Status != ct.Finalized {
		s.FailNow("Epoch 2 should be finalized")
	}

	// Check we have finalized epoch info for opposing chain and some basic assertions
	fininfo, err := nonValidatorNode.QueryFinalizedChainInfo(initialization.ChainBID)
	s.NoError(err)
	// TODO Add more assertion here. Maybe check proofs ?
	s.Equal(fininfo.FinalizedChainInfo.ChainId, initialization.ChainBID)
	s.Equal(fininfo.EpochInfo.EpochNumber, uint64(2))

	currEpoch, err := nonValidatorNode.QueryCurrentEpoch()
	s.NoError(err)

	heightAtFinishedEpoch, err := nonValidatorNode.QueryLightClientHeightEpochEnd(currEpoch - 1)
	s.NoError(err)

	if heightAtFinishedEpoch == 0 {
		// we can only assert, that btc lc height is larger than 0.
		s.FailNow(fmt.Sprintf("Light client height should be  > 0 on epoch %d", currEpoch-1))
	}

	chainB := s.configurer.GetChainConfig(1)
	_, err = chainB.GetDefaultNode()
	s.NoError(err)
}
