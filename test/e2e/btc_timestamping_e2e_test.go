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
	itypes "github.com/babylonchain/babylon/x/incentive/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	// 3. Run IBC relayer between the two chains.
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
func (s *BTCTimestampingTestSuite) Test1ConnectIbc() {
	chainA := s.configurer.GetChainConfig(0)
	chainB := s.configurer.GetChainConfig(1)
	_, err := chainA.GetDefaultNode()
	s.NoError(err)
	_, err = chainB.GetDefaultNode()
	s.NoError(err)
}

func (s *BTCTimestampingTestSuite) Test2BTCBaseHeader() {
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

func (s *BTCTimestampingTestSuite) Test3SendTx() {
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

func (s *BTCTimestampingTestSuite) Test4IbcCheckpointing() {
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

	// submitter/reporter address should not have any rewards yet
	submitterReporterAddr := sdk.MustAccAddressFromBech32(nonValidatorNode.PublicAddress)
	_, err = nonValidatorNode.QueryRewardGauge(submitterReporterAddr)
	s.Error(err)

	nonValidatorNode.FinalizeSealedEpochs(startEpochNum, endEpochNum)

	endEpoch, err := nonValidatorNode.QueryRawCheckpoint(endEpochNum)
	s.NoError(err)
	s.Equal(endEpoch.Status, ct.Finalized)

	// Wait for a some time to ensure that the checkpoint is included in the chain
	time.Sleep(20 * time.Second)
	// Wait for next block
	nonValidatorNode.WaitForNextBlock()

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

	// ensure balance has increased after finalising some epochs
	rewardGauges, err := nonValidatorNode.QueryRewardGauge(submitterReporterAddr)
	s.NoError(err)
	submitterRewardGauge, ok := rewardGauges[itypes.SubmitterType.String()]
	s.True(ok)
	s.True(submitterRewardGauge.Coins.IsAllPositive())
	reporterRewardGauge, ok := rewardGauges[itypes.ReporterType.String()]
	s.True(ok)
	s.True(reporterRewardGauge.Coins.IsAllPositive())

	chainB := s.configurer.GetChainConfig(1)
	_, err = chainB.GetDefaultNode()
	s.NoError(err)
}

func (s *BTCTimestampingTestSuite) Test5WithdrawReward() {
	chainA := s.configurer.GetChainConfig(0)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	// NOTE: nonValidatorNode.PublicAddress is the address associated with key name `val`
	// and is both the submitter and reporter
	submitterReporterAddr := sdk.MustAccAddressFromBech32(nonValidatorNode.PublicAddress)

	// balance before withdraw
	balance, err := nonValidatorNode.QueryBalances(submitterReporterAddr.String())
	s.NoError(err)
	// submitter/reporter reward gauges before withdraw should not be fully withdrawn
	rgs, err := nonValidatorNode.QueryRewardGauge(submitterReporterAddr)
	s.NoError(err)
	submitterRg, reporterRg := rgs[itypes.SubmitterType.String()], rgs[itypes.ReporterType.String()]
	s.T().Logf("submitter witdhrawable reward: %s, reporter witdhrawable reward: %s before withdrawing", submitterRg.GetWithdrawableCoins().String(), reporterRg.GetWithdrawableCoins().String())
	s.False(submitterRg.IsFullyWithdrawn())
	s.False(reporterRg.IsFullyWithdrawn())

	// withdraw submitter reward
	nonValidatorNode.WithdrawReward(itypes.SubmitterType.String(), initialization.ValidatorWalletName)
	nonValidatorNode.WaitForNextBlock()

	// balance after withdrawing submitter reward
	balance2, err := nonValidatorNode.QueryBalances(submitterReporterAddr.String())
	s.NoError(err)
	s.T().Logf("balance2: %s; balance: %s", balance2.String(), balance.String())
	s.True(balance2.IsAllGT(balance))

	// submitter reward gauge should be fully withdrawn
	rgs2, err := nonValidatorNode.QueryRewardGauge(submitterReporterAddr)
	s.NoError(err)
	submitterRg2 := rgs2[itypes.SubmitterType.String()]
	s.T().Logf("submitter withdrawable reward: %s after withdrawing", submitterRg2.GetWithdrawableCoins().String())
	s.True(rgs2[itypes.SubmitterType.String()].IsFullyWithdrawn())

	// withdraw reporter reward
	nonValidatorNode.WithdrawReward(itypes.ReporterType.String(), initialization.ValidatorWalletName)
	nonValidatorNode.WaitForNextBlock()

	// balance after withdrawing reporter reward
	balance3, err := nonValidatorNode.QueryBalances(submitterReporterAddr.String())
	s.NoError(err)
	s.T().Logf("balance3: %s; balance2: %s", balance3.String(), balance2.String())
	s.True(balance3.IsAllGT(balance2))

	// reporter reward gauge should be fully withdrawn
	rgs3, err := nonValidatorNode.QueryRewardGauge(submitterReporterAddr)
	s.NoError(err)
	reporterRg3 := rgs3[itypes.SubmitterType.String()]
	s.T().Logf("reporter withdrawable reward: %s after withdrawing", reporterRg3.GetWithdrawableCoins().String())
	s.True(rgs3[itypes.ReporterType.String()].IsFullyWithdrawn())
}

func (s *BTCTimestampingTestSuite) Test6Wasm() {
	contractPath := "/bytecode/storage_contract.wasm"
	chainA := s.configurer.GetChainConfig(0)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	// store the wasm code
	latestWasmId := int(nonValidatorNode.QueryLatestWasmCodeID())
	nonValidatorNode.StoreWasmCode(contractPath, initialization.ValidatorWalletName)
	s.Eventually(func() bool {
		newLatestWasmId := int(nonValidatorNode.QueryLatestWasmCodeID())
		if latestWasmId+1 > newLatestWasmId {
			return false
		}
		latestWasmId = newLatestWasmId
		return true
	}, time.Second*20, time.Second)

	// instantiate the wasm contract
	var contracts []string
	nonValidatorNode.InstantiateWasmContract(
		strconv.Itoa(latestWasmId),
		`{}`,
		initialization.ValidatorWalletName,
	)
	s.Eventually(func() bool {
		contracts, err = nonValidatorNode.QueryContractsFromId(latestWasmId)
		return err == nil && len(contracts) == 1
	}, time.Second*10, time.Second)
	contractAddr := contracts[0]

	// execute contract
	data := []byte{1, 2, 3, 4, 5}
	dataHex := hex.EncodeToString(data)
	channelId := "1"

	// This is just accepted and ignored atm
	storeMsg := fmt.Sprintf(`{ "save_data": { "data": "%s" } }`, dataHex)
	nonValidatorNode.WasmExecute(contractAddr, storeMsg, initialization.ValidatorWalletName)
	nonValidatorNode.WaitForNextBlock()
	queryMsg := fmt.Sprintf(`{ "account": { "channel_id": "%s" } }`, channelId)
	queryResult, err := nonValidatorNode.QueryWasmSmartObject(contractAddr, queryMsg)
	require.NoError(s.T(), err)
	accountResponse := queryResult["account"].(string)

	s.Equal("TODO: replace me", accountResponse)
}

func (s *BTCTimestampingTestSuite) Test7InterceptFeeCollector() {
	chainA := s.configurer.GetChainConfig(0)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	// ensure incentive module account has positive balance
	incentiveModuleAddr, err := nonValidatorNode.QueryModuleAddress(itypes.ModuleName)
	s.NoError(err)
	incentiveBalance, err := nonValidatorNode.QueryBalances(incentiveModuleAddr.String())
	s.NoError(err)
	s.NotEmpty(incentiveBalance)
	s.T().Logf("incentive module account's balance: %s", incentiveBalance.String())
	s.True(incentiveBalance.IsAllPositive())

	// ensure BTC staking gauge at the current height is eventually non-empty
	// NOTE: sometimes incentive module's BeginBlock is not triggered yet. If this
	// happens, we might need to wait for some time.
	curHeight, err := nonValidatorNode.QueryCurrentHeight()
	s.NoError(err)
	s.Eventually(func() bool {
		btcStakingGauge, err := nonValidatorNode.QueryBTCStakingGauge(uint64(curHeight))
		if err != nil {
			return false
		}
		s.T().Logf("BTC staking gauge at current height %d: %s", curHeight, btcStakingGauge.String())
		return len(btcStakingGauge.Coins) >= 1 && btcStakingGauge.Coins[0].Amount.IsPositive()
	}, time.Second*10, time.Second)

	// ensure BTC timestamping gauge at the current epoch is non-empty
	curEpoch, err := nonValidatorNode.QueryCurrentEpoch()
	s.NoError(err)
	// at the 1st block of an epoch, the gauge does not exist since incentive's BeginBlock
	// at this block accumulates rewards for BTC timestamping gauge for the previous block
	// need to wait for a block to ensure the gauge is created
	var btcTimestampingGauge *itypes.Gauge
	s.Eventually(func() bool {
		btcTimestampingGauge, err = nonValidatorNode.QueryBTCTimestampingGauge(curEpoch)
		if err != nil {
			return false
		}
		s.T().Logf("BTC timestamping gauge at current epoch %d: %s", curEpoch, btcTimestampingGauge.String())
		return !btcTimestampingGauge.Coins.Empty()
	}, time.Second*10, time.Second)

	// wait for 1 block to see if BTC timestamp gauge has accumulated
	nonValidatorNode.WaitForNextBlock()
	btcTimestampingGauge2, err := nonValidatorNode.QueryBTCTimestampingGauge(curEpoch)
	s.NoError(err)
	s.T().Logf("BTC timestamping gauge after a block at current epoch %d: %s", curEpoch, btcTimestampingGauge2.String())
	s.NotEmpty(btcTimestampingGauge2.Coins)
	s.True(btcTimestampingGauge2.Coins.IsAllGTE(btcTimestampingGauge.Coins))

	// after 1 block, incentive's balance has to be accumulated
	incentiveBalance2, err := nonValidatorNode.QueryBalances(incentiveModuleAddr.String())
	s.NoError(err)
	s.T().Logf("incentive module account's balance after a block: %s", incentiveBalance2.String())
	s.True(incentiveBalance2.IsAllGTE(incentiveBalance))
}
