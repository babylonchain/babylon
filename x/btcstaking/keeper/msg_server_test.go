package keeper_test

import (
	"context"
	"errors"
	"math/rand"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

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
	// mocking stuff for BTC checkpoint keeper
	h.BTCCheckpointKeeper.EXPECT().GetPowLimit().Return(h.Net.PowLimit).AnyTimes()
	h.BTCCheckpointKeeper.EXPECT().GetParams(gomock.Any()).Return(btcctypes.DefaultParams()).AnyTimes()

	// randomise covenant committee
	covenantSKs, covenantPKs, err := datagen.GenRandomBTCKeyPairs(r, 5)
	h.NoError(err)
	slashingAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
	h.NoError(err)
	err = h.BTCStakingKeeper.SetParams(h.Ctx, types.Params{
		CovenantPks:                bbn.NewBIP340PKsFromBTCPKs(covenantPKs),
		CovenantQuorum:             3,
		SlashingAddress:            slashingAddress.EncodeAddress(),
		MinSlashingTxFeeSat:        10,
		MinCommissionRate:          sdkmath.LegacyMustNewDecFromStr("0.01"),
		SlashingRate:               sdkmath.LegacyNewDecWithPrec(int64(datagen.RandomInt(r, 41)+10), 2),
		MaxActiveFinalityProviders: 100,
	})
	h.NoError(err)
	return covenantSKs, covenantPKs
}

func (h *Helper) CreateFinalityProvider(r *rand.Rand) (*btcec.PrivateKey, *btcec.PublicKey, *types.FinalityProvider) {
	fpSK, fpPK, err := datagen.GenRandomBTCKeyPair(r)
	h.NoError(err)
	fp, err := datagen.GenRandomFinalityProviderWithBTCSK(r, fpSK)
	h.NoError(err)
	msgNewFp := types.MsgCreateFinalityProvider{
		Signer:      datagen.GenRandomAccount().Address,
		Description: fp.Description,
		Commission:  fp.Commission,
		BabylonPk:   fp.BabylonPk,
		BtcPk:       fp.BtcPk,
		Pop:         fp.Pop,
	}
	_, err = h.MsgServer.CreateFinalityProvider(h.Ctx, &msgNewFp)
	h.NoError(err)
	return fpSK, fpPK, fp
}

func (h *Helper) CreateDelegation(
	r *rand.Rand,
	fpPK *btcec.PublicKey,
	changeAddress string,
	stakingValue int64,
	stakingTime uint16,
) (string, *btcec.PrivateKey, *btcec.PublicKey, *types.MsgCreateBTCDelegation) {
	delSK, delPK, err := datagen.GenRandomBTCKeyPair(r)
	h.NoError(err)
	stakingTimeBlocks := stakingTime
	bsParams := h.BTCStakingKeeper.GetParams(h.Ctx)
	covPKs, err := bbn.NewBTCPKsFromBIP340PKs(bsParams.CovenantPks)
	h.NoError(err)
	testStakingInfo := datagen.GenBTCStakingSlashingInfo(
		r,
		h.t,
		h.Net,
		delSK,
		[]*btcec.PublicKey{fpPK},
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
	h.BTCLightClientKeeper.EXPECT().GetHeaderByHash(gomock.Any(), gomock.Eq(btcHeader.Hash())).Return(&btclctypes.BTCHeaderInfo{Header: &btcHeader, Height: 10}).AnyTimes()
	h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Any()).Return(&btclctypes.BTCHeaderInfo{Height: 30}).AnyTimes()

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

	/*
		logics related to on-demand unbonding
	*/
	stkTxHash := testStakingInfo.StakingTx.TxHash()
	stkOutputIdx := uint32(0)
	defaultParams := btcctypes.DefaultParams()

	unbondingTime := uint16(defaultParams.CheckpointFinalizationTimeout) + 1
	unbondingValue := stakingValue - 1000
	testUnbondingInfo := datagen.GenBTCUnbondingSlashingInfo(
		r,
		h.t,
		h.Net,
		delSK,
		[]*btcec.PublicKey{fpPK},
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

	delSlashingTxSig, err := testUnbondingInfo.GenDelSlashingTxSig(delSK)
	h.NoError(err)

	serializedUnbondingTx, err := bbn.SerializeBTCTx(testUnbondingInfo.UnbondingTx)
	h.NoError(err)

	// all good, construct and send MsgCreateBTCDelegation message
	msgCreateBTCDel := &types.MsgCreateBTCDelegation{
		Signer:                        signer,
		BabylonPk:                     delBabylonPK.(*secp256k1.PubKey),
		BtcPk:                         stPk,
		FpBtcPkList:                   []bbn.BIP340PubKey{*bbn.NewBIP340PubKeyFromBTCPK(fpPK)},
		Pop:                           pop,
		StakingTime:                   uint32(stakingTimeBlocks),
		StakingValue:                  stakingValue,
		StakingTx:                     txInfo,
		SlashingTx:                    testStakingInfo.SlashingTx,
		DelegatorSlashingSig:          delegatorSig,
		UnbondingTx:                   serializedUnbondingTx,
		UnbondingTime:                 uint32(unbondingTime),
		UnbondingValue:                unbondingValue,
		UnbondingSlashingTx:           testUnbondingInfo.SlashingTx,
		DelegatorUnbondingSlashingSig: delSlashingTxSig,
	}

	_, err = h.MsgServer.CreateBTCDelegation(h.Ctx, msgCreateBTCDel)
	h.NoError(err)
	return stakingTxHash, delSK, delPK, msgCreateBTCDel
}

func (h *Helper) CreateCovenantSigs(
	r *rand.Rand,
	covenantSKs []*btcec.PrivateKey,
	msgCreateBTCDel *types.MsgCreateBTCDelegation,
	del *types.BTCDelegation,
) {
	stakingTx, err := bbn.NewBTCTxFromBytes(del.StakingTx)
	h.NoError(err)
	stakingTxHash := stakingTx.TxHash().String()

	bsParams := h.BTCStakingKeeper.GetParams(h.Ctx)

	vPKs, err := bbn.NewBTCPKsFromBIP340PKs(del.FpBtcPkList)
	h.NoError(err)

	stakingInfo, err := del.GetStakingInfo(&bsParams, h.Net)
	h.NoError(err)

	unbondingPathInfo, err := stakingInfo.UnbondingPathSpendInfo()
	h.NoError(err)
	slashingPathInfo, err := stakingInfo.SlashingPathSpendInfo()
	h.NoError(err)

	// generate all covenant signatures from all covenant members
	covenantSlashingTxSigs, err := datagen.GenCovenantAdaptorSigs(
		covenantSKs,
		vPKs,
		stakingTx,
		slashingPathInfo.GetPkScriptPath(),
		msgCreateBTCDel.SlashingTx,
	)
	h.NoError(err)

	/*
		Logics about on-demand unbonding
	*/

	// slash unbonding tx spends unbonding tx
	unbondingTx, err := bbn.NewBTCTxFromBytes(del.BtcUndelegation.UnbondingTx)
	h.NoError(err)
	unbondingInfo, err := del.GetUnbondingInfo(&bsParams, h.Net)
	h.NoError(err)
	unbondingSlashingPathInfo, err := unbondingInfo.SlashingPathSpendInfo()
	h.NoError(err)

	// generate all covenant signatures from all covenant members
	covenantUnbondingSlashingTxSigs, err := datagen.GenCovenantAdaptorSigs(
		covenantSKs,
		vPKs,
		unbondingTx,
		unbondingSlashingPathInfo.GetPkScriptPath(),
		del.BtcUndelegation.SlashingTx,
	)
	h.NoError(err)

	// each covenant member submits signatures
	covUnbondingSigs, err := datagen.GenCovenantUnbondingSigs(covenantSKs, stakingTx, del.StakingOutputIdx, unbondingPathInfo.GetPkScriptPath(), unbondingTx)
	h.NoError(err)

	for i := 0; i < len(bsParams.CovenantPks); i++ {
		msgAddCovenantSig := &types.MsgAddCovenantSigs{
			Signer:                  msgCreateBTCDel.Signer,
			Pk:                      covenantSlashingTxSigs[i].CovPk,
			StakingTxHash:           stakingTxHash,
			SlashingTxSigs:          covenantSlashingTxSigs[i].AdaptorSigs,
			UnbondingTxSig:          bbn.NewBIP340SignatureFromBTCSig(covUnbondingSigs[i]),
			SlashingUnbondingTxSigs: covenantUnbondingSlashingTxSigs[i].AdaptorSigs,
		}
		_, err = h.MsgServer.AddCovenantSigs(h.Ctx, msgAddCovenantSig)
		h.NoError(err)
	}

	/*
		ensure covenant sig is added successfully
	*/
	actualDelWithCovenantSigs, err := h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
	h.NoError(err)
	require.Equal(h.t, len(actualDelWithCovenantSigs.CovenantSigs), int(bsParams.CovenantQuorum)) // TODO: fix
	require.True(h.t, actualDelWithCovenantSigs.HasCovenantQuorums(h.BTCStakingKeeper.GetParams(h.Ctx).CovenantQuorum))

	require.NotNil(h.t, actualDelWithCovenantSigs.BtcUndelegation)
	require.NotNil(h.t, actualDelWithCovenantSigs.BtcUndelegation.CovenantSlashingSigs)
	require.NotNil(h.t, actualDelWithCovenantSigs.BtcUndelegation.CovenantUnbondingSigList)
	require.Len(h.t, actualDelWithCovenantSigs.BtcUndelegation.CovenantUnbondingSigList, int(bsParams.CovenantQuorum))
	require.Len(h.t, actualDelWithCovenantSigs.BtcUndelegation.CovenantSlashingSigs, int(bsParams.CovenantQuorum))
	require.Len(h.t, actualDelWithCovenantSigs.BtcUndelegation.CovenantSlashingSigs[0].AdaptorSigs, 1)

}

func (h *Helper) GetDelegationAndCheckValues(
	r *rand.Rand,
	msgCreateBTCDel *types.MsgCreateBTCDelegation,
	fpPK *btcec.PublicKey,
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
	require.False(h.t, actualDel.HasCovenantQuorums(h.BTCStakingKeeper.GetParams(h.Ctx).CovenantQuorum))
	return actualDel
}

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

func FuzzCreateBTCDelegationAndAddCovenantSigs(f *testing.F) {
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
		stakingTxHash, _, _, msgCreateBTCDel := h.CreateDelegation(
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

		//generate and insert new covenant signatures
		h.CreateCovenantSigs(r, covenantSKs, msgCreateBTCDel, actualDel)

		// ensure the BTC delegation now has voting power
		actualDel, err = h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
		h.NoError(err)
		require.True(h.t, actualDel.HasCovenantQuorums(h.BTCStakingKeeper.GetParams(h.Ctx).CovenantQuorum))
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
		stakingTxHash, delSK, _, msgCreateBTCDel := h.CreateDelegation(
			r,
			fpPK,
			changeAddress.EncodeAddress(),
			stakingValue,
			1000,
		)

		// add covenant signatures to this BTC delegation
		actualDel, err := h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
		h.NoError(err)
		h.CreateCovenantSigs(r, covenantSKs, msgCreateBTCDel, actualDel)

		// ensure the BTC delegation is bonded right now
		actualDel, err = h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
		h.NoError(err)
		btcTip := h.BTCLightClientKeeper.GetTipInfo(h.Ctx).Height
		status := actualDel.GetStatus(btcTip, wValue, bsParams.CovenantQuorum)
		require.Equal(t, types.BTCDelegationStatus_ACTIVE, status)

		// delegator wants to unbond
		delUnbondingSig, err := actualDel.SignUnbondingTx(&bsParams, h.Net, delSK)
		h.NoError(err)
		_, err = h.MsgServer.BTCUndelegate(h.Ctx, &types.MsgBTCUndelegate{
			Signer:         datagen.GenRandomAccount().Address,
			StakingTxHash:  stakingTxHash,
			UnbondingTxSig: bbn.NewBIP340SignatureFromBTCSig(delUnbondingSig),
		})
		h.NoError(err)

		// ensure the BTC delegation is unbonded
		actualDel, err = h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
		h.NoError(err)
		status = actualDel.GetStatus(btcTip, wValue, bsParams.CovenantQuorum)
		require.Equal(t, types.BTCDelegationStatus_UNBONDED, status)
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

	changeAddress, err := datagen.GenRandomBTCAddress(r, h.Net)
	require.NoError(t, err)

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
		changeAddress.EncodeAddress(),
		bsParams.SlashingRate,
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
