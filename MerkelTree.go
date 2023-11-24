package main

import (
	"fmt"
	"strings"
)

// type MerkelTreeNode holds information about either a leaf or non leaf node of a merkel tree
type MerkelTreeNode struct {
	Transaction Transaction
	Self_Hash   Hash
	Left_Child  Hash
	Right_Child Hash
}

func (merkel_tree_node MerkelTreeNode) hashed() Hash {
	if merkel_tree_node.Transaction.Not_Null {
		return merkel_tree_node.Transaction.hashed()
	} else {
		return concat_hash(merkel_tree_node.Left_Child, merkel_tree_node.Right_Child)
	}
}

// Type MerkelTree stores a list of transactions in form of an tree in array form
type MerkelTree struct {
	Transactions map[Hash]Transaction
	Tree         []MerkelTreeNode
	Is_Built     bool
}

func create_merkel_tree() MerkelTree {
	return MerkelTree{Transactions: make(map[Hash]Transaction)}
}

func (merkel_tree MerkelTree) hashed() Hash {
	if !merkel_tree.Is_Built {
		merkel_tree.build()
	}
	return merkel_tree.Tree[0].Self_Hash
}

func (merkel_tree *MerkelTree) add_transaction(transaction Transaction) bool {
	transaction_hash := transaction.hashed()
	_, ok := merkel_tree.Transactions[transaction_hash]
	if ok {
		return false
	}
	merkel_tree.Transactions[transaction_hash] = transaction
	merkel_tree.Is_Built = false
	return true
}

func (merkel_tree *MerkelTree) remove_transaction(transaction Transaction) bool {
	transaction_hash := transaction.hashed()
	_, ok := merkel_tree.Transactions[transaction_hash]
	if !ok {
		return false
	}
	delete(merkel_tree.Transactions, transaction_hash)
	merkel_tree.Is_Built = false
	return true
}

func (merkel_tree MerkelTree) get_transactions() map[Hash]Transaction {
	return merkel_tree.Transactions
}

// MerkelTree's function build creates the tree after all the transactions have been added
func (merkel_tree *MerkelTree) build() {
	if merkel_tree.Is_Built {
		return
	}
	tree_base_size := 1
	for tree_base_size < len(merkel_tree.Transactions) {
		tree_base_size <<= 1
	}
	merkel_tree.Tree = make([]MerkelTreeNode, 2*tree_base_size-1)
	idx := tree_base_size - 1
	for _, transaction := range merkel_tree.Transactions {
		merkel_tree.Tree[idx] = MerkelTreeNode{Transaction: transaction}
		idx++
	}
	for ; idx < 2*tree_base_size-1; idx++ {
		merkel_tree.Tree[idx] = merkel_tree.Tree[idx-1]
	}
	for i := 2*tree_base_size - 2; i >= 0; i-- {
		if i < tree_base_size-1 {
			merkel_tree.Tree[i].Left_Child = merkel_tree.Tree[i*2+1].Self_Hash
			merkel_tree.Tree[i].Right_Child = merkel_tree.Tree[i*2+2].Self_Hash
		}
		merkel_tree.Tree[i].Self_Hash = merkel_tree.Tree[i].hashed()
	}
	merkel_tree.Is_Built = true
}

// MerkelTree function is_valid determines whether the built merkel tree is valid.
//
// For a MerkelTree to be valid, the tree should be a perfect binary tree and hold
// the property that all non leaf nodes have the hash value of their children's hash
func (merkel_tree MerkelTree) is_valid() bool {
	if !merkel_tree.Is_Built {
		return false
	}
	non_leaf := (len(merkel_tree.Tree) - 3) / 2
	for i := 0; i < len(merkel_tree.Tree); i++ {
		if merkel_tree.Tree[i].Self_Hash != merkel_tree.Tree[i].hashed() {
			return false
		}
		if i <= non_leaf &&
			(merkel_tree.Tree[i].Left_Child != merkel_tree.Tree[i*2+1].Self_Hash ||
				merkel_tree.Tree[i].Right_Child != merkel_tree.Tree[i*2+2].Self_Hash) {
			return false
		}
	}
	return true
}

func (merkel_tree MerkelTree) pretty_print() string {
	out := ""
	leaves := (len(merkel_tree.Tree) + 1) / 2
	space, nodes_on_depth, idx, to_subtract := 4*leaves-2, 1, 0, 2*leaves
	out += "> Merkel Tree Node Hashes (transaction string after colon for leaf nodes):\n"
	for idx, merkel_node := range merkel_tree.Tree {
		out += fmt.Sprintf("%.2d) %s", idx+1, merkel_node.Self_Hash.to_string())
		if idx >= leaves-1 {
			out += fmt.Sprintf(":%s", merkel_node.Transaction.Value)
		}
		out += "\n"
	}
	out += "> Merkel Tree Hierarchy:\n"
	for idx < len(merkel_tree.Tree) {
		out += strings.Repeat(" ", space-to_subtract)
		for i := 0; i < nodes_on_depth; i++ {
			out += fmt.Sprintf("%.2d", idx+1)
			if i+1 < nodes_on_depth {
				out += strings.Repeat(" ", space)
			}
			idx++
		}
		out += "\n"
		space -= to_subtract
		to_subtract >>= 1
		nodes_on_depth <<= 1
	}
	out += "\n"
	return out
}
