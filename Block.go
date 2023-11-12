package main

type Block struct {
	Merkel_Root    Hash
	Prev_Block     Hash
	Nonce          Hash
	Trailing_Zeros int
}

func create_block(merkel_root Hash, prev_block Hash, trailing_zeros int) Block {
	return Block{Merkel_Root: merkel_root, Prev_Block: prev_block, Trailing_Zeros: trailing_zeros}
}

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
