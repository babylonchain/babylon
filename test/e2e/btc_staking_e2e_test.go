package e2e

import (
	"math/rand"
	"time"

	"github.com/babylonchain/babylon/test/e2e/configurer"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/stretchr/testify/suite"
)

type BTCStakingTestSuite struct {
	suite.Suite

	configurer configurer.Configurer
}

func (s *BTCStakingTestSuite) SetupSuite() {
	s.T().Log("setting up e2e integration test suite...")
	var err error

	// The e2e test flow is as follows:
	//
	// 1. Configure 1 chain with some validator nodes
	// 2. Execute various e2e tests
	s.configurer, err = configurer.NewBTCStakingConfigurer(s.T(), true)
	s.NoError(err)
	err = s.configurer.ConfigureChains()
	s.NoError(err)
	err = s.configurer.RunSetup()
	s.NoError(err)
}

func (s *BTCStakingTestSuite) TearDownSuite() {
	err := s.configurer.ClearResources()
	s.Require().NoError(err)
}

func (s *BTCStakingTestSuite) TestCreateBTCValidator() {
	chainA := s.configurer.GetChainConfig(0)
	chainA.WaitUntilHeight(1)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	// generate a random BTC validator
	r := rand.New(rand.NewSource(time.Now().Unix()))
	btcVal, err := datagen.GenRandomBTCValidator(r)
	s.NoError(err)
	// create a BTC validator
	nonValidatorNode.CreateBTCValidator(btcVal.BabylonPk, btcVal.BtcPk, btcVal.Pop)

	// TODO: query the existence of BTC validator. need RPC query here
}
