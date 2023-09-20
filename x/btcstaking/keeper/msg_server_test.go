package keeper_test

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

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
	return k, keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
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

func getJuryInfo(t *testing.T,
	r *rand.Rand,
	goCtx context.Context,
	ms types.MsgServer,
	net *chaincfg.Params,
	bsKeeper *keeper.Keeper,
	sdkCtx sdk.Context) (*btcec.PrivateKey, *btcec.PublicKey, btcutil.Address) {
	jurySK, juryPK, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)
	slashingAddr, err := datagen.GenRandomBTCAddress(r, net)
	require.NoError(t, err)
	err = bsKeeper.SetParams(sdkCtx, types.Params{
		JuryPk:              bbn.NewBIP340PubKeyFromBTCPK(juryPK),
		SlashingAddress:     slashingAddr.String(),
		MinSlashingTxFeeSat: 10,
		MinCommissionRate:   sdk.MustNewDecFromStr("0.01"),
	})
	require.NoError(t, err)
	return jurySK, juryPK, slashingAddr

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
	juryPK *btcec.PublicKey,
	slashingAddr string,
	stakingTime uint16,
) (string, *btcec.PrivateKey, *btcec.PublicKey, *types.MsgCreateBTCDelegation) {
	delSK, delPK, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)
	stakingTimeBlocks := stakingTime
	stakingValue := int64(2 * 10e8)
	stakingTx, slashingTx, err := datagen.GenBTCStakingSlashingTx(
		r,
		net,
		delSK,
		validatorPK,
		juryPK,
		stakingTimeBlocks,
		stakingValue,
		slashingAddr,
	)
	require.NoError(t, err)
	// get msgTx
	stakingMsgTx, err := stakingTx.ToMsgTx()
	require.NoError(t, err)
	stakingTxHash := stakingTx.MustGetTxHash()

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
	txInfo := btcctypes.NewTransactionInfo(&btcctypes.TransactionKey{Index: 1, Hash: btcHeader.Hash()}, stakingTx.Tx, btcHeaderWithProof.SpvProof.MerkleNodes)

	// mock for testing k-deep stuff
	btccKeeper.EXPECT().GetPowLimit().Return(net.PowLimit).AnyTimes()
	btccKeeper.EXPECT().GetParams(gomock.Any()).Return(btcctypes.DefaultParams()).AnyTimes()
	btclcKeeper.EXPECT().GetHeaderByHash(gomock.Any(), gomock.Eq(btcHeader.Hash())).Return(&btclctypes.BTCHeaderInfo{Header: &btcHeader, Height: 10}).AnyTimes()
	btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 30})

	// generate proper delegator sig
	delegatorSig, err := slashingTx.Sign(
		stakingMsgTx,
		stakingTx.Script,
		delSK,
		net,
	)
	require.NoError(t, err)

	// all good, construct and send MsgCreateBTCDelegation message
	msgCreateBTCDel := &types.MsgCreateBTCDelegation{
		Signer:        signer,
		BabylonPk:     delBabylonPK.(*secp256k1.PubKey),
		Pop:           pop,
		StakingTx:     stakingTx,
		StakingTxInfo: txInfo,
		SlashingTx:    slashingTx,
		DelegatorSig:  delegatorSig,
	}
	_, err = ms.CreateBTCDelegation(goCtx, msgCreateBTCDel)
	require.NoError(t, err)
	return stakingTxHash, delSK, delPK, msgCreateBTCDel
}

func createJurySig(
	t *testing.T,
	r *rand.Rand,
	goCtx context.Context,
	ms types.MsgServer,
	bsKeeper *keeper.Keeper,
	sdkCtx sdk.Context,
	net *chaincfg.Params,
	jurySK *btcec.PrivateKey,
	msgCreateBTCDel *types.MsgCreateBTCDelegation,
	delegation *types.BTCDelegation,
) {
	stakingMsgTx, err := msgCreateBTCDel.StakingTx.ToMsgTx()
	require.NoError(t, err)
	stakingTxHash := msgCreateBTCDel.StakingTx.MustGetTxHash()
	jurySig, err := msgCreateBTCDel.SlashingTx.Sign(
		stakingMsgTx,
		msgCreateBTCDel.StakingTx.Script,
		jurySK,
		net,
	)
	require.NoError(t, err)
	msgAddJurySig := &types.MsgAddJurySig{
		Signer:        msgCreateBTCDel.Signer,
		ValPk:         delegation.ValBtcPk,
		DelPk:         delegation.BtcPk,
		StakingTxHash: stakingTxHash,
		Sig:           jurySig,
	}
	_, err = ms.AddJurySig(goCtx, msgAddJurySig)
	require.NoError(t, err)

	/*
		ensure jury sig is added successfully
	*/
	actualDelWithJurySig, err := bsKeeper.GetBTCDelegation(sdkCtx, delegation.ValBtcPk, delegation.BtcPk, stakingTxHash)
	require.NoError(t, err)
	require.Equal(t, actualDelWithJurySig.JurySig.MustMarshal(), jurySig.MustMarshal())
	require.True(t, actualDelWithJurySig.HasJurySig())
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
	actualDel, err := bsKeeper.GetBTCDelegation(sdkCtx, bbn.NewBIP340PubKeyFromBTCPK(validatorPK), bbn.NewBIP340PubKeyFromBTCPK(delegatorPK), stakingTxHash)
	require.NoError(t, err)
	require.Equal(t, msgCreateBTCDel.BabylonPk, actualDel.BabylonPk)
	require.Equal(t, msgCreateBTCDel.Pop, actualDel.Pop)
	require.Equal(t, msgCreateBTCDel.StakingTx, actualDel.StakingTx)
	require.Equal(t, msgCreateBTCDel.SlashingTx, actualDel.SlashingTx)
	// ensure the BTC delegation in DB is correctly formatted
	err = actualDel.ValidateBasic()
	require.NoError(t, err)
	// delegation is not activated by jury yet
	require.False(t, actualDel.HasJurySig())
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
	juryPK *btcec.PublicKey,
	slashingAddr string,
) *types.MsgBTCUndelegate {
	stkTxHash, err := chainhash.NewHashFromStr(stakingTxHash)
	require.NoError(t, err)
	stkOutputIdx := uint32(0)
	defaultParams := btcctypes.DefaultParams()

	unbondingTx, slashUnbondingTx, err := datagen.GenBTCUnbondingSlashingTx(
		r,
		net,
		delSK,
		validatorPK,
		juryPK,
		wire.NewOutPoint(stkTxHash, stkOutputIdx),
		uint16(defaultParams.CheckpointFinalizationTimeout)+1,
		int64(actualDel.TotalSat)-1000,
		slashingAddr,
	)
	require.NoError(t, err)
	// random signer
	signer := datagen.GenRandomAccount().Address
	unbondingTxMsg, err := unbondingTx.ToMsgTx()
	require.NoError(t, err)

	sig, err := slashUnbondingTx.Sign(
		unbondingTxMsg,
		unbondingTx.Script,
		delSK,
		net,
	)
	require.NoError(t, err)

	msg := &types.MsgBTCUndelegate{
		Signer:               signer,
		UnbondingTx:          unbondingTx,
		SlashingTx:           slashUnbondingTx,
		DelegatorSlashingSig: sig,
	}
	btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: actualDel.StartHeight + 1})
	_, err = ms.BTCUndelegate(goCtx, msg)
	require.NoError(t, err)
	return msg
}

func FuzzCreateBTCDelegationAndAddJurySig(f *testing.F) {
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
		goCtx := sdk.WrapSDKContext(ctx)

		// set jury PK to params
		jurySK, juryPK, slashingAddr := getJuryInfo(t, r, goCtx, ms, net, bsKeeper, ctx)

		/*
			generate and insert new BTC validator
		*/
		_, validatorPK, _ := createValidator(t, r, goCtx, ms)

		/*
			generate and insert new BTC delegation
		*/
		stakingTxHash, _, delPK, msgCreateBTCDel := createDelegation(
			t,
			r,
			goCtx,
			ms,
			btccKeeper,
			btclcKeeper,
			net,
			validatorPK,
			juryPK,
			slashingAddr.String(),
			1000,
		)

		/*
			verify the new BTC delegation
		*/
		// check existence
		actualDel := getDelegationAndCheckValues(t, r, ms, bsKeeper, ctx, msgCreateBTCDel, validatorPK, delPK, stakingTxHash)

		/*
			generate and insert new jury signature
		*/
		createJurySig(t, r, goCtx, ms, bsKeeper, ctx, net, jurySK, msgCreateBTCDel, actualDel)
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
	goCtx := sdk.WrapSDKContext(ctx)

	// set jury PK to params
	_, juryPK, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)
	slashingAddr, err := datagen.GenRandomBTCAddress(r, net)
	require.NoError(t, err)
	err = bsKeeper.SetParams(ctx, types.Params{
		JuryPk:              bbn.NewBIP340PubKeyFromBTCPK(juryPK),
		SlashingAddress:     slashingAddr.String(),
		MinSlashingTxFeeSat: 10,
		MinCommissionRate:   sdk.MustNewDecFromStr("0.01"),
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
	stakingTx, slashingTx, err := datagen.GenBTCStakingSlashingTx(r, net, delSK, validatorPK, juryPK, stakingTimeBlocks, stakingValue, slashingAddr.String())
	require.NoError(t, err)
	// get msgTx
	stakingMsgTx, err := stakingTx.ToMsgTx()
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
	txInfo := btcctypes.NewTransactionInfo(&btcctypes.TransactionKey{Index: 1, Hash: btcHeader.Hash()}, stakingTx.Tx, btcHeaderWithProof.SpvProof.MerkleNodes)

	// generate proper delegator sig
	delegatorSig, err := slashingTx.Sign(
		stakingMsgTx,
		stakingTx.Script,
		delSK,
		net,
	)
	require.NoError(t, err)

	// all good, construct and send MsgCreateBTCDelegation message
	msgCreateBTCDel := &types.MsgCreateBTCDelegation{
		Signer:        signer,
		BabylonPk:     delBabylonPK.(*secp256k1.PubKey),
		Pop:           pop,
		StakingTx:     stakingTx,
		StakingTxInfo: txInfo,
		SlashingTx:    slashingTx,
		DelegatorSig:  delegatorSig,
	}
	_, err = ms.CreateBTCDelegation(goCtx, msgCreateBTCDel)
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
		goCtx := sdk.WrapSDKContext(ctx)

		jurySK, juryPK, slashingAddr := getJuryInfo(t, r, goCtx, ms, net, bsKeeper, ctx)
		_, validatorPK, _ := createValidator(t, r, goCtx, ms)
		stakingTxHash, delSK, delPK, msgCreateBTCDel := createDelegation(
			t,
			r,
			goCtx,
			ms,
			btccKeeper,
			btclcKeeper,
			net,
			validatorPK,
			juryPK,
			slashingAddr.String(),
			1000,
		)
		actualDel := getDelegationAndCheckValues(t, r, ms, bsKeeper, ctx, msgCreateBTCDel, validatorPK, delPK, stakingTxHash)
		createJurySig(t, r, goCtx, ms, bsKeeper, ctx, net, jurySK, msgCreateBTCDel, actualDel)

		undelegateMsg := createUndelegation(
			t,
			r,
			goCtx,
			ms,
			net,
			btclcKeeper,
			actualDel,
			stakingTxHash,
			delSK,
			validatorPK,
			juryPK,
			slashingAddr.String(),
		)

		actualDelegationWithUnbonding, err := bsKeeper.GetBTCDelegation(ctx, actualDel.ValBtcPk, actualDel.BtcPk, stakingTxHash)
		require.NoError(t, err)

		require.NotNil(t, actualDelegationWithUnbonding.BtcUndelegation)
		require.Equal(t, actualDelegationWithUnbonding.BtcUndelegation.UnbondingTx, undelegateMsg.UnbondingTx)
		require.Equal(t, actualDelegationWithUnbonding.BtcUndelegation.SlashingTx, undelegateMsg.SlashingTx)
		require.Equal(t, actualDelegationWithUnbonding.BtcUndelegation.DelegatorSlashingSig, undelegateMsg.DelegatorSlashingSig)
		require.Nil(t, actualDelegationWithUnbonding.BtcUndelegation.JurySlashingSig)
		require.Nil(t, actualDelegationWithUnbonding.BtcUndelegation.JuryUnbondingSig)
		require.Nil(t, actualDelegationWithUnbonding.BtcUndelegation.ValidatorUnbondingSig)
	})
}

func FuzzAddJuryAndValidatorSignaturesToUnbondind(f *testing.F) {
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
		goCtx := sdk.WrapSDKContext(ctx)

		jurySK, juryPK, slashingAddr := getJuryInfo(t, r, goCtx, ms, net, bsKeeper, ctx)
		valSk, validatorPK, _ := createValidator(t, r, goCtx, ms)
		stakingTxHash, delSK, delPK, msgCreateBTCDel := createDelegation(
			t,
			r,
			goCtx,
			ms,
			btccKeeper,
			btclcKeeper,
			net,
			validatorPK,
			juryPK,
			slashingAddr.String(),
			1000,
		)
		actualDel := getDelegationAndCheckValues(t, r, ms, bsKeeper, ctx, msgCreateBTCDel, validatorPK, delPK, stakingTxHash)
		createJurySig(t, r, goCtx, ms, bsKeeper, ctx, net, jurySK, msgCreateBTCDel, actualDel)

		undelegateMsg := createUndelegation(
			t,
			r,
			goCtx,
			ms,
			net,
			btclcKeeper,
			actualDel,
			stakingTxHash,
			delSK,
			validatorPK,
			juryPK,
			slashingAddr.String(),
		)

		del, err := bsKeeper.GetBTCDelegation(ctx, actualDel.ValBtcPk, actualDel.BtcPk, stakingTxHash)
		require.NoError(t, err)
		require.NotNil(t, del.BtcUndelegation)

		// Check sending validator signature
		stakingTxMsg, err := del.StakingTx.ToMsgTx()
		require.NoError(t, err)
		ubondingTxSignatureValidator, err := undelegateMsg.UnbondingTx.Sign(
			stakingTxMsg,
			del.StakingTx.Script,
			valSk,
			net,
		)
		require.NoError(t, err)
		msg := types.MsgAddValidatorUnbondingSig{
			Signer:         datagen.GenRandomAccount().Address,
			ValPk:          del.ValBtcPk,
			DelPk:          del.BtcPk,
			StakingTxHash:  stakingTxHash,
			UnbondingTxSig: ubondingTxSignatureValidator,
		}
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: actualDel.StartHeight + 1})
		_, err = ms.AddValidatorUnbondingSig(goCtx, &msg)
		require.NoError(t, err)

		delWithValSig, err := bsKeeper.GetBTCDelegation(ctx, actualDel.ValBtcPk, actualDel.BtcPk, stakingTxHash)
		require.NoError(t, err)
		require.NotNil(t, delWithValSig.BtcUndelegation)
		require.NotNil(t, delWithValSig.BtcUndelegation.ValidatorUnbondingSig)

		// Check sending jury signatures
		// unbonding tx spends staking tx
		unbondingTxSignatureJury, err := undelegateMsg.UnbondingTx.Sign(
			stakingTxMsg,
			del.StakingTx.Script,
			jurySK,
			net,
		)
		require.NoError(t, err)

		unbondingTxMsg, err := undelegateMsg.UnbondingTx.ToMsgTx()
		require.NoError(t, err)

		// slash unbodning tx spends unbonding tx
		slashUnbondingTxSignatureJury, err := undelegateMsg.SlashingTx.Sign(
			unbondingTxMsg,
			undelegateMsg.UnbondingTx.Script,
			jurySK,
			net,
		)
		require.NoError(t, err)

		jurySigsMsg := types.MsgAddJuryUnbondingSigs{
			Signer:                 datagen.GenRandomAccount().Address,
			ValPk:                  del.ValBtcPk,
			DelPk:                  del.BtcPk,
			StakingTxHash:          stakingTxHash,
			UnbondingTxSig:         unbondingTxSignatureJury,
			SlashingUnbondingTxSig: slashUnbondingTxSignatureJury,
		}

		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: actualDel.StartHeight + 1})
		_, err = ms.AddJuryUnbondingSigs(goCtx, &jurySigsMsg)
		require.NoError(t, err)

		delWithUnbondingSigs, err := bsKeeper.GetBTCDelegation(ctx, actualDel.ValBtcPk, actualDel.BtcPk, stakingTxHash)
		require.NoError(t, err)
		require.NotNil(t, delWithUnbondingSigs.BtcUndelegation)
		require.NotNil(t, delWithUnbondingSigs.BtcUndelegation.ValidatorUnbondingSig)
		require.NotNil(t, delWithUnbondingSigs.BtcUndelegation.JurySlashingSig)
		require.NotNil(t, delWithUnbondingSigs.BtcUndelegation.JuryUnbondingSig)

	})
}
