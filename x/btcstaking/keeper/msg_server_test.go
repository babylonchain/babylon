package keeper_test

import (
	"context"
	sdkmath "cosmossdk.io/math"
	"errors"
	"math/rand"
	"testing"
	"time"

	"github.com/babylonchain/babylon/btcstaking"
	"github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/babylonchain/babylon/x/btcstaking/keeper"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func setupMsgServer(t testing.TB) (*keeper.Keeper, types.MsgServer, context.Context) {
	k, ctx := keepertest.BTCStakingKeeper(t, nil, nil)
	return k, keeper.NewMsgServerImpl(*k), ctx
}

func TestMsgServer(t *testing.T) {
	_, ms, ctx := setupMsgServer(t)
	require.NotNil(t, ms)
	require.NotNil(t, ctx)
}

func FuzzMsgCreateBTCValidator(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		keeper, ms, goCtx := setupMsgServer(t)
		ctx := sdk.UnwrapSDKContext(goCtx)

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
			_, err = ms.CreateBTCValidator(goCtx, msg)
			require.NoError(t, err)

			btcVals = append(btcVals, btcVal)
		}
		// assert these validators exist in KVStore
		for _, btcVal := range btcVals {
			btcPK := *btcVal.BtcPk
			require.True(t, keeper.HasBTCValidator(ctx, btcPK))
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
			_, err := ms.CreateBTCValidator(goCtx, msg)
			require.Error(t, err)
		}
	})
}

func getCovenantInfo(t *testing.T,
	r *rand.Rand,
	goCtx context.Context,
	ms types.MsgServer,
	net *chaincfg.Params,
	bsKeeper *keeper.Keeper,
	sdkCtx sdk.Context) (*btcec.PrivateKey, *btcec.PublicKey, btcutil.Address) {
	covenantSK, covenantPK, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)
	slashingAddress, err := datagen.GenRandomBTCAddress(r, net)
	require.NoError(t, err)
	err = bsKeeper.SetParams(sdkCtx, types.Params{
		CovenantPks:            []bbn.BIP340PubKey{*bbn.NewBIP340PubKeyFromBTCPK(covenantPK)},
		CovenantQuorum:         1,
		SlashingAddress:        slashingAddress.String(),
		MinSlashingTxFeeSat:    10,
		MinCommissionRate:      sdkmath.LegacyMustNewDecFromStr("0.01"),
		SlashingRate:           sdkmath.LegacyNewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2),
		MaxActiveBtcValidators: 100,
	})
	require.NoError(t, err)
	return covenantSK, covenantPK, slashingAddress

}

func createValidator(
	t *testing.T,
	r *rand.Rand,
	goCtx context.Context,
	ms types.MsgServer,
) (*btcec.PrivateKey, *btcec.PublicKey, *types.BTCValidator) {
	validatorSK, validatorPK, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)
	btcVal, err := datagen.GenRandomBTCValidatorWithBTCSK(r, validatorSK)
	require.NoError(t, err)
	msgNewVal := types.MsgCreateBTCValidator{
		Signer:      datagen.GenRandomAccount().Address,
		Description: btcVal.Description,
		Commission:  btcVal.Commission,
		BabylonPk:   btcVal.BabylonPk,
		BtcPk:       btcVal.BtcPk,
		Pop:         btcVal.Pop,
	}
	_, err = ms.CreateBTCValidator(goCtx, &msgNewVal)
	require.NoError(t, err)
	return validatorSK, validatorPK, btcVal
}

func createDelegation(
	t *testing.T,
	r *rand.Rand,
	goCtx context.Context,
	ms types.MsgServer,
	btccKeeper *types.MockBtcCheckpointKeeper,
	btclcKeeper *types.MockBTCLightClientKeeper,
	net *chaincfg.Params,
	validatorPK *btcec.PublicKey,
	covenantPK *btcec.PublicKey,
	slashingAddress, changeAddress string,
	slashingRate sdkmath.LegacyDec,
	stakingTime uint16,
) (string, *btcec.PrivateKey, *btcec.PublicKey, *types.MsgCreateBTCDelegation) {
	delSK, delPK, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)
	stakingTimeBlocks := stakingTime
	stakingValue := int64(2 * 10e8)

	testStakingInfo := datagen.GenBTCStakingSlashingTx(
		r,
		t,
		net,
		delSK,
		[]*btcec.PublicKey{validatorPK},
		[]*btcec.PublicKey{covenantPK},
		1,
		stakingTimeBlocks,
		stakingValue,
		slashingAddress, changeAddress,
		slashingRate,
	)
	require.NoError(t, err)
	stakingTxHash := testStakingInfo.StakingTx.TxHash().String()

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
	btcHeaderWithProof := datagen.CreateBlockWithTransaction(r, &prevBlock.Header, testStakingInfo.StakingTx)
	btcHeader := btcHeaderWithProof.HeaderBytes
	serializedStakingTx, err := types.SerializeBtcTx(testStakingInfo.StakingTx)
	require.NoError(t, err)

	txInfo := btcctypes.NewTransactionInfo(&btcctypes.TransactionKey{Index: 1, Hash: btcHeader.Hash()}, serializedStakingTx, btcHeaderWithProof.SpvProof.MerkleNodes)

	// mock for testing k-deep stuff
	btccKeeper.EXPECT().GetPowLimit().Return(net.PowLimit).AnyTimes()
	btccKeeper.EXPECT().GetParams(gomock.Any()).Return(btcctypes.DefaultParams()).AnyTimes()
	btclcKeeper.EXPECT().GetHeaderByHash(gomock.Any(), gomock.Eq(btcHeader.Hash())).Return(&btclctypes.BTCHeaderInfo{Header: &btcHeader, Height: 10}).AnyTimes()
	btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 30})

	slashignSpendInfo, err := testStakingInfo.StakingInfo.SlashingPathSpendInfo()
	require.NoError(t, err)

	// generate proper delegator sig
	delegatorSig, err := testStakingInfo.SlashingTx.Sign(
		testStakingInfo.StakingTx,
		0,
		slashignSpendInfo.RevealedLeaf.Script,
		delSK,
		net,
	)
	require.NoError(t, err)

	stakerPk := delSK.PubKey()
	stPk := bbn.NewBIP340PubKeyFromBTCPK(stakerPk)

	// all good, construct and send MsgCreateBTCDelegation message
	msgCreateBTCDel := &types.MsgCreateBTCDelegation{
		Signer:       signer,
		BabylonPk:    delBabylonPK.(*secp256k1.PubKey),
		StakerBtcPk:  stPk,
		ValBtcPkList: []bbn.BIP340PubKey{*bbn.NewBIP340PubKeyFromBTCPK(validatorPK)},
		Pop:          pop,
		StakingTime:  uint32(stakingTimeBlocks),
		StakingValue: stakingValue,
		StakingTx:    txInfo,
		SlashingTx:   testStakingInfo.SlashingTx,
		DelegatorSig: delegatorSig,
	}
	_, err = ms.CreateBTCDelegation(goCtx, msgCreateBTCDel)
	require.NoError(t, err)
	return stakingTxHash, delSK, delPK, msgCreateBTCDel
}

func createCovenantSig(
	t *testing.T,
	r *rand.Rand,
	goCtx context.Context,
	ms types.MsgServer,
	bsKeeper *keeper.Keeper,
	sdkCtx sdk.Context,
	net *chaincfg.Params,
	covenantSK *btcec.PrivateKey,
	msgCreateBTCDel *types.MsgCreateBTCDelegation,
	delegation *types.BTCDelegation,
) {
	stakingTx, err := types.ParseBtcTx(delegation.StakingTx)
	require.NoError(t, err)
	stakingTxHash := stakingTx.TxHash().String()

	cPk := covenantSK.PubKey()

	vPK := delegation.ValBtcPkList[0].MustToBTCPK()

	info, err := btcstaking.BuildStakingInfo(
		delegation.BtcPk.MustToBTCPK(),
		[]*btcec.PublicKey{vPK},
		[]*btcec.PublicKey{cPk},
		1,
		delegation.GetStakingTime(),
		btcutil.Amount(delegation.TotalSat),
		net,
	)

	require.NoError(t, err)
	slashingPathInfo, err := info.SlashingPathSpendInfo()
	require.NoError(t, err)

	covenantSig, err := msgCreateBTCDel.SlashingTx.Sign(
		stakingTx,
		0,
		slashingPathInfo.RevealedLeaf.Script,
		covenantSK,
		net,
	)
	require.NoError(t, err)
	msgAddCovenantSig := &types.MsgAddCovenantSig{
		Signer: msgCreateBTCDel.Signer,
		// TODO: this will be removed after all
		ValPk:         &delegation.ValBtcPkList[0],
		DelPk:         delegation.BtcPk,
		StakingTxHash: stakingTxHash,
		Sig:           covenantSig,
	}
	_, err = ms.AddCovenantSig(goCtx, msgAddCovenantSig)
	require.NoError(t, err)

	/*
		ensure covenant sig is added successfully
	*/
	actualDelWithCovenantSig, err := bsKeeper.GetBTCDelegation(sdkCtx, stakingTxHash)
	require.NoError(t, err)
	require.Equal(t, actualDelWithCovenantSig.CovenantSig.MustMarshal(), covenantSig.MustMarshal())
	require.True(t, actualDelWithCovenantSig.HasCovenantSig())
}

func getDelegationAndCheckValues(
	t *testing.T,
	r *rand.Rand,
	ms types.MsgServer,
	bsKeeper *keeper.Keeper,
	sdkCtx sdk.Context,
	msgCreateBTCDel *types.MsgCreateBTCDelegation,
	validatorPK *btcec.PublicKey,
	delegatorPK *btcec.PublicKey,
	stakingTxHash string,
) *types.BTCDelegation {
	actualDel, err := bsKeeper.GetBTCDelegation(sdkCtx, stakingTxHash)
	require.NoError(t, err)
	require.Equal(t, msgCreateBTCDel.BabylonPk, actualDel.BabylonPk)
	require.Equal(t, msgCreateBTCDel.Pop, actualDel.Pop)
	require.Equal(t, msgCreateBTCDel.StakingTx.Transaction, actualDel.StakingTx)
	require.Equal(t, msgCreateBTCDel.SlashingTx, actualDel.SlashingTx)
	// ensure the BTC delegation in DB is correctly formatted
	err = actualDel.ValidateBasic()
	require.NoError(t, err)
	// delegation is not activated by covenant yet
	require.False(t, actualDel.HasCovenantSig())
	return actualDel
}

func createUndelegation(
	t *testing.T,
	r *rand.Rand,
	goCtx context.Context,
	ms types.MsgServer,
	net *chaincfg.Params,
	btclcKeeper *types.MockBTCLightClientKeeper,
	actualDel *types.BTCDelegation,
	stakingTxHash string,
	delSK *btcec.PrivateKey,
	validatorPK *btcec.PublicKey,
	covenantPK *btcec.PublicKey,
	slashingAddress, changeAddress string,
	slashingRate sdkmath.LegacyDec,
) *types.MsgBTCUndelegate {
	stkTxHash, err := chainhash.NewHashFromStr(stakingTxHash)
	require.NoError(t, err)
	stkOutputIdx := uint32(0)
	defaultParams := btcctypes.DefaultParams()

	unbondingTime := uint16(defaultParams.CheckpointFinalizationTimeout) + 1
	unbondingValue := int64(actualDel.TotalSat) - 1000

	testUnbondingInfo := datagen.GenBTCUnbondingSlashingTx(
		r,
		t,
		net,
		delSK,
		[]*btcec.PublicKey{validatorPK},
		[]*btcec.PublicKey{covenantPK},
		1,
		wire.NewOutPoint(stkTxHash, stkOutputIdx),
		unbondingTime,
		unbondingValue,
		slashingAddress, changeAddress,
		slashingRate,
	)
	require.NoError(t, err)
	// random signer
	signer := datagen.GenRandomAccount().Address
	unbondingTxMsg := testUnbondingInfo.UnbondingTx

	unbondingSlashingPathInfo, err := testUnbondingInfo.UnbondingInfo.SlashingPathSpendInfo()
	require.NoError(t, err)

	sig, err := testUnbondingInfo.SlashingTx.Sign(
		unbondingTxMsg,
		0,
		unbondingSlashingPathInfo.RevealedLeaf.Script,
		delSK,
		net,
	)
	require.NoError(t, err)

	serializedUnbondingTx, err := types.SerializeBtcTx(testUnbondingInfo.UnbondingTx)
	require.NoError(t, err)

	msg := &types.MsgBTCUndelegate{
		Signer:               signer,
		UnbondingTx:          serializedUnbondingTx,
		UnbondingTime:        uint32(unbondingTime),
		UnbondingValue:       unbondingValue,
		SlashingTx:           testUnbondingInfo.SlashingTx,
		DelegatorSlashingSig: sig,
	}
	btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: actualDel.StartHeight + 1})
	_, err = ms.BTCUndelegate(goCtx, msg)
	require.NoError(t, err)
	return msg
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
		bsKeeper, ctx := keepertest.BTCStakingKeeper(t, btclcKeeper, btccKeeper)
		ms := keeper.NewMsgServerImpl(*bsKeeper)

		// set covenant PK to params
		covenantSK, covenantPK, slashingAddress := getCovenantInfo(t, r, ctx, ms, net, bsKeeper, ctx)

		changeAddress, err := datagen.GenRandomBTCAddress(r, net)
		require.NoError(t, err)

		/*
			generate and insert new BTC validator
		*/
		_, validatorPK, _ := createValidator(t, r, ctx, ms)

		/*
			generate and insert new BTC delegation
		*/
		stakingTxHash, _, delPK, msgCreateBTCDel := createDelegation(
			t,
			r,
			ctx,
			ms,
			btccKeeper,
			btclcKeeper,
			net,
			validatorPK,
			covenantPK,
			slashingAddress.String(), changeAddress.String(),
			bsKeeper.GetParams(ctx).SlashingRate,
			1000,
		)

		/*
			verify the new BTC delegation
		*/
		// check existence
		actualDel := getDelegationAndCheckValues(t, r, ms, bsKeeper, ctx, msgCreateBTCDel, validatorPK, delPK, stakingTxHash)

		/*
			generate and insert new covenant signature
		*/
		createCovenantSig(t, r, ctx, ms, bsKeeper, ctx, net, covenantSK, msgCreateBTCDel, actualDel)
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
	bsKeeper, ctx := keepertest.BTCStakingKeeper(t, btclcKeeper, btccKeeper)
	ms := keeper.NewMsgServerImpl(*bsKeeper)

	// set covenant PK to params
	_, covenantPK, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)
	slashingAddress, err := datagen.GenRandomBTCAddress(r, net)
	require.NoError(t, err)
	changeAddress, err := datagen.GenRandomBTCAddress(r, net)
	require.NoError(t, err)
	err = bsKeeper.SetParams(ctx, types.Params{
		CovenantPks:            []bbn.BIP340PubKey{*bbn.NewBIP340PubKeyFromBTCPK(covenantPK)},
		CovenantQuorum:         1,
		SlashingAddress:        slashingAddress.String(),
		MinSlashingTxFeeSat:    10,
		MinCommissionRate:      sdkmath.LegacyMustNewDecFromStr("0.01"),
		SlashingRate:           sdkmath.LegacyNewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2),
		MaxActiveBtcValidators: 100,
	})
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
	testStakingInfo := datagen.GenBTCStakingSlashingTx(
		r,
		t,
		net,
		delSK,
		[]*btcec.PublicKey{validatorPK},
		[]*btcec.PublicKey{covenantPK},
		1,
		stakingTimeBlocks,
		stakingValue,
		slashingAddress.String(), changeAddress.String(),
		bsKeeper.GetParams(ctx).SlashingRate,
	)
	// get msgTx
	stakingMsgTx := testStakingInfo.StakingTx
	serializedStakingTx, err := types.SerializeBtcTx(stakingMsgTx)
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
		slashingPathInfo.RevealedLeaf.Script,
		delSK,
		net,
	)
	require.NoError(t, err)

	// all good, construct and send MsgCreateBTCDelegation message
	msgCreateBTCDel := &types.MsgCreateBTCDelegation{
		Signer:       signer,
		BabylonPk:    delBabylonPK.(*secp256k1.PubKey),
		ValBtcPkList: []bbn.BIP340PubKey{*bbn.NewBIP340PubKeyFromBTCPK(validatorPK)},
		StakerBtcPk:  bbn.NewBIP340PubKeyFromBTCPK(delSK.PubKey()),
		Pop:          pop,
		StakingTime:  uint32(stakingTimeBlocks),
		StakingValue: stakingValue,
		StakingTx:    txInfo,
		SlashingTx:   testStakingInfo.SlashingTx,
		DelegatorSig: delegatorSig,
	}
	_, err = ms.CreateBTCDelegation(ctx, msgCreateBTCDel)
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
		bsKeeper, ctx := keepertest.BTCStakingKeeper(t, btclcKeeper, btccKeeper)
		ms := keeper.NewMsgServerImpl(*bsKeeper)

		covenantSK, covenantPK, slashingAddress := getCovenantInfo(t, r, ctx, ms, net, bsKeeper, ctx)
		changeAddress, err := datagen.GenRandomBTCAddress(r, net)
		require.NoError(t, err)
		_, validatorPK, _ := createValidator(t, r, ctx, ms)
		stakingTxHash, delSK, delPK, msgCreateBTCDel := createDelegation(
			t,
			r,
			ctx,
			ms,
			btccKeeper,
			btclcKeeper,
			net,
			validatorPK,
			covenantPK,
			slashingAddress.String(), changeAddress.String(),
			bsKeeper.GetParams(ctx).SlashingRate,
			1000,
		)
		actualDel := getDelegationAndCheckValues(t, r, ms, bsKeeper, ctx, msgCreateBTCDel, validatorPK, delPK, stakingTxHash)
		createCovenantSig(t, r, ctx, ms, bsKeeper, ctx, net, covenantSK, msgCreateBTCDel, actualDel)

		undelegateMsg := createUndelegation(
			t,
			r,
			ctx,
			ms,
			net,
			btclcKeeper,
			actualDel,
			stakingTxHash,
			delSK,
			validatorPK,
			covenantPK,
			slashingAddress.String(), changeAddress.String(),
			bsKeeper.GetParams(ctx).SlashingRate,
		)

		actualDelegationWithUnbonding, err := bsKeeper.GetBTCDelegation(ctx, stakingTxHash)
		require.NoError(t, err)

		require.NotNil(t, actualDelegationWithUnbonding.BtcUndelegation)
		require.Equal(t, actualDelegationWithUnbonding.BtcUndelegation.UnbondingTx, undelegateMsg.UnbondingTx)
		require.Equal(t, actualDelegationWithUnbonding.BtcUndelegation.SlashingTx, undelegateMsg.SlashingTx)
		require.Equal(t, actualDelegationWithUnbonding.BtcUndelegation.DelegatorSlashingSig, undelegateMsg.DelegatorSlashingSig)
		require.Nil(t, actualDelegationWithUnbonding.BtcUndelegation.CovenantSlashingSig)
		require.Nil(t, actualDelegationWithUnbonding.BtcUndelegation.CovenantUnbondingSig)
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
		bsKeeper, ctx := keepertest.BTCStakingKeeper(t, btclcKeeper, btccKeeper)
		ms := keeper.NewMsgServerImpl(*bsKeeper)

		covenantSK, covenantPK, slashingAddress := getCovenantInfo(t, r, ctx, ms, net, bsKeeper, ctx)
		changeAddress, err := datagen.GenRandomBTCAddress(r, net)
		require.NoError(t, err)
		_, validatorPK, _ := createValidator(t, r, ctx, ms)
		stakingTxHash, delSK, delPK, msgCreateBTCDel := createDelegation(
			t,
			r,
			ctx,
			ms,
			btccKeeper,
			btclcKeeper,
			net,
			validatorPK,
			covenantPK,
			slashingAddress.String(), changeAddress.String(),
			bsKeeper.GetParams(ctx).SlashingRate,
			1000,
		)
		actualDel := getDelegationAndCheckValues(t, r, ms, bsKeeper, ctx, msgCreateBTCDel, validatorPK, delPK, stakingTxHash)
		createCovenantSig(t, r, ctx, ms, bsKeeper, ctx, net, covenantSK, msgCreateBTCDel, actualDel)

		undelegateMsg := createUndelegation(
			t,
			r,
			ctx,
			ms,
			net,
			btclcKeeper,
			actualDel,
			stakingTxHash,
			delSK,
			validatorPK,
			covenantPK,
			slashingAddress.String(), changeAddress.String(),
			bsKeeper.GetParams(ctx).SlashingRate,
		)

		del, err := bsKeeper.GetBTCDelegation(ctx, stakingTxHash)
		require.NoError(t, err)
		require.NotNil(t, del.BtcUndelegation)

		stakingTx, err := types.ParseBtcTx(del.StakingTx)
		require.NoError(t, err)

		unbondingTx, err := types.ParseBtcTx(del.BtcUndelegation.UnbondingTx)
		require.NoError(t, err)

		// Check sending covenant signatures
		// unbonding tx spends staking tx
		stakingInfo, err := btcstaking.BuildStakingInfo(
			del.BtcPk.MustToBTCPK(),
			[]*btcec.PublicKey{validatorPK},
			[]*btcec.PublicKey{covenantPK},
			1,
			uint16(del.GetStakingTime()),
			btcutil.Amount(del.TotalSat),
			net,
		)
		require.NoError(t, err)

		stakingUnbondingPathInfo, err := stakingInfo.UnbondingPathSpendInfo()
		require.NoError(t, err)

		unbondingTxSignatureCovenant, err := btcstaking.SignTxWithOneScriptSpendInputStrict(
			unbondingTx,
			stakingTx,
			del.StakingOutputIdx,
			stakingUnbondingPathInfo.RevealedLeaf.Script,
			covenantSK,
			net,
		)

		covenantUnbondingSig := bbn.NewBIP340SignatureFromBTCSig(unbondingTxSignatureCovenant)
		require.NoError(t, err)

		// slash unbodning tx spends unbonding tx
		unbondingInfo, err := btcstaking.BuildUnbondingInfo(
			del.BtcPk.MustToBTCPK(),
			[]*btcec.PublicKey{validatorPK},
			[]*btcec.PublicKey{covenantPK},
			1,
			uint16(del.BtcUndelegation.GetUnbondingTime()),
			btcutil.Amount(unbondingTx.TxOut[0].Value),
			net,
		)
		require.NoError(t, err)

		unbondingSlashingPathInfo, err := unbondingInfo.SlashingPathSpendInfo()
		require.NoError(t, err)

		slashUnbondingTxSignatureCovenant, err := undelegateMsg.SlashingTx.Sign(
			unbondingTx,
			0,
			unbondingSlashingPathInfo.RevealedLeaf.Script,
			covenantSK,
			net,
		)
		require.NoError(t, err)

		covenantSigsMsg := types.MsgAddCovenantUnbondingSigs{
			Signer:                 datagen.GenRandomAccount().Address,
			ValPk:                  &del.ValBtcPkList[0],
			DelPk:                  del.BtcPk,
			StakingTxHash:          stakingTxHash,
			UnbondingTxSig:         &covenantUnbondingSig,
			SlashingUnbondingTxSig: slashUnbondingTxSignatureCovenant,
		}

		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: actualDel.StartHeight + 1})
		_, err = ms.AddCovenantUnbondingSigs(ctx, &covenantSigsMsg)
		require.NoError(t, err)

		delWithUnbondingSigs, err := bsKeeper.GetBTCDelegation(ctx, stakingTxHash)
		require.NoError(t, err)
		require.NotNil(t, delWithUnbondingSigs.BtcUndelegation)
		require.NotNil(t, delWithUnbondingSigs.BtcUndelegation.CovenantSlashingSig)
		require.NotNil(t, delWithUnbondingSigs.BtcUndelegation.CovenantUnbondingSig)
	})
}
