package utils

import (
	"crypto/rand"
	"math/big"
)

// Shuffle performs a Fisher–Yates shuffle on a slice using crypto/rand.
// It returns the shuffled slice and the permutation applied (original index → new index).
func Shuffle[T any](s []T) ([]T, []int) {
	n := len(s)
	perm := make([]int, n)
	for i := range perm {
		perm[i] = i
	}

	result := make([]T, n)
	copy(result, s)

	for i := n - 1; i > 0; i-- {
		j, err := cryptoRandN(i + 1)
		if err != nil {
			// Fallback: keep order if rand fails
			break
		}
		result[i], result[j] = result[j], result[i]
		perm[i], perm[j] = perm[j], perm[i]
	}
	return result, perm
}

// RandN returns a cryptographically random int in [0, n).
func RandN(n int) (int, error) {
	return cryptoRandN(n)
}

func cryptoRandN(n int) (int, error) {
	if n <= 0 {
		return 0, nil
	}
	max := big.NewInt(int64(n))
	v, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}
	return int(v.Int64()), nil
}

// Sample returns up to k random elements from s without replacement.
func Sample[T any](s []T, k int) []T {
	if k <= 0 {
		return nil
	}
	shuffled, _ := Shuffle(s)
	if k > len(shuffled) {
		return shuffled
	}
	return shuffled[:k]
}
