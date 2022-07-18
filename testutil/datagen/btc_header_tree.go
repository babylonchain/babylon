package datagen

import (
	blctypes "github.com/babylonchain/babylon/x/btclightclient/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BTCHeaderTree struct {
	// headers is a map of unique identifies to BTCHeaderInfo objects
	headers map[string]*blctypes.BTCHeaderInfo
	// children is a map of unique identifies to unique identifiers for children
	children map[string][]string
}

func NewBTCHeaderTree() *BTCHeaderTree {
	headers := make(map[string]*blctypes.BTCHeaderInfo, 0)
	children := make(map[string][]string, 0)
	return &BTCHeaderTree{headers: headers, children: children}
}

// Add adds a node into storage. If the `parent` is set,
// it is also added to the list of `parent`.
func (t *BTCHeaderTree) Add(node *blctypes.BTCHeaderInfo, parent *blctypes.BTCHeaderInfo) {
	t.headers[node.Hash.String()] = node
	if parent != nil {
		t.children[parent.Hash.String()] = append(t.children[parent.Hash.String()], node.Hash.String())
	}
}

// Contains checks whether a node is maintained in the internal storage
func (t *BTCHeaderTree) Contains(node *blctypes.BTCHeaderInfo) bool {
	if _, ok := t.headers[node.Hash.String()]; ok {
		return true
	}
	return false
}

// GetRoot returns the root of the tree -- i.e. the node without an existing parent
func (t *BTCHeaderTree) GetRoot() *blctypes.BTCHeaderInfo {
	for _, header := range t.headers {
		if t.getParent(header) == nil {
			return header
		}
	}
	return nil
}

// GetTip returns the header in the tree with the most work
func (t *BTCHeaderTree) GetTip() *blctypes.BTCHeaderInfo {
	maxWork := sdk.NewUint(0)
	var tip *blctypes.BTCHeaderInfo
	for _, node := range t.headers {
		if node.Work.GT(maxWork) {
			maxWork = *node.Work
			tip = node
		}
	}
	return tip
}

// GetMainChain returns the tree fork with the most work
func (t *BTCHeaderTree) GetMainChain() []*blctypes.BTCHeaderInfo {
	tip := t.GetTip()
	return t.GetNodeAncestry(tip)
}

// RandNumChildren randomly generates 0-2 children with the following probabilities:
// If zeroChildrenAllowed is not set:
// 		1 child:    75%
// 		2 children: 25%
// Otherwise,
// 		0 children: 25%
// 		1 child:    50%
// 		2 children: 25%
func (t *BTCHeaderTree) RandNumChildren(zeroChildrenAllowed bool) int {
	// Randomly identify the number of children
	numChildren := 0
	// If the flag is not set, then we need to generate a child for sure
	if !zeroChildrenAllowed {
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

// GenRandomBTCHeaderTree recursively generates a random tree of BTCHeaderInfo objects rooted at `parent`.
// The tree generation is accomplished by randomly selecting the number of children using the `RandNumChildren()`.
// Then, for each child, a random BTCHeaderInfo object is generated and a new tree rooted
// at that child is recursively generated.
// For each node that is generated, the callback function is invoked in order to identify
// whether we should continue generating or not as well as help with maintenance
// tasks (e.g. inserting headers into keeper storage).
func (t *BTCHeaderTree) GenRandomBTCHeaderTree(minHeight uint64, maxHeight uint64,
	parent *blctypes.BTCHeaderInfo, callback func(info *blctypes.BTCHeaderInfo) bool) {

	if maxHeight == 0 {
		// If we generate more, we exceed the maximum height
		return
	}

	const maxRetries = 3
	retries := 0
	// Generate the children of the parent
	for i := 0; i < t.RandNumChildren(minHeight <= 1); i++ {
		childInfo := GenRandomBTCHeaderInfoWithParent(parent)

		// Rare occasion that we get the same hash, skip
		if t.Contains(childInfo) {
			// Only retry up to 3 times to generate the child
			if retries < maxRetries {
				i -= 1
			}
			retries += 1
			continue
		}

		// Only generate `minHeight-1` subtrees for the first child
		childMinHeight := uint64(0)
		if i == 0 && minHeight-1 > 0 {
			childMinHeight = minHeight - 1
		}
		if callback(childInfo) {
			t.Add(childInfo, parent)
			t.GenRandomBTCHeaderTree(childMinHeight, maxHeight-1, childInfo, callback)
		}
	}
}

// RandomNode selects a random header from the list of nodes
func (t *BTCHeaderTree) RandomNode() *blctypes.BTCHeaderInfo {
	randIdx := RandomInt(len(t.headers))
	var idx uint64 = 0
	for _, node := range t.headers {
		if idx == randIdx {
			return node
		}
		idx += 1
	}
	return nil
}

// getNodeAncestryUpToUtil recursively iterates the parents of the node until the root node is reached
func (t *BTCHeaderTree) getNodeAncestryUpToUtil(ancestry *[]*blctypes.BTCHeaderInfo,
	node *blctypes.BTCHeaderInfo, upTo *blctypes.BTCHeaderInfo) {

	if upTo != nil && node.Eq(upTo) {
		return
	}
	*ancestry = append(*ancestry, node)
	parent := t.getParent(node)
	if parent != nil {
		t.getNodeAncestryUpToUtil(ancestry, parent, upTo)
	}
}

// GetNodeAncestryUpTo returns an ancestry list starting from the tree node and
// leading to a child of the `upTo` parameter if it is not nil.
func (t *BTCHeaderTree) GetNodeAncestryUpTo(node *blctypes.BTCHeaderInfo,
	upTo *blctypes.BTCHeaderInfo) []*blctypes.BTCHeaderInfo {

	ancestry := make([]*blctypes.BTCHeaderInfo, 0)
	t.getNodeAncestryUpToUtil(&ancestry, node, upTo)
	return ancestry
}

// GetNodeAncestry returns an ancestry list starting from the tree node and
// leading to the root of the tree.
func (t *BTCHeaderTree) GetNodeAncestry(node *blctypes.BTCHeaderInfo) []*blctypes.BTCHeaderInfo {
	return t.GetNodeAncestryUpTo(node, nil)
}

// GetRandomAncestor retrieves the ancestry list and returns an ancestor from it.
// Can include the node itself.
func (t *BTCHeaderTree) GetRandomAncestor(node *blctypes.BTCHeaderInfo) *blctypes.BTCHeaderInfo {
	ancestry := t.GetNodeAncestry(node)
	idx := RandomInt(len(ancestry))
	return ancestry[idx]
}

// IsOnNodeChain returns true or false depending on whether the node
// is equal or a descendant of the `ancestor` parameter.
func (t *BTCHeaderTree) IsOnNodeChain(node *blctypes.BTCHeaderInfo, ancestor *blctypes.BTCHeaderInfo) bool {
	if node.Eq(ancestor) {
		return true
	}
	ancestryUpTo := t.GetNodeAncestryUpTo(node, ancestor)
	lastElement := ancestryUpTo[len(ancestryUpTo)-1]
	parent := t.getParent(lastElement)
	if parent != nil && parent.Eq(ancestor) {
		return true
	}
	return false
}

// GetChildren returns the children of a node as a list of BTCHeaderInfo objects
func (t *BTCHeaderTree) GetChildren(node *blctypes.BTCHeaderInfo) []*blctypes.BTCHeaderInfo {
	if !t.Contains(node) {
		panic("Retrieving children of non existent node")
	}
	childrenHash := t.children[node.Hash.String()]
	children := make([]*blctypes.BTCHeaderInfo, 0)
	for _, childHash := range childrenHash {
		children = append(children, t.headers[childHash])
	}
	return children
}

// getNodeDescendantsUtil recursively iterates the descendants of a node and adds them to a list
func (t *BTCHeaderTree) getNodeDescendantsUtil(descendants *[]*blctypes.BTCHeaderInfo, node *blctypes.BTCHeaderInfo) {
	*descendants = append(*descendants, node)
	for _, child := range t.GetChildren(node) {
		t.getNodeDescendantsUtil(descendants, child)
	}
}

// GetNodeDescendants returns a list of the descendants of a node
func (t *BTCHeaderTree) GetNodeDescendants(node *blctypes.BTCHeaderInfo) []*blctypes.BTCHeaderInfo {
	descendants := make([]*blctypes.BTCHeaderInfo, 0)
	t.getNodeDescendantsUtil(&descendants, node)
	return descendants
}

// RandomDescendant returns a random descendant of the node
func (t *BTCHeaderTree) RandomDescendant(node *blctypes.BTCHeaderInfo) *blctypes.BTCHeaderInfo {
	descendants := t.GetNodeDescendants(node)
	idx := RandomInt(len(descendants))
	return descendants[idx]
}

// getParent returns the parent of the node, or nil if it doesn't exist
func (t *BTCHeaderTree) getParent(node *blctypes.BTCHeaderInfo) *blctypes.BTCHeaderInfo {
	if header, ok := t.headers[node.Header.ParentHash().String()]; ok {
		return header
	}
	return nil
}
