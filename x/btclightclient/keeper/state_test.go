package keeper_test

import (
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
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)

		// Create base header and test whether the tip and storages are set
		baseHeader := datagen.GenRandomBTCHeaderInfo()
		blcKeeper.HeadersState(ctx).CreateHeader(baseHeader)
		tip := blcKeeper.HeadersState(ctx).GetTip()
		if tip == nil {
			t.Errorf("Creation of base header did not lead to creation of tip")
		}
		if !baseHeader.Eq(tip) {
			t.Errorf("Tip does not correspond to the one submitted")
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
	f.Add(int64(42))
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
	f.Add(int64(42))
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
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)

		nilBaseHeader := blcKeeper.HeadersState(ctx).GetBaseBTCHeader()
		if nilBaseHeader != nil {
			t.Errorf("Non-existent base BTC header led to non-nil return")
		}

		headersMap := datagen.GenRandomHeaderInfoTree()
		for _, headerInfo := range headersMap {
			blcKeeper.HeadersState(ctx).CreateHeader(headerInfo)
		}

		tip := blcKeeper.HeadersState(ctx).GetTip()
		mainChain := getMainChain(headersMap, tip)

		expectedBaseHeader := mainChain[len(mainChain)-1]
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
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)
		maxHeaders := 256 // maximum 255 headers with particular height
		numHeaders := datagen.RandomInt(maxHeaders)
		height := rand.Uint64() // the height for those headers

		// This will contain a mapping between all the header hashes that were created
		// and a boolean value.
		hashCount := make(map[string]bool)

		// Generate numHeaders with particular height
		var i uint64
		for i = 0; i < numHeaders; i++ {
			headerInfo := datagen.GenRandomBTCHeaderInfoWithHeight(height)
			hashCount[headerInfo.Hash.MarshalHex()] = false
			blcKeeper.HeadersState(ctx).CreateHeader(headerInfo)
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
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)
		headersMap := datagen.GenRandomHeaderInfoTree()
		maxAccPow := sdk.NewUint(0)
		var tip *types.BTCHeaderInfo = nil
		// Add all headers to storage
		for _, headerInfo := range headersMap {
			blcKeeper.HeadersState(ctx).CreateHeader(headerInfo)
			if headerInfo.Work.GT(maxAccPow) {
				maxAccPow = *headerInfo.Work
				tip = headerInfo
			}
		}

		expectedMainChain := getMainChain(headersMap, tip)
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
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)
		// Generate a tree of at least a depth of two, since we need at least two nodes
		headersMap := datagen.GenRandomBTCHeaderInfoTreeMinDepth(uint64(2))
		// Add all headers to storage
		for _, headerInfo := range headersMap {
			blcKeeper.HeadersState(ctx).CreateHeader(headerInfo)
		}
		// Get two random headers from the tree. Use random indexes to identify those.
		// Get the random indexes
		header1Idx := datagen.RandomInt(len(headersMap))
		header2Idx := datagen.RandomIntOtherThan(int(header1Idx), len(headersMap))
		headers := selectRandomHeaders(headersMap, []uint64{header1Idx, header2Idx})
		header1 := headers[0]
		header2 := headers[1]

		// Identify the highest common ancestor for the two headers:
		// Do a BFS starting from both headers and maintain a hashmap denoting whether
		// something has been encountered. If we get to something that has been encountered,
		// then that's the highest common ancestor.
		var highestCommonAncestor *types.BTCHeaderInfo = nil
		visited := make(map[string]bool, 0)
		queue := make([]*types.BTCHeaderInfo, 0)
		queue = append(queue, header1)
		queue = append(queue, header2)
		for len(queue) > 0 {
			top := queue[0]
			queue = queue[1:] // Not that performant, O(N^2) complexity
			// If the node has been visited, it is the highest common ancestor
			if _, ok := visited[top.Hash.String()]; ok {
				highestCommonAncestor = top
				break
			}
			visited[top.Hash.String()] = true
			// Check if parent exists, we might be in the base node for which its parent does not exist.
			if parent, ok := headersMap[top.Header.ParentHash().String()]; ok {
				queue = append(queue, parent)
			}
		}

		if highestCommonAncestor == nil {
			t.Fatalf("Could not find a highest common ancestor")
		}

		retrievedHighestCommonAncestor := blcKeeper.HeadersState(ctx).GetHighestCommonAncestor(header1, header2)
		if retrievedHighestCommonAncestor == nil {
			t.Fatalf("No common ancestor found between the nodes %s and %s. Expected ancestor: %s", header1.Hash, header2.Hash, highestCommonAncestor.Hash)
		}
		if !highestCommonAncestor.Eq(retrievedHighestCommonAncestor) {
			t.Errorf("Did not retrieve the correct highest common ancestor. Got %s, expected %s", retrievedHighestCommonAncestor.Hash, highestCommonAncestor.Hash)
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
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)
		// Generate a tree of any size.
		// We can work with even one header, since this should lead to an empty result.
		headersMap := datagen.GenRandomHeaderInfoTree()
		// Insert the headers into storage
		for _, headerInfo := range headersMap {
			blcKeeper.HeadersState(ctx).CreateHeader(headerInfo)
		}
		// Get a random descendant and insert headers into storage
		descendantIdx := datagen.RandomInt(len(headersMap))
		descendant := selectRandomHeaders(headersMap, []uint64{descendantIdx})[0]

		// get the chain ending starting from the base header and ending on descendant
		chain := getChain(headersMap, descendant)
		// get a random ancestor (or the same node)
		ancestorIdx := datagen.RandomInt(len(chain))
		ancestor := chain[ancestorIdx]

		expectedAncestors := chain[ancestorIdx+1:]

		gotAncestors := blcKeeper.HeadersState(ctx).GetInOrderAncestorsUntil(descendant, ancestor)
		if len(gotAncestors) != len(expectedAncestors) {
			t.Errorf("Got different ancestor list sizes. Expected %d got %d", len(expectedAncestors), len(gotAncestors))
		}

		for i := 0; i < len(expectedAncestors); i++ {
			if !expectedAncestors[i].Eq(gotAncestors[i]) {
				t.Errorf("Ancestors do not match. Expected %s got %s", expectedAncestors[i].Hash, gotAncestors[i].Hash)
			}
		}
	})
}

func setupRandomChain(k *keeper.Keeper, ctx sdk.Context) {
	// TODO
}

func selectRandomHeaders(headers map[string]*types.BTCHeaderInfo, idxs []uint64) []*types.BTCHeaderInfo {
	res := make([]*types.BTCHeaderInfo, len(idxs))
	var idx uint64 = 0
	for _, headerInfo := range headers {
		for intIdx, pos := range idxs {
			if idx == pos {
				res[intIdx] = headerInfo
			}
		}
		idx += 1
	}
	return res
}

// getChain retrieves the chain starting from the descendant up to the base header.
func getChain(headers map[string]*types.BTCHeaderInfo, descendant *types.BTCHeaderInfo) []*types.BTCHeaderInfo {
	var chain []*types.BTCHeaderInfo
	if parent, ok := headers[descendant.Header.ParentHash().String()]; ok {
		chain = getChain(headers, parent)
	}
	chain = append(chain, descendant)
	return chain
}

// getMainChain finds the tip of the chain and retrieves all its ancestors until the base header
// 				The chain starts from the tip and leads to the base header
func getMainChain(headers map[string]*types.BTCHeaderInfo, tip *types.BTCHeaderInfo) []*types.BTCHeaderInfo {
	// This chain starts from the base header and leads to the tip
	// We want the reverse of it
	chain := getChain(headers, tip)
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain
}
