package keeper_test

import (
	"context"
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	keepertest "github.com/babylonchain/babylon/testutil/keeper"
	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	btclctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/babylonchain/babylon/x/btcstaking/keeper"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg"
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
				Signer:    datagen.GenRandomAccount().Address,
				BabylonPk: btcVal.BabylonPk,
				BtcPk:     btcVal.BtcPk,
				Pop:       btcVal.Pop,
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
				Signer:    datagen.GenRandomAccount().Address,
				BabylonPk: btcVal2.BabylonPk,
				BtcPk:     btcVal2.BtcPk,
				Pop:       btcVal2.Pop,
			}
			_, err := ms.CreateBTCValidator(goCtx, msg)
			require.Error(t, err)
		}
	})
}

func FuzzCreateBTCDelegationAndAddJurySig(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// mock BTC light client and BTC checkpoint modules
		btclcKeeper := types.NewMockBTCLightClientKeeper(ctrl)
		btccKeeper := types.NewMockBtcCheckpointKeeper(ctrl)
		bsKeeper, ctx := keepertest.BTCStakingKeeper(t, btclcKeeper, btccKeeper)
		ms := keeper.NewMsgServerImpl(*bsKeeper)
		goCtx := sdk.WrapSDKContext(ctx)

		// set jury PK to params
		jurySK, juryPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		slashingAddr, err := datagen.GenRandomBTCAddress(r, &chaincfg.SimNetParams)
		require.NoError(t, err)
		err = bsKeeper.SetParams(ctx, types.Params{
			JuryPk:              bbn.NewBIP340PubKeyFromBTCPK(juryPK),
			SlashingAddress:     slashingAddr,
			MinSlashingTxFeeSat: 10,
		})
		require.NoError(t, err)

		/*
			generate and insert new BTC validator
		*/
		validatorSK, validatorPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		btcVal, err := datagen.GenRandomBTCValidatorWithBTCSK(r, validatorSK)
		require.NoError(t, err)
		msgNewVal := types.MsgCreateBTCValidator{
			Signer:    datagen.GenRandomAccount().Address,
			BabylonPk: btcVal.BabylonPk,
			BtcPk:     btcVal.BtcPk,
			Pop:       btcVal.Pop,
		}
		_, err = ms.CreateBTCValidator(goCtx, &msgNewVal)
		require.NoError(t, err)

		/*
			generate and insert new BTC delegation
		*/
		// key pairs, staking tx and slashing tx
		delSK, delPK, err := datagen.GenRandomBTCKeyPair(r)
		require.NoError(t, err)
		stakingTimeBlocks := uint16(5)
		stakingValue := int64(2 * 10e8)
		stakingTx, slashingTx, err := datagen.GenBTCStakingSlashingTx(r, delSK, validatorPK, juryPK, stakingTimeBlocks, stakingValue, slashingAddr)
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
		// mock for testing k-deep stuff
		btccKeeper.EXPECT().GetPowLimit().Return(chaincfg.SimNetParams.PowLimit).AnyTimes()
		btccKeeper.EXPECT().GetParams(gomock.Any()).Return(btcctypes.DefaultParams()).AnyTimes()
		btclcKeeper.EXPECT().GetHeaderByHash(gomock.Any(), gomock.Eq(btcHeader.Hash())).Return(&btclctypes.BTCHeaderInfo{Header: &btcHeader, Height: 10}).AnyTimes()
		btclcKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 30})

		// generate proper delegator sig
		delegatorSig, err := slashingTx.Sign(
			stakingMsgTx,
			stakingTx.StakingScript,
			delSK,
			&chaincfg.SimNetParams,
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

		/*
			verify the new BTC delegation
		*/
		// check existence
		actualDel, err := bsKeeper.GetBTCDelegation(ctx, *bbn.NewBIP340PubKeyFromBTCPK(validatorPK), *bbn.NewBIP340PubKeyFromBTCPK(delPK))
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

		/*
			generate and insert new jury signature
		*/
		jurySig, err := slashingTx.Sign(
			stakingMsgTx,
			stakingTx.StakingScript,
			jurySK,
			&chaincfg.SimNetParams,
		)
		require.NoError(t, err)
		msgAddJurySig := &types.MsgAddJurySig{
			Signer: signer,
			ValPk:  btcVal.BtcPk,
			DelPk:  actualDel.BtcPk,
			Sig:    jurySig,
		}
		_, err = ms.AddJurySig(ctx, msgAddJurySig)
		require.NoError(t, err)

		/*
			ensure jury sig is added successfully
		*/
		actualDelWithJurySig, err := bsKeeper.GetBTCDelegation(ctx, *bbn.NewBIP340PubKeyFromBTCPK(validatorPK), *bbn.NewBIP340PubKeyFromBTCPK(delPK))
		require.NoError(t, err)
		require.Equal(t, actualDelWithJurySig.JurySig.MustMarshal(), jurySig.MustMarshal())
		require.True(t, actualDelWithJurySig.HasJurySig())
	})
}
