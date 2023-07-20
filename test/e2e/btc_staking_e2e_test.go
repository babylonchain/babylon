package e2e

import (
	"math"
	"math/rand"
	"time"

	"github.com/babylonchain/babylon/test/e2e/configurer"
	"github.com/babylonchain/babylon/test/e2e/initialization"
	"github.com/babylonchain/babylon/test/e2e/util"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
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

// TestCreateBTCValidatorAndDelegation is an end-to-end test for
// user story 1/2: user creates BTC validator and BTC delegation, then
// jury approves the BTC delegation
func (s *BTCStakingTestSuite) TestCreateBTCValidatorAndDelegation() {
	chainA := s.configurer.GetChainConfig(0)
	chainA.WaitUntilHeight(1)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	/*
		create a random BTC validator on Babylon
	*/
	// generate a random BTC validator
	r := rand.New(rand.NewSource(time.Now().Unix()))
	btcVal, err := datagen.GenRandomBTCValidator(r)
	s.NoError(err)
	// create this BTC validator
	nonValidatorNode.CreateBTCValidator(btcVal.BabylonPk, btcVal.BtcPk, btcVal.Pop)

	// wait for a block so that above txs take effect
	nonValidatorNode.WaitForNextBlock()

	// query the existence of BTC validator and assert equivalence
	actualBtcVals := nonValidatorNode.QueryBTCValidators()
	s.Len(actualBtcVals, 1)
	s.Equal(util.Cdc.MustMarshal(btcVal), util.Cdc.MustMarshal(actualBtcVals[0]))

	/*
		create a random BTC delegation under this bTC validator
	*/
	// BTC staking params, BTC delegation key pairs and PoP
	params := nonValidatorNode.QueryBTCStakingParams()
	delBabylonSK, delBabylonPK, err := datagen.GenRandomSecp256k1KeyPair(r)
	s.NoError(err)
	delBTCSK, delBTCPK, err := datagen.GenRandomBTCKeyPair(r)
	s.NoError(err)
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

	pendingDels := nonValidatorNode.QueryBTCValidatorDelegations(btcVal.BtcPk.ToHexStr(), bstypes.BTCDelegationStatus_PENDING)
	s.Len(pendingDels, 1)
	s.Equal(delBTCPK.SerializeCompressed()[1:], pendingDels[0].BtcPk.MustToBTCPK().SerializeCompressed()[1:])

	/*
		generate and insert new jury signature, in order to activate the BTC delegation
	*/
	jurySKBytes := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	jurySK, _ := btcec.PrivKeyFromBytes(jurySKBytes)
	jurySig, err := slashingTx.Sign(
		stakingMsgTx,
		stakingTx.StakingScript,
		jurySK,
		&chaincfg.SimNetParams,
	)
	s.NoError(err)
	nonValidatorNode.AddJurySig(btcVal.BtcPk, bbn.NewBIP340PubKeyFromBTCPK(delBTCPK), jurySig)

	// wait for a block so that above txs take effect
	nonValidatorNode.WaitForNextBlock()

	// query the existence of BTC delegation and assert equivalence
	actualDels := nonValidatorNode.QueryBTCValidatorDelegations(btcVal.BtcPk.ToHexStr(), bstypes.BTCDelegationStatus_ACTIVE)
	s.Len(actualDels, 1)
	s.Equal(delBTCPK.SerializeCompressed()[1:], actualDels[0].BtcPk.MustToBTCPK().SerializeCompressed()[1:])
}
