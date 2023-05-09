package keeper_test

import (
	"fmt"
	"github.com/babylonchain/babylon/testutil/datagen"
	"github.com/babylonchain/babylon/x/btclightclient/keeper"
	"github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math/rand"
)

// Mock hooks interface
// We create a custom type of hooks that just store what they were called with
// to aid with testing.
var _ types.BTCLightClientHooks = &MockHooks{}

type MockHooks struct {
	AfterBTCRollForwardStore    []*types.BTCHeaderInfo
	AfterBTCRollBackStore       []*types.BTCHeaderInfo
	AfterBTCHeaderInsertedStore []*types.BTCHeaderInfo
}

func NewMockHooks() *MockHooks {
	rollForwardStore := make([]*types.BTCHeaderInfo, 0)
	rollBackwardStore := make([]*types.BTCHeaderInfo, 0)
	headerInsertedStore := make([]*types.BTCHeaderInfo, 0)
	return &MockHooks{
		AfterBTCRollForwardStore:    rollForwardStore,
		AfterBTCRollBackStore:       rollBackwardStore,
		AfterBTCHeaderInsertedStore: headerInsertedStore,
	}
}

func (m *MockHooks) AfterBTCRollForward(_ sdk.Context, headerInfo *types.BTCHeaderInfo) {
	m.AfterBTCRollForwardStore = append(m.AfterBTCRollForwardStore, headerInfo)
}

func (m *MockHooks) AfterBTCRollBack(_ sdk.Context, headerInfo *types.BTCHeaderInfo) {
	m.AfterBTCRollBackStore = append(m.AfterBTCRollBackStore, headerInfo)
}

func (m *MockHooks) AfterBTCHeaderInserted(_ sdk.Context, headerInfo *types.BTCHeaderInfo) {
	m.AfterBTCHeaderInsertedStore = append(m.AfterBTCHeaderInsertedStore, headerInfo)
}

// Methods for generating trees

// genRandomTree generates a tree of headers. It accomplishes this by generating a root
// which will serve as the base header and then invokes the `genRandomTreeWithRoot` utility.
// The `minTreeHeight` and `maxTreeHeight` parameters denote the minimum and maximum height
// of the tree that is generated. For example, a `minTreeHeight` of 1,
// means that the tree should have at least one node (the root), while
// a `maxTreeHeight` of 4, denotes that the maximum height of the tree should be 4.
func genRandomTree(r *rand.Rand, k *keeper.Keeper, ctx sdk.Context, minHeight uint64, maxHeight uint64) *datagen.BTCHeaderTree {
	tree := datagen.NewBTCHeaderTree()
	// Generate the root for the tree
	root := datagen.GenRandomBTCHeaderInfo(r)
	tree.Add(root, nil)
	k.SetBaseBTCHeader(ctx, *root)

	genRandomTreeWithParent(r, k, ctx, minHeight-1, maxHeight-1, root, tree)

	return tree
}

// genRandomTreeWithParent is a utility function for inserting the headers
// While the tree is generated, the headers that are generated for it are inserted into storage.
func genRandomTreeWithParent(r *rand.Rand, k *keeper.Keeper, ctx sdk.Context, minHeight uint64,
	maxHeight uint64, root *types.BTCHeaderInfo, tree *datagen.BTCHeaderTree) {

	if minHeight > maxHeight {
		panic("Min height more than max height")
	}

	tree.GenRandomBTCHeaderTree(r, minHeight, maxHeight, root, func(header *types.BTCHeaderInfo) bool {
		err := k.InsertHeader(ctx, header.Header)
		if err != nil {
			panic(fmt.Sprintf("header insertion failed: %s", err))
		}
		return true
	})
}
