package keeper_test

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/babylonchain/babylon/btcstaking"
	asig "github.com/babylonchain/babylon/crypto/schnorr-adaptor-signature"
	"github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/babylonchain/babylon/x/btcstaking/keeper"
	"github.com/babylonchain/babylon/x/btcstaking/types"
)

type Helper struct {
	t *testing.T

	Ctx                  context.Context
	BTCStakingKeeper     *keeper.Keeper
	BTCLightClientKeeper *types.MockBTCLightClientKeeper
	BTCCheckpointKeeper  *types.MockBtcCheckpointKeeper
	MsgServer            types.MsgServer
	Net                  *chaincfg.Params
}

func NewHelper(t *testing.T, btclcKeeper *types.MockBTCLightClientKeeper, btccKeeper *types.MockBtcCheckpointKeeper) *Helper {
	k, ctx := keepertest.BTCStakingKeeper(t, btclcKeeper, btccKeeper)
	msgSrvr := keeper.NewMsgServerImpl(*k)

	return &Helper{
		t:                    t,
		Ctx:                  ctx,
		BTCStakingKeeper:     k,
		BTCLightClientKeeper: btclcKeeper,
		BTCCheckpointKeeper:  btccKeeper,
		MsgServer:            msgSrvr,
		Net:                  &chaincfg.SimNetParams,
	}
}

func (h *Helper) NoError(err error) {
	require.NoError(h.t, err)
}

func (h *Helper) GenAndApplyParams(r *rand.Rand) ([]*btcec.PrivateKey, []*btcec.PublicKey) {
	// TODO: randomise covenant committee and quorum?
	covenantSKs, covenantPKs, err := datagen.GenRandomBTCKeyPairs(r, 5)
	h.NoError(err)
	slashingAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
	h.NoError(err)
	err = h.BTCStakingKeeper.SetParams(h.Ctx, types.Params{
		CovenantPks:            bbn.NewBIP340PKsFromBTCPKs(covenantPKs),
		CovenantQuorum:         3,
		SlashingAddress:        slashingAddress.EncodeAddress(),
		MinSlashingTxFeeSat:    10,
		MinCommissionRate:      sdkmath.LegacyMustNewDecFromStr("0.01"),
		SlashingRate:           sdkmath.LegacyNewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2),
		MaxActiveBtcValidators: 100,
	})
	h.NoError(err)
	return covenantSKs, covenantPKs
}

func (h *Helper) CreateValidator(r *rand.Rand) (*btcec.PrivateKey, *btcec.PublicKey, *types.BTCValidator) {
	validatorSK, validatorPK, err := datagen.GenRandomBTCKeyPair(r)
	h.NoError(err)
	btcVal, err := datagen.GenRandomBTCValidatorWithBTCSK(r, validatorSK)
	h.NoError(err)
	msgNewVal := types.MsgCreateBTCValidator{
		Signer:      datagen.GenRandomAccount().Address,
		Description: btcVal.Description,
		Commission:  btcVal.Commission,
		BabylonPk:   btcVal.BabylonPk,
		BtcPk:       btcVal.BtcPk,
		Pop:         btcVal.Pop,
	}
	_, err = h.MsgServer.CreateBTCValidator(h.Ctx, &msgNewVal)
	h.NoError(err)
	return validatorSK, validatorPK, btcVal
}

func (h *Helper) CreateDelegation(
	r *rand.Rand,
	validatorPK *btcec.PublicKey,
	changeAddress string,
	stakingTime uint16,
) (string, *btcec.PrivateKey, *btcec.PublicKey, *types.MsgCreateBTCDelegation) {
	delSK, delPK, err := datagen.GenRandomBTCKeyPair(r)
	h.NoError(err)
	stakingTimeBlocks := stakingTime
	stakingValue := int64(2 * 10e8)
	bsParams := h.BTCStakingKeeper.GetParams(h.Ctx)
	covPKs, err := bbn.NewBTCPKsFromBIP340PKs(bsParams.CovenantPks)
	h.NoError(err)
	testStakingInfo := datagen.GenBTCStakingSlashingInfo(
		r,
		h.t,
		h.Net,
		delSK,
		[]*btcec.PublicKey{validatorPK},
		covPKs,
		bsParams.CovenantQuorum,
		stakingTimeBlocks,
		stakingValue,
		bsParams.SlashingAddress,
		changeAddress,
		bsParams.SlashingRate,
	)
	h.NoError(err)
	stakingTxHash := testStakingInfo.StakingTx.TxHash().String()

	// random signer
	signer := datagen.GenRandomAccount().Address
	// random Babylon SK
	delBabylonSK, delBabylonPK, err := datagen.GenRandomSecp256k1KeyPair(r)
	h.NoError(err)
	// PoP
	pop, err := types.NewPoP(delBabylonSK, delSK)
	h.NoError(err)
	// generate staking tx info
	prevBlock, _ := datagen.GenRandomBtcdBlock(r, 0, nil)
	btcHeaderWithProof := datagen.CreateBlockWithTransaction(r, &prevBlock.Header, testStakingInfo.StakingTx)
	btcHeader := btcHeaderWithProof.HeaderBytes
	serializedStakingTx, err := bbn.SerializeBTCTx(testStakingInfo.StakingTx)
	h.NoError(err)

	txInfo := btcctypes.NewTransactionInfo(&btcctypes.TransactionKey{Index: 1, Hash: btcHeader.Hash()}, serializedStakingTx, btcHeaderWithProof.SpvProof.MerkleNodes)

	// mock for testing k-deep stuff
	h.BTCCheckpointKeeper.EXPECT().GetPowLimit().Return(h.Net.PowLimit).AnyTimes()
	h.BTCCheckpointKeeper.EXPECT().GetParams(gomock.Any()).Return(btcctypes.DefaultParams()).AnyTimes()
	h.BTCLightClientKeeper.EXPECT().GetHeaderByHash(gomock.Any(), gomock.Eq(btcHeader.Hash())).Return(&btclctypes.BTCHeaderInfo{Header: &btcHeader, Height: 10}).AnyTimes()
	h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 30})

	slashingSpendInfo, err := testStakingInfo.StakingInfo.SlashingPathSpendInfo()
	h.NoError(err)

	// generate proper delegator sig
	delegatorSig, err := testStakingInfo.SlashingTx.Sign(
		testStakingInfo.StakingTx,
		0,
		slashingSpendInfo.GetPkScriptPath(),
		delSK,
	)
	h.NoError(err)

	stakerPk := delSK.PubKey()
	stPk := bbn.NewBIP340PubKeyFromBTCPK(stakerPk)

	// all good, construct and send MsgCreateBTCDelegation message
	msgCreateBTCDel := &types.MsgCreateBTCDelegation{
		Signer:       signer,
		BabylonPk:    delBabylonPK.(*secp256k1.PubKey),
		BtcPk:        stPk,
		ValBtcPkList: []bbn.BIP340PubKey{*bbn.NewBIP340PubKeyFromBTCPK(validatorPK)},
		Pop:          pop,
		StakingTime:  uint32(stakingTimeBlocks),
		StakingValue: stakingValue,
		StakingTx:    txInfo,
		SlashingTx:   testStakingInfo.SlashingTx,
		DelegatorSig: delegatorSig,
	}
	_, err = h.MsgServer.CreateBTCDelegation(h.Ctx, msgCreateBTCDel)
	h.NoError(err)
	return stakingTxHash, delSK, delPK, msgCreateBTCDel
}

func (h *Helper) CreateCovenantSigs(
	r *rand.Rand,
	covenantSKs []*btcec.PrivateKey,
	msgCreateBTCDel *types.MsgCreateBTCDelegation,
	delegation *types.BTCDelegation,
) {
	stakingTx, err := bbn.NewBTCTxFromBytes(delegation.StakingTx)
	h.NoError(err)
	stakingTxHash := stakingTx.TxHash().String()

	bsParams := h.BTCStakingKeeper.GetParams(h.Ctx)
	cPks, err := bbn.NewBTCPKsFromBIP340PKs(bsParams.CovenantPks)
	h.NoError(err)

	vPKs, err := bbn.NewBTCPKsFromBIP340PKs(delegation.ValBtcPkList)
	h.NoError(err)

	info, err := btcstaking.BuildStakingInfo(
		delegation.BtcPk.MustToBTCPK(),
		vPKs,
		cPks,
		bsParams.CovenantQuorum,
		delegation.GetStakingTime(),
		btcutil.Amount(delegation.TotalSat),
		h.Net,
	)
	h.NoError(err)

	slashingPathInfo, err := info.SlashingPathSpendInfo()
	h.NoError(err)

	// generate all covenant signatures from all covenant members
	covenantSigs, err := datagen.GenCovenantAdaptorSigs(
		covenantSKs,
		vPKs,
		stakingTx,
		slashingPathInfo.GetPkScriptPath(),
		msgCreateBTCDel.SlashingTx,
	)
	h.NoError(err)

	// each covenant member submits signatures
	for i := 0; i < len(covenantSigs); i++ {
		msgAddCovenantSig := &types.MsgAddCovenantSig{
			Signer:        msgCreateBTCDel.Signer,
			Pk:            covenantSigs[i].CovPk,
			StakingTxHash: stakingTxHash,
			Sigs:          covenantSigs[i].AdaptorSigs,
		}
		_, err = h.MsgServer.AddCovenantSig(h.Ctx, msgAddCovenantSig)
		h.NoError(err)
	}

	/*
		ensure covenant sig is added successfully
	*/
	actualDelWithCovenantSig, err := h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
	h.NoError(err)
	require.Equal(h.t, len(actualDelWithCovenantSig.CovenantSigs), int(bsParams.CovenantQuorum)) // TODO: fix
	require.True(h.t, actualDelWithCovenantSig.HasCovenantQuorum(h.BTCStakingKeeper.GetParams(h.Ctx).CovenantQuorum))
}

func (h *Helper) GetDelegationAndCheckValues(
	r *rand.Rand,
	msgCreateBTCDel *types.MsgCreateBTCDelegation,
	validatorPK *btcec.PublicKey,
	delegatorPK *btcec.PublicKey,
	stakingTxHash string,
) *types.BTCDelegation {
	actualDel, err := h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
	h.NoError(err)
	require.Equal(h.t, msgCreateBTCDel.BabylonPk, actualDel.BabylonPk)
	require.Equal(h.t, msgCreateBTCDel.Pop, actualDel.Pop)
	require.Equal(h.t, msgCreateBTCDel.StakingTx.Transaction, actualDel.StakingTx)
	require.Equal(h.t, msgCreateBTCDel.SlashingTx, actualDel.SlashingTx)
	// ensure the BTC delegation in DB is correctly formatted
	err = actualDel.ValidateBasic()
	h.NoError(err)
	// delegation is not activated by covenant yet
	require.False(h.t, actualDel.HasCovenantQuorum(h.BTCStakingKeeper.GetParams(h.Ctx).CovenantQuorum))
	return actualDel
}

func (h *Helper) CreateUndelegation(
	r *rand.Rand,
	actualDel *types.BTCDelegation,
	delSK *btcec.PrivateKey,
	validatorPK *btcec.PublicKey,
	changeAddress string,
) *types.MsgBTCUndelegate {
	stkTxHash, err := actualDel.GetStakingTxHash()
	h.NoError(err)
	stkOutputIdx := uint32(0)
	defaultParams := btcctypes.DefaultParams()

	unbondingTime := uint16(defaultParams.CheckpointFinalizationTimeout) + 1
	unbondingValue := int64(actualDel.TotalSat) - 1000
	bsParams := h.BTCStakingKeeper.GetParams(h.Ctx)
	covPKs, err := bbn.NewBTCPKsFromBIP340PKs(bsParams.CovenantPks)
	h.NoError(err)
	testUnbondingInfo := datagen.GenBTCUnbondingSlashingInfo(
		r,
		h.t,
		h.Net,
		delSK,
		[]*btcec.PublicKey{validatorPK},
		covPKs,
		bsParams.CovenantQuorum,
		wire.NewOutPoint(&stkTxHash, stkOutputIdx),
		unbondingTime,
		unbondingValue,
		bsParams.SlashingAddress,
		changeAddress,
		bsParams.SlashingRate,
	)
	h.NoError(err)
	// random signer
	signer := datagen.GenRandomAccount().Address
	unbondingTxMsg := testUnbondingInfo.UnbondingTx

	unbondingSlashingPathInfo, err := testUnbondingInfo.UnbondingInfo.SlashingPathSpendInfo()
	h.NoError(err)

	sig, err := testUnbondingInfo.SlashingTx.Sign(
		unbondingTxMsg,
		0,
		unbondingSlashingPathInfo.GetPkScriptPath(),
		delSK,
	)
	h.NoError(err)

	serializedUnbondingTx, err := bbn.SerializeBTCTx(testUnbondingInfo.UnbondingTx)
	h.NoError(err)

	msg := &types.MsgBTCUndelegate{
		Signer:               signer,
		UnbondingTx:          serializedUnbondingTx,
		UnbondingTime:        uint32(unbondingTime),
		UnbondingValue:       unbondingValue,
		SlashingTx:           testUnbondingInfo.SlashingTx,
		DelegatorSlashingSig: sig,
	}
	h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: actualDel.StartHeight + 1})
	_, err = h.MsgServer.BTCUndelegate(h.Ctx, msg)
	h.NoError(err)
	return msg
}

func FuzzMsgCreateBTCValidator(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		h := NewHelper(t, nil, nil)

		// generate new BTC validators
		btcVals := []*types.BTCValidator{}
		for i := 0; i < int(datagen.RandomInt(r, 10)); i++ {
			btcVal, err := datagen.GenRandomBTCValidator(r)
			require.NoError(t, err)
			msg := &types.MsgCreateBTCValidator{
				Signer:      datagen.GenRandomAccount().Address,
				Description: btcVal.Description,
				Commission:  btcVal.Commission,
				BabylonPk:   btcVal.BabylonPk,
				BtcPk:       btcVal.BtcPk,
				Pop:         btcVal.Pop,
			}
			_, err = h.MsgServer.CreateBTCValidator(h.Ctx, msg)
			require.NoError(t, err)

			btcVals = append(btcVals, btcVal)
		}
		// assert these validators exist in KVStore
		for _, btcVal := range btcVals {
			btcPK := *btcVal.BtcPk
			require.True(t, h.BTCStakingKeeper.HasBTCValidator(h.Ctx, btcPK))
		}

		// duplicated BTC validators should not pass
		for _, btcVal2 := range btcVals {
			msg := &types.MsgCreateBTCValidator{
				Signer:      datagen.GenRandomAccount().Address,
				Description: btcVal2.Description,
				Commission:  btcVal2.Commission,
				BabylonPk:   btcVal2.BabylonPk,
				BtcPk:       btcVal2.BtcPk,
				Pop:         btcVal2.Pop,
			}
			_, err := h.MsgServer.CreateBTCValidator(h.Ctx, msg)
			require.Error(t, err)
		}
	})
}

func FuzzCreateBTCDelegationAndAddCovenantSig(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		net := &chaincfg.SimNetParams
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock BTC light client and BTC checkpoint modules
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		h := NewHelper(t, btclcKeeper, btccKeeper)

		// set covenant PKs to params
		covenantSKs, _ := h.GenAndApplyParams(r)

		changeAddress, err := datagen.GenRandomBTCAddress(r, net)
		require.NoError(t, err)

		/*
			generate and insert new BTC validator
		*/
		_, validatorPK, _ := h.CreateValidator(r)

		/*
			generate and insert new BTC delegation
		*/
		stakingTxHash, _, delPK, msgCreateBTCDel := h.CreateDelegation(
			r,
			validatorPK,
			changeAddress.EncodeAddress(),
			1000,
		)

		/*
			verify the new BTC delegation
		*/
		// check existence
		actualDel := h.GetDelegationAndCheckValues(r, msgCreateBTCDel, validatorPK, delPK, stakingTxHash)

		/*
			generate and insert new covenant signature
		*/
		h.CreateCovenantSigs(r, covenantSKs, msgCreateBTCDel, actualDel)
	})
}

func TestDoNotAllowDelegationWithoutValidator(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	net := &chaincfg.SimNetParams
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// mock BTC light client and BTC checkpoint modules
	btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
	btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
	btccKeeper.EXPECT().GetParams(gomock.Any()).Return(btcctypes.DefaultParams()).AnyTimes()
	h := NewHelper(t, btclcKeeper, btccKeeper)

	// set covenant PK to params
	_, covenantPKs := h.GenAndApplyParams(r)
	bsParams := h.BTCStakingKeeper.GetParams(h.Ctx)

	changeAddress, err := datagen.GenRandomBTCAddress(r, net)
	require.NoError(t, err)

	// We only generate a validator, but not insert it into KVStore. So later
	// insertion of delegation should fail.
	_, validatorPK, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)

	/*
		generate and insert valid new BTC delegation
	*/
	delSK, _, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)
	stakingTimeBlocks := uint16(5)
	stakingValue := int64(2 * 10e8)
	testStakingInfo := datagen.GenBTCStakingSlashingInfo(
		r,
		t,
		net,
		delSK,
		[]*btcec.PublicKey{validatorPK},
		covenantPKs,
		bsParams.CovenantQuorum,
		stakingTimeBlocks,
		stakingValue,
		bsParams.SlashingAddress,
		changeAddress.EncodeAddress(),
		bsParams.SlashingRate,
	)
	// get msgTx
	stakingMsgTx := testStakingInfo.StakingTx
	serializedStakingTx, err := bbn.SerializeBTCTx(stakingMsgTx)
	require.NoError(t, err)
	// random signer
	signer := datagen.GenRandomAccount().Address
	// random Babylon SK
	delBabylonSK, delBabylonPK, err := datagen.GenRandomSecp256k1KeyPair(r)
	require.NoError(t, err)
	// PoP
	pop, err := types.NewPoP(delBabylonSK, delSK)
	require.NoError(t, err)
	// generate staking tx info
	prevBlock, _ := datagen.GenRandomBtcdBlock(r, 0, nil)
	btcHeaderWithProof := datagen.CreateBlockWithTransaction(r, &prevBlock.Header, stakingMsgTx)
	btcHeader := btcHeaderWithProof.HeaderBytes
	txInfo := btcctypes.NewTransactionInfo(
		&btcctypes.TransactionKey{Index: 1, Hash: btcHeader.Hash()},
		serializedStakingTx,
		btcHeaderWithProof.SpvProof.MerkleNodes,
	)

	slashingPathInfo, err := testStakingInfo.StakingInfo.SlashingPathSpendInfo()
	require.NoError(t, err)

	// generate proper delegator sig
	delegatorSig, err := testStakingInfo.SlashingTx.Sign(
		stakingMsgTx,
		0,
		slashingPathInfo.GetPkScriptPath(),
		delSK,
	)
	require.NoError(t, err)

	// all good, construct and send MsgCreateBTCDelegation message
	msgCreateBTCDel := &types.MsgCreateBTCDelegation{
		Signer:       signer,
		BabylonPk:    delBabylonPK.(*secp256k1.PubKey),
		ValBtcPkList: []bbn.BIP340PubKey{*bbn.NewBIP340PubKeyFromBTCPK(validatorPK)},
		BtcPk:        bbn.NewBIP340PubKeyFromBTCPK(delSK.PubKey()),
		Pop:          pop,
		StakingTime:  uint32(stakingTimeBlocks),
		StakingValue: stakingValue,
		StakingTx:    txInfo,
		SlashingTx:   testStakingInfo.SlashingTx,
		DelegatorSig: delegatorSig,
	}
	_, err = h.MsgServer.CreateBTCDelegation(h.Ctx, msgCreateBTCDel)
	require.Error(t, err)
	require.True(t, errors.Is(err, types.ErrBTCValNotFound))
}

func FuzzCreateBTCDelegationAndUndelegation(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		net := &chaincfg.SimNetParams
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock BTC light client and BTC checkpoint modules
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		h := NewHelper(t, btclcKeeper, btccKeeper)

		covenantSKs, _ := h.GenAndApplyParams(r)
		changeAddress, err := datagen.GenRandomBTCAddress(r, net)
		require.NoError(t, err)
		_, validatorPK, _ := h.CreateValidator(r)
		stakingTxHash, delSK, delPK, msgCreateBTCDel := h.CreateDelegation(
			r,
			validatorPK,
			changeAddress.EncodeAddress(),
			1000,
		)
		actualDel := h.GetDelegationAndCheckValues(r, msgCreateBTCDel, validatorPK, delPK, stakingTxHash)
		h.CreateCovenantSigs(r, covenantSKs, msgCreateBTCDel, actualDel)

		undelegateMsg := h.CreateUndelegation(
			r,
			actualDel,
			delSK,
			validatorPK,
			changeAddress.EncodeAddress(),
		)

		actualDelegationWithUnbonding, err := h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
		require.NoError(t, err)

		require.NotNil(t, actualDelegationWithUnbonding.BtcUndelegation)
		require.Equal(t, actualDelegationWithUnbonding.BtcUndelegation.UnbondingTx, undelegateMsg.UnbondingTx)
		require.Equal(t, actualDelegationWithUnbonding.BtcUndelegation.SlashingTx, undelegateMsg.SlashingTx)
		require.Equal(t, actualDelegationWithUnbonding.BtcUndelegation.DelegatorSlashingSig, undelegateMsg.DelegatorSlashingSig)
		require.Nil(t, actualDelegationWithUnbonding.BtcUndelegation.CovenantSlashingSigs)
		require.Equal(t, actualDelegationWithUnbonding.BtcUndelegation.UnbondingTime, undelegateMsg.UnbondingTime)
		require.Nil(t, actualDelegationWithUnbonding.BtcUndelegation.CovenantUnbondingSigList)
	})
}

func FuzzAddCovenantSigToUnbonding(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		net := &chaincfg.SimNetParams
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock BTC light client and BTC checkpoint modules
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		h := NewHelper(t, btclcKeeper, btccKeeper)

		covenantSKs, covenantPKs := h.GenAndApplyParams(r)
		bsParams := h.BTCStakingKeeper.GetParams(h.Ctx)

		changeAddress, err := datagen.GenRandomBTCAddress(r, net)
		require.NoError(t, err)
		_, validatorPK, _ := h.CreateValidator(r)
		stakingTxHash, delSK, delPK, msgCreateBTCDel := h.CreateDelegation(
			r,
			validatorPK,
			changeAddress.EncodeAddress(),
			1000,
		)
		actualDel := h.GetDelegationAndCheckValues(r, msgCreateBTCDel, validatorPK, delPK, stakingTxHash)
		h.CreateCovenantSigs(r, covenantSKs, msgCreateBTCDel, actualDel)

		undelegateMsg := h.CreateUndelegation(
			r,
			actualDel,
			delSK,
			validatorPK,
			changeAddress.EncodeAddress(),
		)

		del, err := h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
		require.NoError(t, err)
		require.NotNil(t, del.BtcUndelegation)

		stakingTx, err := bbn.NewBTCTxFromBytes(del.StakingTx)
		require.NoError(t, err)

		unbondingTx, err := bbn.NewBTCTxFromBytes(del.BtcUndelegation.UnbondingTx)
		require.NoError(t, err)

		// Check sending covenant signatures
		// unbonding tx spends staking tx
		stakingInfo, err := btcstaking.BuildStakingInfo(
			del.BtcPk.MustToBTCPK(),
			[]*btcec.PublicKey{validatorPK},
			covenantPKs,
			bsParams.CovenantQuorum,
			uint16(del.GetStakingTime()),
			btcutil.Amount(del.TotalSat),
			net,
		)
		require.NoError(t, err)

		unbondingPathInfo, err := stakingInfo.UnbondingPathSpendInfo()
		require.NoError(t, err)

		// slash unbonding tx spends unbonding tx
		unbondingInfo, err := btcstaking.BuildUnbondingInfo(
			del.BtcPk.MustToBTCPK(),
			[]*btcec.PublicKey{validatorPK},
			covenantPKs,
			bsParams.CovenantQuorum,
			uint16(del.BtcUndelegation.GetUnbondingTime()),
			btcutil.Amount(unbondingTx.TxOut[0].Value),
			net,
		)
		require.NoError(t, err)

		unbondingSlashingPathInfo, err := unbondingInfo.SlashingPathSpendInfo()
		require.NoError(t, err)

		enckey, err := asig.NewEncryptionKeyFromBTCPK(validatorPK)
		require.NoError(t, err)

		// submit covenant signatures for each covenant member
		for i := 0; i < int(bsParams.CovenantQuorum); i++ {
			covenantUnbondingBTCSig, err := btcstaking.SignTxWithOneScriptSpendInputStrict(
				unbondingTx,
				stakingTx,
				del.StakingOutputIdx,
				unbondingPathInfo.GetPkScriptPath(),
				covenantSKs[i],
			)

			covenantUnbondingSig := bbn.NewBIP340SignatureFromBTCSig(covenantUnbondingBTCSig)
			require.NoError(t, err)

			slashUnbondingTxSig, err := undelegateMsg.SlashingTx.EncSign(
				unbondingTx,
				0,
				unbondingSlashingPathInfo.GetPkScriptPath(),
				covenantSKs[i],
				enckey,
			)
			require.NoError(t, err)

			covenantSigsMsg := types.MsgAddCovenantUnbondingSigs{
				Signer:                  datagen.GenRandomAccount().Address,
				Pk:                      bbn.NewBIP340PubKeyFromBTCPK(covenantSKs[i].PubKey()),
				StakingTxHash:           stakingTxHash,
				UnbondingTxSig:          &covenantUnbondingSig,
				SlashingUnbondingTxSigs: [][]byte{slashUnbondingTxSig.MustMarshal()},
			}

			btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: actualDel.StartHeight + 1})
			_, err = h.MsgServer.AddCovenantUnbondingSigs(h.Ctx, &covenantSigsMsg)
			require.NoError(t, err)
		}

		delWithUnbondingSigs, err := h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
		require.NoError(t, err)
		require.NotNil(t, delWithUnbondingSigs.BtcUndelegation)
		require.NotNil(t, delWithUnbondingSigs.BtcUndelegation.CovenantSlashingSigs)
		require.NotNil(t, delWithUnbondingSigs.BtcUndelegation.CovenantUnbondingSigList)
		require.Len(t, delWithUnbondingSigs.BtcUndelegation.CovenantUnbondingSigList, int(bsParams.CovenantQuorum))
		require.Len(t, delWithUnbondingSigs.BtcUndelegation.CovenantSlashingSigs, int(bsParams.CovenantQuorum))
		require.Len(t, delWithUnbondingSigs.BtcUndelegation.CovenantSlashingSigs[0].AdaptorSigs, 1)

		covPKMap := map[string]struct{}{}
		for _, covPK := range covenantPKs {
			covPKMap[bbn.NewBIP340PubKeyFromBTCPK(covPK).MarshalHex()] = struct{}{}
		}
		for _, actualCovSigs := range delWithUnbondingSigs.BtcUndelegation.CovenantSlashingSigs {
			_, ok := covPKMap[actualCovSigs.CovPk.MarshalHex()]
			require.True(t, ok)
		}
	})
}
