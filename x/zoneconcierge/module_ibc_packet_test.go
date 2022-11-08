package zoneconcierge_test

import (
	"encoding/json"
	"math/rand"

	"github.com/babylonchain/babylon/app"
	zctypes "github.com/babylonchain/babylon/x/zoneconcierge/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
)

// SetupTest creates a coordinator with 2 test chains.
func (suite *ZoneConciergeTestSuite) SetupTestForIBCPackets() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	// replace the first test chain with a Babylon chain
	ibctesting.DefaultTestingAppInit = func() (ibctesting.TestingApp, map[string]json.RawMessage) {
		babylonApp := app.Setup(suite.T(), false)
		suite.zcKeeper = babylonApp.ZoneConciergeKeeper
		encCdc := app.MakeTestEncodingConfig()
		genesis := app.NewDefaultGenesisState(encCdc.Marshaler)
		return babylonApp, genesis
	}
	babylonChainID := ibctesting.GetChainID(1)
	suite.coordinator.Chains[babylonChainID] = ibctesting.NewTestChain(suite.T(), suite.coordinator, babylonChainID)

	suite.babylonChain = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.czChain = suite.coordinator.GetChain(ibctesting.GetChainID(2))

	// commit some blocks so that QueryProof returns valid proof (cannot return valid query if height <= 1)
	suite.coordinator.CommitNBlocks(suite.babylonChain, 2)
	suite.coordinator.CommitNBlocks(suite.czChain, 2)
}

func (suite *ZoneConciergeTestSuite) TestSetChannel() {
	suite.SetupTestForIBCPackets()

	path := ibctesting.NewPath(suite.babylonChain, suite.czChain)

	// set the port ID to be consistent with ZoneConcierge
	path.EndpointA.ChannelConfig.PortID = zctypes.PortID
	path.EndpointB.ChannelConfig.PortID = zctypes.PortID

	// create client and connections on both chains
	suite.coordinator.SetupConnections(path)

	// check for channel to be created on chainA
	_, found := suite.babylonChain.App.GetIBCKeeper().ChannelKeeper.GetChannel(suite.babylonChain.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
	suite.False(found)

	// establish channel
	path.SetChannelOrdered()
	suite.coordinator.CreateChannels(path)

	storedChannel, found := suite.babylonChain.App.GetIBCKeeper().ChannelKeeper.GetChannel(suite.babylonChain.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
	expectedCounterparty := channeltypes.NewCounterparty(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID)

	// assert channel states and properties
	suite.True(found)
	suite.Equal(channeltypes.OPEN, storedChannel.State)
	suite.Equal(channeltypes.ORDERED, storedChannel.Ordering)
	suite.Equal(expectedCounterparty, storedChannel.Counterparty)

	err := path.EndpointA.UpdateClient()
	suite.Require().NoError(err)
	err = path.EndpointB.UpdateClient()
	suite.Require().NoError(err)

	numTests := 10
	for k := 0; k < numTests; k++ { // test suite does not support fuzz tests so we simulate it here
		// retrieve the send sequence number in Babylon
		nextSeqSend, found := suite.babylonChain.App.GetIBCKeeper().ChannelKeeper.GetNextSequenceSend(suite.babylonChain.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		suite.True(found)

		// commit blocks to create some pending heartbeat packets
		numBlocks := rand.Intn(10)
		for i := 0; i < numBlocks; i++ {
			suite.coordinator.CommitBlock(suite.babylonChain)
		}

		// retrieve the send sequence number in Babylon again
		newNextSeqSend, found := suite.babylonChain.App.GetIBCKeeper().ChannelKeeper.GetNextSequenceSend(suite.babylonChain.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
		suite.True(found)

		// Assert the gap between two sequence numbers
		// Note that CommitBlock triggers 2 times of BeginBlock
		suite.Equal(uint64(numBlocks*2), newNextSeqSend-nextSeqSend)

		// update clients to ensure no panic happens
		err = path.EndpointA.UpdateClient()
		suite.Require().NoError(err)
		err = path.EndpointB.UpdateClient()
		suite.Require().NoError(err)
	}

}
