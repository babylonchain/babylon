package btcstaking

import (
	"encoding/hex"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"

	bbn "github.com/babylonchain/babylon/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// Point with unknown discrete logarithm defined in: https://github.com/bitcoin/bips/blob/master/bip-0341.mediawiki#constructing-and-spending-taproot-outputs
	// using it as internal public key effectively disables taproot key spends
	unspendableKeyPath = "0250929b74c1a04954b78b4b6035e97a5e078a5a0f28ec96d547bfee9ace803ac0"
)

var (
	unspendableKeyPathKey = unspendableKeyPathInternalPubKeyInternal(unspendableKeyPath)
)

func unspendableKeyPathInternalPubKeyInternal(keyHex string) btcec.PublicKey {
	keyBytes, err := hex.DecodeString(keyHex)

	if err != nil {
		panic(fmt.Sprintf("unexpected error: %v", err))
	}

	// We are using btcec here, as key is 33 byte compressed format.
	pubKey, err := btcec.ParsePubKey(keyBytes)

	if err != nil {
		panic(fmt.Sprintf("unexpected error: %v", err))
	}
	return *pubKey
}

func unspendableKeyPathInternalPubKey() btcec.PublicKey {
	return unspendableKeyPathKey
}

func NewTaprootTreeFromScripts(
	scripts [][]byte,
) *txscript.IndexedTapScriptTree {
	var tapLeafs []txscript.TapLeaf
	for _, script := range scripts {
		scr := script
		tapLeafs = append(tapLeafs, txscript.NewBaseTapLeaf(scr))
	}
	return txscript.AssembleTaprootScriptTree(tapLeafs...)
}

func DeriveTaprootAddress(
	tapScriptTree *txscript.IndexedTapScriptTree,
	internalPubKey *btcec.PublicKey,
	net *chaincfg.Params) (*btcutil.AddressTaproot, error) {

	tapScriptRootHash := tapScriptTree.RootNode.TapHash()

	outputKey := txscript.ComputeTaprootOutputKey(
		internalPubKey, tapScriptRootHash[:],
	)

	address, err := btcutil.NewAddressTaproot(
		schnorr.SerializePubKey(outputKey), net)

	if err != nil {
		return nil, fmt.Errorf("error encoding Taproot address: %v", err)
	}

	return address, nil
}

func DeriveTaprootPkScript(
	tapScriptTree *txscript.IndexedTapScriptTree,
	internalPubKey *btcec.PublicKey,
	net *chaincfg.Params,
) ([]byte, error) {
	taprootAddress, err := DeriveTaprootAddress(
		tapScriptTree,
		&unspendableKeyPathKey,
		net,
	)

	if err != nil {
		return nil, err
	}

	taprootPkScript, err := txscript.PayToAddrScript(taprootAddress)

	if err != nil {
		return nil, err
	}

	return taprootPkScript, nil
}

type taprootScriptHolder struct {
	internalPubKey *btcec.PublicKey
	scriptTree     *txscript.IndexedTapScriptTree
}

func newTaprootScriptHolder(
	internalPubKey *btcec.PublicKey,
	scripts [][]byte,
) (*taprootScriptHolder, error) {
	if internalPubKey == nil {
		return nil, fmt.Errorf("internal public key is nil")
	}

	if len(scripts) == 0 {
		return &taprootScriptHolder{
			scriptTree: txscript.NewIndexedTapScriptTree(0),
		}, nil
	}

	createdLeafs := make(map[chainhash.Hash]bool)
	tapLeafs := make([]txscript.TapLeaf, len(scripts))

	for i, s := range scripts {
		script := s
		if len(script) == 0 {
			return nil, fmt.Errorf("cannot build tree with empty script")
		}

		tapLeaf := txscript.NewBaseTapLeaf(script)
		leafHash := tapLeaf.TapHash()

		if _, ok := createdLeafs[leafHash]; ok {
			return nil, fmt.Errorf("duplicate script in provided scripts")
		}

		createdLeafs[leafHash] = true
		tapLeafs[i] = tapLeaf
	}

	scriptTree := txscript.AssembleTaprootScriptTree(tapLeafs...)

	return &taprootScriptHolder{
		internalPubKey: internalPubKey,
		scriptTree:     scriptTree,
	}, nil
}

func (t *taprootScriptHolder) scriptSpendInfoByName(
	leafHash chainhash.Hash,
) (*SpendInfo, error) {
	scriptIdx, ok := t.scriptTree.LeafProofIndex[leafHash]

	if !ok {
		return nil, fmt.Errorf("script not found in script tree")
	}

	merkleProof := t.scriptTree.LeafMerkleProofs[scriptIdx]

	return &SpendInfo{
		ControlBlock: merkleProof.ToControlBlock(t.internalPubKey),
		RevealedLeaf: merkleProof.TapLeaf,
	}, nil
}

func (t *taprootScriptHolder) taprootPkScript(net *chaincfg.Params) ([]byte, error) {
	return DeriveTaprootPkScript(
		t.scriptTree,
		t.internalPubKey,
		net,
	)
}

// Package responsible for different kinds of btc scripts used by babylon
// Staking script has 3 spending paths:
// 1. Staker can spend after relative time lock - staking
// 2. Staker can spend with covenat cooperation any time
// 3. Staker can spend with finality provider and covenant cooperation any time.
type StakingInfo struct {
	StakingOutput         *wire.TxOut
	scriptHolder          *taprootScriptHolder
	timeLockPathLeafHash  chainhash.Hash
	unbondingPathLeafHash chainhash.Hash
	slashingPathLeafHash  chainhash.Hash
}

// GetPkScript returns the full staking taproot pkscript in the corresponding staking tx
func (sti *StakingInfo) GetPkScript() []byte {
	return sti.StakingOutput.PkScript
}

// GetOutputFetcher returns the fetcher of the staking tx's output
func (sti *StakingInfo) GetOutputFetcher() *txscript.CannedPrevOutputFetcher {
	return txscript.NewCannedPrevOutputFetcher(
		sti.GetPkScript(), sti.StakingOutput.Value,
	)
}

// SpendInfo contains information necessary to create witness for given script
type SpendInfo struct {
	// Control block contains merkle proof of inclusion of revealed script path
	ControlBlock txscript.ControlBlock
	// RevealedLeaf is the leaf of the script tree which is revealed i.e scriptpath
	// which is being executed
	RevealedLeaf txscript.TapLeaf
}

// GetPkScriptPath returns the path of the taproot pkscript corresponding
// to the triggered spending condition of the tx associated with the SpendInfo
func (si *SpendInfo) GetPkScriptPath() []byte {
	return si.RevealedLeaf.Script
}

func SpendInfoFromRevealedScript(
	revealedScript []byte,
	internalKey *btcec.PublicKey,
	tree *txscript.IndexedTapScriptTree) (*SpendInfo, error) {

	revealedLeaf := txscript.NewBaseTapLeaf(revealedScript)
	leafHash := revealedLeaf.TapHash()

	scriptIdx, ok := tree.LeafProofIndex[leafHash]

	if !ok {
		return nil, fmt.Errorf("script not found in script tree")
	}

	merkleProof := tree.LeafMerkleProofs[scriptIdx]

	return &SpendInfo{
		ControlBlock: merkleProof.ToControlBlock(internalKey),
		RevealedLeaf: revealedLeaf,
	}, nil
}

func aggregateScripts(scripts ...[]byte) []byte {
	if len(scripts) == 0 {
		return []byte{}
	}

	var finalScript []byte

	for _, script := range scripts {
		finalScript = append(finalScript, script...)
	}
	return finalScript
}

// babylonScriptPaths contains all possible babylon script paths
// not every babylon output will contain all of those paths
type babylonScriptPaths struct {
	// timeLockPathScript is the script path for normal unbonding
	// <Staker_PK> OP_CHECKSIGVERIFY  <Staking_Time_Blocks> OP_CHECKSEQUENCEVERIFY
	timeLockPathScript []byte
	// unbondingPathScript is the script path for on-demand early unbonding
	// <Staker_PK> OP_CHECKSIGVERIFY
	// <Covenant_PK1> OP_CHECKSIG ... <Covenant_PKN> OP_CHECKSIGADD M OP_GREATERTHANOREQUAL OP_VERIFY
	unbondingPathScript []byte
	// slashingPathScript is the script path for slashing
	// <Staker_PK> OP_CHECKSIGVERIFY
	// <Covenant_PK1> OP_CHECKSIG ... <Covenant_PKN> OP_CHECKSIGADD M OP_GREATERTHANOREQUAL OP_VERIFY
	// <FP_PK1> OP_CHECKSIG ... <FP_PKN> OP_CHECKSIGADD 1 OP_GREATERTHANOREQUAL OP_VERIFY
	slashingPathScript []byte
}

func newBabylonScriptPaths(
	stakerKey *btcec.PublicKey,
	fpKeys []*btcec.PublicKey,
	covenantKeys []*btcec.PublicKey,
	covenantQuorum uint32,
	lockTime uint16,
) (*babylonScriptPaths, error) {
	if stakerKey == nil {
		return nil, fmt.Errorf("staker key is nil")
	}

	timeLockPathScript, err := buildTimeLockScript(stakerKey, lockTime)

	if err != nil {
		return nil, err
	}

	covenantMultisigScript, err := buildMultiSigScript(
		covenantKeys,
		covenantQuorum,
		// covenant multisig is always last in script so we do not run verify and leave
		// last value on the stack. If we do not leave at least one element on the stack
		// script will always error
		false,
	)

	if err != nil {
		return nil, err
	}

	stakerSigScript, err := buildSingleKeySigScript(stakerKey, true)

	if err != nil {
		return nil, err
	}

	fpMultisigScript, err := buildMultiSigScript(
		fpKeys,
		// we always require only one finality provider to sign
		1,
		// we need to run verify to clear the stack, as finality provider multisig is in the middle of the script
		true,
	)

	if err != nil {
		return nil, err
	}

	unbondingPathScript := aggregateScripts(
		stakerSigScript,
		covenantMultisigScript,
	)

	slashingPathScript := aggregateScripts(
		stakerSigScript,
		fpMultisigScript,
		covenantMultisigScript,
	)

	return &babylonScriptPaths{
		timeLockPathScript:  timeLockPathScript,
		unbondingPathScript: unbondingPathScript,
		slashingPathScript:  slashingPathScript,
	}, nil
}

func BuildStakingInfo(
	stakerKey *btcec.PublicKey,
	fpKeys []*btcec.PublicKey,
	covenantKeys []*btcec.PublicKey,
	covenantQuorum uint32,
	stakingTime uint16,
	stakingAmount btcutil.Amount,
	net *chaincfg.Params,
) (*StakingInfo, error) {
	unspendableKeyPathKey := unspendableKeyPathInternalPubKey()

	babylonScripts, err := newBabylonScriptPaths(
		stakerKey,
		fpKeys,
		covenantKeys,
		covenantQuorum,
		stakingTime,
	)

	if err != nil {
		return nil, err
	}

	var unbondingPaths [][]byte
	unbondingPaths = append(unbondingPaths, babylonScripts.timeLockPathScript)
	unbondingPaths = append(unbondingPaths, babylonScripts.unbondingPathScript)
	unbondingPaths = append(unbondingPaths, babylonScripts.slashingPathScript)

	timeLockLeafHash := txscript.NewBaseTapLeaf(babylonScripts.timeLockPathScript).TapHash()
	unbondingPathLeafHash := txscript.NewBaseTapLeaf(babylonScripts.unbondingPathScript).TapHash()
	slashingLeafHash := txscript.NewBaseTapLeaf(babylonScripts.slashingPathScript).TapHash()

	sh, err := newTaprootScriptHolder(
		&unspendableKeyPathKey,
		unbondingPaths,
	)

	if err != nil {
		return nil, err
	}

	taprootPkScript, err := sh.taprootPkScript(net)

	if err != nil {
		return nil, err
	}

	stakingOutput := wire.NewTxOut(int64(stakingAmount), taprootPkScript)

	return &StakingInfo{
		StakingOutput:         stakingOutput,
		scriptHolder:          sh,
		timeLockPathLeafHash:  timeLockLeafHash,
		unbondingPathLeafHash: unbondingPathLeafHash,
		slashingPathLeafHash:  slashingLeafHash,
	}, nil
}

func (i *StakingInfo) TimeLockPathSpendInfo() (*SpendInfo, error) {
	return i.scriptHolder.scriptSpendInfoByName(i.timeLockPathLeafHash)
}

func (i *StakingInfo) UnbondingPathSpendInfo() (*SpendInfo, error) {
	return i.scriptHolder.scriptSpendInfoByName(i.unbondingPathLeafHash)
}

func (i *StakingInfo) SlashingPathSpendInfo() (*SpendInfo, error) {
	return i.scriptHolder.scriptSpendInfoByName(i.slashingPathLeafHash)
}

// Unbonding script has 2 spending paths:
// 1. Staker can spend after relative time lock - staking
// 2. Staker can spend with finality provider and covenant cooperation any time.
type UnbondingInfo struct {
	UnbondingOutput      *wire.TxOut
	scriptHolder         *taprootScriptHolder
	timeLockPathLeafHash chainhash.Hash
	slashingPathLeafHash chainhash.Hash
}

func BuildUnbondingInfo(
	stakerKey *btcec.PublicKey,
	fpKeys []*btcec.PublicKey,
	covenantKeys []*btcec.PublicKey,
	covenantQuorum uint32,
	unbondingTime uint16,
	unbondingAmount btcutil.Amount,
	net *chaincfg.Params,
) (*UnbondingInfo, error) {
	unspendableKeyPathKey := unspendableKeyPathInternalPubKey()

	babylonScripts, err := newBabylonScriptPaths(
		stakerKey,
		fpKeys,
		covenantKeys,
		covenantQuorum,
		unbondingTime,
	)

	if err != nil {
		return nil, err
	}

	var unbondingPaths [][]byte
	unbondingPaths = append(unbondingPaths, babylonScripts.timeLockPathScript)
	unbondingPaths = append(unbondingPaths, babylonScripts.slashingPathScript)

	timeLockLeafHash := txscript.NewBaseTapLeaf(babylonScripts.timeLockPathScript).TapHash()
	slashingLeafHash := txscript.NewBaseTapLeaf(babylonScripts.slashingPathScript).TapHash()

	sh, err := newTaprootScriptHolder(
		&unspendableKeyPathKey,
		unbondingPaths,
	)

	if err != nil {
		return nil, err
	}

	taprootPkScript, err := sh.taprootPkScript(net)

	if err != nil {
		return nil, err
	}

	unbondingOutput := wire.NewTxOut(int64(unbondingAmount), taprootPkScript)

	return &UnbondingInfo{
		UnbondingOutput:      unbondingOutput,
		scriptHolder:         sh,
		timeLockPathLeafHash: timeLockLeafHash,
		slashingPathLeafHash: slashingLeafHash,
	}, nil
}

func (i *UnbondingInfo) TimeLockPathSpendInfo() (*SpendInfo, error) {
	return i.scriptHolder.scriptSpendInfoByName(i.timeLockPathLeafHash)
}

func (i *UnbondingInfo) SlashingPathSpendInfo() (*SpendInfo, error) {
	return i.scriptHolder.scriptSpendInfoByName(i.slashingPathLeafHash)
}

// IsRateValid checks if the given rate is between the valid range i.e., (0,1) with a precision of at most 2 decimal places.
func IsRateValid(rate sdkmath.LegacyDec) bool {
	// Check if the slashing rate is between 0 and 1
	if rate.LTE(sdkmath.LegacyZeroDec()) || rate.GTE(sdkmath.LegacyOneDec()) {
		return false
	}

	// Multiply by 100 to move the decimal places and check if precision is at most 2 decimal places
	multipliedRate := rate.Mul(sdkmath.LegacyNewDec(100))

	// Truncate the rate to remove decimal places
	truncatedRate := multipliedRate.TruncateDec()

	// Check if the truncated rate is equal to the original rate
	return multipliedRate.Equal(truncatedRate)
}

type RelativeTimeLockTapScriptInfo struct {
	// data necessary to build witness for given script
	SpendInfo *SpendInfo
	// lock time in script, required to set proper sequence number when spending output
	// with relative time lock
	LockTime uint16
	// taproot address of the script
	TapAddress btcutil.Address
	// pkscript in output which commits to the given script/leaf
	PkScript []byte
}

func BuildRelativeTimelockTaprootScript(
	pk *btcec.PublicKey,
	lockTime uint16,
	net *chaincfg.Params,
) (*RelativeTimeLockTapScriptInfo, error) {
	unspendableKeyPathKey := unspendableKeyPathInternalPubKey()

	script, err := buildTimeLockScript(pk, lockTime)

	if err != nil {
		return nil, err
	}

	sh, err := newTaprootScriptHolder(
		&unspendableKeyPathKey,
		[][]byte{script},
	)

	if err != nil {
		return nil, err
	}

	// there is only one script path in tree, so we can use index 0
	proof := sh.scriptTree.LeafMerkleProofs[0]

	spendInfo := &SpendInfo{
		ControlBlock: proof.ToControlBlock(&unspendableKeyPathKey),
		RevealedLeaf: proof.TapLeaf,
	}

	taprootAddress, err := DeriveTaprootAddress(
		sh.scriptTree,
		&unspendableKeyPathKey,
		net,
	)

	if err != nil {
		return nil, err
	}

	taprootPkScript, err := txscript.PayToAddrScript(taprootAddress)

	if err != nil {
		return nil, err
	}

	return &RelativeTimeLockTapScriptInfo{
		SpendInfo:  spendInfo,
		LockTime:   lockTime,
		TapAddress: taprootAddress,
		PkScript:   taprootPkScript,
	}, nil
}

// ParseBlkHeightAndPubKeyFromStoreKey expects to receive a key with
// BigEndianUint64(blkHeight) || BIP340PubKey(fpBTCPK)
func ParseBlkHeightAndPubKeyFromStoreKey(key []byte) (blkHeight uint64, fpBTCPK *bbn.BIP340PubKey, err error) {
	sizeBigEndian := 8
	if len(key) < sizeBigEndian+1 {
		return 0, nil, fmt.Errorf("key not long enough to parse block height and BIP340PubKey: %s", key)
	}

	fpBTCPK, err = bbn.NewBIP340PubKey(key[sizeBigEndian:])
	if err != nil {
		return 0, nil, fmt.Errorf("failed to parse pub key from key %w: %w", bbn.ErrUnmarshal, err)
	}

	blkHeight = sdk.BigEndianToUint64(key[:sizeBigEndian])
	return blkHeight, fpBTCPK, nil
}
