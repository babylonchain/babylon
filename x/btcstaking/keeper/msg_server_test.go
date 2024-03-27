package keeper_test

import (
	"encoding/hex"
	"errors"
	"math"
	"math/rand"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	asig "github.com/babylonchain/babylon/crypto/schnorr-adaptor-signature"
	"github.com/babylonchain/babylon/testutil/datagen"
	testhelper "github.com/babylonchain/babylon/testutil/helper"
	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/babylonchain/babylon/x/btcstaking/keeper"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func FuzzMsgCreateFinalityProvider(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		h := NewHelper(t, nil, nil)

		// generate new finality providers
		fps := []*types.FinalityProvider{}
		for i := 0; i < int(datagen.RandomInt(r, 10)); i++ {
			fp, err := datagen.GenRandomFinalityProvider(r)
			require.NoError(t, err)
			msg := &types.MsgCreateFinalityProvider{
				Signer:      datagen.GenRandomAccount().Address,
				Description: fp.Description,
				Commission:  fp.Commission,
				BabylonPk:   fp.BabylonPk,
				BtcPk:       fp.BtcPk,
				Pop:         fp.Pop,
			}
			_, err = h.MsgServer.CreateFinalityProvider(h.Ctx, msg)
			require.NoError(t, err)

			fps = append(fps, fp)
		}
		// assert these finality providers exist in KVStore
		for _, fp := range fps {
			btcPK := *fp.BtcPk
			require.True(t, h.BTCStakingKeeper.HasFinalityProvider(h.Ctx, btcPK))
		}

		// duplicated finality providers should not pass
		for _, fp2 := range fps {
			msg := &types.MsgCreateFinalityProvider{
				Signer:      datagen.GenRandomAccount().Address,
				Description: fp2.Description,
				Commission:  fp2.Commission,
				BabylonPk:   fp2.BabylonPk,
				BtcPk:       fp2.BtcPk,
				Pop:         fp2.Pop,
			}
			_, err := h.MsgServer.CreateFinalityProvider(h.Ctx, msg)
			require.Error(t, err)
		}
	})
}

func FuzzMsgEditFinalityProvider(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		h := testhelper.NewHelper(t)
		bsKeeper := h.App.BTCStakingKeeper
		msgSrvr := keeper.NewMsgServerImpl(bsKeeper)

		// generate new finality provider
		fp, err := datagen.GenRandomFinalityProvider(r)
		fpAddr := sdk.AccAddress(fp.BabylonPk.Address())
		require.NoError(t, err)
		// insert the finality provider
		h.AddFinalityProvider(fp)
		// assert the finality providers exist in KVStore
		require.True(t, bsKeeper.HasFinalityProvider(h.Ctx, *fp.BtcPk))

		// updated commission and description
		newCommission := datagen.GenRandomCommission(r)
		newDescription := datagen.GenRandomDescription(r)

		// scenario 1: editing finality provider should succeed
		msg := &types.MsgEditFinalityProvider{
			Signer:      fpAddr.String(),
			BtcPk:       *fp.BtcPk,
			Description: newDescription,
			Commission:  &newCommission,
		}
		_, err = msgSrvr.EditFinalityProvider(h.Ctx, msg)
		h.NoError(err)
		editedFp, err := bsKeeper.GetFinalityProvider(h.Ctx, *fp.BtcPk)
		h.NoError(err)
		require.Equal(t, newCommission, *editedFp.Commission)
		require.Equal(t, newDescription, editedFp.Description)

		// scenario 2: message from an unauthorised signer should fail
		newCommission = datagen.GenRandomCommission(r)
		newDescription = datagen.GenRandomDescription(r)
		msg = &types.MsgEditFinalityProvider{
			Signer:      datagen.GenRandomAccount().Address,
			BtcPk:       *fp.BtcPk,
			Description: newDescription,
			Commission:  &newCommission,
		}
		_, err = msgSrvr.EditFinalityProvider(h.Ctx, msg)
		h.Error(err)
		errStatus := status.Convert(err)
		require.Equal(t, codes.PermissionDenied, errStatus.Code())
	})
}

func FuzzCreateBTCDelegation(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock BTC light client and BTC checkpoint modules
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		h := NewHelper(t, btclcKeeper, btccKeeper)

		// set all parameters
		h.GenAndApplyParams(r)

		changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
		require.NoError(t, err)

		// generate and insert new finality provider
		_, fpPK, _ := h.CreateFinalityProvider(r)

		// generate and insert new BTC delegation
		stakingValue := int64(2 * 10e8)
		stakingTxHash, _, _, msgCreateBTCDel, _ := h.CreateDelegation(
			r,
			fpPK,
			changeAddress.EncodeAddress(),
			stakingValue,
			1000,
		)

		// ensure consistency between the msg and the BTC delegation in DB
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
		require.False(h.t, actualDel.HasCovenantQuorums(h.BTCStakingKeeper.GetParams(h.Ctx).CovenantQuorum))
	})
}

func TestProperVersionInDelegation(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// mock BTC light client and BTC checkpoint modules
	btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
	btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
	h := NewHelper(t, btclcKeeper, btccKeeper)

	// set all parameters
	h.GenAndApplyParams(r)

	changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
	require.NoError(t, err)

	// generate and insert new finality provider
	_, fpPK, _ := h.CreateFinalityProvider(r)

	// generate and insert new BTC delegation
	stakingValue := int64(2 * 10e8)
	stakingTxHash, _, _, _, _ := h.CreateDelegation(
		r,
		fpPK,
		changeAddress.EncodeAddress(),
		stakingValue,
		1000,
	)

	// ensure consistency between the msg and the BTC delegation in DB
	actualDel, err := h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
	h.NoError(err)
	err = actualDel.ValidateBasic()
	h.NoError(err)
	// Current version will be `1` as:
	// - version `0` is initialized by `NewHelper`
	// - version `1` is set by `GenAndApplyParams`
	require.Equal(t, uint32(1), actualDel.ParamsVersion)

	customMinUnbondingTime := uint32(2000)
	currentParams := h.BTCStakingKeeper.GetParams(h.Ctx)
	currentParams.MinUnbondingTime = 2000
	// Update new params
	err = h.BTCStakingKeeper.SetParams(h.Ctx, currentParams)
	require.NoError(t, err)
	// create new delegation
	stakingTxHash1, _, _, _, err := h.CreateDelegationCustom(
		r,
		fpPK,
		changeAddress.EncodeAddress(),
		stakingValue,
		1000,
		stakingValue-1000,
		uint16(customMinUnbondingTime)+1,
	)
	require.NoError(t, err)
	actualDel1, err := h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash1)
	h.NoError(err)
	err = actualDel1.ValidateBasic()
	h.NoError(err)
	// Assert that the new delegation has the updated params version
	require.Equal(t, uint32(2), actualDel1.ParamsVersion)
}

func FuzzAddCovenantSigs(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock BTC light client and BTC checkpoint modules
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		h := NewHelper(t, btclcKeeper, btccKeeper)

		// set all parameters
		covenantSKs, _ := h.GenAndApplyParams(r)

		changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
		require.NoError(t, err)

		// generate and insert new finality provider
		_, fpPK, _ := h.CreateFinalityProvider(r)

		// generate and insert new BTC delegation
		stakingValue := int64(2 * 10e8)
		stakingTxHash, _, _, msgCreateBTCDel, _ := h.CreateDelegation(
			r,
			fpPK,
			changeAddress.EncodeAddress(),
			stakingValue,
			1000,
		)

		// ensure consistency between the msg and the BTC delegation in DB
		actualDel, err := h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
		h.NoError(err)
		// delegation is not activated by covenant yet
		require.False(h.t, actualDel.HasCovenantQuorums(h.BTCStakingKeeper.GetParams(h.Ctx).CovenantQuorum))

		msgs := h.GenerateCovenantSignaturesMessages(r, covenantSKs, msgCreateBTCDel, actualDel)

		// ensure the system does not panick due to a bogus covenant sig request
		bogusMsg := *msgs[0]
		bogusMsg.StakingTxHash = datagen.GenRandomBtcdHash(r).String()
		_, err = h.MsgServer.AddCovenantSigs(h.Ctx, &bogusMsg)
		h.Error(err)

		for _, msg := range msgs {
			_, err = h.MsgServer.AddCovenantSigs(h.Ctx, msg)
			h.NoError(err)
			// check that submitting the same covenant signature does not produce an error
			_, err = h.MsgServer.AddCovenantSigs(h.Ctx, msg)
			h.NoError(err)
		}

		// ensure the BTC delegation now has voting power
		actualDel, err = h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
		h.NoError(err)
		require.True(h.t, actualDel.HasCovenantQuorums(h.BTCStakingKeeper.GetParams(h.Ctx).CovenantQuorum))
		require.True(h.t, actualDel.BtcUndelegation.HasCovenantQuorums(h.BTCStakingKeeper.GetParams(h.Ctx).CovenantQuorum))
		votingPower := actualDel.VotingPower(h.BTCLightClientKeeper.GetTipInfo(h.Ctx).Height, h.BTCCheckpointKeeper.GetParams(h.Ctx).CheckpointFinalizationTimeout, h.BTCStakingKeeper.GetParams(h.Ctx).CovenantQuorum)
		require.Equal(t, uint64(stakingValue), votingPower)
	})
}

func FuzzBTCUndelegate(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock BTC light client and BTC checkpoint modules
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		h := NewHelper(t, btclcKeeper, btccKeeper)

		// set all parameters
		covenantSKs, _ := h.GenAndApplyParams(r)

		bsParams := h.BTCStakingKeeper.GetParams(h.Ctx)
		wValue := h.BTCCheckpointKeeper.GetParams(h.Ctx).CheckpointFinalizationTimeout

		changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
		require.NoError(t, err)

		// generate and insert new finality provider
		_, fpPK, _ := h.CreateFinalityProvider(r)

		// generate and insert new BTC delegation
		stakingValue := int64(2 * 10e8)
		stakingTxHash, delSK, _, msgCreateBTCDel, actualDel := h.CreateDelegation(
			r,
			fpPK,
			changeAddress.EncodeAddress(),
			stakingValue,
			1000,
		)

		// add covenant signatures to this BTC delegation
		h.CreateCovenantSigs(r, covenantSKs, msgCreateBTCDel, actualDel)

		// ensure the BTC delegation is bonded right now
		actualDel, err = h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
		h.NoError(err)
		btcTip := h.BTCLightClientKeeper.GetTipInfo(h.Ctx).Height
		status := actualDel.GetStatus(btcTip, wValue, bsParams.CovenantQuorum)
		require.Equal(t, types.BTCDelegationStatus_ACTIVE, status)

		// construct unbonding msg
		delUnbondingSig, err := actualDel.SignUnbondingTx(&bsParams, h.Net, delSK)
		h.NoError(err)
		msg := &types.MsgBTCUndelegate{
			Signer:         datagen.GenRandomAccount().Address,
			StakingTxHash:  stakingTxHash,
			UnbondingTxSig: bbn.NewBIP340SignatureFromBTCSig(delUnbondingSig),
		}

		// ensure the system does not panick due to a bogus unbonding msg
		bogusMsg := *msg
		bogusMsg.StakingTxHash = datagen.GenRandomBtcdHash(r).String()
		_, err = h.MsgServer.BTCUndelegate(h.Ctx, &bogusMsg)
		h.Error(err)

		// unbond
		_, err = h.MsgServer.BTCUndelegate(h.Ctx, msg)
		h.NoError(err)

		// ensure the BTC delegation is unbonded
		actualDel, err = h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
		h.NoError(err)
		status = actualDel.GetStatus(btcTip, wValue, bsParams.CovenantQuorum)
		require.Equal(t, types.BTCDelegationStatus_UNBONDED, status)
	})
}

func FuzzSelectiveSlashing(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock BTC light client and BTC checkpoint modules
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		h := NewHelper(t, btclcKeeper, btccKeeper)

		// set all parameters
		covenantSKs, _ := h.GenAndApplyParams(r)
		bsParams := h.BTCStakingKeeper.GetParams(h.Ctx)

		changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
		require.NoError(t, err)

		// generate and insert new finality provider
		fpSK, fpPK, _ := h.CreateFinalityProvider(r)
		fpBtcPk := bbn.NewBIP340PubKeyFromBTCPK(fpPK)

		// generate and insert new BTC delegation
		stakingValue := int64(2 * 10e8)
		stakingTxHash, _, _, msgCreateBTCDel, actualDel := h.CreateDelegation(
			r,
			fpPK,
			changeAddress.EncodeAddress(),
			stakingValue,
			1000,
		)

		// add covenant signatures to this BTC delegation
		// so that the BTC delegation becomes bonded
		h.CreateCovenantSigs(r, covenantSKs, msgCreateBTCDel, actualDel)
		// now BTC delegation has all covenant signatures
		actualDel, err = h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
		h.NoError(err)
		require.True(t, actualDel.HasCovenantQuorums(bsParams.CovenantQuorum))

		// construct message for the evidence of selective slashing
		msg := &types.MsgSelectiveSlashingEvidence{
			Signer:           datagen.GenRandomAccount().Address,
			StakingTxHash:    actualDel.MustGetStakingTxHash().String(),
			RecoveredFpBtcSk: fpSK.Serialize(),
		}

		// ensure the system does not panick due to a bogus unbonding msg
		bogusMsg := *msg
		bogusMsg.StakingTxHash = datagen.GenRandomBtcdHash(r).String()
		_, err = h.MsgServer.SelectiveSlashingEvidence(h.Ctx, &bogusMsg)
		h.Error(err)

		// submit evidence of selective slashing
		_, err = h.MsgServer.SelectiveSlashingEvidence(h.Ctx, msg)
		h.NoError(err)

		// ensure the finality provider is slashed
		slashedFp, err := h.BTCStakingKeeper.GetFinalityProvider(h.Ctx, fpBtcPk.MustMarshal())
		h.NoError(err)
		require.True(t, slashedFp.IsSlashed())
	})
}

func FuzzSelectiveSlashing_StakingTx(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock BTC light client and BTC checkpoint modules
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		h := NewHelper(t, btclcKeeper, btccKeeper)

		// set all parameters
		covenantSKs, _ := h.GenAndApplyParams(r)
		bsParams := h.BTCStakingKeeper.GetParams(h.Ctx)

		changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
		require.NoError(t, err)

		// generate and insert new finality provider
		fpSK, fpPK, _ := h.CreateFinalityProvider(r)
		fpBtcPk := bbn.NewBIP340PubKeyFromBTCPK(fpPK)

		// generate and insert new BTC delegation
		stakingValue := int64(2 * 10e8)
		stakingTxHash, _, _, msgCreateBTCDel, actualDel := h.CreateDelegation(
			r,
			fpPK,
			changeAddress.EncodeAddress(),
			stakingValue,
			1000,
		)

		// add covenant signatures to this BTC delegation
		// so that the BTC delegation becomes bonded
		h.CreateCovenantSigs(r, covenantSKs, msgCreateBTCDel, actualDel)
		// now BTC delegation has all covenant signatures
		actualDel, err = h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
		h.NoError(err)
		require.True(t, actualDel.HasCovenantQuorums(bsParams.CovenantQuorum))

		// finality provider pulls off selective slashing by decrypting covenant's adaptor signature
		// on the slashing tx
		// choose a random covenant adaptor signature
		covIdx := datagen.RandomInt(r, int(bsParams.CovenantQuorum))
		covPK := bbn.NewBIP340PubKeyFromBTCPK(covenantSKs[covIdx].PubKey())
		fpIdx := datagen.RandomInt(r, len(actualDel.FpBtcPkList))
		covASig, err := actualDel.GetCovSlashingAdaptorSig(covPK, int(fpIdx), bsParams.CovenantQuorum)
		h.NoError(err)

		// finality provider decrypts the covenant signature
		decKey, err := asig.NewDecyptionKeyFromBTCSK(fpSK)
		h.NoError(err)
		decryptedCovenantSig := bbn.NewBIP340SignatureFromBTCSig(covASig.Decrypt(decKey))

		// recover the fpSK by using adaptor signature and decrypted Schnorr signature
		recoveredFPDecKey := covASig.Recover(decryptedCovenantSig.MustToBTCSig())
		recoveredFPSK := recoveredFPDecKey.ToBTCSK()
		// ensure the recovered finality provider SK is same as the real one
		require.Equal(t, fpSK.Serialize(), recoveredFPSK.Serialize())

		// submit evidence of selective slashing
		msg := &types.MsgSelectiveSlashingEvidence{
			Signer:           datagen.GenRandomAccount().Address,
			StakingTxHash:    actualDel.MustGetStakingTxHash().String(),
			RecoveredFpBtcSk: recoveredFPSK.Serialize(),
		}
		_, err = h.MsgServer.SelectiveSlashingEvidence(h.Ctx, msg)
		h.NoError(err)

		// ensure the finality provider is slashed
		slashedFp, err := h.BTCStakingKeeper.GetFinalityProvider(h.Ctx, fpBtcPk.MustMarshal())
		h.NoError(err)
		require.True(t, slashedFp.IsSlashed())
	})
}

func TestDoNotAllowDelegationWithoutFinalityProvider(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
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
	bcParams := h.BTCCheckpointKeeper.GetParams(h.Ctx)

	minUnbondingTime := types.MinimumUnbondingTime(
		bsParams,
		bcParams,
	)

	slashingChangeLockTime := uint16(minUnbondingTime) + 1

	// We only generate a finality provider, but not insert it into KVStore. So later
	// insertion of delegation should fail.
	_, fpPK, err := datagen.GenRandomBTCKeyPair(r)
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
		h.Net,
		delSK,
		[]*btcec.PublicKey{fpPK},
		covenantPKs,
		bsParams.CovenantQuorum,
		stakingTimeBlocks,
		stakingValue,
		bsParams.SlashingAddress,
		bsParams.SlashingRate,
		slashingChangeLockTime,
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

	stkTxHash := testStakingInfo.StakingTx.TxHash()
	unbondingTime := 100 + 1
	unbondingValue := stakingValue - datagen.UnbondingTxFee // TODO: parameterise fee
	testUnbondingInfo := datagen.GenBTCUnbondingSlashingInfo(
		r,
		t,
		h.Net,
		delSK,
		[]*btcec.PublicKey{fpPK},
		covenantPKs,
		bsParams.CovenantQuorum,
		wire.NewOutPoint(&stkTxHash, datagen.StakingOutIdx),
		uint16(unbondingTime),
		unbondingValue,
		bsParams.SlashingAddress,
		bsParams.SlashingRate,
		slashingChangeLockTime,
	)
	unbondingTx, err := bbn.SerializeBTCTx(testUnbondingInfo.UnbondingTx)
	h.NoError(err)
	delUnbondingSlashingSig, err := testUnbondingInfo.GenDelSlashingTxSig(delSK)
	h.NoError(err)

	// all good, construct and send MsgCreateBTCDelegation message
	msgCreateBTCDel := &types.MsgCreateBTCDelegation{
		Signer:                        signer,
		BabylonPk:                     delBabylonPK.(*secp256k1.PubKey),
		FpBtcPkList:                   []bbn.BIP340PubKey{*bbn.NewBIP340PubKeyFromBTCPK(fpPK)},
		BtcPk:                         bbn.NewBIP340PubKeyFromBTCPK(delSK.PubKey()),
		Pop:                           pop,
		StakingTime:                   uint32(stakingTimeBlocks),
		StakingValue:                  stakingValue,
		StakingTx:                     txInfo,
		SlashingTx:                    testStakingInfo.SlashingTx,
		DelegatorSlashingSig:          delegatorSig,
		UnbondingTx:                   unbondingTx,
		UnbondingTime:                 uint32(unbondingTime),
		UnbondingValue:                unbondingValue,
		UnbondingSlashingTx:           testUnbondingInfo.SlashingTx,
		DelegatorUnbondingSlashingSig: delUnbondingSlashingSig,
	}
	_, err = h.MsgServer.CreateBTCDelegation(h.Ctx, msgCreateBTCDel)
	require.Error(t, err)
	require.True(t, errors.Is(err, types.ErrFpNotFound))
}

func TestCorrectUnbondingTimeInDelegation(t *testing.T) {
	tests := []struct {
		name                      string
		finalizationTimeout       uint64
		minUnbondingTime          uint32
		unbondingTimeInDelegation uint16
		err                       error
	}{
		{
			name:                      "successful delegation when ubonding time in delegation is larger than finalization timeout when finalization timeout is larger than min unbonding time",
			unbondingTimeInDelegation: 101,
			minUnbondingTime:          99,
			finalizationTimeout:       100,
			err:                       nil,
		},
		{
			name:                      "failed delegation when ubonding time in delegation is not larger than finalization time when min unbonding time is lower than finalization timeout",
			unbondingTimeInDelegation: 100,
			minUnbondingTime:          99,
			finalizationTimeout:       100,
			err:                       types.ErrInvalidUnbondingTx,
		},
		{
			name:                      "successful delegation when ubonding time ubonding time in delegation is larger than min unbonding time when min unbonding time is larger than finalization timeout",
			unbondingTimeInDelegation: 151,
			minUnbondingTime:          150,
			finalizationTimeout:       100,
			err:                       nil,
		},
		{
			name:                      "failed delegation when ubonding time in delegation is not larger than minUnbondingTime when min unbonding time is larger than finalization timeout",
			unbondingTimeInDelegation: 150,
			minUnbondingTime:          150,
			finalizationTimeout:       100,
			err:                       types.ErrInvalidUnbondingTx,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rand.New(rand.NewSource(time.Now().Unix()))
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// mock BTC light client and BTC checkpoint modules
			btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
			btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
			h := NewHelper(t, btclcKeeper, btccKeeper)

			// set all parameters
			_, _ = h.GenAndApplyCustomParams(r, tt.finalizationTimeout, tt.minUnbondingTime)

			changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
			require.NoError(t, err)

			// generate and insert new finality provider
			_, fpPK, _ := h.CreateFinalityProvider(r)

			// generate and insert new BTC delegation
			stakingValue := int64(2 * 10e8)
			stakingTxHash, _, _, _, err := h.CreateDelegationCustom(
				r,
				fpPK,
				changeAddress.EncodeAddress(),
				stakingValue,
				1000,
				stakingValue-1000,
				tt.unbondingTimeInDelegation,
			)
			if tt.err != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.err))
			} else {
				require.NoError(t, err)
				// Retrieve delegation from keeper
				delegation, err := h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
				require.NoError(t, err)
				require.Equal(t, tt.unbondingTimeInDelegation, uint16(delegation.UnbondingTime))
			}
		})
	}
}

func TestMinimalUnbondingRate(t *testing.T) {
	tests := []struct {
		name                       string
		stakingValue               int64
		unbondingValueInDelegation int64
		err                        error
	}{
		{
			name:                       "successful delegation when unbonding value is >=80% of staking value",
			stakingValue:               10000,
			unbondingValueInDelegation: 8000,
			err:                        nil,
		},
		{
			name:                       "failed delegation when unbonding value is <80% of staking value",
			stakingValue:               10000,
			unbondingValueInDelegation: 7999,
			err:                        types.ErrInvalidUnbondingTx,
		},
		{
			name:                       "failed delegation when unbonding value >= stake value",
			stakingValue:               10000,
			unbondingValueInDelegation: 10000,
			err:                        types.ErrInvalidUnbondingTx,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rand.New(rand.NewSource(time.Now().Unix()))
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// mock BTC light client and BTC checkpoint modules
			btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
			btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
			h := NewHelper(t, btclcKeeper, btccKeeper)

			// set all parameters, by default minimal unbonding value is 80% of staking value
			_, _ = h.GenAndApplyParams(r)

			changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
			require.NoError(t, err)

			// generate and insert new finality provider
			_, fpPK, _ := h.CreateFinalityProvider(r)

			// generate and insert new BTC delegation
			stakingTxHash, _, _, _, err := h.CreateDelegationCustom(
				r,
				fpPK,
				changeAddress.EncodeAddress(),
				tt.stakingValue,
				1000,
				tt.unbondingValueInDelegation,
				1000,
			)
			if tt.err != nil {
				require.Error(t, err)
				require.True(t, errors.Is(err, tt.err))
			} else {
				require.NoError(t, err)
				// Retrieve delegation from keeper
				delegation, err := h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
				require.NoError(t, err)
				require.NotNil(t, delegation)
			}
		})
	}
}

func createNDelegationsForFinalityProvider(
	r *rand.Rand,
	t *testing.T,
	fpPK *btcec.PublicKey,
	stakingValue int64,
	numDelegations int,
	quorum uint32,
) []*types.BTCDelegation {
	var delegations []*types.BTCDelegation
	for i := 0; i < numDelegations; i++ {
		covenatnSks, _, err := datagen.GenRandomBTCKeyPairs(r, int(quorum))
		require.NoError(t, err)

		delSK, _, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)

		net := &chaincfg.SimNetParams
		slashingAddress, err := datagen.GenRandomBTCAddress(r, net)
		require.NoError(t, err)

		slashingRate := sdkmath.LegacyNewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2)

		del, err := datagen.GenRandomBTCDelegation(
			r,
			t,
			[]bbn.BIP340PubKey{*bbn.NewBIP340PubKeyFromBTCPK(fpPK)},
			delSK,
			covenatnSks,
			quorum,
			slashingAddress.EncodeAddress(),
			0,
			0+math.MaxUint16,
			uint64(stakingValue),
			slashingRate,
			math.MaxUint16,
		)
		require.NoError(t, err)

		delegations = append(delegations, del)
	}
	return delegations
}

type ExpectedProviderData struct {
	numDelegations int32
	stakingValue   int32
}

func FuzzDeterminismBtcstakingBeginBlocker(f *testing.F) {
	// less seeds than usual as this is pretty long running test
	datagen.AddRandomSeedsToFuzzer(f, 5)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		valSet, privSigner, err := datagen.GenesisValidatorSetWithPrivSigner(1)
		require.NoError(t, err)

		var expectedProviderData map[string]*ExpectedProviderData = make(map[string]*ExpectedProviderData)

		// Create two test apps from the same set of validators
		h := testhelper.NewHelperWithValSet(t, valSet, privSigner)
		h1 := testhelper.NewHelperWithValSet(t, valSet, privSigner)

		// Default params are the same in both apps
		covQuorum := h.App.BTCStakingKeeper.GetParams(h.Ctx).CovenantQuorum
		maxFinalityProviders := int32(h.App.BTCStakingKeeper.GetParams(h.Ctx).MaxActiveFinalityProviders)

		// Number of finality providers from 10 to maxFinalityProviders + 10
		numFinalityProviders := int(r.Int31n(maxFinalityProviders) + 10)

		fps := datagen.CreateNFinalityProviders(r, t, numFinalityProviders)

		// Fill the databse of both apps with the same finality providers and delegations
		for _, fp := range fps {
			h.AddFinalityProvider(fp)
			h1.AddFinalityProvider(fp)
		}

		for _, fp := range fps {
			// each finality provider has different amount of delegations with different amount
			stakingValue := r.Int31n(200000) + 10000
			numDelegations := r.Int31n(10)

			if numDelegations > 0 {
				expectedProviderData[fp.BtcPk.MarshalHex()] = &ExpectedProviderData{
					numDelegations: numDelegations,
					stakingValue:   stakingValue,
				}
			}

			delegations := createNDelegationsForFinalityProvider(
				r,
				t,
				fp.BtcPk.MustToBTCPK(),
				int64(stakingValue),
				int(numDelegations),
				covQuorum,
			)

			for _, del := range delegations {
				h.AddDelegation(del)
				h1.AddDelegation(del)
			}
		}

		// Execute block for both apps
		ctx, err := h.ApplyEmptyBlockWithVoteExtension(r)
		require.NoError(t, err)

		ctx1, err := h1.ApplyEmptyBlockWithVoteExtension(r)
		require.NoError(t, err)

		// Given that there is no transactions and the data in db is the same
		// app hash produced by both apps should be the same
		appHash1 := hex.EncodeToString(ctx.BlockHeader().AppHash)
		appHash2 := hex.EncodeToString(ctx1.BlockHeader().AppHash)
		require.Equal(t, appHash1, appHash2)
	})
}
