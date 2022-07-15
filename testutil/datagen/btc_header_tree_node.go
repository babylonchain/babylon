package datagen

import (
	btclightclienttypes "github.com/babylonchain/babylon/x/btclightclient/types"
)

type BTCHeaderTreeNode struct {
	Header   *btclightclienttypes.BTCHeaderInfo
	Parent   *BTCHeaderTreeNode
	Children []*BTCHeaderTreeNode
}

func NewBTCHeaderTreeNode(header *btclightclienttypes.BTCHeaderInfo, parent *BTCHeaderTreeNode) *BTCHeaderTreeNode {
	children := make([]*BTCHeaderTreeNode, 0)
	return &BTCHeaderTreeNode{Header: header, Parent: parent, Children: children}
}

// getAncestryUpToUtil recursively iterates the parents of the node until the root node is reached
func (n *BTCHeaderTreeNode) getAncestryUpToUtil(ancestry *[]*BTCHeaderTreeNode, upTo *BTCHeaderTreeNode) {
	if upTo != nil && n.Eq(upTo) {
		return
	}
	*ancestry = append(*ancestry, n)
	if n.Parent != nil {
		n.Parent.getAncestryUpToUtil(ancestry, upTo)
	}
}

// GetHeaderAncestryUpTo returns an ancestry list starting from the tree node and
// leading to the `upTo` parameter if it is not nil.
func (n *BTCHeaderTreeNode) GetHeaderAncestryUpTo(upTo *BTCHeaderTreeNode) []*BTCHeaderTreeNode {
	ancestry := make([]*BTCHeaderTreeNode, 0)
	n.getAncestryUpToUtil(&ancestry, upTo)
	return ancestry
}

// GetHeaderAncestry returns an ancestry list starting from the tree node and
// leading to the root of the tree.
func (n *BTCHeaderTreeNode) GetHeaderAncestry() []*BTCHeaderTreeNode {
	return n.GetHeaderAncestryUpTo(nil)
}

// GetRandomAncestor retrieves the ancestry list and returns an ancestor from it.
// Can include the node itself.
func (n *BTCHeaderTreeNode) GetRandomAncestor() *BTCHeaderTreeNode {
	ancestry := n.GetHeaderAncestry()
	idx := RandomInt(len(ancestry))
	return ancestry[idx]
}

// getNodesUtil recursively iterates the children of a node in order to traverse the entire tree
func (n *BTCHeaderTreeNode) getNodesUtil(nodes *[]*BTCHeaderTreeNode) {
	*nodes = append(*nodes, n)
	for _, node := range n.Children {
		node.getNodesUtil(nodes)
	}
}

// GetNodes returns a list of all the nodes rooted at the tree specified by this node
func (n *BTCHeaderTreeNode) GetNodes() []*BTCHeaderTreeNode {
	nodes := make([]*BTCHeaderTreeNode, 0)
	n.getNodesUtil(&nodes)
	return nodes
}

// Eq checks whether two BTCHeaderTreeNode instances are equal
func (n *BTCHeaderTreeNode) Eq(other *BTCHeaderTreeNode) bool {
	return n.Header.Eq(other.Header)
}

// InsertChild inserts a child into the children list
func (n *BTCHeaderTreeNode) InsertChild(child *BTCHeaderTreeNode) {
	n.Children = append(n.Children, child)
}
