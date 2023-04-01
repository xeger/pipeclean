package nlp_test

import (
	"testing"

	"github.com/xeger/sqlstream/nlp"
)

func TestClean(t *testing.T) {
	assert := func(input, expected string) {
		if actual := nlp.Clean(input); actual != expected {
			t.Errorf("nlp.Clean: expected %q, got %q", expected, actual)
		}
	}
	assert("Hello, world!", "hello, world!")
	assert("   aHaHaHH    Ahah Hah", "ahahahh ahah hah")
}
