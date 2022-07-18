package keeper_test

import (
	"fmt"
	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/rand"
	"testing"
)

func FuzzHeadersStateCreateHeader(f *testing.F) {
	/*
		 Checks:
		 1. A headerInfo provided as an argument leads to the following storage objects being created:
			 - A (height, headerHash) -> headerInfo object
			 - A (headerHash) -> height object
			 - A (headerHash) -> work object
			 - A () -> tip object depending on conditions:
				 * If the tip does not exist, then the headerInfo is the tip
				 * If the tip exists, and the header inserted has greater work than it, then it becomes the tip

		 Data generation:
		 - Create four headers:
			 1. The Base header. This will test whether the tip is set.
			 2. A header that builds on top of the base header.
				This will test whether the tip is properly updated once more work is added.
			 3. A header that builds on top of the base header but with less work than (2).
				This will test whether the tip is not updated when less work than it is added.
			 4. A header that builds on top of the base header but with equal work to (2).
				This will test whether the tip is not updated when equal work to it is added.
		 - No need to create a tree, since this function does not consider the existence or the chain,
		   it just inserts into state and updates the tip based on a simple work comparison.
	*/
	datagen.AddRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)

		// Generate a tree with a single root node
		tree := genRandomTree(blcKeeper, ctx, 1, 1)
		baseHeader := tree.GetRoot()

		// Test whether the tip and storages are set
		tip := blcKeeper.HeadersState(ctx).GetTip()
		if tip == nil {
			t.Errorf("Creation of base header did not lead to creation of tip")
		}
		if !baseHeader.Eq(tip) {
			t.Errorf("Tip does not correspond to the one submitted %s %s", baseHeader.Hash, tip.Hash)
		}
		headerObj, err := blcKeeper.HeadersState(ctx).GetHeader(baseHeader.Height, baseHeader.Hash)
		if err != nil {
			t.Errorf("Could not retrieve created header")
		}
		if !baseHeader.Eq(headerObj) {
			t.Errorf("Created object does not correspond to the one submitted")
		}
		work, err := blcKeeper.HeadersState(ctx).GetHeaderWork(baseHeader.Hash)
		if err != nil {
			t.Errorf("Could not retrieve work of created header")
		}
		if !baseHeader.Work.Equal(*work) {
			t.Errorf("Created object work does not correspond to the one submitted")
		}
		height, err := blcKeeper.HeadersState(ctx).GetHeaderHeight(baseHeader.Hash)
		if err != nil {
			t.Errorf("Could not retrieve height of created header")
		}
		if height != baseHeader.Height {
			t.Errorf("Created object height does not correspond to the one submitted")
		}

		// Test whether a new header updates the tip.
		// The smaller number, the bigger the difficulty
		mostDifficulty := sdk.NewUint(10)
		lessDifficulty := mostDifficulty.Add(sdk.NewUint(1))
		// Create an object that builds on top of base header
		childMostWork := datagen.GenRandomBTCHeaderInfoWithParentAndBits(baseHeader, &mostDifficulty)
		blcKeeper.HeadersState(ctx).CreateHeader(childMostWork)
		// Check whether the tip was updated
		tip = blcKeeper.HeadersState(ctx).GetTip()
		if tip == nil {
			t.Errorf("Tip became nil instead of getting updated")
		}
		if !childMostWork.Eq(tip) {
			t.Errorf("Tip did not get properly updated")
		}

		childEqualWork := datagen.GenRandomBTCHeaderInfoWithParentAndBits(baseHeader, &mostDifficulty)
		blcKeeper.HeadersState(ctx).CreateHeader(childEqualWork)
		// Check whether the tip was updated
		tip = blcKeeper.HeadersState(ctx).GetTip()
		if !childMostWork.Eq(tip) {
			t.Errorf("Tip got updated when it shouldn't")
		}

		childLessWork := datagen.GenRandomBTCHeaderInfoWithParentAndBits(baseHeader, &lessDifficulty)
		blcKeeper.HeadersState(ctx).CreateHeader(childLessWork)
		// Check whether the tip was updated
		tip = blcKeeper.HeadersState(ctx).GetTip()
		if !childMostWork.Eq(tip) {
			t.Errorf("Tip got updated when it shouldn't")
		}

	})
}

func FuzzHeadersStateTipOps(f *testing.F) {
	/*
		Functions Tested:
		1. CreateTip
		2. GetTip
		3. TipExists

		Checks:
		* CreateTip
			1. The `headerInfo` object passed is set as the tip.
		* GetTip
			1. If the tip does not exist, nil is returned.
			2. The element maintained in the tip storage is returned.
		* TipExists
			1. Returns true/false depending on the existence of a tip.

		Data generation:
		- Create two headers:
			1. A header that will be set as the tip.
			2. A header that will override it.
	*/
	datagen.AddRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)

		headerInfo1 := datagen.GenRandomBTCHeaderInfo()
		headerInfo2 := datagen.GenRandomBTCHeaderInfo()

		retrievedHeaderInfo := blcKeeper.HeadersState(ctx).GetTip()
		if retrievedHeaderInfo != nil {
			t.Errorf("GetTip did not return nil for empty tip")
		}

		if blcKeeper.HeadersState(ctx).TipExists() {
			t.Errorf("TipExists returned true when no tip has been set")
		}

		blcKeeper.HeadersState(ctx).CreateTip(headerInfo1)
		retrievedHeaderInfo = blcKeeper.HeadersState(ctx).GetTip()

		if !headerInfo1.Eq(retrievedHeaderInfo) {
			t.Errorf("HeaderInfo object did not get stored in tip")
		}

		if !blcKeeper.HeadersState(ctx).TipExists() {
			t.Errorf("TipExists returned false when a tip had been set")
		}

		blcKeeper.HeadersState(ctx).CreateTip(headerInfo2)
		retrievedHeaderInfo = blcKeeper.HeadersState(ctx).GetTip()
		if !headerInfo2.Eq(retrievedHeaderInfo) {
			t.Errorf("Tip did not get overriden")
		}
		if !blcKeeper.HeadersState(ctx).TipExists() {
			t.Errorf("TipExists returned false when a tip had been set")
		}
	})
}

func FuzzHeadersStateGetHeaderOps(f *testing.F) {
	/*
		Functions tested:
		1. GetHeader
		2. GetHeaderHeight
		3. GetHeaderWork
		4. GetHeaderByHash
		5. HeaderExists

		Checks:
		* GetHeader
			1. If the header specified by a height and a hash does not exist, (nil, error) is returned
			2. If the header specified by a height and a hash exists, (headerInfo, nil) is returned
		* GetHeaderHeight
			1. If the header specified by the hash does not exist, (0, error) is returned
			2. If the header specified by the hash exists, (height, nil) is returned
		* GetHeaderWork
			1. If the header specified by the hash does not exist, (nil, error) is returned
			2. If the header specified by the hash exists, (work, nil) is returned.
		* GetHeaderByHash
			1. If the header specified by the hash does not exist (nil, error) is returned
			2. If the header specified by the hash exists (headerInfo, nil) is returned.
		* HeaderExists
			1. Returns false if the header passed is nil.
			2. Returns true/false depending on the existence of the header.

		Data generation:
		- Create a header and store it using the `CreateHeader` method. Do retrievals to check conditions.
	*/
	datagen.AddRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)
		headerInfo := datagen.GenRandomBTCHeaderInfo()
		wrongHash := datagen.MutateHash(headerInfo.Hash)
		wrongHeight := headerInfo.Height + datagen.RandomInt(10) + 1

		// ****** HeaderExists tests ******
		if blcKeeper.HeadersState(ctx).HeaderExists(nil) {
			t.Errorf("HeaderExists returned true for nil input")
		}
		if blcKeeper.HeadersState(ctx).HeaderExists(headerInfo.Hash) {
			t.Errorf("HeaderExists returned true for not created header")
		}
		blcKeeper.HeadersState(ctx).CreateHeader(headerInfo)
		if !blcKeeper.HeadersState(ctx).HeaderExists(headerInfo.Hash) {
			t.Errorf("HeaderExists returned false for created header")
		}
		// ****** GetHeader tests ******
		// correct retrieval
		retrievedHeaderInfo, err := blcKeeper.HeadersState(ctx).GetHeader(headerInfo.Height, headerInfo.Hash)
		if err != nil {
			t.Errorf("GetHeader returned error for valid retrieval: %s", err)
		}
		if retrievedHeaderInfo == nil || !retrievedHeaderInfo.Eq(headerInfo) {
			t.Errorf("GetHeader returned a header that is nil or does not equal the one inserted")
		}
		retrievedHeaderInfo, err = blcKeeper.HeadersState(ctx).GetHeader(headerInfo.Height, wrongHash)
		if retrievedHeaderInfo != nil || err == nil {
			t.Errorf("GetHeader returned a filled HeaderInfo or the error is nil for invalid input")
		}

		retrievedHeaderInfo, err = blcKeeper.HeadersState(ctx).GetHeader(wrongHeight, headerInfo.Hash)
		if retrievedHeaderInfo != nil || err == nil {
			t.Errorf("GetHeader returned a filled HeaderInfo or the error is nil for invalid input")
		}

		retrievedHeaderInfo, err = blcKeeper.HeadersState(ctx).GetHeader(wrongHeight, wrongHash)
		if retrievedHeaderInfo != nil || err == nil {
			t.Errorf("GetHeader returned a filled HeaderInfo or the error is nil for invalid input")
		}

		// ****** GetHeaderHeight tests ******
		height, err := blcKeeper.HeadersState(ctx).GetHeaderHeight(headerInfo.Hash)
		if err != nil {
			t.Errorf("GetHeaderHeight returned an error for valid retrieval: %s", err)
		}
		if height != headerInfo.Height {
			t.Errorf("GetHeaderHeight returned incorrect height")
		}
		height, err = blcKeeper.HeadersState(ctx).GetHeaderHeight(wrongHash)
		if err == nil || height != 0 {
			t.Errorf("GetHeaderHeight returned nil error or a height different than zero for invalid input")
		}

		// ****** GetHeaderWork tests ******
		work, err := blcKeeper.HeadersState(ctx).GetHeaderWork(headerInfo.Hash)
		if err != nil {
			t.Errorf("GetHeaderWork returned an error for valid retrieval: %s", err)
		}
		if work == nil || !work.Equal(*headerInfo.Work) {
			t.Errorf("GetHeaderWork returned nil or incorrect work")
		}
		work, err = blcKeeper.HeadersState(ctx).GetHeaderWork(wrongHash)
		if err == nil || work != nil {
			t.Errorf("GetHeaderWork returned nil error or a work different than nil for invalid input")
		}

		// ****** GetHeaderByHash tests ******
		retrievedHeaderInfo, err = blcKeeper.HeadersState(ctx).GetHeaderByHash(headerInfo.Hash)
		if err != nil {
			t.Errorf("GetHeaderByHash returned an error for valid retrieval: %s", err)
		}
		if retrievedHeaderInfo == nil || !retrievedHeaderInfo.Eq(headerInfo) {
			t.Errorf("GetHeaderByHash returned a header that is nil or does not equal the one inserted")
		}

		retrievedHeaderInfo, err = blcKeeper.HeadersState(ctx).GetHeaderByHash(wrongHash)
		if retrievedHeaderInfo != nil || err == nil {
			t.Errorf("GetHeaderByHash returned a filled HeaderInfo or the error is nil for invalid input")
		}
	})
}

func FuzzHeadersStateGetBaseBTCHeader(f *testing.F) {
	/*
		Checks:
		1. If no headers exist, nil is returned
		2. The oldest element of the main chain is returned.

		Data generation:
		- Generate a random tree and retrieve the main chain from it.
	*/
	datagen.AddRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)

		nilBaseHeader := blcKeeper.HeadersState(ctx).GetBaseBTCHeader()
		if nilBaseHeader != nil {
			t.Errorf("Non-existent base BTC header led to non-nil return")
		}

		tree := genRandomTree(blcKeeper, ctx, 1, 10)
		expectedBaseHeader := tree.GetRoot()

		gotBaseHeader := blcKeeper.HeadersState(ctx).GetBaseBTCHeader()

		if !expectedBaseHeader.Eq(gotBaseHeader) {
			t.Errorf("Expected base header %s got %s", expectedBaseHeader.Hash, gotBaseHeader.Hash)
		}
	})
}

func FuzzHeadersStateHeadersByHeight(f *testing.F) {
	/*
		Checks:
		1. If the height does not correspond to any headers, the function parameter is never invoked.
		2. If the height corresponds to headers, the function is invoked for all of those headers.
		3. If the height corresponds to headers, the function is invoked until a stop signal is given.

		Data generation:
		- Generate a `rand.Intn(N)` number of headers with a particular height and insert them into storage.
		- The randomness of the number of headers should guarantee that (1) and (2) are observed.
		- Generate a random stop signal 1/N times.
	*/
	datagen.AddRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)
		maxHeaders := 256 // maximum 255 headers with particular height
		numHeaders := datagen.RandomInt(maxHeaders)

		// This will contain a mapping between all the header hashes that were created
		// and a boolean value.
		hashCount := make(map[string]bool)
		// Setup a tree with a single header
		tree := genRandomTree(blcKeeper, ctx, 1, 1)
		baseHeader := tree.GetRoot()
		height := baseHeader.Height + 1

		// Generate numHeaders with particular height
		var i uint64
		for i = 0; i < numHeaders; i++ {
			headerInfo := datagen.GenRandomBTCHeaderInfoWithParent(baseHeader)
			hashCount[headerInfo.Hash.MarshalHex()] = true
			blcKeeper.InsertHeader(ctx, headerInfo.Header)
		}

		var headersAdded uint64 = 0
		var stopHeight uint64 = 0
		blcKeeper.HeadersState(ctx).HeadersByHeight(height, func(header *types.BTCHeaderInfo) bool {
			headersAdded += 1
			if _, ok := hashCount[header.Hash.MarshalHex()]; !ok {
				t.Errorf("HeadersByHeight returned header that was not created")
			}
			hashCount[header.Hash.MarshalHex()] = true
			if datagen.OneInN(maxHeaders) {
				// Only set it once
				if stopHeight != 0 {
					stopHeight = headersAdded
				}
				return true
			}
			return false
		})
		if stopHeight != 0 && stopHeight != headersAdded {
			t.Errorf("Stop signal was not respected. %d headers were added while %d were expected", stopHeight, headersAdded)
		}

		for _, cnt := range hashCount {
			if !cnt && headersAdded == numHeaders {
				// If there is a header hash that the count is not set
				// and all the headers were iterated, then something went wrong
				t.Errorf("Function did not iterate all headers")
			}
		}
	})
}

func FuzzHeadersStateGetMainChain(f *testing.F) {
	/*
		Functions Tested:
		1. GetMainChain
		2. GetMainChainUpTo

		Checks:
		* GetMainChain
			1. We get the entire main chain.
		* GetMainChainUpTo
			1. We get the main chain containing `depth + 1` elements.

		Data generation:
		- Generate a random tree and retrieve the main chain from it.
		- Randomly generate the depth
	*/
	datagen.AddRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)

		tree := genRandomTree(blcKeeper, ctx, 1, 10)
		expectedMainChain := tree.GetMainChain()
		gotMainChain := blcKeeper.HeadersState(ctx).GetMainChain()

		if len(expectedMainChain) != len(gotMainChain) {
			t.Fatalf("Expected main chain length of %d, got %d", len(expectedMainChain), len(gotMainChain))
		}

		for i := 0; i < len(expectedMainChain); i++ {
			if !expectedMainChain[i].Eq(gotMainChain[i]) {
				t.Errorf("Expected header %s at position %d, got %s", expectedMainChain[i].Hash, i, gotMainChain[i].Hash)
			}
		}

		// depth is a random integer
		upToDepth := datagen.RandomInt(len(expectedMainChain))
		expectedMainChainUpTo := expectedMainChain[:upToDepth+1]
		gotMainChainUpTo := blcKeeper.HeadersState(ctx).GetMainChainUpTo(upToDepth)
		if len(expectedMainChainUpTo) != len(gotMainChainUpTo) {
			t.Fatalf("Expected main chain length of %d, got %d", len(expectedMainChainUpTo), len(gotMainChainUpTo))
		}

		for i := 0; i < len(expectedMainChainUpTo); i++ {
			if !expectedMainChainUpTo[i].Eq(gotMainChainUpTo[i]) {
				t.Errorf("Expected header %s at position %d, got %s", expectedMainChainUpTo[i].Hash, i, gotMainChainUpTo[i].Hash)
			}
		}
	})
}

func FuzzHeadersStateGetHighestCommonAncestor(f *testing.F) {
	/*
		Checks:
		1. The header returned is an ancestor of both headers.
		2. There is no header that is an ancestor of both headers that has a higher height
		   than the one returned.
		3. There is always a header that is returned, since all headers are built on top of the same root.

		Data generation:
		- Generate a random tree of headers and store it.
		- Select two random headers and call `GetHighestCommonAncestor` for them.
	*/
	datagen.AddRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)
		// Generate a random tree with at least one node
		tree := genRandomTree(blcKeeper, ctx, 1, 10)
		// Retrieve a random common ancestor
		commonAncestor := tree.RandomNode()

		// Generate two child nodes for the common ancestor
		childRoot1 := datagen.GenRandomBTCHeaderInfoWithParent(commonAncestor)
		childRoot2 := datagen.GenRandomBTCHeaderInfoWithParent(commonAncestor)
		if tree.Contains(childRoot1) || tree.Contains(childRoot2) {
			// Unlucky case where we get the same hash. Should be extremely rare.
			// Instead of adding extra complexity to this test case, just skip it
			t.Skip()
		}
		// Insert them into storage
		blcKeeper.InsertHeader(ctx, childRoot1.Header)
		blcKeeper.InsertHeader(ctx, childRoot2.Header)
		// Add them into the data structures maintained by the tree
		tree.Add(childRoot1, commonAncestor)
		tree.Add(childRoot2, commonAncestor)

		// Generate subtrees rooted at the descendant nodes
		genRandomTreeWithParent(blcKeeper, ctx, 1, 10, childRoot1, tree)
		genRandomTreeWithParent(blcKeeper, ctx, 1, 10, childRoot2, tree)

		// Get a random descendant from each of the subtrees
		descendant1 := tree.RandomDescendant(childRoot1)
		descendant2 := tree.RandomDescendant(childRoot2)

		retrievedHighestCommonAncestor := blcKeeper.HeadersState(ctx).GetHighestCommonAncestor(descendant1, descendant2)
		if retrievedHighestCommonAncestor == nil {
			t.Fatalf("No common ancestor found between the nodes %s and %s. Expected ancestor: %s", descendant1.Hash, descendant2.Hash, commonAncestor.Hash)
		}
		if !commonAncestor.Eq(retrievedHighestCommonAncestor) {
			fmt.Println("Failed")
			t.Errorf("Did not retrieve the correct highest common ancestor. Got %s, expected %s", retrievedHighestCommonAncestor.Hash, commonAncestor.Hash)
		}
	})
}

func FuzzHeadersStateGetInOrderAncestorsUntil(f *testing.F) {
	/*
		Checks:
		1. All the ancestors are contained in the returned list.
		2. The ancestors do not include the `ancestor` parameter.
		3. The ancestors are in order starting from the `ancestor`'s child and leading to the `descendant` parameter.

		Data generation:
		- Generate a random tree of headers and store it.
		- Select a random header which will serve as the `descendant`. Cannot be the base header.
		- Select a random header that is an ancestor of `descendant`.
	*/
	datagen.AddRandomSeedsToFuzzer(f, 100)
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)

		// Generate a tree of any size.
		// We can work with even one header, since this should lead to an empty result.
		tree := genRandomTree(blcKeeper, ctx, 1, 10)

		// Get a random header from the tree
		descendant := tree.RandomNode()
		// Get a random ancestor from it
		ancestor := tree.GetRandomAncestor(descendant)
		// Get the ancestry of the descendant.
		// It is in reverse order from the one that GetInOrderAncestorsUntil returns, since it starts with the descendant.
		expectedAncestorsReverse := tree.GetNodeAncestryUpTo(descendant, ancestor)
		gotAncestors := blcKeeper.HeadersState(ctx).GetInOrderAncestorsUntil(descendant, ancestor)
		if len(gotAncestors) != len(expectedAncestorsReverse) {
			t.Errorf("Got different ancestor list sizes. Expected %d got %d", len(expectedAncestorsReverse), len(gotAncestors))
		}

		for i := 0; i < len(expectedAncestorsReverse); i++ {
			reverseIdx := len(expectedAncestorsReverse) - i - 1
			if !expectedAncestorsReverse[i].Eq(gotAncestors[reverseIdx]) {
				t.Errorf("Ancestors do not match. Expected %s got %s", expectedAncestorsReverse[i].Hash, gotAncestors[reverseIdx].Hash)
			}
		}
	})
}

// genRandomTree generates a tree of headers. It accomplishes this by generating a root
// which will serve as the base header and then invokes the `genRandomTreeWithRoot` utility.
// The `minTreeHeight` and `maxTreeHeight` parameters denote the minimum and maximum height
// of the tree that is generated. For example, a `minTreeHeight` of 1,
// means that the tree should have at least one node (the root), while
// a `maxTreeHeight` of 4, denotes that the maximum height of the tree should be 4.
func genRandomTree(k *keeper.Keeper, ctx sdk.Context, minHeight uint64, maxHeight uint64) *datagen.BTCHeaderTree {
	tree := datagen.NewBTCHeaderTree()
	// Generate the root for the tree
	root := datagen.GenRandomBTCHeaderInfo()
	tree.Add(root, nil)
	k.SetBaseBTCHeader(ctx, *root)

	genRandomTreeWithParent(k, ctx, minHeight-1, maxHeight-1, root, tree)

	return tree
}

// genRandomTreeWithParent is a utility function for inserting the headers
// While the tree is generated, the headers that are generated for it are inserted into storage.
func genRandomTreeWithParent(k *keeper.Keeper, ctx sdk.Context, minHeight uint64,
	maxHeight uint64, root *types.BTCHeaderInfo, tree *datagen.BTCHeaderTree) {

	if minHeight > maxHeight {
		panic("Min height more than max height")
	}

	tree.GenRandomBTCHeaderTree(minHeight, maxHeight, root, func(header *types.BTCHeaderInfo) bool {
		err := k.InsertHeader(ctx, header.Header)
		if err != nil {
			panic(fmt.Sprintf("header insertion failed: %s", err))
		}
		return true
	})
}
