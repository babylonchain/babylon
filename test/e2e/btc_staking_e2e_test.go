package e2e

import (
	"math"
	"math/rand"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/babylonchain/babylon/crypto/eots"
	"github.com/babylonchain/babylon/test/e2e/configurer"
	"github.com/babylonchain/babylon/test/e2e/initialization"
	"github.com/babylonchain/babylon/test/e2e/util"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	ftypes "github.com/babylonchain/babylon/x/finality/types"
)

var (
	r = rand.New(rand.NewSource(time.Now().Unix()))
	// BTC validator
	valSK, _, _ = datagen.GenRandomBTCKeyPair(r)
	btcVal, _   = datagen.GenRandomBTCValidatorWithBTCSK(r, valSK)
	// BTC delegation
	delBabylonSK, delBabylonPK, _ = datagen.GenRandomSecp256k1KeyPair(r)
	delBTCSK, delBTCPK, _         = datagen.GenRandomBTCKeyPair(r)
	// jury
	jurySK, _ = btcec.PrivKeyFromBytes(
		[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	)
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

// TestCreateBTCValidatorAndDelegation is an end-to-end test for
// user story 1: user creates BTC validator and BTC delegation
func (s *BTCStakingTestSuite) Test1CreateBTCValidatorAndDelegation() {
	chainA := s.configurer.GetChainConfig(0)
	chainA.WaitUntilHeight(1)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	/*
		create a random BTC validator on Babylon
	*/
	nonValidatorNode.CreateBTCValidator(btcVal.BabylonPk, btcVal.BtcPk, btcVal.Pop)

	// wait for a block so that above txs take effect
	nonValidatorNode.WaitForNextBlock()

	// query the existence of BTC validator and assert equivalence
	actualBtcVals := nonValidatorNode.QueryBTCValidators()
	s.Len(actualBtcVals, 1)
	s.Equal(util.Cdc.MustMarshal(btcVal), util.Cdc.MustMarshal(actualBtcVals[0]))

	/*
		create a random BTC delegation under this BTC validator
	*/
	// BTC staking params, BTC delegation key pairs and PoP
	params := nonValidatorNode.QueryBTCStakingParams()
	pop, err := bstypes.NewPoP(delBabylonSK, delBTCSK)
	s.NoError(err)
	// generate staking tx and slashing tx
	stakingTimeBlocks := uint16(math.MaxUint16)
	stakingValue := int64(2 * 10e8)
	stakingTx, slashingTx, err := datagen.GenBTCStakingSlashingTx(
		r,
		delBTCSK,
		btcVal.BtcPk.MustToBTCPK(),
		params.JuryPk.MustToBTCPK(),
		stakingTimeBlocks,
		stakingValue,
		params.SlashingAddress,
	)
	s.NoError(err)
	stakingMsgTx, err := stakingTx.ToMsgTx()
	s.NoError(err)
	// generate proper delegator sig
	delegatorSig, err := slashingTx.Sign(
		stakingMsgTx,
		stakingTx.StakingScript,
		delBTCSK,
		&chaincfg.SimNetParams,
	)
	s.NoError(err)

	// submit staking tx to Bitcoin and get inclusion proof
	currentBtcTip, err := nonValidatorNode.QueryTip()
	s.NoError(err)
	blockWithStakingTx := datagen.CreateBlockWithTransaction(r, currentBtcTip.Header.ToBlockHeader(), stakingMsgTx)
	nonValidatorNode.InsertHeader(&blockWithStakingTx.HeaderBytes)
	// make block k-deep
	for i := 0; i < initialization.BabylonBtcConfirmationPeriod; i++ {
		nonValidatorNode.InsertNewEmptyBtcHeader(r)
	}
	stakingTxInfo := btcctypes.NewTransactionInfoFromSpvProof(blockWithStakingTx.SpvProof)

	// submit the message for creating BTC delegation
	nonValidatorNode.CreateBTCDelegation(delBabylonPK.(*secp256k1.PubKey), pop, stakingTx, stakingTxInfo, slashingTx, delegatorSig)

	// wait for a block so that above txs take effect
	nonValidatorNode.WaitForNextBlock()

	pendingDelSet := nonValidatorNode.QueryBTCValidatorDelegations(btcVal.BtcPk.MarshalHex())
	s.Len(pendingDelSet, 1)
	pendingDels := pendingDelSet[0]
	s.Len(pendingDels.Dels, 1)
	s.Equal(delBTCPK.SerializeCompressed()[1:], pendingDels.Dels[0].BtcPk.MustToBTCPK().SerializeCompressed()[1:])
	s.Nil(pendingDels.Dels[0].JurySig)
}

// Test2SubmitJurySignature is an end-to-end test for user
// story 2: jury approves the BTC delegation
func (s *BTCStakingTestSuite) Test2SubmitJurySignature() {
	chainA := s.configurer.GetChainConfig(0)
	chainA.WaitUntilHeight(1)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	// get last BTC delegation
	pendingDelsSet := nonValidatorNode.QueryBTCValidatorDelegations(btcVal.BtcPk.MarshalHex())
	s.Len(pendingDelsSet, 1)
	pendingDels := pendingDelsSet[0]
	s.Len(pendingDels.Dels, 1)
	pendingDel := pendingDels.Dels[0]
	s.Nil(pendingDel.JurySig)

	slashingTx := pendingDel.SlashingTx
	stakingTx := pendingDel.StakingTx
	stakingMsgTx, err := stakingTx.ToMsgTx()
	s.NoError(err)
	stakingTxHash := stakingTx.MustGetTxHash()

	/*
		generate and insert new jury signature, in order to activate the BTC delegation
	*/
	jurySig, err := slashingTx.Sign(
		stakingMsgTx,
		stakingTx.StakingScript,
		jurySK,
		&chaincfg.SimNetParams,
	)
	s.NoError(err)
	nonValidatorNode.AddJurySig(btcVal.BtcPk, bbn.NewBIP340PubKeyFromBTCPK(delBTCPK), stakingTxHash, jurySig)

	// wait for a block so that above txs take effect
	nonValidatorNode.WaitForNextBlock()

	// ensure the BTC delegation has jury sig now
	activeDelsSet := nonValidatorNode.QueryBTCValidatorDelegations(btcVal.BtcPk.MarshalHex())
	s.Len(activeDelsSet, 1)
	activeDels := activeDelsSet[0]
	s.Len(activeDels.Dels, 1)
	activeDel := activeDels.Dels[0]
	s.NotNil(activeDel.JurySig)

	// ensure BTC staking is activated
	activatedHeight := nonValidatorNode.QueryActivatedHeight()
	s.Positive(activatedHeight)
	// ensure BTC validator has voting power at activated height
	currentBtcTip, err := nonValidatorNode.QueryTip()
	s.NoError(err)
	activeBTCVals := nonValidatorNode.QueryActiveBTCValidatorsAtHeight(activatedHeight)
	s.Len(activeBTCVals, 1)
	s.Equal(activeBTCVals[0].VotingPower, activeDels.VotingPower(currentBtcTip.Height, initialization.BabylonBtcFinalizationPeriod))
	s.Equal(activeBTCVals[0].VotingPower, activeDel.VotingPower(currentBtcTip.Height, initialization.BabylonBtcFinalizationPeriod))
}

// Test2CommitPublicRandomnessAndSubmitFinalitySignature is an end-to-end
// test for user story 3: BTC validator commits public randomness and submits
// finality signature, such that blocks can be finalised.
func (s *BTCStakingTestSuite) Test3CommitPublicRandomnessAndSubmitFinalitySignature() {
	chainA := s.configurer.GetChainConfig(0)
	chainA.WaitUntilHeight(1)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	// get activated height
	activatedHeight := nonValidatorNode.QueryActivatedHeight()
	s.Positive(activatedHeight)

	/*
		commit a number of public randomness since activatedHeight
	*/
	// commit public randomness list
	srList, msgCommitPubRandList, err := datagen.GenRandomMsgCommitPubRandList(r, valSK, activatedHeight, 100)
	s.NoError(err)
	nonValidatorNode.CommitPubRandList(
		msgCommitPubRandList.ValBtcPk,
		msgCommitPubRandList.StartHeight,
		msgCommitPubRandList.PubRandList,
		msgCommitPubRandList.Sig,
	)

	// ensure public randomness list is eventually committed
	nonValidatorNode.WaitForNextBlock()
	var pubRandMap map[uint64]*bbn.SchnorrPubRand
	s.Eventually(func() bool {
		pubRandMap = nonValidatorNode.QueryListPublicRandomness(btcVal.BtcPk)
		return len(pubRandMap) > 0
	}, time.Minute, time.Second*5)
	s.Equal(pubRandMap[activatedHeight].MustMarshal(), msgCommitPubRandList.PubRandList[0].MustMarshal())

	/*
		submit finality signature
	*/
	// get block to vote
	blockToVote, err := nonValidatorNode.QueryBlock(int64(activatedHeight))
	s.NoError(err)
	msgToSign := append(sdk.Uint64ToBigEndian(activatedHeight), blockToVote.LastCommitHash...)
	// generate EOTS signature
	sig, err := eots.Sign(valSK, srList[0], msgToSign)
	s.NoError(err)
	eotsSig := bbn.NewSchnorrEOTSSigFromModNScalar(sig)
	// submit finality signature
	nonValidatorNode.AddFinalitySig(btcVal.BtcPk, activatedHeight, blockToVote.LastCommitHash, eotsSig)

	// ensure vote is eventually cast
	nonValidatorNode.WaitForNextBlock()
	var votes []bbn.BIP340PubKey
	s.Eventually(func() bool {
		votes = nonValidatorNode.QueryVotesAtHeight(activatedHeight)
		return len(votes) > 0
	}, time.Minute, time.Second*5)
	s.Equal(votes[0].MarshalHex(), btcVal.BtcPk.MarshalHex())
	// once the vote is cast, ensure block is finalised
	finalizedBlock := nonValidatorNode.QueryIndexedBlock(activatedHeight)
	s.NotEmpty(finalizedBlock)
	s.Equal(blockToVote.LastCommitHash.Bytes(), finalizedBlock.LastCommitHash)
	finalizedBlocks := nonValidatorNode.QueryListBlocks(ftypes.QueriedBlockStatus_FINALIZED)
	s.NotEmpty(finalizedBlocks)
	s.Equal(blockToVote.LastCommitHash.Bytes(), finalizedBlocks[0].LastCommitHash)
}
