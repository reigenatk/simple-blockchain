package main

import "crypto/sha256"

type MerkleTree struct {
	RootNode *MerkleNode
}

// each node of the Merkle tree has a left/right node, as well as some value
// which is the hash of the concatenated values of its left and right nodes.
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Hash  []byte
}

// given two MerkleNodes and some data to add into the hash
// this function returns a new MerkleNode
// put in their words:
// Every node contains some data.
// When a node is a leaf, the data is passed from the outside
// (a serialized transaction in our case).
// When a node is linked to other nodes, it takes their
// data and concatenates and hashes it.
func NewMerkleNode(left, right *MerkleNode, data []byte) MerkleNode {
	ret := MerkleNode{}

	// so if left and right both have no data, just hash the data
	if left == nil && right == nil {
		hashval := sha256.Sum256(data)
		ret.Hash = hashval[:]
	} else {
		// otherwise append the data from left and right
		// then hash
		prevHashes := append(left.Hash, right.Hash...)
		hashval := sha256.Sum256(prevHashes)
		ret.Hash = hashval[:]
	}

	ret.Left = left
	ret.Right = right

	return ret
}

// we pass in a list of 32-byte transaction IDs (hence the 2D byte slice)
func NewMerkleTree(data [][]byte) MerkleTree {
	var LeafNodes []MerkleNode
	for _, bytes := range data {
		LeafNodes = append(LeafNodes, NewMerkleNode(nil, nil, bytes))
	}

	// Merkle Tree must have even number of leaf nodes
	// so if its odd, duplicate the last one
	if len(LeafNodes)%2 != 0 {
		LeafNodes = append(LeafNodes, LeafNodes[len(LeafNodes)-1:]...)
	}

	for len(LeafNodes) != 1 {
		var nextLevel []MerkleNode
		for i := 0; i < len(LeafNodes)/2; i++ {

			newNode := NewMerkleNode(&LeafNodes[2*i], &LeafNodes[2*i+1], []byte{})
			nextLevel = append(nextLevel, newNode)
		}
		LeafNodes = nextLevel
	}
	return MerkleTree{
		RootNode: &LeafNodes[0],
	}
}
