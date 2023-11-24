package main

import (
	"crypto/rand"
	"math/big"

	"golang.org/x/exp/constraints"
)

// function random_int returns a random integer in the range (mn, mx)
func random_int(mn int64, mx int64) int64 {
	if mn > mx {
		panic("mn should be less than mx")
	}
	num, _ := rand.Int(rand.Reader, big.NewInt(mx-mn+1))
	return num.Int64() + mn
}

// function random_string returns a random string of length n consisting of lowercase english alphabets
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
		hash.Value[i] = byte(tmp.Int64())
	}
	return hash
}

// function CollectChanOne reads a value from a channel in a non-blocking manner
func CollectChanOne[T any](ch <-chan T) (T, bool) {
	select {
	case val, stillOpen := <-ch:
		return val, stillOpen
	default:
		var zeroT T
		return zeroT, false
	}
}

func min[T constraints.Ordered](obj1 T, obj2 T) T {
	if obj1 <= obj2 {
		return obj1
	}
	return obj2
}

func max[T constraints.Ordered](obj1 T, obj2 T) T {
	if obj1 >= obj2 {
		return obj1
	}
	return obj2
}

// function get_map_keys returns a slice containing all the keys of the provided map
func get_map_keys[K comparable, V any](mymap map[K]V) []K {
	keys := make([]K, 0, len(mymap))
	for k := range mymap {
		keys = append(keys, k)
	}
	return keys
}

// function get_map_values returns a slice containing all the values of the provided map
func get_map_values[K comparable, V any](mymap map[K]V) []V {
	values := make([]V, 0, len(mymap))
	for _, v := range mymap {
		values = append(values, v)
	}
	return values
}

// function reverse_slice returns the reverse of the given slice
func reverse_slice[T comparable](myslice []T) []T {
	for i := 0; i < len(myslice)/2; i++ {
		myslice[i], myslice[len(myslice)-i-1] = myslice[len(myslice)-i-1], myslice[i]
	}
	return myslice
}
