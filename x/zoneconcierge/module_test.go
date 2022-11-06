package zoneconcierge_test

import (
	"testing"

	client "github.com/cosmos/ibc-go/v5/modules/core/02-client"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v5/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v5/modules/light-clients/07-tendermint/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
	"github.com/stretchr/testify/suite"
)

type ZoneConciergeTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	babylonChain *ibctesting.TestChain
	czChain      *ibctesting.TestChain
}

func (suite *ZoneConciergeTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 3)
	suite.babylonChain = suite.coordinator.GetChain(ibctesting.GetChainID(1)) // TODO: make it Babylon
	suite.czChain = suite.coordinator.GetChain(ibctesting.GetChainID(2))

	// set Tendermint client for Babylon chain
	revision := clienttypes.ParseChainID(suite.babylonChain.GetContext().ChainID())
	tmClient := ibctmtypes.NewClientState(
		suite.babylonChain.ChainID,
		ibctmtypes.DefaultTrustLevel,
		ibctesting.TrustingPeriod,
		ibctesting.UnbondingPeriod,
		ibctesting.MaxClockDrift,
		clienttypes.NewHeight(revision, uint64(suite.babylonChain.GetContext().BlockHeight())),
		commitmenttypes.GetSDKSpecs(),
		ibctesting.UpgradePath,
		false,
		false,
	)
	suite.babylonChain.App.GetIBCKeeper().ClientKeeper.SetClientState(suite.babylonChain.GetContext(), exported.Tendermint, tmClient)
}

func (suite *ZoneConciergeTestSuite) TestHandleLightClientHeader() {
	prevHeight := clienttypes.GetSelfHeight(suite.babylonChain.GetContext())

	tmClient := suite.babylonChain.GetClientState(exported.Tendermint)
	suite.Require().Equal(prevHeight, tmClient.GetLatestHeight())

	for i := 0; i < 10; i++ {
		// increment height
		suite.coordinator.CommitBlock(suite.babylonChain, suite.czChain)

		suite.Require().NotPanics(func() {
			client.BeginBlocker(suite.babylonChain.GetContext(), suite.babylonChain.App.GetIBCKeeper().ClientKeeper)
		}, "BeginBlocker shouldn't panic")

		tmClient = suite.babylonChain.GetClientState(exported.Tendermint)
		suite.Require().Equal(prevHeight.Increment(), tmClient.GetLatestHeight())
		prevHeight = tmClient.GetLatestHeight().(clienttypes.Height)
	}
}

func TestZoneConciergeTestSuite(t *testing.T) {
	suite.Run(t, new(ZoneConciergeTestSuite))
}
