package datagen

import (
	"fmt"
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	"github.com/babylonchain/babylon/btcstaking"
	bbn "github.com/babylonchain/babylon/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
)

const (
	StakingOutIdx  = uint32(0)
	UnbondingTxFee = int64(1000)
)

func GenRandomBTCValidator(r *rand.Rand) (*bstypes.BTCValidator, error) {
	// key pairs
	btcSK, _, err := GenRandomBTCKeyPair(r)
	if err != nil {
		return nil, err
	}
	return GenRandomBTCValidatorWithBTCSK(r, btcSK)
}

func GenRandomBTCValidatorWithBTCSK(r *rand.Rand, btcSK *btcec.PrivateKey) (*bstypes.BTCValidator, error) {
	bbnSK, _, err := GenRandomSecp256k1KeyPair(r)
	if err != nil {
		return nil, err
	}
	return GenRandomBTCValidatorWithBTCBabylonSKs(r, btcSK, bbnSK)
}

func GenRandomBTCValidatorWithBTCBabylonSKs(r *rand.Rand, btcSK *btcec.PrivateKey, bbnSK cryptotypes.PrivKey) (*bstypes.BTCValidator, error) {
	// commission
	commission := sdkmath.LegacyNewDecWithPrec(int64(RandomInt(r, 49)+1), 2) // [1/100, 50/100]
	// description
	description := stakingtypes.Description{}
	// key pairs
	btcPK := btcSK.PubKey()
	bip340PK := bbn.NewBIP340PubKeyFromBTCPK(btcPK)
	bbnPK := bbnSK.PubKey()
	secp256k1PK, ok := bbnPK.(*secp256k1.PubKey)
	if !ok {
		return nil, fmt.Errorf("failed to assert bbnPK to *secp256k1.PubKey")
	}
	// pop
	pop, err := bstypes.NewPoP(bbnSK, btcSK)
	if err != nil {
		return nil, err
	}
	return &bstypes.BTCValidator{
		Description: &description,
		Commission:  &commission,
		BabylonPk:   secp256k1PK,
		BtcPk:       bip340PK,
		Pop:         pop,
	}, nil
}

// TODO: accomodate presign unbonding flow
func GenRandomBTCDelegation(
	r *rand.Rand,
	t *testing.T,
	valBTCPKs []bbn.BIP340PubKey,
	delSK *btcec.PrivateKey,
	covenantSKs []*btcec.PrivateKey,
	covenantQuorum uint32,
	slashingAddress, changeAddress string,
	startHeight, endHeight, totalSat uint64,
	slashingRate sdkmath.LegacyDec,
) (*bstypes.BTCDelegation, error) {
	net := &chaincfg.SimNetParams
	delPK := delSK.PubKey()
	delBTCPK := bbn.NewBIP340PubKeyFromBTCPK(delPK)
	// list of covenant PKs
	covenantBTCPKs := []*btcec.PublicKey{}
	for _, covenantSK := range covenantSKs {
		covenantBTCPKs = append(covenantBTCPKs, covenantSK.PubKey())
	}
	// list of validator PKs
	valPKs := []*btcec.PublicKey{}
	for _, valBTCPK := range valBTCPKs {
		valPK, err := valBTCPK.ToBTCPK()
		if err != nil {
			return nil, err
		}
		valPKs = append(valPKs, valPK)
	}

	// BTC delegation Babylon key pairs
	bbnSK, bbnPK, err := GenRandomSecp256k1KeyPair(r)
	if err != nil {
		return nil, err
	}
	secp256k1PK, ok := bbnPK.(*secp256k1.PubKey)
	if !ok {
		return nil, fmt.Errorf("failed to assert bbnPK to *secp256k1.PubKey")
	}
	// pop
	pop, err := bstypes.NewPoP(bbnSK, delSK)
	if err != nil {
		return nil, err
	}
	// staking/slashing tx
	stakingSlashingInfo := GenBTCStakingSlashingInfo(
		r,
		t,
		net,
		delSK,
		valPKs,
		covenantBTCPKs,
		covenantQuorum,
		uint16(endHeight-startHeight),
		int64(totalSat),
		slashingAddress, changeAddress,
		slashingRate,
	)

	slashingPathSpendInfo, err := stakingSlashingInfo.StakingInfo.SlashingPathSpendInfo()
	require.NoError(t, err)

	stakingMsgTx := stakingSlashingInfo.StakingTx

	// delegator sig
	delegatorSig, err := stakingSlashingInfo.SlashingTx.Sign(
		stakingMsgTx,
		StakingOutIdx,
		slashingPathSpendInfo.GetPkScriptPath(),
		delSK,
	)
	require.NoError(t, err)

	// covenant sigs
	covenantSigs, err := GenCovenantAdaptorSigs(
		covenantSKs,
		valPKs,
		stakingMsgTx,
		slashingPathSpendInfo.GetPkScriptPath(),
		stakingSlashingInfo.SlashingTx,
	)
	require.NoError(t, err)

	serializedStakingTx, err := bbn.SerializeBTCTx(stakingSlashingInfo.StakingTx)
	require.NoError(t, err)

	del := &bstypes.BTCDelegation{
		BabylonPk:        secp256k1PK,
		BtcPk:            delBTCPK,
		Pop:              pop,
		ValBtcPkList:     valBTCPKs,
		StartHeight:      startHeight,
		EndHeight:        endHeight,
		TotalSat:         totalSat,
		StakingOutputIdx: StakingOutIdx,
		DelegatorSig:     delegatorSig,
		CovenantSigs:     covenantSigs,
		StakingTx:        serializedStakingTx,
		SlashingTx:       stakingSlashingInfo.SlashingTx,
	}

	/*
		construct BTC undelegation
	*/

	// construct unbonding info
	stkTxHash := stakingSlashingInfo.StakingTx.TxHash()
	unbondingValue := totalSat - uint64(UnbondingTxFee)
	w := uint16(100) // TODO: parameterise w
	unbondingSlashingInfo := GenBTCUnbondingSlashingInfo(
		r,
		t,
		net,
		delSK,
		valPKs,
		covenantBTCPKs,
		covenantQuorum,
		wire.NewOutPoint(&stkTxHash, StakingOutIdx),
		w+1,
		int64(unbondingValue),
		slashingAddress, changeAddress,
		slashingRate,
	)

	unbondingTxBytes, err := bbn.SerializeBTCTx(unbondingSlashingInfo.UnbondingTx)
	require.NoError(t, err)
	delSlashingTxSig, err := unbondingSlashingInfo.GenDelSlashingTxSig(delSK)
	require.NoError(t, err)
	del.BtcUndelegation = &bstypes.BTCUndelegation{
		UnbondingTx:          unbondingTxBytes,
		UnbondingTime:        uint32(w + 1),
		SlashingTx:           unbondingSlashingInfo.SlashingTx,
		DelegatorSlashingSig: delSlashingTxSig,
	}

	/*
		covenant signs BTC undelegation
	*/

	unbondingPathSpendInfo, err := stakingSlashingInfo.StakingInfo.UnbondingPathSpendInfo()
	require.NoError(t, err)

	covUnbondingSlashingSigs, covUnbondingSigs, err := unbondingSlashingInfo.GenCovenantSigs(
		covenantSKs,
		valPKs,
		stakingMsgTx,
		unbondingPathSpendInfo.GetPkScriptPath(),
	)
	require.NoError(t, err)

	del.BtcUndelegation.CovenantSlashingSigs = covUnbondingSlashingSigs
	del.BtcUndelegation.CovenantUnbondingSigList = covUnbondingSigs

	return del, nil
}

type TestStakingSlashingInfo struct {
	StakingTx   *wire.MsgTx
	SlashingTx  *bstypes.BTCSlashingTx
	StakingInfo *btcstaking.StakingInfo
}

type TestUnbondingSlashingInfo struct {
	UnbondingTx   *wire.MsgTx
	SlashingTx    *bstypes.BTCSlashingTx
	UnbondingInfo *btcstaking.UnbondingInfo
}

func GenBTCStakingSlashingInfoWithOutPoint(
	r *rand.Rand,
	t *testing.T,
	btcNet *chaincfg.Params,
	outPoint *wire.OutPoint,
	stakerSK *btcec.PrivateKey,
	validatorPKs []*btcec.PublicKey,
	covenantPKs []*btcec.PublicKey,
	covenantQuorum uint32,
	stakingTimeBlocks uint16,
	stakingValue int64,
	slashingAddress, changeAddress string,
	slashingRate sdkmath.LegacyDec,
) *TestStakingSlashingInfo {

	stakingInfo, err := btcstaking.BuildStakingInfo(
		stakerSK.PubKey(),
		validatorPKs,
		covenantPKs,
		covenantQuorum,
		stakingTimeBlocks,
		btcutil.Amount(stakingValue),
		btcNet,
	)

	require.NoError(t, err)
	tx := wire.NewMsgTx(2)
	// add the given tx input
	txIn := wire.NewTxIn(outPoint, nil, nil)
	tx.AddTxIn(txIn)
	tx.AddTxOut(stakingInfo.StakingOutput)

	// 2 outputs for changes and staking output
	changeAddrScript, err := GenRandomPubKeyHashScript(r, btcNet)
	require.NoError(t, err)
	require.False(t, txscript.GetScriptClass(changeAddrScript) == txscript.NonStandardTy)

	tx.AddTxOut(wire.NewTxOut(10000, changeAddrScript)) // output for change

	// construct slashing tx
	slashingAddrBtc, err := btcutil.DecodeAddress(slashingAddress, btcNet)
	require.NoError(t, err)
	changeAddrBtc, err := btcutil.DecodeAddress(changeAddress, btcNet)
	require.NoError(t, err)
	slashingMsgTx, err := btcstaking.BuildSlashingTxFromStakingTxStrict(
		tx,
		StakingOutIdx,
		slashingAddrBtc, changeAddrBtc,
		2000,
		slashingRate,
		btcNet)
	require.NoError(t, err)
	slashingTx, err := bstypes.NewBTCSlashingTxFromMsgTx(slashingMsgTx)
	require.NoError(t, err)

	return &TestStakingSlashingInfo{
		StakingTx:   tx,
		SlashingTx:  slashingTx,
		StakingInfo: stakingInfo,
	}
}

func GenBTCStakingSlashingInfo(
	r *rand.Rand,
	t *testing.T,
	btcNet *chaincfg.Params,
	stakerSK *btcec.PrivateKey,
	validatorPKs []*btcec.PublicKey,
	covenantPKs []*btcec.PublicKey,
	covenantQuorum uint32,
	stakingTimeBlocks uint16,
	stakingValue int64,
	slashingAddress, changeAddress string,
	slashingRate sdkmath.LegacyDec,
) *TestStakingSlashingInfo {
	// an arbitrary input
	spend := makeSpendableOutWithRandOutPoint(r, btcutil.Amount(stakingValue+UnbondingTxFee))
	outPoint := &spend.prevOut
	return GenBTCStakingSlashingInfoWithOutPoint(
		r,
		t,
		btcNet,
		outPoint,
		stakerSK,
		validatorPKs,
		covenantPKs,
		covenantQuorum,
		stakingTimeBlocks,
		stakingValue,
		slashingAddress, changeAddress,
		slashingRate)
}

func GenBTCUnbondingSlashingInfo(
	r *rand.Rand,
	t *testing.T,
	btcNet *chaincfg.Params,
	stakerSK *btcec.PrivateKey,
	validatorPKs []*btcec.PublicKey,
	covenantPKs []*btcec.PublicKey,
	covenantQuorum uint32,
	stakingTransactionOutpoint *wire.OutPoint,
	stakingTimeBlocks uint16,
	stakingValue int64,
	slashingAddress, changeAddress string,
	slashingRate sdkmath.LegacyDec,
) *TestUnbondingSlashingInfo {

	unbondingInfo, err := btcstaking.BuildUnbondingInfo(
		stakerSK.PubKey(),
		validatorPKs,
		covenantPKs,
		covenantQuorum,
		stakingTimeBlocks,
		btcutil.Amount(stakingValue),
		btcNet,
	)

	require.NoError(t, err)
	tx := wire.NewMsgTx(2)
	// add the given tx input
	txIn := wire.NewTxIn(stakingTransactionOutpoint, nil, nil)
	tx.AddTxIn(txIn)
	tx.AddTxOut(unbondingInfo.UnbondingOutput)

	// construct slashing tx
	slashingAddrBtc, err := btcutil.DecodeAddress(slashingAddress, btcNet)
	require.NoError(t, err)
	changeAddrBtc, err := btcutil.DecodeAddress(changeAddress, btcNet)
	require.NoError(t, err)
	slashingMsgTx, err := btcstaking.BuildSlashingTxFromStakingTxStrict(
		tx,
		StakingOutIdx,
		slashingAddrBtc, changeAddrBtc,
		2000,
		slashingRate,
		btcNet)
	require.NoError(t, err)
	slashingTx, err := bstypes.NewBTCSlashingTxFromMsgTx(slashingMsgTx)
	require.NoError(t, err)

	return &TestUnbondingSlashingInfo{
		UnbondingTx:   tx,
		SlashingTx:    slashingTx,
		UnbondingInfo: unbondingInfo,
	}
}

func (info *TestUnbondingSlashingInfo) GenDelSlashingTxSig(sk *btcec.PrivateKey) (*bbn.BIP340Signature, error) {
	unbondingTxMsg := info.UnbondingTx
	unbondingTxSlashingPathInfo, err := info.UnbondingInfo.SlashingPathSpendInfo()
	if err != nil {
		return nil, err
	}
	slashingTxSig, err := info.SlashingTx.Sign(
		unbondingTxMsg,
		StakingOutIdx,
		unbondingTxSlashingPathInfo.GetPkScriptPath(),
		sk,
	)
	if err != nil {
		return nil, err
	}
	return slashingTxSig, nil
}

func (info *TestUnbondingSlashingInfo) GenCovenantSigs(
	covSKs []*btcec.PrivateKey,
	valPKs []*btcec.PublicKey,
	stakingTx *wire.MsgTx,
	unbondingPkScriptPath []byte,
) ([]*bstypes.CovenantAdaptorSignatures, []*bstypes.SignatureInfo, error) {
	unbondingSlashingPathInfo, err := info.UnbondingInfo.SlashingPathSpendInfo()
	if err != nil {
		return nil, nil, err
	}

	covUnbondingSlashingSigs, err := GenCovenantAdaptorSigs(
		covSKs,
		valPKs,
		info.UnbondingTx,
		unbondingSlashingPathInfo.GetPkScriptPath(),
		info.SlashingTx,
	)
	if err != nil {
		return nil, nil, err
	}
	covUnbondingSigs, err := GenCovenantUnbondingSigs(
		covSKs,
		stakingTx,
		StakingOutIdx,
		unbondingPkScriptPath,
		info.UnbondingTx,
	)
	if err != nil {
		return nil, nil, err
	}
	covUnbondingSigList := []*bstypes.SignatureInfo{}
	for i := range covUnbondingSigs {
		covUnbondingSigList = append(covUnbondingSigList, &bstypes.SignatureInfo{
			Pk:  bbn.NewBIP340PubKeyFromBTCPK(covSKs[i].PubKey()),
			Sig: bbn.NewBIP340SignatureFromBTCSig(covUnbondingSigs[i]),
		})
	}
	return covUnbondingSlashingSigs, covUnbondingSigList, nil
}
