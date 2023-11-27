package btcstaking

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

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
// 3. Staker can spend with validator and covenant cooperation any time.
type StakingInfo struct {
	StakingOutput         *wire.TxOut
	scriptHolder          *taprootScriptHolder
	timeLockPathLeafHash  chainhash.Hash
	unbondingPathLeafHash chainhash.Hash
	slashingPathLeafHash  chainhash.Hash
}

// SpendInfo contains information necessary to create witness for given script
type SpendInfo struct {
	// Control block contains merkle proof of inclusion of revealed script path
	ControlBlock txscript.ControlBlock
	// RevealedLeaf is the leaf of the script tree which is revealed i.e scriptpath
	// which is being executed
	RevealedLeaf txscript.TapLeaf
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

// babylonScriptPaths is and aggregate of all possible babylon script paths
// not every babylon output will contain all of those paths
type babylonScriptPaths struct {
	timeLockPathScript  []byte
	unbondingPathScript []byte
	slashingPathScript  []byte
}

func newBabylonScriptPaths(
	stakerKey *btcec.PublicKey,
	validatorKeys []*btcec.PublicKey,
	covenantKeys []*btcec.PublicKey,
	covenantThreshold uint32,
	lockTime uint16,
) (*babylonScriptPaths, error) {
	if stakerKey == nil {
		return nil, fmt.Errorf("staker key is nil")
	}

	timeLockPathScript, err := BuildTimeLockScript(stakerKey, lockTime)

	if err != nil {
		return nil, err
	}

	covenantMultisigScript, err := BuildMultiSigScript(
		covenantKeys,
		covenantThreshold,
		// covenant multisig is always last in script so we do not run verify and leave
		// last value on the stack. If we do not leave at least one element on the stack
		// script will always error
		false,
	)

	if err != nil {
		return nil, err
	}

	stakerSigScript, err := BuildSingleKeySigScript(stakerKey, true)

	if err != nil {
		return nil, err
	}

	validatorSigScript, err := BuildMultiSigScript(
		validatorKeys,
		// we always require only one validator to sign
		1,
		// we need to run verify to clear the stack, as validator multisig is in the middle of the script
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
		validatorSigScript,
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
	validatorKeys []*btcec.PublicKey,
	covenantKeys []*btcec.PublicKey,
	covenantThreshold uint32,
	stakingTime uint16,
	stakingAmount btcutil.Amount,
	net *chaincfg.Params,
) (*StakingInfo, error) {
	unspendableKeyPathKey := UnspendableKeyPathInternalPubKey()

	babylonScripts, err := newBabylonScriptPaths(
		stakerKey,
		validatorKeys,
		covenantKeys,
		covenantThreshold,
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
// 2. Staker can spend with validator and covenant cooperation any time.
type UnbondingInfo struct {
	UnbondingOutput      *wire.TxOut
	scriptHolder         *taprootScriptHolder
	timeLockPathLeafHash chainhash.Hash
	slashingPathLeafHash chainhash.Hash
}

func BuildUnbondingInfo(
	stakerKey *btcec.PublicKey,
	validatorKeys []*btcec.PublicKey,
	covenantKeys []*btcec.PublicKey,
	covenantThreshold uint32,
	unbondingTime uint16,
	unbondingAmount btcutil.Amount,
	net *chaincfg.Params,
) (*UnbondingInfo, error) {
	unspendableKeyPathKey := UnspendableKeyPathInternalPubKey()

	babylonScripts, err := newBabylonScriptPaths(
		stakerKey,
		validatorKeys,
		covenantKeys,
		covenantThreshold,
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
