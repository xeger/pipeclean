package nlp_test

import (
	"testing"

	"github.com/xeger/pipeclean/nlp"
)

func TestIsLower(t *testing.T) {
	assert := func(input string, expected bool) {
		if actual := nlp.IsLower(input); actual != expected {
			t.Errorf("nlp.IsLower(%q): expected %v, got %v", input, expected, actual)
		}
	}
	assert("hi", true)
	assert("hi world", true)
	assert("12873", true)
	assert("#@$*&", true)

	assert("Hi", false)
	assert("HI", false)
	assert("HI world", false)
	assert("hi WORLD", false)
}

func TestIsTitle(t *testing.T) {
	assert := func(input string, expected bool) {
		if actual := nlp.IsTitle(input); actual != expected {
			t.Errorf("nlp.IsTitle(%q): expected %v, got %v", input, expected, actual)
		}
	}
	assert("Hi", true)
	assert("Hi World", true)
	assert("12873", true)
	assert("#@$*&", true)

	assert("HI world", false)
	assert("hi World", false)
	assert("HI World", false)
	assert("hi World", false)
}

func TestIsUpper(t *testing.T) {
	assert := func(input string, expected bool) {
		if actual := nlp.IsUpper(input); actual != expected {
			t.Errorf("nlp.IsUpper(%q): expected %v, got %v", input, expected, actual)
		}
	}
	assert("HI", true)
	assert("HI WORLD", true)
	assert("12873", true)
	assert("#@$*&", true)

	assert("Hi", false)
	assert("hi", false)
	assert("HI world", false)
	assert("hi WORLD", false)
}
