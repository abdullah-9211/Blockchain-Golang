package main

import (
	"fmt"
	"strings"
)

type MerkelTreeNode struct {
	__transaction Transaction
	__self_hash   Hash
	__left_child  Hash
	__right_child Hash
}

func (merkel_tree_node *MerkelTreeNode) hashed() Hash {
	if merkel_tree_node.__transaction.__not_null {
		return merkel_tree_node.__transaction.hashed()
	} else {
		return concat_hash(merkel_tree_node.__left_child, merkel_tree_node.__right_child)
	}
}

type MerkelTree struct {
	__transactions map[Hash]Transaction
	__tree         []MerkelTreeNode
	__is_built     bool
}

func create_merkel_tree() *MerkelTree {
	return &MerkelTree{__transactions: make(map[Hash]Transaction)}
}

func (merkel_tree *MerkelTree) hashed() Hash {
	if !merkel_tree.__is_built {
		merkel_tree.build()
	}
	return merkel_tree.__tree[0].__self_hash
}

func (merkel_tree *MerkelTree) add_transaction(transaction Transaction) bool {
	transaction_hash := transaction.hashed()
	_, ok := merkel_tree.__transactions[transaction_hash]
	if ok {
		return false
	}
	merkel_tree.__transactions[transaction_hash] = transaction
	merkel_tree.__is_built = false
	return true
}

func (merkel_tree *MerkelTree) remove_transaction(transaction Transaction) bool {
	transaction_hash := transaction.hashed()
	_, ok := merkel_tree.__transactions[transaction_hash]
	if !ok {
		return false
	}
	delete(merkel_tree.__transactions, transaction_hash)
	merkel_tree.__is_built = false
	return true
}

func (merkel_tree *MerkelTree) get_transactions() map[Hash]Transaction {
	return merkel_tree.__transactions
}

func (merkel_tree *MerkelTree) build() {
	if merkel_tree.__is_built {
		return
	}
	tree_base_size := 1
	for tree_base_size < len(merkel_tree.__transactions) {
		tree_base_size <<= 1
	}
	merkel_tree.__tree = make([]MerkelTreeNode, 2*tree_base_size-1)
	idx := tree_base_size - 1
	for _, transaction := range merkel_tree.__transactions {
		merkel_tree.__tree[idx] = MerkelTreeNode{__transaction: transaction}
		idx++
	}
	for ; idx < 2*tree_base_size-1; idx++ {
		merkel_tree.__tree[idx] = merkel_tree.__tree[idx-1]
	}
	for i := 2*tree_base_size - 2; i >= 0; i-- {
		if i < tree_base_size-1 {
			merkel_tree.__tree[i].__left_child = merkel_tree.__tree[i*2+1].__self_hash
			merkel_tree.__tree[i].__right_child = merkel_tree.__tree[i*2+2].__self_hash
		}
		merkel_tree.__tree[i].__self_hash = merkel_tree.__tree[i].hashed()
	}
	merkel_tree.__is_built = true
}

func (merkel_tree *MerkelTree) is_valid() bool {
	if !merkel_tree.__is_built {
		return false
	}
	non_leaf := (len(merkel_tree.__tree) - 3) / 2
	for i := 0; i < len(merkel_tree.__tree); i++ {
		if merkel_tree.__tree[i].__self_hash != merkel_tree.__tree[i].hashed() {
			return false
		}
		if i <= non_leaf &&
			(merkel_tree.__tree[i].__left_child != merkel_tree.__tree[i*2+1].__self_hash ||
				merkel_tree.__tree[i].__right_child != merkel_tree.__tree[i*2+2].__self_hash) {
			return false
		}
	}
	return true
}

func (merkel_tree *MerkelTree) pretty_print() string {
	out := ""
	leaves := (len(merkel_tree.__tree) + 1) / 2
	space, nodes_on_depth, idx, to_subtract := 4*leaves-2, 1, 0, 2*leaves
	out += "> Merkel Tree Node Hashes (transaction string after colon for leaf nodes):\n"
	for idx, merkel_node := range merkel_tree.__tree {
		out += fmt.Sprintf("%.2d) %s", idx+1, merkel_node.__self_hash.to_string())
		if idx >= leaves-1 {
			out += fmt.Sprintf(":%s", merkel_node.__transaction.__value)
		}
		out += "\n"
	}
	out += "> Merkel Tree Hierarchy:\n"
	for idx < len(merkel_tree.__tree) {
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
