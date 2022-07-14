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

func GenRandomBtcdHeader(version int32, bits uint32, nonce uint32,
	timeInt int64, prevBlockStr string, merkleRootStr string) *wire.BlockHeader {
	if !ValidHex(prevBlockStr, bbl.BTCHeaderHashLen) {
		prevBlockStr = GenRandomHexStr(bbl.BTCHeaderHashLen)
	}
	if !ValidHex(merkleRootStr, bbl.BTCHeaderHashLen) {
		merkleRootStr = GenRandomHexStr(bbl.BTCHeaderHashLen)
	}

	// Get the chainhash versions
	prevBlock, _ := chainhash.NewHashFromStr(prevBlockStr)
	merkleRoot, _ := chainhash.NewHashFromStr(merkleRootStr)
	time := time.Unix(timeInt, 0)

	// Construct a header
	header := wire.BlockHeader{
		Version:    version,
		Bits:       bits,
		Nonce:      nonce,
		PrevBlock:  *prevBlock,
		MerkleRoot: *merkleRoot,
		Timestamp:  time,
	}

	return &header
}

func GenRandomBits() uint32 {
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

func GenRandomPrevBlockChainhash() *chainhash.Hash {
	chHash, _ := chainhash.NewHashFromStr(GenRandomHexStr(bbl.BTCHeaderHashLen))
	return chHash
}

func GenRandomMerkleRootChainhash() *chainhash.Hash {
	// TODO: use a constant for this
	chHash, _ := chainhash.NewHashFromStr(GenRandomHexStr(32))
	return chHash
}

func GenRandomTime() time.Time {
	// TODO: Do not use the current time
	return time.Now()
}

func GenRandomVersion() int32 {
	return rand.Int31()
}

func GenRandomHeaderBytes(parent *btclightclienttypes.BTCHeaderInfo, bitsBig *sdk.Uint) bbl.BTCHeaderBytes {
	// TODO: add prefixes to func names
	headerBits := GenRandomBits()
	parentHash := GenRandomPrevBlockChainhash()
	merkleRoot := GenRandomMerkleRootChainhash()
	time := GenRandomTime()
	version := GenRandomVersion()

	if bitsBig != nil {
		headerBits = blockchain.BigToCompact(bitsBig.BigInt())
	}
	if parent != nil {
		parentHash = parent.Hash.ToChainhash()
	}

	btcdHeader := &wire.BlockHeader{}
	btcdHeader.Bits = headerBits
	btcdHeader.PrevBlock = *parentHash
	btcdHeader.Version = version
	btcdHeader.Timestamp = time
	btcdHeader.MerkleRoot = *merkleRoot
	return bbl.NewBTCHeaderBytesFromBlockHeader(btcdHeader)
}

func GenRandomHeight() uint64 {
	return rand.Uint64()
}

// GenRandomHeaderInfoWithParentAndBits generates a BTCHeaderInfo object in which the `header.PrevBlock` points to the `parent`
// and the `Work` property points to the accumulated work (parent.Work + header.Work). Less bits as a parameter, means more difficulty.
func GenRandomHeaderInfoWithParentAndBits(parent *btclightclienttypes.BTCHeaderInfo, bits *sdk.Uint) *btclightclienttypes.BTCHeaderInfo {
	header := GenRandomHeaderBytes(parent, bits)
	height := GenRandomHeight()
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

func GenRandomHeaderInfoWithParent(parent *btclightclienttypes.BTCHeaderInfo) *btclightclienttypes.BTCHeaderInfo {
	return GenRandomHeaderInfoWithParentAndBits(parent, nil)
}

// GenRandomHeaderInfo generates a random BTCHeaderInfo object
func GenRandomHeaderInfo() *btclightclienttypes.BTCHeaderInfo {
	return GenRandomHeaderInfoWithParent(nil)
}

// GenRandomHeaderInfoWithHeight generates a random BTCHeaderInfo object with a particular height
func GenRandomHeaderInfoWithHeight(height uint64) *btclightclienttypes.BTCHeaderInfo {
	headerInfo := GenRandomHeaderInfo()
	headerInfo.Height = height
	return headerInfo
}

// genRandomHeaderInfoChildren recursivelly generates a random tree of BTCHeaderInfo objects rooted at `parent`.
// 							   It accomplishes this by randomly selecting the number of children for the `parent` block, and
// 							   then generating trees rooted at the children block.
//							   A `depth` argument is provided that specifies the maximum depth for the tree rooted at `parent`.
// 							   The generated BTCHeaderInfo objects are inserted into a hashmap, for future efficient retrieval.
func genRandomHeaderInfoChildren(headersMap map[string]*btclightclienttypes.BTCHeaderInfo, parent *btclightclienttypes.BTCHeaderInfo, minDepth uint64, depth uint64) {
	// Base condition
	if depth == 0 {
		return
	}

	// Randomly identify the number of children
	numChildren := 0
	if minDepth > 0 {
		numChildren = 1 // 75% chance of 1 child now
	}
	if OneInN(2) {
		// 50% of the times, one child
		numChildren = 1
	} else if OneInN(2) {
		// 25% of the times, 2 children
		// Implies that 25% of the times 0 children
		numChildren = 2
	}

	// Generate the children, insert them into the hashmap, and generate the grandchildren.
	for i := 0; i < numChildren; i++ {
		child := GenRandomHeaderInfoWithParent(parent)
		if _, ok := headersMap[child.Hash.String()]; ok {
			// Extraordinary chance that we got the same hash
			continue
		}
		// Insert the child into the hash map
		headersMap[child.Hash.String()] = child
		// Generate the grandchildren
		genRandomHeaderInfoChildren(headersMap, child, minDepth-1, depth-1)
	}
}

// GenRandomHeaderInfoTreeMinDepth recursivelly generates a random tree of BTCHeaderInfo objects that has a minimum depth.
func GenRandomHeaderInfoTreeMinDepth(minDepth uint64) map[string]*btclightclienttypes.BTCHeaderInfo {
	headers := make(map[string]*btclightclienttypes.BTCHeaderInfo, 0)
	depth := RandomInt(10) + 1 // Maximum depth: 10
	if depth < minDepth {
		depth = minDepth
	}
	root := GenRandomHeaderInfo()

	headers[root.Hash.String()] = root

	genRandomHeaderInfoChildren(headers, root, minDepth-1, depth-1)

	return headers
}

// GenRandomHeaderInfoTree recursivelly generates a random tree of BTCHeaderInfo objects.
func GenRandomHeaderInfoTree() map[string]*btclightclienttypes.BTCHeaderInfo {
	return GenRandomHeaderInfoTreeMinDepth(0)
}

// MutateHash takes a hash as a parameter, copies it, modifies the copy, and returns the copy.
func MutateHash(hash *bbl.BTCHeaderHashBytes) *bbl.BTCHeaderHashBytes {
	mutatedBytes := make([]byte, bbl.BTCHeaderHashLen)
	copy(mutatedBytes, hash.MustMarshal())
	mutatedBytes[0] -= 1
	mutated, _ := bbl.NewBTCHeaderHashBytesFromBytes(mutatedBytes)
	return &mutated
}
