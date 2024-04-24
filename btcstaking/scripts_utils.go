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
// SCRIPT: <Pk1> OP_CHEKCSIG <Pk2> OP_CHECKSIGADD <Pk3> OP_CHECKSIGADD ... <PkN> OP_CHECKSIGADD <threshold> OP_GREATERTHANOREQUAL OP_VERIFY
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
// - whether there are at lest 2 keys
// - whether there are no duplicate keys
// returns copy of the slice of keys sorted lexicographically
func prepareKeysForMultisigScript(keys []*btcec.PublicKey) ([]*btcec.PublicKey, error) {
	if len(keys) < 2 {
		return nil, fmt.Errorf("cannot create multisig script with less than 2 keys")
	}

	sortedKeys := SortKeys(keys)

	for i := 0; i < len(sortedKeys)-1; i++ {
		if bytes.Equal(schnorr.SerializePubKey(sortedKeys[i]), schnorr.SerializePubKey(sortedKeys[i+1])) {
			return nil, fmt.Errorf("duplicate key in list of keys")
		}
	}

	return sortedKeys, nil
}

// buildMultiSigScript creates multisig script with given keys and signer threshold to
// successfully execute script
// it validates whether provided keys are unique and the threshold is not greater than number of keys
// If there is only one key provided it will return single key sig script
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
