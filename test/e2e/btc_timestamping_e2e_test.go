package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/babylonchain/babylon/test/e2e/configurer"
	"github.com/babylonchain/babylon/test/e2e/initialization"
	bbn "github.com/babylonchain/babylon/types"
	ct "github.com/babylonchain/babylon/x/checkpointing/types"
	incentivetypes "github.com/babylonchain/babylon/x/incentive/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type BTCTimestampingTestSuite struct {
	suite.Suite

	configurer configurer.Configurer
}

func (s *BTCTimestampingTestSuite) SetupSuite() {
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
	s.configurer, err = configurer.NewBTCTimestampingConfigurer(s.T(), true)

	s.Require().NoError(err)

	err = s.configurer.ConfigureChains()
	s.Require().NoError(err)

	err = s.configurer.RunSetup()
	s.Require().NoError(err)
}

func (s *BTCTimestampingTestSuite) TearDownSuite() {
	err := s.configurer.ClearResources()
	s.Require().NoError(err)
}

// Most simple test, just checking that two chains are up and connected through
// ibc
func (s *BTCTimestampingTestSuite) TestConnectIbc() {
	chainA := s.configurer.GetChainConfig(0)
	chainB := s.configurer.GetChainConfig(1)
	_, err := chainA.GetDefaultNode()
	s.NoError(err)
	_, err = chainB.GetDefaultNode()
	s.NoError(err)
}

func (s *BTCTimestampingTestSuite) TestBTCBaseHeader() {
	hardcodedHeader, _ := bbn.NewBTCHeaderBytesFromHex("0100000000000000000000000000000000000000000000000000000000000000000000003ba3edfd7a7b12b27ac72c3e67768f617fc81bc3888a51323a9fb8aa4b1e5e4a45068653ffff7f2002000000")
	hardcodedHeaderHeight := uint64(0)

	chainA := s.configurer.GetChainConfig(0)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)
	baseHeader, err := nonValidatorNode.QueryBtcBaseHeader()
	s.NoError(err)
	s.True(baseHeader.Hash.Eq(hardcodedHeader.Hash()))
	s.Equal(hardcodedHeaderHeight, baseHeader.Height)
}

func (s *BTCTimestampingTestSuite) TestSendTx() {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	chainA := s.configurer.GetChainConfig(0)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	tip1, err := nonValidatorNode.QueryTip()
	s.NoError(err)

	nonValidatorNode.InsertNewEmptyBtcHeader(r)

	tip2, err := nonValidatorNode.QueryTip()
	s.NoError(err)

	s.Equal(tip1.Height+1, tip2.Height)

	// check that light client properly updates its state
	tip1Depth, err := nonValidatorNode.QueryHeaderDepth(tip1.Hash.MarshalHex())
	s.NoError(err)
	s.Equal(tip1Depth, uint64(1))

	tip2Depth, err := nonValidatorNode.QueryHeaderDepth(tip2.Hash.MarshalHex())
	s.NoError(err)
	// tip should have 0 depth
	s.Equal(tip2Depth, uint64(0))
}

func (s *BTCTimestampingTestSuite) TestIbcCheckpointing() {
	chainA := s.configurer.GetChainConfig(0)
	chainA.WaitUntilHeight(35)

	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	// Query checkpoint chain info for opposing chain
	chainsInfo, err := nonValidatorNode.QueryChainsInfo([]string{initialization.ChainBID})
	s.NoError(err)
	s.Equal(chainsInfo[0].ChainId, initialization.ChainBID)

	// Finalize epoch 1,2,3 , as first headers of opposing chain are in epoch 3
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
	_, err = chainB.GetDefaultNode()
	s.NoError(err)
}

func (s *BTCTimestampingTestSuite) TestWasm() {
	contractPath := "/bytecode/storage_contract.wasm"
	chainA := s.configurer.GetChainConfig(0)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	require.NoError(s.T(), err)
	nonValidatorNode.StoreWasmCode(contractPath, initialization.ValidatorWalletName)
	nonValidatorNode.WaitForNextBlock()
	latestWasmId := int(nonValidatorNode.QueryLatestWasmCodeID())
	nonValidatorNode.InstantiateWasmContract(
		strconv.Itoa(latestWasmId),
		`{}`,
		initialization.ValidatorWalletName,
	)
	nonValidatorNode.WaitForNextBlock()
	contracts, err := nonValidatorNode.QueryContractsFromId(1)
	s.NoError(err)
	s.Require().Len(contracts, 1, "Wrong number of contracts for the counter")
	contractAddr := contracts[0]

	data := []byte{1, 2, 3, 4, 5}
	dataHex := hex.EncodeToString(data)
	dataHash := sha256.Sum256(data)
	dataHashHex := hex.EncodeToString(dataHash[:])

	storeMsg := fmt.Sprintf(`{"save_data":{"data":"%s"}}`, dataHex)
	nonValidatorNode.WasmExecute(contractAddr, storeMsg, initialization.ValidatorWalletName)
	nonValidatorNode.WaitForNextBlock()
	queryMsg := fmt.Sprintf(`{"check_data": {"data_hash":"%s"}}`, dataHashHex)
	queryResult, err := nonValidatorNode.QueryWasmSmartObject(contractAddr, queryMsg)
	require.NoError(s.T(), err)
	finalized := queryResult["finalized"].(bool)
	latestFinalizedEpoch := int(queryResult["latest_finalized_epoch"].(float64))
	saveEpoch := int(queryResult["save_epoch"].(float64))

	require.False(s.T(), finalized)
	// in previous test we already finalized epoch 3
	require.Equal(s.T(), 3, latestFinalizedEpoch)
	// data is not finalized yet, so save epoch should be strictly greater than latest finalized epoch
	require.Greater(s.T(), saveEpoch, latestFinalizedEpoch)
}

func (s *BTCTimestampingTestSuite) TestInterceptFeeCollector() {
	chainA := s.configurer.GetChainConfig(0)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	// ensure incentive module account has positive balance
	incentiveModuleAddr, err := nonValidatorNode.QueryModuleAddress(incentivetypes.ModuleName)
	s.NoError(err)
	incentiveBalance, err := nonValidatorNode.QueryBalances(incentiveModuleAddr.String())
	s.NoError(err)
	s.NotEmpty(incentiveBalance)
	s.True(incentiveBalance.IsAllPositive())

	// ensure BTC staking gauge at the current height is non-empty
	curHeight, err := nonValidatorNode.QueryCurrentHeight()
	s.NoError(err)
	btcStakingGauge, err := nonValidatorNode.QueryBTCStakingGauge(uint64(curHeight))
	s.NoError(err)
	s.True(len(btcStakingGauge.Coins) >= 1)
	s.True(btcStakingGauge.Coins[0].Amount.IsPositive())

	// ensure BTC timestamping gauge at the current epoch is non-empty
	curEpoch, err := nonValidatorNode.QueryCurrentEpoch()
	s.NoError(err)
	// at the 1st block of an epoch, the gauge does not exist since incentive's BeginBlock
	// at this block accumulates rewards for BTC timestamping gauge for the previous block
	// need to wait for a block to ensure the gauge is created
	nonValidatorNode.WaitForNextBlock()
	btcTimestampingGauge, err := nonValidatorNode.QueryBTCTimestampingGauge(curEpoch)
	s.NoError(err)
	s.NotEmpty(btcTimestampingGauge.Coins)

	// wait for 1 block to see if BTC timestamp gauge has accumulated
	nonValidatorNode.WaitForNextBlock()
	btcTimestampingGauge2, err := nonValidatorNode.QueryBTCTimestampingGauge(curEpoch)
	s.NoError(err)
	s.NotEmpty(btcTimestampingGauge2.Coins)
	s.True(btcTimestampingGauge2.Coins.IsAllGTE(btcTimestampingGauge.Coins))

	// after 1 block, incentive's balance has to be accumulated
	incentiveBalance2, err := nonValidatorNode.QueryBalances(incentiveModuleAddr.String())
	s.NoError(err)
	s.True(incentiveBalance2.IsAllGTE(incentiveBalance))
}
