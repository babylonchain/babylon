package zoneconcierge_test

import (
	"encoding/json"

	"github.com/babylonchain/babylon/app"
	zctypes "github.com/babylonchain/babylon/x/zoneconcierge/types"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
)

var (
	disabledTimeoutTimestamp = uint64(0)
	disabledTimeoutHeight    = clienttypes.ZeroHeight()
	timeoutHeight            = clienttypes.NewHeight(0, 100)

	// for when the testing package cannot be used
	connIDA = "connA"
	connIDB = "connB"
)

// SetupTest creates a coordinator with 2 test chains.
func (suite *ZoneConciergeTestSuite) SetupTestForTestingIBCPackets() {
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
	// create client and connections on both chains
	path := ibctesting.NewPath(suite.babylonChain, suite.czChain)
	suite.coordinator.SetupConnections(path)

	// set the port ID to be consistent with ZoneConcierge
	path.EndpointA.ChannelConfig.PortID = zctypes.PortID

	// check for channel to be created on chainA
	_, found := suite.babylonChain.App.GetIBCKeeper().ChannelKeeper.GetChannel(suite.babylonChain.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
	suite.False(found)

	path.SetChannelOrdered()

	// init channel
	err := path.EndpointA.ChanOpenInit()
	suite.NoError(err)

	storedChannel, found := suite.babylonChain.App.GetIBCKeeper().ChannelKeeper.GetChannel(suite.babylonChain.GetContext(), path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID)
	// counterparty channel id is empty after open init
	expectedCounterparty := channeltypes.NewCounterparty(path.EndpointB.ChannelConfig.PortID, "")

	suite.True(found)
	suite.Equal(channeltypes.INIT, storedChannel.State)
	suite.Equal(channeltypes.ORDERED, storedChannel.Ordering)
	suite.Equal(expectedCounterparty, storedChannel.Counterparty)
}
