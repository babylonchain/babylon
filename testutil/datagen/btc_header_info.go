package datagen

import (
	bbl "github.com/babylonchain/babylon/types"
	btclightclienttypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/rand"
	"time"
)

func GenRandomBtcdHeader() *wire.BlockHeader {
	version := GenRandomBTCHeaderVersion()
	bits := GenRandomBTCHeaderBits()
	nonce := GenRandomBTCHeaderNonce()
	prevBlock := GenRandomBTCHeaderPrevBlock()
	merkleRoot := GenRandomBTCHeaderMerkleRoot()
	timestamp := GenRandomBTCHeaderTimestamp()

	header := &wire.BlockHeader{
		Version:    version,
		Bits:       bits,
		Nonce:      nonce,
		PrevBlock:  *prevBlock.ToChainhash(),
		MerkleRoot: *merkleRoot,
		Timestamp:  timestamp,
	}
	return header
}

// GenRandomBTCHeaderBits constructs a random uint32 corresponding to BTC header difficulty bits
func GenRandomBTCHeaderBits() uint32 {
	// Instead of navigating through all the different signs and bit constructing
	// of the workBits, we can resort to having a uint64 (instead of the maximum of 2^256).
	// First, generate an integer, convert it into a big.Int and then into compact form.

	difficulty := rand.Uint64()
	if difficulty == 0 {
		difficulty += 1
	}
	bigDifficulty := sdk.NewUint(difficulty)

	workBits := blockchain.BigToCompact(bigDifficulty.BigInt())
	return workBits
}

// GenRandomBTCHeaderPrevBlock constructs a random BTCHeaderHashBytes instance
func GenRandomBTCHeaderPrevBlock() *bbl.BTCHeaderHashBytes {
	hex := GenRandomHexStr(bbl.BTCHeaderHashLen)
	hashBytes, _ := bbl.NewBTCHeaderHashBytesFromHex(hex)
	return &hashBytes
}

// GenRandomBTCHeaderMerkleRoot generates a random hex string corresponding to a merkle root
func GenRandomBTCHeaderMerkleRoot() *chainhash.Hash {
	// TODO: the length should become a constant and the merkle root should have a custom type
	chHash, _ := chainhash.NewHashFromStr(GenRandomHexStr(32))
	return chHash
}

// GenRandomBTCHeaderTimestamp generates a random BTC header timestamp
func GenRandomBTCHeaderTimestamp() time.Time {
	randomTime := rand.Int63n(time.Now().Unix())
	return time.Unix(randomTime, 0)
}

// GenRandomBTCHeaderVersion generates a random version integer
func GenRandomBTCHeaderVersion() int32 {
	return rand.Int31()
}

// GenRandomBTCHeaderNonce generates a random BTC header nonce
func GenRandomBTCHeaderNonce() uint32 {
	return rand.Uint32()
}

// GenRandomBTCHeaderBytes generates a random BTCHeaderBytes object
// based on randomly generated BTC header attributes
// If the `parent` argument is not `nil`, then the `PrevBlock`
// attribute of the BTC header will point to the hash of the parent and the
// `Timestamp` attribute will be later than the parent's `Timestamp`.
// If the `bitsBig` argument is not `nil`, then the `Bits` attribute
// of the BTC header will point to the compact form of big integer.
func GenRandomBTCHeaderBytes(parent *btclightclienttypes.BTCHeaderInfo, bitsBig *sdk.Uint) bbl.BTCHeaderBytes {
	btcdHeader := GenRandomBtcdHeader()

	if bitsBig != nil {
		btcdHeader.Bits = blockchain.BigToCompact(bitsBig.BigInt())
	}
	if parent != nil {
		// Set the parent hash
		btcdHeader.PrevBlock = *parent.Hash.ToChainhash()
		// The time should be more recent than the parent time
		// Typical BTC header difference is 10 mins with some fluctuations
		// The header timestamp is going to be the time of the parent + 10 mins +- 0-59 seconds
		seconds := rand.Intn(60)
		if OneInN(2) { // 50% of the times subtract the seconds
			seconds = -1 * seconds
		}
		btcdHeader.Timestamp = parent.Header.Time().Add(time.Minute*10 + time.Duration(seconds)*time.Second)
	}

	return bbl.NewBTCHeaderBytesFromBlockHeader(btcdHeader)
}

// GenRandomBTCHeight returns a random uint64
func GenRandomBTCHeight() uint64 {
	return rand.Uint64()
}

// GenRandomBTCHeaderInfoWithParentAndBits generates a BTCHeaderInfo object in which the `header.PrevBlock` points to the `parent`
// and the `Work` property points to the accumulated work (parent.Work + header.Work). Less bits as a parameter, means more difficulty.
func GenRandomBTCHeaderInfoWithParentAndBits(parent *btclightclienttypes.BTCHeaderInfo, bits *sdk.Uint) *btclightclienttypes.BTCHeaderInfo {
	header := GenRandomBTCHeaderBytes(parent, bits)
	height := GenRandomBTCHeight()
	if parent != nil {
		height = parent.Height + 1
	}

	accumulatedWork := btclightclienttypes.CalcWork(&header)
	if parent != nil {
		accumulatedWork = btclightclienttypes.CumulativeWork(accumulatedWork, *parent.Work)
	}

	return &btclightclienttypes.BTCHeaderInfo{
		Header: &header,
		Hash:   header.Hash(),
		Height: height,
		Work:   &accumulatedWork,
	}
}

// GenRandomBTCHeaderInfoWithParent generates a random BTCHeaderInfo object
// in which the parent points to the `parent` parameter.
func GenRandomBTCHeaderInfoWithParent(parent *btclightclienttypes.BTCHeaderInfo) *btclightclienttypes.BTCHeaderInfo {
	return GenRandomBTCHeaderInfoWithParentAndBits(parent, nil)
}

func GenRandomBTCHeaderInfoWithBits(bits *sdk.Uint) *btclightclienttypes.BTCHeaderInfo {
	return GenRandomBTCHeaderInfoWithParentAndBits(nil, bits)
}

// GenRandomBTCHeaderInfo generates a random BTCHeaderInfo object
func GenRandomBTCHeaderInfo() *btclightclienttypes.BTCHeaderInfo {
	return GenRandomBTCHeaderInfoWithParent(nil)
}

// MutateHash takes a hash as a parameter, copies it, modifies the copy, and returns the copy.
func MutateHash(hash *bbl.BTCHeaderHashBytes) *bbl.BTCHeaderHashBytes {
	mutatedBytes := make([]byte, bbl.BTCHeaderHashLen)
	// Retrieve a random byte index
	idx := RandomInt(bbl.BTCHeaderHashLen)
	copy(mutatedBytes, hash.MustMarshal())
	// Add one to the index
	mutatedBytes[idx] += 1
	mutated, _ := bbl.NewBTCHeaderHashBytesFromBytes(mutatedBytes)
	return &mutated
}
