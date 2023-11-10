package main

import "crypto/sha256"

type Transaction struct {
	__value    string
	__not_null bool
}

func (transaction *Transaction) hashed() Hash {
	return Hash{__value: sha256.Sum256([]byte(transaction.__value))}
}
