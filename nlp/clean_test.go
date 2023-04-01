package nlp_test

import (
	"testing"

	"github.com/xeger/sqlstream/nlp"
)

func TestClean(t *testing.T) {
	test := func(input, expected string) {
		if output := nlp.Clean(input); output != expected {
			t.Errorf("nlp.Clean: expected '%s', got '%s'", expected, output)
		}
	}
	test("Hello, world!", "hello, world!")
	test("   aHaHaHH    Ahah Hah", "ahahahh ahah hah")
}
