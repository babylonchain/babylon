package e2e

import (
	"encoding/json"
	"fmt"
	"github.com/babylonchain/babylon/test/e2e/configurer"
	"github.com/babylonchain/babylon/test/e2e/initialization"
	ct "github.com/babylonchain/babylon/x/checkpointing/types"
	"github.com/stretchr/testify/suite"
	"time"
)

type ChannelsResponse struct {
	Channels []struct {
		State        string `json:"state"`
		Ordering     string `json:"ordering"`
		Counterparty struct {
			PortID    string `json:"port_id"`
			ChannelID string `json:"channel_id"`
		} `json:"counterparty"`
		ConnectionsHops []string `json:"connection_hops"`
		Version         string   `json:"version"`
		PortID          string   `json:"port_id"`
		ChannelID       string   `json:"channel_id"`
	}
}

type NextSequenceResponse struct {
	NextSequenceRecv string   `json:"next_sequence_receive"`
	Proof            []string `json:"proof"`
	ProofHeight      struct {
		RevisionNumber string `json:"revision_number"`
		RevisionHeight string `json:"revision_height"`
	} `json:"proof_height"`
}

type BTCTimestampingPhase2TestSuite struct {
	suite.Suite

	configurer configurer.Configurer
}

func (s *BTCTimestampingPhase2TestSuite) SetupSuite() {
	s.T().Log("setting up phase 2 integration test suite...")
	var (
		err error
	)

	// The e2e test flow is as follows:
	//
	// 1. Configure two chains - chain A and chain B.
	//   * For each chain, set up several validator nodes
	//   * Initialize configs and genesis for all them.
	// 2. Start both networks.
	// 3. Store and instantiate babylon contract on chain B.
	// 3. Execute various e2e tests, excluding IBC
	s.configurer, err = configurer.NewBTCTimestampingPhase2Configurer(s.T(), true)

	s.Require().NoError(err)

	err = s.configurer.ConfigureChains()
	s.Require().NoError(err)

	err = s.configurer.RunSetup()
	s.Require().NoError(err)
}

func (s *BTCTimestampingPhase2TestSuite) TearDownSuite() {
	err := s.configurer.ClearResources()
	s.Require().NoError(err)
}

func (s *BTCTimestampingPhase2TestSuite) Test1IbcCheckpointingPhase2() {
	chainA := s.configurer.GetChainConfig(0)
	chainA.WaitUntilHeight(35)

	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	// Query checkpoint chain info for opposing chain
	chainsInfo, err := nonValidatorNode.QueryChainsInfo([]string{initialization.ChainBID})
	s.NoError(err)
	s.Equal(chainsInfo[0].ChainId, initialization.ChainBID)

	// Finalize epoch 1, 2, 3, as first headers of opposing chain are in epoch 3
	var (
		startEpochNum uint64 = 1
		endEpochNum   uint64 = 3
	)

	nonValidatorNode.FinalizeSealedEpochs(startEpochNum, endEpochNum)

	endEpoch, err := nonValidatorNode.QueryRawCheckpoint(endEpochNum)
	s.NoError(err)
	s.Equal(endEpoch.Status, ct.Finalized)

	// Check we have epoch info for opposing chain and some basic assertions
	epochChainsInfo, err := nonValidatorNode.QueryEpochChainsInfo(endEpochNum, []string{initialization.ChainBID})
	s.NoError(err)
	s.Equal(epochChainsInfo[0].ChainId, initialization.ChainBID)
	s.Equal(epochChainsInfo[0].LatestHeader.BabylonEpoch, endEpochNum)

	// Check we have finalized epoch info for opposing chain and some basic assertions
	finalizedChainsInfo, err := nonValidatorNode.QueryFinalizedChainsInfo([]string{initialization.ChainBID})
	s.NoError(err)

	// TODO Add more assertion here. Maybe check proofs ?
	s.Equal(finalizedChainsInfo[0].FinalizedChainInfo.ChainId, initialization.ChainBID)
	s.Equal(finalizedChainsInfo[0].EpochInfo.EpochNumber, endEpochNum)

	currEpoch, err := nonValidatorNode.QueryCurrentEpoch()
	s.NoError(err)

	heightAtEndedEpoch, err := nonValidatorNode.QueryLightClientHeightEpochEnd(currEpoch - 1)
	s.NoError(err)

	if heightAtEndedEpoch == 0 {
		// we can only assert, that btc lc height is larger than 0.
		s.FailNow(fmt.Sprintf("Light client height should be  > 0 on epoch %d", currEpoch-1))
	}

	chainB := s.configurer.GetChainConfig(1)
	validatorNode, err := chainB.GetDefaultNode()
	s.NoError(err)

	bz, err := validatorNode.QueryGRPCGateway("/ibc/core/channel/v1/channels", nil)
	s.NoError(err)
	s.T().Logf("channels: %s", bz)
	var channelsResponse ChannelsResponse
	err = json.Unmarshal(bz, &channelsResponse)
	s.NoError(err)

	// Validate channel state and kind
	s.Equal("STATE_OPEN", channelsResponse.Channels[1].State)
	s.Equal("ORDER_ORDERED", channelsResponse.Channels[1].Ordering)
	s.Contains("wasm", channelsResponse.Channels[1].PortID)
	// Define channel and port ids
	channelID := channelsResponse.Channels[1].ChannelID
	portID := channelsResponse.Channels[1].PortID

	s.T().Log("Sleeping for 10 seconds to allow for IBC packets to be sent")
	time.Sleep(10 * time.Second)

	// Query next sequence id
	bz, err = validatorNode.QueryGRPCGateway(fmt.Sprintf("/ibc/core/channel/v1/channels/%s/ports/%s/next_sequence", channelID, portID), nil)
	s.NoError(err)
	var nextSequenceResponse NextSequenceResponse
	err = json.Unmarshal(bz, &nextSequenceResponse)
	s.NoError(err)
	// Check that three IBC packets have been received
	s.Equal("4", nextSequenceResponse.NextSequenceRecv)
}
