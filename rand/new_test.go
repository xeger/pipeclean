package rand_test

import (
	"testing"

	"github.com/xeger/sqlstream/rand"
)

func TestNewSource(t *testing.T) {
	output := make([][]int64, 256)
	for i := 0; i < 256; i++ {
		output[i] = make([]int64, 256)
		for j := 0; j < 256; j++ {
			output[i][j] = rand.NewSource("").Int63()
		}
	}
	tgt := output[0]
	for i := 1; i < len(output); i++ {
		for j := 0; j < 256; j++ {
			if tgt[j] != output[i][j] {
				t.Fatalf("rand.NewSource().Int63() is not deterministic")
				break
			}
		}
	}
}
