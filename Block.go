package main

import "fmt"

// Type Block holds the information for a single block
type Block struct {
	Merkel_Root    Hash
	Prev_Block     Hash
	Nonce          Hash
	Trailing_Zeros int
}

func create_block(merkel_root Hash, prev_block Hash, trailing_zeros int) Block {
	return Block{Merkel_Root: merkel_root, Prev_Block: prev_block, Trailing_Zeros: trailing_zeros}
}

// Block's function mine randomly searches and sets a nonce that gives the given number of trailing
// zeros when Merkel_Root, Prev_Block, and Nonce are hashed together
func (block *Block) mine() {
	for {
		current_hash := concat_hash(block.Prev_Block, block.Merkel_Root, block.Nonce)
		if current_hash.trailing_zeros() >= block.Trailing_Zeros {
			break
		}
		block.Nonce = random_hash()
	}
}

func (block Block) hashed() Hash {
	return concat_hash(block.Prev_Block, block.Merkel_Root, block.Nonce)
}

func (block Block) is_valid() bool {
	current_hash := block.hashed()
	return current_hash.trailing_zeros() >= block.Trailing_Zeros
}

func (block Block) pretty_print() string {
	return fmt.Sprintf("Merkel_Root: %v\nPrev_Block: %v\nNonce: %v\nTrailing_Zeros: %v\n",
		block.Merkel_Root.to_string(),
		block.Prev_Block.to_string(),
		block.Nonce.to_string(),
		block.Trailing_Zeros)
}

// Merkel_Root    Hash
// Prev_Block     Hash
// Nonce          Hash
// Trailing_Zeros int
