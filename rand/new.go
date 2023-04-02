package rand

import (
	"hash/fnv"
	"math/rand"
)

// NewSource creates rand.Source from a given seed string.
func NewSource(seed string) rand.Source {
	h := fnv.New64a()
	h.Write([]byte(seed))
	return rand.NewSource(int64(h.Sum64()))
}

// NewRand creates a rand.Rand from a given seed string.
func NewRand(seed string) *rand.Rand {
	return rand.New(NewSource(seed))
}
