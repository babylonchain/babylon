package btcstaking

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/txscript"
)

// private helper to assemble multisig script
// if `withVerify` is ture script will end with OP_NUMEQUALVERIFY otherwise with OP_NUMEQUAL
// SCRIPT: <Pk1> OP_CHEKCSIG <Pk2> OP_CHECKSIGADD <Pk3> OP_CHECKSIGADD ... <PkN> OP_CHECKSIGADD <threshold> OP_NUMEQUALVERIFY (or OP_NUMEQUAL)
func assembleMultiSigScript(
	pubkeys []*btcec.PublicKey,
	threshold uint32,
	withVerify bool,
) ([]byte, error) {
	builder := txscript.NewScriptBuilder()

	for i, key := range pubkeys {
		builder.AddData(schnorr.SerializePubKey(key))
		if i == 0 {
			builder.AddOp(txscript.OP_CHECKSIG)
		} else {
			builder.AddOp(txscript.OP_CHECKSIGADD)
		}
	}

	builder.AddInt64(int64(threshold))
	if withVerify {
		builder.AddOp(txscript.OP_NUMEQUALVERIFY)
	} else {
		builder.AddOp(txscript.OP_NUMEQUAL)
	}

	return builder.Script()
}

// SortKeys takes a set of schnorr public keys and returns a new slice that is
// a copy of the keys sorted in lexicographical order bytes on the x-only
// pubkey serialization.
func SortKeys(keys []*btcec.PublicKey) []*btcec.PublicKey {
	sortedKeys := make([]*btcec.PublicKey, len(keys))
	copy(sortedKeys, keys)
	sort.SliceStable(sortedKeys, func(i, j int) bool {
		keyIBytes := schnorr.SerializePubKey(sortedKeys[i])
		keyJBytes := schnorr.SerializePubKey(sortedKeys[j])
		return bytes.Compare(keyIBytes, keyJBytes) == -1
	})
	return sortedKeys
}

// prepareKeys prepares keys to be used in multisig script
// Validates:
// - whether there are at least 2 keys
// returns copy of the slice of keys sorted lexicographically
// Note: It is up to the caller to ensure that the keys are unique
func prepareKeysForMultisigScript(keys []*btcec.PublicKey) ([]*btcec.PublicKey, error) {
	if len(keys) < 2 {
		return nil, fmt.Errorf("cannot create multisig script with less than 2 keys")
	}

	sortedKeys := SortKeys(keys)

	return sortedKeys, nil
}

// buildMultiSigScript creates multisig script with given keys and signer threshold to
// successfully execute script
// it validates whether threshold is not greater than number of keys
// If there is only one key provided it will return single key sig script
// Note: It is up to the caller to ensure that the keys are unique
func buildMultiSigScript(
	keys []*btcec.PublicKey,
	threshold uint32,
	withVerify bool,
) ([]byte, error) {
	if len(keys) == 0 {
		return nil, fmt.Errorf("no keys provided")
	}

	if threshold > uint32(len(keys)) {
		return nil, fmt.Errorf("required number of valid signers is greater than number of provided keys")
	}

	if len(keys) == 1 {
		// if we have only one key we can use single key sig script
		return buildSingleKeySigScript(keys[0], withVerify)
	}

	sortedKeys, err := prepareKeysForMultisigScript(keys)

	if err != nil {
		return nil, err
	}

	return assembleMultiSigScript(sortedKeys, threshold, withVerify)
}

// Only holder of private key for given pubKey can spend after relative lock time
// SCRIPT: <StakerPk> OP_CHECKSIGVERIFY <stakingTime> OP_CHECKSEQUENCEVERIFY
func buildTimeLockScript(
	pubKey *btcec.PublicKey,
	lockTime uint16,
) ([]byte, error) {
	builder := txscript.NewScriptBuilder()
	builder.AddData(schnorr.SerializePubKey(pubKey))
	builder.AddOp(txscript.OP_CHECKSIGVERIFY)
	builder.AddInt64(int64(lockTime))
	builder.AddOp(txscript.OP_CHECKSEQUENCEVERIFY)
	return builder.Script()
}

// Only holder of private key for given pubKey can spend
// SCRIPT: <pubKey> OP_CHECKSIGVERIFY
func buildSingleKeySigScript(
	pubKey *btcec.PublicKey,
	withVerify bool,
) ([]byte, error) {
	builder := txscript.NewScriptBuilder()
	builder.AddData(schnorr.SerializePubKey(pubKey))

	if withVerify {
		builder.AddOp(txscript.OP_CHECKSIGVERIFY)
	} else {
		builder.AddOp(txscript.OP_CHECKSIG)
	}

	return builder.Script()
}
