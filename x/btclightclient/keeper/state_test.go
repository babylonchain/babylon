package keeper_test

import (
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
		 - No need to create a tree, since this function does not considers existence or the chain,
		   it just inserts into state and updates the tip based on a simple work comparison.
	*/
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		t.Skip()
	})
}

func FuzzHeadersStateCreateTip(f *testing.F) {
	/*
		Checks:
		1. The `headerInfo` object passed is set as the tip.

		Data generation:
		- Create two headers:
			1. A header that will be set as the tip.
			2. A header that will override it.
	*/
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		t.Skip()
	})
}

func FuzzHeadersStateGetHeader(f *testing.F) {
	/*
		Checks:
		1. If the header specified by a height and a hash does not exist, (nil, error) is returned
		2. If the header specified by a height and a hash exists, (headerInfo, nil) is returned

		Data generation:
		- Create a header and store it using the `CreateHeader` method. Do retrievals using
			* (height, hash)
			* (height, wrongHash)
			* (wrongHeight, hash)
			* (wrongHeight, wrongHash)
	*/
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		t.Skip()
	})
}

func FuzzHeadersStateGetHeaderHeight(f *testing.F) {
	/*
		Checks:
		1. If the header specified by the hash does not exist, (nil, error) is returned
		2. If the header specified by the hash exists, (height, nil) is returned

		Data generation:
		- Create a header and store it using the `CreateHeader` method. Do retrievals:
			* (hash)
			* (wrongHash)
	*/
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		t.Skip()
	})
}

func FuzzHeadersStateGetHeaderWork(f *testing.F) {
	/*
		 Checks:
		 1. If the header specified by the hash does not exist, (nil, error) is returned
		 2. If the header specified by the hash exists, (work, nil) is returned.

		 Data generation:
		 - Create a header and store it using the `CreateHeader` method. Do retrievals:
			* (hash)
			* (wrongHash)
	*/
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		t.Skip()
	})
}

func FuzzHeadersStateGetHeaderByHash(f *testing.F) {
	/*
		Checks:
		1. If the header specified by the hash does not exist (nil, error) is returned
		2. If the header specified by the hash exists (headerInfo, nil) is returned.

		Data generation:
		- Create a header and store it using the `CreateHeader` method. Do retrievals:
			* (hash)
			* (wrongHash)
	*/
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		t.Skip()
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
		t.Skip()
	})
}

func FuzzHeadersStateGetTip(f *testing.F) {
	/*
		Checks:
		1. If the tip does not exist, nil is returned.
		2. The element maintained in the tip storage is returned.

		Data generation:
		- Generate two headers and store them using the `CreateTip` method.
			- First to be inserted first
			- Second to override the first
	*/
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		t.Skip()
	})
}

func FuzzHeadersStateGetHeadersByHeight(f *testing.F) {
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
		t.Skip()
	})
}

func FuzzHeadersStateGetMainChainUpTo(f *testing.F) {
	/*
		Checks:
		1. If the tip does not exist, we have no headers, so an empty list is returned.
		2. We get the main chain containing `depth + 1` elements.

		Data generation:
		- Generate a random tree and retrieve the main chain from it.
		- Randomly generate the depth as `rand.Intn(tipHeight - baseHeight)`.
	*/
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		t.Skip()
	})
}

func FuzzHeadersStateGetMainChain(f *testing.F) {
	/*
		Checks:
		1. We get the entire main chain.

		Data generation:
		- Generate a random tree and retrieve the main chain from it.
	*/
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		t.Skip()
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
		t.Skip()
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
		t.Skip()
	})
}

func FuzzHeadersStateHeaderExists(f *testing.F) {
	/*
		Checks:
		- Returns false if the header passed is nil.
		- Returns true/false depending on the existence of the header.

		Data generation:
		- Generate a header and insert it into storage using `CreateHeader`.
	*/
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		t.Skip()
	})
}

func FuzzHeadersStateTipExists(f *testing.F) {
	/*
		Checks:
		- Returns true/false depending on the existence of a tip.

		Data generation:
		- Generate a header and insert it into storage using `CreateTip`.
	*/
	f.Add(int64(42))
	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)
		t.Skip()
	})
}
