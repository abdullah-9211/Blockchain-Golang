package main

type Transaction struct {
	Value    string
	Not_Null bool
}

func (transaction Transaction) hashed() Hash {
	return hash_string(transaction.Value)
}
