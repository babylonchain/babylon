package keeper_test

import (
	"math/rand"
	"testing"

	"cosmossdk.io/core/header"
	sdkmath "cosmossdk.io/math"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

var (
	net = &chaincfg.SimNetParams
)

type Helper struct {
	t testing.TB

	Ctx                  sdk.Context
	BTCStakingKeeper     *keeper.Keeper
	BTCLightClientKeeper *types.MockBTCLightClientKeeper
	BTCCheckpointKeeper  *types.MockBtcCheckpointKeeper
	CheckpointingKeeper  *types.MockCheckpointingKeeper
	BTCStakingHooks      *types.MockBtcStakingHooks
	MsgServer            types.MsgServer
	Net                  *chaincfg.Params
}

func NewHelper(t testing.TB, btclcKeeper *types.MockBTCLightClientKeeper, btccKeeper *types.MockBtcCheckpointKeeper, ckptKeeper *types.MockCheckpointingKeeper) *Helper {
	k, ctx := keepertest.BTCStakingKeeper(t, btclcKeeper, btccKeeper, ckptKeeper)
	ctx = ctx.WithHeaderInfo(header.Info{Height: 1})
	msgSrvr := keeper.NewMsgServerImpl(*k)

	ctrl := gomock.NewController(t)
	mockedHooks := types.NewMockBtcStakingHooks(ctrl)
	mockedHooks.EXPECT().AfterFinalityProviderActivated(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	k.SetHooks(mockedHooks)

	return &Helper{
		t:                    t,
		Ctx:                  ctx,
		BTCStakingKeeper:     k,
		BTCLightClientKeeper: btclcKeeper,
		BTCCheckpointKeeper:  btccKeeper,
		CheckpointingKeeper:  ckptKeeper,
		MsgServer:            msgSrvr,
		Net:                  &chaincfg.SimNetParams,
	}
}

func (h *Helper) NoError(err error) {
	require.NoError(h.t, err)
}

func (h *Helper) Error(err error) {
	require.Error(h.t, err)
}

func (h *Helper) GenAndApplyParams(r *rand.Rand) ([]*btcec.PrivateKey, []*btcec.PublicKey) {
	return h.GenAndApplyCustomParams(r, 100, 0)
}

func (h *Helper) SetCtxHeight(height uint64) {
	h.Ctx = datagen.WithCtxHeight(h.Ctx, height)
}

func (h *Helper) GenAndApplyCustomParams(
	r *rand.Rand,
	finalizationTimeout uint64,
	minUnbondingTime uint32,
) ([]*btcec.PrivateKey, []*btcec.PublicKey) {
	// mock base header
	baseHeader := btclctypes.SimnetGenesisBlock()
	h.BTCLightClientKeeper.EXPECT().GetBaseBTCHeader(gomock.Any()).Return(&baseHeader).AnyTimes()

	// mocking stuff for BTC checkpoint keeper
	h.BTCCheckpointKeeper.EXPECT().GetPowLimit().Return(h.Net.PowLimit).AnyTimes()

	params := btcctypes.DefaultParams()
	params.CheckpointFinalizationTimeout = finalizationTimeout

	h.BTCCheckpointKeeper.EXPECT().GetParams(gomock.Any()).Return(params).AnyTimes()

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
		MinUnbondingTime:           minUnbondingTime,
		MinUnbondingRate:           sdkmath.LegacyMustNewDecFromStr("0.8"),
	})
	h.NoError(err)
	return covenantSKs, covenantPKs
}

func CreateFinalityProvider(r *rand.Rand, t *testing.T) *types.FinalityProvider {
	fpSK, _, err := datagen.GenRandomBTCKeyPair(r)
	require.NoError(t, err)
	fp, err := datagen.GenRandomFinalityProviderWithBTCSK(r, fpSK)
	require.NoError(t, err)

	return &types.FinalityProvider{
		Description: fp.Description,
		Commission:  fp.Commission,
		Addr:        fp.Addr,
		BtcPk:       fp.BtcPk,
		Pop:         fp.Pop,
	}
}

func (h *Helper) CreateFinalityProvider(r *rand.Rand) (*btcec.PrivateKey, *btcec.PublicKey, *types.FinalityProvider) {
	fpSK, fpPK, err := datagen.GenRandomBTCKeyPair(r)
	h.NoError(err)
	fp, err := datagen.GenRandomFinalityProviderWithBTCSK(r, fpSK)
	h.NoError(err)
	msgNewFp := types.MsgCreateFinalityProvider{
		Addr:        fp.Addr,
		Description: fp.Description,
		Commission:  fp.Commission,
		BtcPk:       fp.BtcPk,
		Pop:         fp.Pop,
	}

	_, err = h.MsgServer.CreateFinalityProvider(h.Ctx, &msgNewFp)
	h.NoError(err)
	return fpSK, fpPK, fp
}

func (h *Helper) CreateDelegationCustom(
	r *rand.Rand,
	fpPK *btcec.PublicKey,
	changeAddress string,
	stakingValue int64,
	stakingTime uint16,
	unbondingValue int64,
	unbondingTime uint16,
) (string, *btcec.PrivateKey, *btcec.PublicKey, *types.MsgCreateBTCDelegation, error) {
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
		bsParams.SlashingRate,
		unbondingTime,
	)
	h.NoError(err)
	stakingTxHash := testStakingInfo.StakingTx.TxHash().String()

	// random signer
	staker := sdk.MustAccAddressFromBech32(datagen.GenRandomAccount().Address)

	// PoP
	pop, err := types.NewPoPBTC(staker, delSK)
	h.NoError(err)
	// generate staking tx info
	prevBlock, _ := datagen.GenRandomBtcdBlock(r, 0, nil)
	btcHeaderWithProof := datagen.CreateBlockWithTransaction(r, &prevBlock.Header, testStakingInfo.StakingTx)
	btcHeader := btcHeaderWithProof.HeaderBytes
	serializedStakingTx, err := bbn.SerializeBTCTx(testStakingInfo.StakingTx)
	h.NoError(err)

	txInfo := btcctypes.NewTransactionInfo(&btcctypes.TransactionKey{Index: 1, Hash: btcHeader.Hash()}, serializedStakingTx, btcHeaderWithProof.SpvProof.MerkleNodes)

	// mock for testing k-deep stuff
	h.BTCLightClientKeeper.EXPECT().GetHeaderByHash(gomock.Eq(h.Ctx), gomock.Eq(btcHeader.Hash())).Return(&btclctypes.BTCHeaderInfo{Header: &btcHeader, Height: 10}).AnyTimes()
	h.BTCLightClientKeeper.EXPECT().GetTipInfo(gomock.Eq(h.Ctx)).Return(&btclctypes.BTCHeaderInfo{Height: 30}).AnyTimes()

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
		bsParams.SlashingRate,
		unbondingTime,
	)
	h.NoError(err)

	delSlashingTxSig, err := testUnbondingInfo.GenDelSlashingTxSig(delSK)
	h.NoError(err)

	serializedUnbondingTx, err := bbn.SerializeBTCTx(testUnbondingInfo.UnbondingTx)
	h.NoError(err)

	// all good, construct and send MsgCreateBTCDelegation message
	fpBTCPK := bbn.NewBIP340PubKeyFromBTCPK(fpPK)
	msgCreateBTCDel := &types.MsgCreateBTCDelegation{
		StakerAddr:                    staker.String(),
		BtcPk:                         stPk,
		FpBtcPkList:                   []bbn.BIP340PubKey{*fpBTCPK},
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
	if err != nil {
		return "", nil, nil, nil, err
	}

	return stakingTxHash, delSK, delPK, msgCreateBTCDel, nil
}

func (h *Helper) CreateDelegation(
	r *rand.Rand,
	fpPK *btcec.PublicKey,
	changeAddress string,
	stakingValue int64,
	stakingTime uint16,
) (string, *btcec.PrivateKey, *btcec.PublicKey, *types.MsgCreateBTCDelegation, *types.BTCDelegation) {
	bsParams := h.BTCStakingKeeper.GetParams(h.Ctx)
	bcParams := h.BTCCheckpointKeeper.GetParams(h.Ctx)

	minUnbondingTime := types.MinimumUnbondingTime(
		bsParams,
		bcParams,
	)

	stakingTxHash, delSK, delPK, msgCreateBTCDel, err := h.CreateDelegationCustom(
		r,
		fpPK,
		changeAddress,
		stakingValue,
		stakingTime,
		stakingValue-1000,
		uint16(minUnbondingTime)+1,
	)

	h.NoError(err)

	stakingMsgTx, err := bbn.NewBTCTxFromBytes(msgCreateBTCDel.StakingTx.Transaction)
	h.NoError(err)
	btcDel, err := h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingMsgTx.TxHash().String())
	h.NoError(err)

	return stakingTxHash, delSK, delPK, msgCreateBTCDel, btcDel
}

func (h *Helper) GenerateCovenantSignaturesMessages(
	r *rand.Rand,
	covenantSKs []*btcec.PrivateKey,
	msgCreateBTCDel *types.MsgCreateBTCDelegation,
	del *types.BTCDelegation,
) []*types.MsgAddCovenantSigs {
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

	msgs := make([]*types.MsgAddCovenantSigs, len(bsParams.CovenantPks))

	for i := 0; i < len(bsParams.CovenantPks); i++ {
		msgAddCovenantSig := &types.MsgAddCovenantSigs{
			Signer:                  msgCreateBTCDel.StakerAddr,
			Pk:                      covenantSlashingTxSigs[i].CovPk,
			StakingTxHash:           stakingTxHash,
			SlashingTxSigs:          covenantSlashingTxSigs[i].AdaptorSigs,
			UnbondingTxSig:          bbn.NewBIP340SignatureFromBTCSig(covUnbondingSigs[i]),
			SlashingUnbondingTxSigs: covenantUnbondingSlashingTxSigs[i].AdaptorSigs,
		}
		msgs[i] = msgAddCovenantSig
	}
	return msgs
}

func (h *Helper) CreateCovenantSigs(
	r *rand.Rand,
	covenantSKs []*btcec.PrivateKey,
	msgCreateBTCDel *types.MsgCreateBTCDelegation,
	del *types.BTCDelegation,
) {
	stakingTx, err := bbn.NewBTCTxFromBytes(del.StakingTx)
	stakingTxHash := stakingTx.TxHash().String()

	bsParams := h.BTCStakingKeeper.GetParams(h.Ctx)

	h.NoError(err)
	covenantMsgs := h.GenerateCovenantSignaturesMessages(r, covenantSKs, msgCreateBTCDel, del)
	for _, msg := range covenantMsgs {
		msgCopy := msg
		_, err := h.MsgServer.AddCovenantSigs(h.Ctx, msgCopy)
		h.NoError(err)
	}
	/*
		ensure covenant sig is added successfully
	*/
	actualDelWithCovenantSigs, err := h.BTCStakingKeeper.GetBTCDelegation(h.Ctx, stakingTxHash)
	h.NoError(err)
	require.Equal(h.t, len(actualDelWithCovenantSigs.CovenantSigs), int(bsParams.CovenantQuorum))
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
	// TODO: update pop in BTC delegation
	require.Equal(h.t, msgCreateBTCDel.StakerAddr, actualDel.StakerAddr)
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
