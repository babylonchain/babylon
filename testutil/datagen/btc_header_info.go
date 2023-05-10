package datagen

import (
	sdkmath "cosmossdk.io/math"
	bbn "github.com/babylonchain/babylon/types"
	btclightclienttypes "github.com/babylonchain/babylon/x/btclightclient/types"
	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/big"
	"math/rand"
	"time"
)

func GenRandomBtcdHeader(r *rand.Rand) *wire.BlockHeader {
	version := GenRandomBTCHeaderVersion(r)
	bits := GenRandomBTCHeaderBits(r)
	nonce := GenRandomBTCHeaderNonce(r)
	prevBlock := GenRandomBTCHeaderPrevBlock(r)
	merkleRoot := GenRandomBTCHeaderMerkleRoot(r)
	timestamp := GenRandomBTCHeaderTimestamp(r)

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
func GenRandomBTCHeaderBits(r *rand.Rand) uint32 {
	// Instead of navigating through all the different signs and bit constructing
	// of the workBits, we can resort to having a uint64 (instead of the maximum of 2^256).
	// First, generate an integer, convert it into a big.Int and then into compact form.

	difficulty := r.Uint64()
	if difficulty == 0 {
		difficulty += 1
	}
	bigDifficulty := sdk.NewUint(difficulty)

	workBits := blockchain.BigToCompact(bigDifficulty.BigInt())
	return workBits
}

// GenRandomBTCHeaderPrevBlock constructs a random BTCHeaderHashBytes instance
func GenRandomBTCHeaderPrevBlock(r *rand.Rand) *bbn.BTCHeaderHashBytes {
	hex := GenRandomHexStr(r, bbn.BTCHeaderHashLen)
	hashBytes, _ := bbn.NewBTCHeaderHashBytesFromHex(hex)
	return &hashBytes
}

// GenRandomBTCHeaderMerkleRoot generates a random hex string corresponding to a merkle root
func GenRandomBTCHeaderMerkleRoot(r *rand.Rand) *chainhash.Hash {
	// TODO: the length should become a constant and the merkle root should have a custom type
	chHash, _ := chainhash.NewHashFromStr(GenRandomHexStr(r, 32))
	return chHash
}

// GenRandomBTCHeaderTimestamp generates a random BTC header timestamp
func GenRandomBTCHeaderTimestamp(r *rand.Rand) time.Time {
	randomTime := r.Int63n(time.Now().Unix())
	return time.Unix(randomTime, 0)
}

// GenRandomBTCHeaderVersion generates a random version integer
func GenRandomBTCHeaderVersion(r *rand.Rand) int32 {
	return r.Int31()
}

// GenRandomBTCHeaderNonce generates a random BTC header nonce
func GenRandomBTCHeaderNonce(r *rand.Rand) uint32 {
	return r.Uint32()
}

// GenRandomBTCHeaderBytes generates a random BTCHeaderBytes object
// based on randomly generated BTC header attributes
// If the `parent` argument is not `nil`, then the `PrevBlock`
// attribute of the BTC header will point to the hash of the parent and the
// `Timestamp` attribute will be later than the parent's `Timestamp`.
// If the `bitsBig` argument is not `nil`, then the `Bits` attribute
// of the BTC header will point to the compact form of big integer.
func GenRandomBTCHeaderBytes(r *rand.Rand, parent *btclightclienttypes.BTCHeaderInfo, bitsBig *sdkmath.Uint) bbn.BTCHeaderBytes {
	btcdHeader := GenRandomBtcdHeader(r)

	if bitsBig != nil {
		btcdHeader.Bits = blockchain.BigToCompact(bitsBig.BigInt())
	}
	if parent != nil {
		// Set the parent hash
		btcdHeader.PrevBlock = *parent.Hash.ToChainhash()
		// The time should be more recent than the parent time
		// Typical BTC header difference is 10 mins with some fluctuations
		// The header timestamp is going to be the time of the parent + 10 mins +- 0-59 seconds
		seconds := r.Intn(60)
		if OneInN(r, 2) { // 50% of the times subtract the seconds
			seconds = -1 * seconds
		}
		btcdHeader.Timestamp = parent.Header.Time().Add(time.Minute*10 + time.Duration(seconds)*time.Second)
	}

	return bbn.NewBTCHeaderBytesFromBlockHeader(btcdHeader)
}

// GenRandomBTCHeight returns a random uint64
func GenRandomBTCHeight(r *rand.Rand) uint64 {
	return r.Uint64()
}

// GenRandomBTCHeaderInfoWithParentAndBits generates a BTCHeaderInfo object in which the `header.PrevBlock` points to the `parent`
// and the `Work` property points to the accumulated work (parent.Work + header.Work). Less bits as a parameter, means more difficulty.
func GenRandomBTCHeaderInfoWithParentAndBits(r *rand.Rand, parent *btclightclienttypes.BTCHeaderInfo, bits *sdkmath.Uint) *btclightclienttypes.BTCHeaderInfo {
	header := GenRandomBTCHeaderBytes(r, parent, bits)
	height := GenRandomBTCHeight(r)
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
func GenRandomBTCHeaderInfoWithParent(r *rand.Rand, parent *btclightclienttypes.BTCHeaderInfo) *btclightclienttypes.BTCHeaderInfo {
	return GenRandomBTCHeaderInfoWithParentAndBits(r, parent, nil)
}

// GenRandomValidBTCHeaderInfoWithParent generates random BTCHeaderInfo object
// with valid proof of work.
// WARNING: if parent is from network with a lot of work (mainnet) it may never finish
// use only with simnet headers
func GenRandomValidBTCHeaderInfoWithParent(r *rand.Rand, parent btclightclienttypes.BTCHeaderInfo) *btclightclienttypes.BTCHeaderInfo {
	randHeader := GenRandomBtcdHeader(r)
	parentHeader := parent.Header.ToBlockHeader()

	randHeader.Version = parentHeader.Version
	randHeader.PrevBlock = parentHeader.BlockHash()
	randHeader.Bits = parentHeader.Bits
	randHeader.Timestamp = parentHeader.Timestamp.Add(50 * time.Second)
	SolveBlock(randHeader)

	headerBytes := bbn.NewBTCHeaderBytesFromBlockHeader(randHeader)

	accumulatedWork := btclightclienttypes.CalcWork(&headerBytes)
	accumulatedWork = btclightclienttypes.CumulativeWork(accumulatedWork, *parent.Work)

	return &btclightclienttypes.BTCHeaderInfo{
		Header: &headerBytes,
		Hash:   headerBytes.Hash(),
		Height: parent.Height + 1,
		Work:   &accumulatedWork,
	}
}

func GenRandomBTCHeaderInfoWithBits(r *rand.Rand, bits *sdkmath.Uint) *btclightclienttypes.BTCHeaderInfo {
	return GenRandomBTCHeaderInfoWithParentAndBits(r, nil, bits)
}

// GenRandomBTCHeaderInfo generates a random BTCHeaderInfo object
func GenRandomBTCHeaderInfo(r *rand.Rand) *btclightclienttypes.BTCHeaderInfo {
	return GenRandomBTCHeaderInfoWithParent(r, nil)
}

func GenRandomBTCHeaderInfoWithInvalidHeader(r *rand.Rand, powLimit *big.Int) *btclightclienttypes.BTCHeaderInfo {
	var tries = 0
	for {
		info := GenRandomBTCHeaderInfo(r)

		err := bbn.ValidateBTCHeader(info.Header.ToBlockHeader(), powLimit)

		if err != nil {
			return info
		}

		tries++

		if tries >= 100 {
			panic("Failed to generate invalid btc header in 100 random tries")
		}
	}
}

// MutateHash takes a hash as a parameter, copies it, modifies the copy, and returns the copy.
func MutateHash(r *rand.Rand, hash *bbn.BTCHeaderHashBytes) *bbn.BTCHeaderHashBytes {
	mutatedBytes := make([]byte, bbn.BTCHeaderHashLen)
	// Retrieve a random byte index
	idx := RandomInt(r, bbn.BTCHeaderHashLen)
	copy(mutatedBytes, hash.MustMarshal())
	// Add one to the index
	mutatedBytes[idx] += 1
	mutated, _ := bbn.NewBTCHeaderHashBytesFromBytes(mutatedBytes)
	return &mutated
}
