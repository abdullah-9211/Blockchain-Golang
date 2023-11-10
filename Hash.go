package main

import (
	"crypto/sha256"
	"encoding/hex"
)

type Hash struct {
	__value [32]byte
}

func (hash *Hash) to_string() string {
	return hex.EncodeToString(hash.__value[:])
}

func (hash *Hash) trailing_zeros() int {
	count := 0
	for i := 31; i >= 0; i-- {
		bt := hash.__value[i]
		for j := 0; j < 8 && bt > 0; j++ {
			if bt&1 == 1 {
				return count + j
			}
			bt >>= 1
		}
		count += 8
	}
	return count
}

func concat_hash(elements ...Hash) Hash {
	acc := make([]byte, 0, 32*len(elements))
	for _, element := range elements {
		acc = append(acc, element.__value[:]...)
	}
	return Hash{__value: sha256.Sum256(acc)}
}