package main

import "fmt"

type Blockchain struct {
	__blocks       map[Hash]Block
	__merkel_trees map[Hash]MerkelTree
}

func (blockchain *Blockchain) __calc_block_length_values() *map[Hash]int {
	dp := make(map[Hash]int)
	for block_hash := range blockchain.__blocks {
		dp[block_hash] = -2
	}
	var calc func(block *Block) int
	calc = func(block *Block) int {
		block_hash := block.hashed()
		if dp[block_hash] == -2 {
			if block.__prev_block == (Hash{}) {
				dp[block_hash] = 1
			} else {
				prev_block, ok := blockchain.__blocks[block.__prev_block]
				length := -1
				if ok {
					length = calc(&prev_block)
				}
				if length == -1 {
					dp[block_hash] = -1
				} else {
					dp[block_hash] = length + 1
				}
			}
		}
		return dp[block_hash]
	}
	for block_hash, block := range blockchain.__blocks {
		dp[block_hash] = calc(&block)
	}
	return &dp
}

func create_blockchain() *Blockchain {
	return &Blockchain{__blocks: make(map[Hash]Block), __merkel_trees: make(map[Hash]MerkelTree)}
}

func (blockchain *Blockchain) add_block(block Block) {
	block_hash := block.hashed()
	blockchain.__blocks[block_hash] = block
}

func (blockchain *Blockchain) remove_block(block Block) {
	block_hash := block.hashed()
	delete(blockchain.__blocks, block_hash)
}

func (blockchain *Blockchain) add_merkel_tree(merkel_tree MerkelTree) {
	merkel_tree_hash := merkel_tree.hashed()
	blockchain.__merkel_trees[merkel_tree_hash] = merkel_tree
}

func (blockchain *Blockchain) remove_merkel_tree(merkel_tree MerkelTree) {
	merkel_tree_hash := merkel_tree.hashed()
	delete(blockchain.__merkel_trees, merkel_tree_hash)
}

func (blockchain *Blockchain) is_valid_blocks() bool {
	dp := blockchain.__calc_block_length_values()
	max_length := -1
	for block_hash, block := range blockchain.__blocks {
		if !block.is_valid() {
			return false
		}
		length := (*dp)[block_hash]
		if length > max_length {
			max_length = length
		}
	}
	return max_length == len(blockchain.__blocks)
}

func (blockchain *Blockchain) is_valid_merkel_trees() bool {
	for _, block := range blockchain.__blocks {
		merkel_tree, ok := blockchain.__merkel_trees[block.__merkel_root]
		if !ok || !merkel_tree.is_valid() {
			return false
		}
	}
	return true
}

func (blockchain *Blockchain) is_valid() bool {
	return blockchain.is_valid_blocks() && blockchain.is_valid_merkel_trees()
}

func (blockchain *Blockchain) remove_short_chains() {
	dp := blockchain.__calc_block_length_values()
	max_length, best_block := -1, Block{}
	for block_hash, block := range blockchain.__blocks {
		length := (*dp)[block_hash]
		if length > max_length {
			max_length = length
			best_block = block
		}
	}
	to_keep := make(map[Hash]Block)
	for {
		to_keep[best_block.hashed()] = best_block
		if best_block.__prev_block == (Hash{}) {
			break
		}
		best_block = blockchain.__blocks[best_block.__prev_block]
	}
	blockchain.__blocks = to_keep
}

func (blockchain *Blockchain) remove_extra_merkel_trees() {
	new_merkel_trees := make(map[Hash]MerkelTree)
	for _, block := range blockchain.__blocks {
		tree, ok := blockchain.__merkel_trees[block.__merkel_root]
		if ok {
			new_merkel_trees[tree.hashed()] = tree
		}
	}
	blockchain.__merkel_trees = new_merkel_trees
}

func (blockchain *Blockchain) pretty_print_blocks() string {
	if !blockchain.is_valid_blocks() {
		return "-- Invalid Blockchain --"
	}
	dp := blockchain.__calc_block_length_values()
	max_length, best_block := -1, Block{}
	for block_hash, block := range blockchain.__blocks {
		length := (*dp)[block_hash]
		if length > max_length {
			max_length = length
			best_block = block
		}
	}
	ordered_blocks := make([]Block, max_length)
	for {
		ordered_blocks[max_length-1] = best_block
		if best_block.__prev_block == (Hash{}) {
			break
		}
		best_block = blockchain.__blocks[best_block.__prev_block]
		max_length--
	}
	out := "Blocks in order:\n"
	for idx, block := range ordered_blocks {
		block_hash := block.hashed()
		out += fmt.Sprintf("%.2d) Prev Block: %s\n    Merkel Root: %s\n    Nonce: %s\n    Hash: %s\n\n", idx+1, block.__prev_block.to_string(), block.__merkel_root.to_string(), block.__nonce.to_string(), block_hash.to_string())
	}
	return out
}

func (blockchain *Blockchain) pretty_print() string {
	if !blockchain.is_valid() {
		return "-- Invalid Blockchain --"
	}
	dp := blockchain.__calc_block_length_values()
	max_length, best_block := -1, Block{}
	for block_hash, block := range blockchain.__blocks {
		length := (*dp)[block_hash]
		if length > max_length {
			max_length = length
			best_block = block
		}
	}
	ordered_blocks := make([]Block, max_length)
	for {
		ordered_blocks[max_length-1] = best_block
		if best_block.__prev_block == (Hash{}) {
			break
		}
		best_block = blockchain.__blocks[best_block.__prev_block]
		max_length--
	}
	out := "Blocks in order:\n"
	for idx, block := range ordered_blocks {
		block_hash := block.hashed()
		out += fmt.Sprintf("%.2d) Prev Block: %s\n    Merkel Root: %s\n    Nonce: %s\n    Hash: %s\n", idx+1, block.__prev_block.to_string(), block.__merkel_root.to_string(), block.__nonce.to_string(), block_hash.to_string())
		merkel_tree := blockchain.__merkel_trees[block.__merkel_root]
		out += merkel_tree.pretty_print()
	}
	return out
}
