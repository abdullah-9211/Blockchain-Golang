package main

import "fmt"

type Blockchain struct {
	Blocks       map[Hash]Block
	Merkel_Trees map[Hash]MerkelTree
}

func (blockchain *Blockchain) __calc_block_length_values() *map[Hash]int {
	dp := make(map[Hash]int)
	for block_hash := range blockchain.Blocks {
		dp[block_hash] = -2
	}
	var calc func(block *Block) int
	calc = func(block *Block) int {
		block_hash := block.hashed()
		if dp[block_hash] == -2 {
			if block.Prev_Block == (Hash{}) {
				dp[block_hash] = 1
			} else {
				prev_block, ok := blockchain.Blocks[block.Prev_Block]
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
	for block_hash, block := range blockchain.Blocks {
		dp[block_hash] = calc(&block)
	}
	return &dp
}

func create_blockchain() Blockchain {
	return Blockchain{Blocks: make(map[Hash]Block), Merkel_Trees: make(map[Hash]MerkelTree)}
}

func (blockchain *Blockchain) add_block(block Block) {
	block_hash := block.hashed()
	blockchain.Blocks[block_hash] = block
}

func (blockchain *Blockchain) remove_block(block Block) {
	block_hash := block.hashed()
	delete(blockchain.Blocks, block_hash)
}

func (blockchain *Blockchain) add_merkel_tree(merkel_tree MerkelTree) {
	merkel_tree_hash := merkel_tree.hashed()
	blockchain.Merkel_Trees[merkel_tree_hash] = merkel_tree
}

func (blockchain *Blockchain) remove_merkel_tree(merkel_tree MerkelTree) {
	merkel_tree_hash := merkel_tree.hashed()
	delete(blockchain.Merkel_Trees, merkel_tree_hash)
}

func (blockchain Blockchain) get_last_hash() Hash {
	if len(blockchain.Blocks) == 0 {
		return Hash{}
	}
	dp := blockchain.__calc_block_length_values()
	max_length, best_block := -1, Block{}
	for block_hash, block := range blockchain.Blocks {
		length := (*dp)[block_hash]
		if length > max_length {
			max_length = length
			best_block = block
		}
	}
	return best_block.hashed()
}

func (blockchain Blockchain) is_valid_blocks() bool {
	dp := blockchain.__calc_block_length_values()
	max_length := -1
	for block_hash, block := range blockchain.Blocks {
		if !block.is_valid() {
			return false
		}
		length := (*dp)[block_hash]
		if length > max_length {
			max_length = length
		}
	}
	return max_length == len(blockchain.Blocks)
}

func (blockchain Blockchain) is_valid_merkel_trees() bool {
	for _, block := range blockchain.Blocks {
		merkel_tree, ok := blockchain.Merkel_Trees[block.Merkel_Root]
		if !ok || !merkel_tree.is_valid() {
			return false
		}
	}
	return true
}

func (blockchain Blockchain) is_valid() bool {
	return blockchain.is_valid_blocks() && blockchain.is_valid_merkel_trees()
}

func (blockchain *Blockchain) remove_short_chains() {
	dp := blockchain.__calc_block_length_values()
	max_length, best_block := -1, Block{}
	for block_hash, block := range blockchain.Blocks {
		length := (*dp)[block_hash]
		if length > max_length {
			max_length = length
			best_block = block
		}
	}
	to_keep_blocks := make(map[Hash]Block)
	to_keep_merkel_trees := make(map[Hash]MerkelTree)
	for {
		to_keep_blocks[best_block.hashed()] = best_block
		to_keep_merkel_trees[best_block.Merkel_Root] = blockchain.Merkel_Trees[best_block.Merkel_Root]
		if best_block.Prev_Block == (Hash{}) {
			break
		}
		best_block = blockchain.Blocks[best_block.Prev_Block]
	}
	blockchain.Blocks = to_keep_blocks
	blockchain.Merkel_Trees = to_keep_merkel_trees
}

func (blockchain Blockchain) pretty_print_blocks() string {
	if !blockchain.is_valid_blocks() {
		return "-- Invalid Blockchain --"
	}
	dp := blockchain.__calc_block_length_values()
	max_length, best_block := -1, Block{}
	for block_hash, block := range blockchain.Blocks {
		length := (*dp)[block_hash]
		if length > max_length {
			max_length = length
			best_block = block
		}
	}
	ordered_blocks := make([]Block, max_length)
	for {
		ordered_blocks[max_length-1] = best_block
		if best_block.Prev_Block == (Hash{}) {
			break
		}
		best_block = blockchain.Blocks[best_block.Prev_Block]
		max_length--
	}
	out := "Blocks in order:\n"
	for idx, block := range ordered_blocks {
		block_hash := block.hashed()
		out += fmt.Sprintf("%.2d) Prev Block: %s\n    Merkel Root: %s\n    Nonce: %s\n    Hash: %s\n\n", idx+1, block.Prev_Block.to_string(), block.Merkel_Root.to_string(), block.Nonce.to_string(), block_hash.to_string())
	}
	return out
}

func (blockchain Blockchain) pretty_print() string {
	if !blockchain.is_valid() {
		return "-- Invalid Blockchain --"
	}
	dp := blockchain.__calc_block_length_values()
	max_length, best_block := -1, Block{}
	for block_hash, block := range blockchain.Blocks {
		length := (*dp)[block_hash]
		if length > max_length {
			max_length = length
			best_block = block
		}
	}
	ordered_blocks := make([]Block, max_length)
	for {
		ordered_blocks[max_length-1] = best_block
		if best_block.Prev_Block == (Hash{}) {
			break
		}
		best_block = blockchain.Blocks[best_block.Prev_Block]
		max_length--
	}
	out := "Blocks in order:\n"
	for idx, block := range ordered_blocks {
		block_hash := block.hashed()
		out += fmt.Sprintf("%.2d) Prev Block: %s\n    Merkel Root: %s\n    Nonce: %s\n    Hash: %s\n", idx+1, block.Prev_Block.to_string(), block.Merkel_Root.to_string(), block.Nonce.to_string(), block_hash.to_string())
		merkel_tree := blockchain.Merkel_Trees[block.Merkel_Root]
		out += merkel_tree.pretty_print()
	}
	return out
}
