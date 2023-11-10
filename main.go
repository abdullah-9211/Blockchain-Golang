package main

import "fmt"

func main() {

	ts := make([]Transaction, 100)
	for i := 0; i < 100; i++ {
		ts[i].__value = random_string(10)
		ts[i].__not_null = true
	}

	blockchain := create_blockchain()

	last_hash := Hash{}
	for i := 0; i < 5; i++ {

		merkel_tree := create_merkel_tree()
		for j := 0; j < 4; j++ {
			merkel_tree.add_transaction(ts[i*4+j])
		}
		merkel_tree.build()
		block := create_block(merkel_tree.hashed(), last_hash, 12)
		block.mine()
		last_hash = block.hashed()
		blockchain.add_block(*block)
		blockchain.add_merkel_tree(*merkel_tree)
	}

	fmt.Println(blockchain.pretty_print())
}
