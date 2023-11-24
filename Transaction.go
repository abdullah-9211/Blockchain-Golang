package main

// Type Transaction holds all the information about a single transaction
type Transaction struct {
	Value    string
	Not_Null bool
}

func (transaction Transaction) hashed() Hash {
	return hash_string(transaction.Value)
}
