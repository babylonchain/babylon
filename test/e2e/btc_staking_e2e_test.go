package e2e

import (
	"encoding/hex"
	"math"
	"math/rand"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/babylonchain/babylon/btcstaking"
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
	r   = rand.New(rand.NewSource(time.Now().Unix()))
	net = &chaincfg.SimNetParams
	// BTC validator
	valSK, _, _       = datagen.GenRandomBTCKeyPair(r)
	btcVal, _         = datagen.GenRandomBTCValidatorWithBTCSK(r, valSK)
	btcValBabylonAddr = sdk.AccAddress(btcVal.BabylonPk.Address()).String()
	// BTC delegation
	delBabylonSK, delBabylonPK, _ = datagen.GenRandomSecp256k1KeyPair(r)
	delBabylonAddr                = sdk.AccAddress(delBabylonPK.Address()).String()
	delBTCSK, delBTCPK, _         = datagen.GenRandomBTCKeyPair(r)
	// jury
	jurySK, _ = btcec.PrivKeyFromBytes(
		[]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	)

	stakingValue = int64(2 * 10e8)
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
	nonValidatorNode.CreateBTCValidator(btcVal.BabylonPk, btcVal.BtcPk, btcVal.Pop, btcVal.Description.Moniker, btcVal.Description.Identity, btcVal.Description.Website, btcVal.Description.SecurityContact, btcVal.Description.Details, btcVal.Commission)

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
	stakingTx, slashingTx, err := datagen.GenBTCStakingSlashingTx(
		r,
		net,
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
		stakingTx.Script,
		delBTCSK,
		net,
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

	// check delegation
	delegation := nonValidatorNode.QueryBtcDelegation(stakingTx.MustGetTxHash())
	s.NotNil(delegation)
	expectedScript := hex.EncodeToString(stakingTx.Script)
	s.Equal(expectedScript, delegation.StakingScript)
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
		stakingTx.Script,
		jurySK,
		net,
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

	// get the balance of BTC validator and delegation before submitting finality signature
	// later we will check if rewards are distributed after they finalise a block
	btcValBalance, err := nonValidatorNode.QueryBalances(btcValBabylonAddr)
	s.NoError(err)
	btcDelBalance, err := nonValidatorNode.QueryBalances(delBabylonAddr)
	s.NoError(err)

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

	// ensure voters have received rewards after the block is finalised
	btcValBalance2, err := nonValidatorNode.QueryBalances(btcValBabylonAddr)
	s.NoError(err)
	s.True(btcValBalance2.IsAllGT(btcValBalance))
	btcDelBalance2, err := nonValidatorNode.QueryBalances(delBabylonAddr)
	s.NoError(err)
	s.True(btcDelBalance2.IsAllGT(btcDelBalance))
}

// Test4SubmitStakerUnbonding is an end-to-end test for user unbodning
func (s *BTCStakingTestSuite) Test4SubmitStakerUnbonding() {
	chainA := s.configurer.GetChainConfig(0)
	chainA.WaitUntilHeight(1)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)
	// wait for a block so that above txs take effect
	nonValidatorNode.WaitForNextBlock()

	activeDelsSet := nonValidatorNode.QueryBTCValidatorDelegations(btcVal.BtcPk.MarshalHex())
	s.Len(activeDelsSet, 1)
	activeDels := activeDelsSet[0]
	s.Len(activeDels.Dels, 1)
	activeDel := activeDels.Dels[0]
	s.NotNil(activeDel.JurySig)

	// params for juryPk and slashing address
	params := nonValidatorNode.QueryBTCStakingParams()

	stakingTx := activeDel.StakingTx
	stakingMsgTx, err := stakingTx.ToMsgTx()
	s.NoError(err)
	stakingTxHash := stakingTx.MustGetTxHash()
	stakingTxChainHash, err := chainhash.NewHashFromStr(stakingTxHash)
	s.NoError(err)

	stakingOutputIdx, err := btcstaking.GetIdxOutputCommitingToScript(
		stakingMsgTx, activeDel.StakingTx.Script, net,
	)
	s.NoError(err)

	fee := int64(1000)
	unbondingTx, slashUnbondingTx, err := datagen.GenBTCUnbondingSlashingTx(
		r,
		net,
		delBTCSK,
		btcVal.BtcPk.MustToBTCPK(),
		params.JuryPk.MustToBTCPK(),
		wire.NewOutPoint(stakingTxChainHash, uint32(stakingOutputIdx)),
		initialization.BabylonBtcFinalizationPeriod+1,
		stakingValue-fee,
		params.SlashingAddress,
	)
	s.NoError(err)

	unbondingTxMsg, err := unbondingTx.ToMsgTx()
	s.NoError(err)

	slashingTxSig, err := slashUnbondingTx.Sign(
		unbondingTxMsg,
		unbondingTx.Script,
		delBTCSK,
		net,
	)
	s.NoError(err)

	// submit the message for creating BTC undelegation
	nonValidatorNode.CreateBTCUndelegation(unbondingTx, slashUnbondingTx, slashingTxSig)
	// wait for a block so that above txs take effect
	nonValidatorNode.WaitForNextBlock()

	valDelegations := nonValidatorNode.QueryBTCValidatorDelegations(btcVal.BtcPk.MarshalHex())
	s.Len(valDelegations, 1)
	s.Len(valDelegations[0].Dels, 1)
	delegation := valDelegations[0].Dels[0]
	s.NotNil(delegation.BtcUndelegation)
}

// Test5SubmitStakerUnbonding is an end-to-end test for jury and validator submitting signatures
// for unbonding transaction
func (s *BTCStakingTestSuite) Test5SubmitUnbondingSignatures() {
	chainA := s.configurer.GetChainConfig(0)
	chainA.WaitUntilHeight(1)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)
	// wait for a block so that above txs take effect
	nonValidatorNode.WaitForNextBlock()

	allDelegations := nonValidatorNode.QueryBTCValidatorDelegations(btcVal.BtcPk.MarshalHex())
	s.Len(allDelegations, 1)
	delegatorDelegations := allDelegations[0]
	s.Len(delegatorDelegations.Dels, 1)
	delegation := delegatorDelegations.Dels[0]

	s.NotNil(delegation.BtcUndelegation)
	s.Nil(delegation.BtcUndelegation.ValidatorUnbondingSig)
	s.Nil(delegation.BtcUndelegation.JuryUnbondingSig)
	s.Nil(delegation.BtcUndelegation.JurySlashingSig)

	// First sent validator signature
	stakingTxMsg, err := delegation.StakingTx.ToMsgTx()
	s.NoError(err)
	stakingTxHash := delegation.StakingTx.MustGetTxHash()

	validatorUnbondingSig, err := delegation.BtcUndelegation.UnbondingTx.Sign(
		stakingTxMsg,
		delegation.StakingTx.Script,
		valSK,
		net,
	)
	s.NoError(err)

	nonValidatorNode.AddValidatorUnbondingSig(btcVal.BtcPk, bbn.NewBIP340PubKeyFromBTCPK(delBTCPK), stakingTxHash, validatorUnbondingSig)
	nonValidatorNode.WaitForNextBlock()

	allDelegationsValSig := nonValidatorNode.QueryBTCValidatorDelegations(btcVal.BtcPk.MarshalHex())
	s.Len(allDelegationsValSig, 1)
	delegationWithValSig := allDelegationsValSig[0].Dels[0]
	s.NotNil(delegationWithValSig.BtcUndelegation)
	s.NotNil(delegationWithValSig.BtcUndelegation.ValidatorUnbondingSig)

	unbodnindDelegations := nonValidatorNode.QueryUnbondingDelegations()
	s.Len(unbodnindDelegations, 1)

	btcTip, err := nonValidatorNode.QueryTip()
	s.NoError(err)
	s.Equal(
		bstypes.BTCDelegationStatus_UNBONDING,
		delegationWithValSig.GetStatus(btcTip.Height, initialization.BabylonBtcFinalizationPeriod),
	)

	// Next send jury signatures
	juryUnbondingSig, err := delegation.BtcUndelegation.UnbondingTx.Sign(
		stakingTxMsg,
		delegation.StakingTx.Script,
		jurySK,
		net,
	)
	s.NoError(err)

	unbondingTxMsg, err := delegation.BtcUndelegation.UnbondingTx.ToMsgTx()
	s.NoError(err)
	jurySlashingSig, err := delegation.BtcUndelegation.SlashingTx.Sign(
		unbondingTxMsg,
		delegation.BtcUndelegation.UnbondingTx.Script,
		jurySK,
		net,
	)
	s.NoError(err)
	nonValidatorNode.AddJuryUnbondingSigs(btcVal.BtcPk, bbn.NewBIP340PubKeyFromBTCPK(delBTCPK), stakingTxHash, juryUnbondingSig, jurySlashingSig)
	nonValidatorNode.WaitForNextBlock()

	// Check all signatures are properly registered
	allDelegationsWithSigs := nonValidatorNode.QueryBTCValidatorDelegations(btcVal.BtcPk.MarshalHex())
	s.Len(allDelegationsWithSigs, 1)
	delegationWithSigs := allDelegationsWithSigs[0].Dels[0]
	s.NotNil(delegationWithSigs.BtcUndelegation)
	s.NotNil(delegationWithSigs.BtcUndelegation.ValidatorUnbondingSig)
	s.NotNil(delegationWithSigs.BtcUndelegation.JuryUnbondingSig)
	s.NotNil(delegationWithSigs.BtcUndelegation.JurySlashingSig)
	btcTip, err = nonValidatorNode.QueryTip()
	s.NoError(err)
	s.Equal(
		bstypes.BTCDelegationStatus_UNBONDED,
		delegationWithSigs.GetStatus(btcTip.Height, initialization.BabylonBtcFinalizationPeriod),
	)
}
