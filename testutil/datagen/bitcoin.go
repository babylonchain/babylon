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

// GenRandomHeaderInfoWithParentAndWork generates a BTCHeaderInfo object in which the `header.PrevBlock` points to the `parent`
// 								 and the `Work` property points to the accumulated work (parent.Work + header.Work)
func GenRandomHeaderInfoWithParentAndWork(parent *btclightclienttypes.BTCHeaderInfo, work *sdk.Uint) *btclightclienttypes.BTCHeaderInfo {
	header, _ := bbl.NewBTCHeaderBytesFromBytes(GenRandomByteArray(bbl.BTCHeaderLen))

	// Random header work
	headerWork := sdk.NewUint(rand.Uint64()).Add(sdk.NewUint(1))
	if work != nil {
		headerWork = *work
	}

	// Retrieve the btcd block to do modifications
	btcdHeader := header.ToBlockHeader()
	btcdHeader.Bits = blockchain.BigToCompact(headerWork.BigInt())

	// Compute the parameters that depend on the parent
	var accumulatedWork sdk.Uint
	var height uint64
	if parent != nil {
		height = parent.Height + 1
		accumulatedWork = btclightclienttypes.CumulativeWork(headerWork, *parent.Work)
		// Set the PrevBlock to the parent's hash
		btcdHeader.PrevBlock = *parent.Hash.ToChainhash()
	} else {
		height = rand.Uint64()
		// if there is no parent, the accumulated work is the same as the work
		accumulatedWork = headerWork
	}

	header = bbl.NewBTCHeaderBytesFromBlockHeader(btcdHeader)

	return &btclightclienttypes.BTCHeaderInfo{
		Header: &header,
		Hash:   header.Hash(),
		Height: height,
		Work:   &accumulatedWork,
	}
}

func GenRandomHeaderInfoWithParent(parent *btclightclienttypes.BTCHeaderInfo) *btclightclienttypes.BTCHeaderInfo {
	return GenRandomHeaderInfoWithParentAndWork(parent, nil)
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
