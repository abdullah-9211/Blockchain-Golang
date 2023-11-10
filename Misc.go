package main

import (
	"crypto/rand"
	"math/big"
)

func random_string(n int) string {
	arr := make([]byte, n)
	for i := 0; i < n; i++ {
		tmp, _ := rand.Int(rand.Reader, big.NewInt(25))
		arr[i] = byte(tmp.Int64() + 65)
	}
	return string(arr)
}

func random_hash() Hash {
	hash := Hash{}
	for i := 0; i < 32; i++ {
		tmp, _ := rand.Int(rand.Reader, big.NewInt((1<<8)-1))
		hash.__value[i] = byte(tmp.Int64())
	}
	return hash
}
