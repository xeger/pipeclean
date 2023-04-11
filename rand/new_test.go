package rand_test

import (
	"testing"

	"github.com/xeger/pipeclean/rand"
)

func TestNewSource(t *testing.T) {
	N := 8
	output := make([][]int64, 256)
	for i := 0; i < 256; i++ {
		output[i] = make([]int64, N)
		for j := 0; j < N; j++ {
			output[i][j] = rand.NewSource("").Int63()
		}
	}
	tgt := []int64{2761073929007780571, 2761073929007780571, 2761073929007780571, 2761073929007780571, 2761073929007780571, 2761073929007780571, 2761073929007780571, 2761073929007780571}
	for i := 1; i < len(output); i++ {
		for j := 0; j < N; j++ {
			if tgt[j] != output[i][j] {
				t.Fatalf("rand.NewSource().Int63() is not deterministic %v", output[i])
				break
			}
		}
	}
}
