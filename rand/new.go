package rand

import (
	"hash/fnv"
	"math/rand"
)

// Hash computes a compact hash of the given string.
func Hash(s string) int64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return int64(h.Sum64())

}

// NewSource creates rand.Source from a given seed string.
func NewSource(seed string) rand.Source {
	return rand.NewSource(Hash(seed))
}

// NewRand creates a rand.Rand from a given seed string.
func NewRand(seed string) *rand.Rand {
	return rand.New(NewSource(seed))
}
