package bip322

import (
	"bytes"
	"fmt"
	"io"

	"github.com/btcsuite/btcd/wire"
)

// this file provides functionality for converting a simple signature to
// a witness stack.
// the entire file is adapted from https://github.com/btcsuite/btcd/blob/v0.23.4/wire/msgtx.go#L568-L599

const (
	// maxWitnessItemsPerInput is the maximum number of witness items to
	// be read for the witness data for a single TxIn. This number is
	// derived using a possible lower bound for the encoding of a witness
	// item: 1 byte for length + 1 byte for the witness item itself, or two
	// bytes. This value is then divided by the currently allowed maximum
	// "cost" for a transaction. We use this for an upper bound for the
	// buffer and consensus makes sure that the weight of a transaction
	// cannot be more than 4000000.
	maxWitnessItemsPerInput = 4_000_000

	// maxWitnessItemSize is the maximum allowed size for an item within
	// an input's witness data. This value is bounded by the largest
	// possible block size, post segwit v1 (taproot).
	maxWitnessItemSize = 4_000_000
)

// readScript reads a variable length byte array that represents a transaction
// script.  It is encoded as a varInt containing the length of the array
// followed by the bytes themselves.  An error is returned if the length is
// greater than the passed maxAllowed parameter which helps protect against
// memory exhaustion attacks and forced panics through malformed messages.  The
// fieldName parameter is only used for the error message so it provides more
// context in the error.
func readScript(r io.Reader, pver uint32, maxAllowed uint32, fieldName string) ([]byte, error) {
	count, err := wire.ReadVarInt(r, pver)
	if err != nil {
		return nil, err
	}

	// Prevent byte array larger than the max message size.  It would
	// be possible to cause memory exhaustion and panics without a sane
	// upper bound on this count.
	if count > uint64(maxAllowed) {
		return nil, fmt.Errorf("%s is larger than the max allowed size "+
			"[count %d, max %d]", fieldName, count, maxAllowed)
	}

	b := make([]byte, count)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// simpleSigToWitness converts a simple signature into a witness stack
// As per the BIP-322 spec:
// "A simple signature consists of a witness stack, consensus encoded as a vector of vectors of bytes"
// https://github.com/bitcoin/bips/blob/e643d247c8bc086745f3031cdee0899803edea2f/bip-0322.mediawiki#simple
// However, the above does not provide much information about its encoding.
// We work on the encoding based on the Leather wallet implementation:
// https://github.com/leather-wallet/extension/blob/068d3cd465e1a642a763fecfa0e3ce5e94b07286/src/shared/crypto/bitcoin/bip322/bip322-utils.ts#L58
// More specifically, the signature is encoded as follows:
// - 1st byte: Elements of the witness stack that are serialized
// - For each element of the stack
//   - The first byte specifies how many bytes it contains
//   - The rest are the bytes of the element
func SimpleSigToWitness(sig []byte) ([][]byte, error) {
	// For each input, the witness is encoded as a stack
	// with one or more items. Therefore, we first read a
	// varint which encodes the number of stack items.
	buf := bytes.NewBuffer(sig)
	witCount, err := wire.ReadVarInt(buf, 0)
	if err != nil {
		return nil, err
	}

	// Prevent a possible memory exhaustion attack by
	// limiting the witCount value to a sane upper bound.
	if witCount > maxWitnessItemsPerInput {
		return nil, fmt.Errorf("too many witness items to fit "+
			"into max message size [count %d, max %d]",
			witCount, maxWitnessItemsPerInput)
	}

	// Then for witCount number of stack items, each item
	// has a varint length prefix, followed by the witness
	// item itself.
	witnessStack := make([][]byte, witCount)
	for j := uint64(0); j < witCount; j++ {
		witnessStack[j], err = readScript(
			buf, 0, maxWitnessItemSize, "script witness item",
		)
		if err != nil {
			return nil, err
		}
	}

	return witnessStack, nil
}

// serialization of witness copied from btcd
func writeTxWitness(
	w io.Writer,
	wit [][]byte,
) error {
	// pver is always 0 (at least in btcd)
	err := wire.WriteVarInt(w, 0, uint64(len(wit)))
	if err != nil {
		return err
	}
	for _, item := range wit {
		err = wire.WriteVarBytes(w, 0, item)
		if err != nil {
			return err
		}
	}
	return nil
}

func SerializeWitness(w wire.TxWitness) ([]byte, error) {
	var buf bytes.Buffer

	if err := writeTxWitness(&buf, w); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
