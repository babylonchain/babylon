package e2e

import (
	"time"

	"github.com/babylonchain/babylon/test/e2e/configurer"
	"github.com/babylonchain/babylon/test/e2e/initialization"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type IBCTransferTestSuite struct {
	suite.Suite

	configurer configurer.Configurer
}

func (s *IBCTransferTestSuite) SetupSuite() {
	s.T().Log("setting up IBC test suite...")
	var (
		err error
	)

	s.configurer, err = configurer.NewIBCTransferConfigurer(s.T(), true)

	s.Require().NoError(err)

	err = s.configurer.ConfigureChains()
	s.Require().NoError(err)

	err = s.configurer.RunSetup()
	s.Require().NoError(err)
}

func (s *IBCTransferTestSuite) TearDownSuite() {
	err := s.configurer.ClearResources()
	s.Require().NoError(err)
}

func (s *IBCTransferTestSuite) Test1IBCTransfer() {
	babylonChain := s.configurer.GetChainConfig(0)

	babylonNode, err := babylonChain.GetNodeAtIndex(2)
	s.NoError(err)

	sender := initialization.ValidatorWalletName
	babylonNode.SendIBCTransfer(sender, sender, "", sdk.NewInt64Coin("ubbn", 1000))

	time.Sleep(1 * time.Minute)

	// TODO: check the transfer is successful. Right now this is done by manually looking at the log
}
