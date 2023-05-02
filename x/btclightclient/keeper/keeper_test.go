package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func FuzzKeeperIsHeaderKDeep(f *testing.F) {
	/*
		Checks:
		1. if the BTCHeaderBytes object is nil, an error is returned
		2. if the header does not exist, an error is returned
		3. if the header exists but it is higher than `depth`, false is returned
		4. if the header exists and is equal to `depth`, true is returned
		5. if the header exists and is higher than `depth`, true is returned
		6. if the header exists and is equal or higher to `depth` but not on the main chain, false is returned

		Data Generation:
		- Generate a random tree of headers.
		- Get the mainchain and select appropriate headers.
	*/
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)

		depth := r.Uint64()

		// Test nil input
		isDeep, err := blcKeeper.IsHeaderKDeep(ctx, nil, depth)
		if err == nil {
			t.Errorf("Nil input led to nil error")
		}
		if isDeep {
			t.Errorf("Nil input led to a true result")
		}

		// Test header not existing
		nonExistentHeader := datagen.GenRandomBTCHeaderBytes(r, nil, nil)
		isDeep, err = blcKeeper.IsHeaderKDeep(ctx, nonExistentHeader.Hash(), depth)
		if err == nil {
			t.Errorf("Non existent header led to nil error")
		}
		if isDeep {
			t.Errorf("Non existent header led to a true result")
		}

		// Generate a random tree of headers with at least one node
		tree := genRandomTree(r, blcKeeper, ctx, 1, 10)
		// Get a random header from the tree
		header := tree.RandomNode(r)
		// Get the tip of the chain and check whether the header is on the chain that it defines
		// In that case, the true/false result depends on the depth parameter that we provide.
		// Otherwise, the result should always be false, regardless of the parameter.
		tip := tree.GetTip()
		if tree.IsOnNodeChain(tip, header) {
			mainchain := tree.GetMainChain()
			// Select a random depth based on the main-chain length
			randDepth := uint64(r.Int63n(int64(len(mainchain))))
			isDeep, err = blcKeeper.IsHeaderKDeep(ctx, header.Hash, randDepth)
			// Identify whether the function should return true or false
			headerDepth := tip.Height - header.Height
			// If the random depth that we chose is more than the headerDepth, then it should return true
			expectedIsDeep := randDepth >= headerDepth
			if err != nil {
				t.Errorf("Existent header led to a non-nil error")
			}
			if expectedIsDeep != isDeep {
				t.Errorf("Expected result %t for header with depth %d when parameter depth is %d", expectedIsDeep, headerDepth, randDepth)
			}
		} else {
			// The depth provided does not matter, we should always get false.
			randDepth := r.Uint64()
			isDeep, err = blcKeeper.IsHeaderKDeep(ctx, header.Hash, randDepth)
			if err != nil {
				t.Errorf("Existent header led to a non-nil error %s", err)
			}
			if isDeep {
				t.Errorf("Got a true result for header that is not part of the mainchain")
			}
		}
	})
}

func FuzzKeeperMainChainDepth(f *testing.F) {
	/*
		Checks:
		1. if the BTCHeaderBytes object is nil, an error is returned and the height is -1
		2. if the BTCHeaderBytes object does not exist in storage, (-1, error) is returned
		3. if the BTCHeaderBytes object has a height that is higher than the tip, (-1, error) is returned
		4. if the header is not on the main chain, (-1, nil) is returned
		5. if the header exists and is on the mainchain, (depth, nil) is returned

		Data Generation:
		- Generate a random tree of headers.
		- Random generation of a header that is not inserted into storage.
		- Random selection of a header from the main chain and outside of it.
	*/
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)

		// Test nil input
		depth, err := blcKeeper.MainChainDepth(ctx, nil)
		if err == nil {
			t.Errorf("Nil input led to nil error")
		}
		if depth != -1 {
			t.Errorf("Nil input led to a result that is not -1")
		}

		// Test header not existing
		nonExistentHeader := datagen.GenRandomBTCHeaderBytes(r, nil, nil)
		depth, err = blcKeeper.MainChainDepth(ctx, nonExistentHeader.Hash())
		if err == nil {
			t.Errorf("Non existent header led to nil error")
		}
		if depth != -1 {
			t.Errorf("Non existing header led to a result that is not -1")
		}

		// Generate a random tree of headers with at least one node
		tree := genRandomTree(r, blcKeeper, ctx, 1, 10)
		// Get a random header from the tree
		header := tree.RandomNode(r)
		// Get the tip of the chain and check whether the header is on the chain that it defines
		// In that case, the depth result depends on the depth of the header on the mainchain.
		// Otherwise, the result should always be -1
		tip := tree.GetTip()
		// Get the depth
		depth, err = blcKeeper.MainChainDepth(ctx, header.Hash)
		if err != nil {
			t.Errorf("Existent and header led to error")
		}
		if tree.IsOnNodeChain(tip, header) {
			expectedDepth := tip.Height - header.Height
			if depth < 0 {
				t.Errorf("Mainchain header led to negative depth")
			}
			if uint64(depth) != expectedDepth {
				t.Errorf("Got depth %d, expected %d", depth, expectedDepth)
			}
		} else {
			if depth >= 0 {
				t.Errorf("Non-mainchain header let to >= 0 result")
			}
		}
	})
}

func FuzzKeeperBlockHeight(f *testing.F) {
	/*
		Checks:
		1. if the BTCHeaderBytes object is nil, a (0, error) is returned
		2. if the BTCHeaderBytes object does not exist in storage, (0, error) is returned.
		3. if the BTCHeaderBytes object exists, (height, nil) is returned.

		Data Generation:
		- Generate a random tree of headers.
		- Random generation of a header that is not inserted into storage.
		- Random selection of a header from the main chain and outside of it.
	*/
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)

		// Test nil input
		height, err := blcKeeper.BlockHeight(ctx, nil)
		if err == nil {
			t.Errorf("Nil input led to nil error")
		}
		if height != 0 {
			t.Errorf("Nil input led to a result that is not -1")
		}

		// Test header not existing
		nonExistentHeader := datagen.GenRandomBTCHeaderBytes(r, nil, nil)
		height, err = blcKeeper.BlockHeight(ctx, nonExistentHeader.Hash())
		if err == nil {
			t.Errorf("Non existent header led to nil error")
		}
		if height != 0 {
			t.Errorf("Non existing header led to a result that is not -1")
		}

		tree := genRandomTree(r, blcKeeper, ctx, 1, 10)
		header := tree.RandomNode(r)
		height, err = blcKeeper.BlockHeight(ctx, header.Hash)
		if err != nil {
			t.Errorf("Existent header led to an error")
		}
		if height != header.Height {
			t.Errorf("BlockHeight returned %d, expected %d", height, header.Height)
		}
	})
}

func FuzzKeeperIsAncestor(f *testing.F) {
	/*
		Checks:
		1. If the child hash or the parent hash are nil, an error is returned
		2. If the child has a lower height than the parent, an error is returned
		3. If the child and the parent are the same, false is returned
		4. If the parent is an ancestor of child then `true` is returned.

		Data generation:
		- Generate a random tree of headers and insert it into storage.
		- Select a random header and select a random descendant and a random ancestor to test (2-4).
	*/
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)

		nonExistentParent := datagen.GenRandomBTCHeaderInfo(r)
		nonExistentChild := datagen.GenRandomBTCHeaderInfo(r)

		// nil inputs test
		isAncestor, err := blcKeeper.IsAncestor(ctx, nil, nil)
		if err == nil {
			t.Errorf("Nil input led to nil error")
		}
		if isAncestor {
			t.Errorf("Nil input led to true result")
		}
		isAncestor, err = blcKeeper.IsAncestor(ctx, nonExistentParent.Hash, nil)
		if err == nil {
			t.Errorf("Nil input led to nil error")
		}
		if isAncestor {
			t.Errorf("Nil input led to true result")
		}
		isAncestor, err = blcKeeper.IsAncestor(ctx, nil, nonExistentChild.Hash)
		if err == nil {
			t.Errorf("Nil input led to nil error")
		}
		if isAncestor {
			t.Errorf("Nil input led to true result")
		}

		// non-existent test
		isAncestor, err = blcKeeper.IsAncestor(ctx, nonExistentParent.Hash, nonExistentChild.Hash)
		if err == nil {
			t.Errorf("Non existent headers led to nil error")
		}
		if isAncestor {
			t.Errorf("Non existent headers led to true result")
		}

		// Generate random tree of headers
		tree := genRandomTree(r, blcKeeper, ctx, 1, 10)
		header := tree.RandomNode(r)
		ancestor := tree.RandomNode(r)

		if ancestor.Eq(header) {
			// Same headers test
			isAncestor, err = blcKeeper.IsAncestor(ctx, ancestor.Hash, header.Hash)
			if err != nil {
				t.Errorf("Valid input led to an error")
			}
			if isAncestor {
				t.Errorf("Same header input led to true result")
			}
		} else if ancestor.Height >= header.Height { // Descendant test
			isAncestor, err = blcKeeper.IsAncestor(ctx, ancestor.Hash, header.Hash)
			if err != nil {
				t.Errorf("Providing a descendant as a parent led to a non-nil error")
			}
			if isAncestor {
				t.Errorf("Providing a descendant as a parent led to a true result")
			}
		} else { // Ancestor test
			isAncestor, err = blcKeeper.IsAncestor(ctx, ancestor.Hash, header.Hash)
			if err != nil {
				t.Errorf("Valid input led to an error")
			}
			if isAncestor != tree.IsOnNodeChain(header, ancestor) { // The result should be whether it is an ancestor or not
				t.Errorf("Got invalid ancestry result. Expected %t, got %t", tree.IsOnNodeChain(header, ancestor), isAncestor)
			}
		}
	})
}

func FuzzKeeperInsertHeader(f *testing.F) {
	/*
		Checks:
		1. if the BTCHeaderBytes object is nil, an error is returned
		2. if the BTCHeaderBytes object corresponds to an existing header, an error is returned
		3. if the BTCHeaderBytes object parent is not maintained, an error is returned
		4. if all the checks pass:
			4a. corresponding objects have been created on the headers, hashToHeight, and hashToWork storages
			4b. the cumulative work of the added header is its own + its parent's
			4c. the height of the added header is its parent's + 1
			4d. the object added to the headers storage corresponds to a header info with the above attributes
			4e. the tip is properly updated and corresponding roll-forward and roll-backward events are triggered. Three cases:
				4e1. The new header builds on top of the existing tip
					 - New header becomes the tip
					 - Roll-Forward event to the new tip
				4e2. The new header builds on a fork and the cumulative work is less than the tip
					 - The tip does not change
					 - No events are triggered
				4e3. The new header builds on a fork and the cumulative work is more than the tip
					 - New header becomes the tip
					 - Roll-backward event to the highest common ancestor (can use the `GetHighestCommonAncestor` function)
					 - Roll-forward event to all the elements of the fork after the highest common ancestor

		Data Generation:
		- Generate a random tree of headers and insert them into storage.
		- Construct BTCHeaderBytes object that corresponds to existing header
		- Construct BTCHeaderBytes object for which its parent is not maintained
		- Construct BTCHeaderBytes objects that:
			* Build on top of the tip
			* Build on top of a header that is `rand.Intn(tipHeight)` headers back from the tip.
				- This should emulate both 4e2 and 4e3.
	*/
	datagen.AddRandomSeedsToFuzzer(f, 10)
	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))
		blcKeeper, ctx := testkeeper.BTCLightClientKeeper(t)
		tree := genRandomTree(r, blcKeeper, ctx, 1, 10)

		// Test nil input
		err := blcKeeper.InsertHeader(ctx, nil)
		if err == nil {
			t.Errorf("Nil input led to nil error")
		}

		existingHeader := tree.RandomNode(r)
		err = blcKeeper.InsertHeader(ctx, existingHeader.Header)
		if err == nil {
			t.Errorf("Existing header led to nil error")
		}

		nonExistentHeader := datagen.GenRandomBTCHeaderInfo(r)
		err = blcKeeper.InsertHeader(ctx, nonExistentHeader.Header)
		if err == nil {
			t.Errorf("Header with non-existent parent led to nil error")
		}

		// Create mock hooks that just store with what they were called
		mockHooks := NewMockHooks()
		blcKeeper.SetHooks(mockHooks)

		// Select a random header and build a header on top of it
		parentHeader := tree.RandomNode(r)
		header := datagen.GenRandomBTCHeaderInfoWithParent(r, parentHeader)

		// Assign a new event manager
		// We do this because the tree building might have led to events getting sent
		// and we want to ignore those.
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		// Get the tip in order to check if the header build on top of the tip
		oldTip := blcKeeper.HeadersState(ctx).GetTip()

		// Insert the header into storage
		err = blcKeeper.InsertHeader(ctx, header.Header)
		if err != nil {
			t.Errorf("Valid header led to an error")
		}

		// Get the new tip
		newTip := blcKeeper.HeadersState(ctx).GetTip()

		// Get event types. Those will be useful to test the types of the emitted events
		rollForwadType, _ := sdk.TypedEventToEvent(&types.EventBTCRollForward{})
		rollBackType, _ := sdk.TypedEventToEvent(&types.EventBTCRollBack{})
		headerInsertedType, _ := sdk.TypedEventToEvent(&types.EventBTCHeaderInserted{})

		// The headerInserted hook call should contain the new header
		if len(mockHooks.AfterBTCHeaderInsertedStore) != 1 {
			t.Fatalf("Expected a single BTCHeaderInserted hook to be invoked. Got %d", len(mockHooks.AfterBTCHeaderInsertedStore))
		}
		if !mockHooks.AfterBTCHeaderInsertedStore[0].Eq(header) {
			t.Errorf("The headerInfo inside the BTCHeaderInserted hook is not the new header")
		}
		// Check that an event has been triggered for the new header
		if len(ctx.EventManager().Events()) == 0 {
			t.Fatalf("No events were triggered")
		}

		// The header creation event should have been the one that was first generated
		if ctx.EventManager().Events()[0].Type != headerInsertedType.Type {
			t.Errorf("The first event does not have the BTCHeaderInserted type")
		}

		// If the new header builds on top of the tip
		if oldTip.Eq(parentHeader) {
			// The new tip should be equal to the new header
			if !newTip.Eq(header) {
				t.Errorf("Inserted header builts on top of the previous tip but does not become the new tip")
			}
			// The rollforward hook must be sent once
			if len(mockHooks.AfterBTCRollForwardStore) != 1 {
				t.Fatalf("Expected a single BTCRollForward hook to be invoked. Got %d", len(mockHooks.AfterBTCRollForwardStore))
			}
			// The rollfoward hook call should contain the new header
			if !mockHooks.AfterBTCRollForwardStore[0].Eq(header) {
				t.Errorf("The headerInfo inside the BTCRollForward hook is not the new header")
			}
			// No rollback hooks must be invoked
			if len(mockHooks.AfterBTCRollBackStore) != 0 {
				t.Fatalf("Expected the BTCRollBack hook to not be invoked")
			}
			// 2 events because the first one is for the header creation
			if len(ctx.EventManager().Events()) != 2 {
				t.Fatalf("We expected only two events. One for header creation and one for rolling forward.")
			}
			// The second event should be the roll forward one
			if ctx.EventManager().Events()[1].Type != rollForwadType.Type {
				t.Errorf("The second event does not have the roll forward type")
			}
		} else if oldTip.Work.GT(*header.Work) {
			// If the tip has a greater work than the newly inserted header
			// no events should be sent and the tip should not change
			if !oldTip.Eq(newTip) {
				t.Errorf("Header with less work inserted but the tip changed")
			}
			// No rollforward hooks should be invoked
			if len(mockHooks.AfterBTCRollForwardStore) != 0 {
				t.Fatalf("Expected the BTCRollForward hook to not be invoked")
			}
			// No rollback hooks should be invoked
			if len(mockHooks.AfterBTCRollBackStore) != 0 {
				t.Fatalf("Expected the BTCRollBack hook to not be invoked")
			}
			// No other events other than BTCHeaderInserted should be invoked
			if len(ctx.EventManager().Events()) != 1 {
				t.Errorf("Extra events have been invoked when the tip wasn't updated")
			}
		} else {
			// The tip has been updated. It should be towards the new header
			if !newTip.Eq(header) {
				t.Errorf("Inserted header has more work than the previous tip but does not become the new tip")
			}
			// Get the highest common ancestor of the old tip and the new header
			hca := blcKeeper.HeadersState(ctx).GetHighestCommonAncestor(header, oldTip)
			// Get the ancestry of the header up to the highest common ancestor
			ancestry := tree.GetNodeAncestryUpTo(header, hca)
			// We should have as many invocations of the roll-forward hook as the ancestors
			// up to the highest common one
			if len(ancestry) != len(mockHooks.AfterBTCRollForwardStore) {
				t.Fatalf("Expected as many invocations of the roll-forward hook as the number of ancestors.")
			}
			for i := 0; i < len(ancestry); i++ {
				// Compare the nodes in reverse order, since the rollfoward events should be in an oldest header first manner.
				if !ancestry[i].Eq(mockHooks.AfterBTCRollForwardStore[len(ancestry)-i-1]) {
					t.Errorf("Headers do not match. Expected %s got %s", ancestry[i].Hash, mockHooks.AfterBTCRollForwardStore[len(ancestry)-i-1].Hash)
				}
			}

			// The rollback hook should be invoked for the highest common ancestor
			if len(mockHooks.AfterBTCRollBackStore) != 1 {
				t.Fatalf("Expected the BTCRollBack hook to be invoked once")
			}
			if !mockHooks.AfterBTCRollBackStore[0].Eq(hca) {
				t.Errorf("Expected the BTCRollBack hook to be invoked for the highest common ancestor")
			}

			// Test the invoked events
			invokedEvents := ctx.EventManager().Events()
			// There should be a total of len(ancestry) + 2 events
			if len(invokedEvents) != len(ancestry)+2 {
				t.Errorf("More events than expected were invoked %d %d", len(invokedEvents), len(ancestry)+2)
			}
			// Only test that there is a certain number of rollForward and rollBack events
			// Testing the attributes is a much more complex approach
			rollForwardCnt := 0
			rollBackCnt := 0
			for i := 0; i < len(invokedEvents); i++ {
				if invokedEvents[i].Type == rollForwadType.Type {
					rollForwardCnt += 1
				}
				if invokedEvents[i].Type == rollBackType.Type {
					rollBackCnt += 1
				}
			}
			if rollForwardCnt != len(ancestry) || rollBackCnt != 1 {
				t.Errorf("Wrong number of roll forward and roll back events")
			}
		}
	})
}
