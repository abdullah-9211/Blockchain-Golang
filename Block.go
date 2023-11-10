package main

type Block struct {
	__merkel_root    Hash
	__prev_block     Hash
	__nonce          Hash
	__trailing_zeros int
}

func create_block(merkel_root Hash, prev_block Hash, trailing_zeros int) *Block {
	return &Block{__merkel_root: merkel_root, __prev_block: prev_block, __trailing_zeros: trailing_zeros}
}

func (block *Block) mine() {
	for {
		current_hash := concat_hash(block.__prev_block, block.__merkel_root, block.__nonce)
		if current_hash.trailing_zeros() >= block.__trailing_zeros {
			break
		}
		block.__nonce = random_hash()
	}
}

func (block *Block) hashed() Hash {
	return concat_hash(block.__prev_block, block.__merkel_root, block.__nonce)
}

func (block *Block) is_valid() bool {
	current_hash := block.hashed()
	return current_hash.trailing_zeros() >= block.__trailing_zeros
}
