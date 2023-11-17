package btcstaking

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

// key sorting code copied from musig2 impl in btcd: https://github.com/btcsuite/btcd/blob/master/btcec/schnorr/musig2/keys.go
type sortableKeys []*btcec.PublicKey

// Less reports whether the element with index i must sort before the element
// with index j.
func (s sortableKeys) Less(i, j int) bool {
	// TODO(roasbeef): more efficient way to compare...
	keyIBytes := schnorr.SerializePubKey(s[i])
	keyJBytes := schnorr.SerializePubKey(s[j])

	return bytes.Compare(keyIBytes, keyJBytes) == -1
}

// Swap swaps the elements with indexes i and j.
func (s sortableKeys) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Len is the number of elements in the collection.
func (s sortableKeys) Len() int {
	return len(s)
}

// sortKeys takes a set of schnorr public keys and returns a new slice that is
// a copy of the keys sorted in lexicographical order bytes on the x-only
// pubkey serialization.
func sortKeys(keys []*btcec.PublicKey) []*btcec.PublicKey {
	keySet := sortableKeys(keys)
	if sort.IsSorted(keySet) {
		return keys
	}

	sort.Sort(keySet)
	return keySet
}

type SignatureInfo struct {
	SignerPubKey *btcec.PublicKey
	Signature    []byte
}

func NewSignatureInfo(
	signerPubKey *btcec.PublicKey,
	signature []byte,
) *SignatureInfo {
	return &SignatureInfo{
		SignerPubKey: signerPubKey,
		Signature:    signature,
	}
}

type sortableSigInfo []*SignatureInfo

// Less reports whether the element with index i must sort before the element
// with index j.
func (s sortableSigInfo) Less(i, j int) bool {
	// TODO(roasbeef): more efficient way to compare...
	keyIBytes := schnorr.SerializePubKey(s[i].SignerPubKey)
	keyJBytes := schnorr.SerializePubKey(s[j].SignerPubKey)

	return bytes.Compare(keyIBytes, keyJBytes) == 1
}

// Swap swaps the elements with indexes i and j.
func (s sortableSigInfo) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Len is the number of elements in the collection.
func (s sortableSigInfo) Len() int {
	return len(s)
}

// Helper function to sort all signatures in reverse lexicographical order of signing public keys
// this way signatures are ready to be used in multisig witness with corresponding public keys
func SortSignatureInfo(infos []*SignatureInfo) []*SignatureInfo {
	keySet := sortableSigInfo(infos)
	if sort.IsSorted(keySet) {
		return infos
	}

	sort.Sort(keySet)
	return keySet
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

// CreateBabylonWitness creates babylon compatible witness, as babylon scripts
// has witness with the same shape
// - first come signatures
// - then whole revealed script
// - then control block
func CreateBabylonWitness(
	signatures [][]byte,
	si *SpendInfo,
) (wire.TxWitness, error) {
	numSignatures := len(signatures)

	if numSignatures == 0 {
		return nil, fmt.Errorf("cannot build witness without signatures")
	}

	if si == nil {
		return nil, fmt.Errorf("cannot build witness without spend info")
	}

	controlBlockBytes, err := si.ControlBlock.ToBytes()

	if err != nil {
		return nil, err
	}

	// witness stack has:
	// all signatures
	// whole revealed script
	// control block
	witnessStack := wire.TxWitness(make([][]byte, numSignatures+2))

	for i, sig := range signatures {
		sc := sig
		witnessStack[i] = sc
	}

	witnessStack[numSignatures] = si.RevealedLeaf.Script
	witnessStack[numSignatures+1] = controlBlockBytes

	return witnessStack, nil
}
