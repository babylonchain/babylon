package datagen

import (
	btclightclienttypes "github.com/babylonchain/babylon/x/btclightclient/types"
)

type BTCHeaderTree struct {
	Root      *BTCHeaderTreeNode
	Nodes     []*BTCHeaderTreeNode
	MinHeight uint64
	MaxHeight uint64
}

func NewBTCHeaderTree(root *BTCHeaderTreeNode, minTreeHeight uint64, maxTreeHeight uint64) *BTCHeaderTree {
	return &BTCHeaderTree{Root: root, MinHeight: minTreeHeight, MaxHeight: maxTreeHeight, Nodes: nil}
}

// RandNumChildren randomly generates 0-2 children with the following probabilities:
// If `MinHeight` is more than 1:
// 		1 child:    75%
// 		2 children: 25%
// Otherwise,
// 		0 children: 25%
// 		1 child:    50%
// 		2 children: 25%
func (t *BTCHeaderTree) RandNumChildren() int {
	// Randomly identify the number of children
	numChildren := 0
	// If we have a minimum height > 1, then we need to generate a child for sure
	if t.MinHeight > 1 {
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
	return numChildren
}

// GenRandomBTCHeaderInfoTree recursively generates a random tree of BTCHeaderInfo objects rooted at `root`.
// The tree generation is accomplished by randomly selecting the number of children using the `RandNumChildren()`.
// Then, for each child, a random BTCHeaderInfo object is generated and a new tree rooted
// at that child is recursively generated.
// For each node that is generated, the callback function is invoked in order to identify
// whether this node should be included in the tree or not. For example, a node might not be included in a tree
// due to its hash already existing (a very rare event). In the case that a node is rejected, we retry to generate
// that node up to a particular limit.
func (t *BTCHeaderTree) GenRandomBTCHeaderInfoTree(callback func(*btclightclienttypes.BTCHeaderInfo) bool) {
	if t.MaxHeight == 1 {
		// We are already at the maximum depth, do not generate any children
		return
	}
	const maxRetries = 3
	retries := 0
	for i := 0; i < t.RandNumChildren(); i++ {
		childInfo := GenRandomBTCHeaderInfoWithParent(t.Root.Header)
		childNode := NewBTCHeaderTreeNode(childInfo, t.Root)
		// Only generate `minHeight-1` subtrees for the first child
		childMinHeight := uint64(0)
		if i != 0 && t.MinHeight-1 > 0 {
			childMinHeight = t.MinHeight - 1
		}

		childTree := NewBTCHeaderTree(childNode, childMinHeight, t.MaxHeight-1)
		// Callback returns `true` if the child is ok for the purposes of the test
		if !callback(childInfo) {
			// Only retry three times
			if retries != maxRetries {
				i -= 1 // Regenerate this child
			}
			retries += 1
			continue
		}
		// All good, add it to children list
		t.Root.InsertChild(childNode)
		childTree.GenRandomBTCHeaderInfoTree(callback)
	}
}

// GetNodes returns an in-order list of the tree nodes
func (t *BTCHeaderTree) GetNodes() []*BTCHeaderTreeNode {
	if t.Nodes != nil {
		return t.Nodes
	}
	t.Nodes = t.Root.GetNodes()
	return t.Nodes
}

// GetTip returns the header in the tree with the most work
func (t *BTCHeaderTree) GetTip() *BTCHeaderTreeNode {
	// We can traverse the tree as is, but GetNodes() already does this for us
	// with caching, so an extra iteration is worth the reduced complexity.
	tip := t.Root
	nodes := t.GetNodes()
	for _, node := range nodes {
		if node.Header.Work.GT(*tip.Header.Work) {
			tip = node
		}
	}
	return tip
}

// GetMainChain returns the tree fork with the most work
func (t *BTCHeaderTree) GetMainChain() []*BTCHeaderTreeNode {
	nodeMostWork := t.GetTip()
	return nodeMostWork.GetHeaderAncestry()
}

// SelectRandomHeader selects a random header from the list of nodes
func (t *BTCHeaderTree) SelectRandomHeader() *BTCHeaderTreeNode {
	nodes := t.GetNodes()
	idx := RandomInt(len(nodes))
	return nodes[idx]
}
