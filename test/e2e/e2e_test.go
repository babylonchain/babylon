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

	chainA.WaitUntilHeight(35)

	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	// Query checkpoint chain info for opposing chain
	chainInfo, err := nonValidatorNode.QueryCheckpointChainsInfo(initialization.ChainBID)
	s.NoError(err)
	s.Equal(chainInfo[0].ChainId, initialization.ChainBID)

	// Finalize epoch 1,2,3 , as first headers of opposing chain are in epoch 3
	nonValidatorNode.FinalizeSealedEpochs(1, 3)

	epoch3, err := nonValidatorNode.QueryCheckpointForEpoch(3)
	s.NoError(err)

	if epoch3.Status != ct.Finalized {
		s.FailNow("Epoch 2 should be finalized")
	}

	// Check we have finalized epoch info for opposing chain and some basic assertions
	finalizedResp, err := nonValidatorNode.QueryFinalizedChainsInfo(initialization.ChainBID)
	s.NoError(err)

	finalizedInfo := finalizedResp.Data[0].FinalizedChainInfo
	// TODO Add more assertion here. Maybe check proofs ?
	s.Equal(finalizedInfo.ChainId, initialization.ChainBID)
	s.Equal(finalizedInfo.EpochInfo.EpochNumber, uint64(3))

	currEpoch, err := nonValidatorNode.QueryCurrentEpoch()
	s.NoError(err)

	heightAtEndedEpoch, err := nonValidatorNode.QueryLightClientHeightEpochEnd(currEpoch - 1)
	s.NoError(err)

	if heightAtEndedEpoch == 0 {
		// we can only assert, that btc lc height is larger than 0.
		s.FailNow(fmt.Sprintf("Light client height should be  > 0 on epoch %d", currEpoch-1))
	}

	chainB := s.configurer.GetChainConfig(1)
	_, err = chainB.GetDefaultNode()
	s.NoError(err)
}
