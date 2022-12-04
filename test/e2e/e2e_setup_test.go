package e2e

import (
	"testing"

	"github.com/stretchr/testify/suite"

	configurer "github.com/babylonchain/babylon/test/e2e/configurer"
)

type IntegrationTestSuite struct {
	suite.Suite

	configurer configurer.Configurer
}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up e2e integration test suite...")
	var (
		err error
	)

	// The e2e test flow is as follows:
	//
	// 1. Configure two chains - chan A and chain B.
	//   * For each chain, set up several validator nodes
	//   * Initialize configs and genesis for all them.
	// 2. Start both networks.
	// 3. Run IBC relayer betweeen the two chains.
	// 4. Execute various e2e tests, including IBC
	s.configurer, err = configurer.New(s.T(), true)

	s.Require().NoError(err)

	err = s.configurer.ConfigureChains()
	s.Require().NoError(err)

	err = s.configurer.RunSetup()
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	err := s.configurer.ClearResources()
	s.Require().NoError(err)
}
