package main

import (
	"testing"
)

func mustChange(t *testing.T, a string) {
	s := NewScrubber()
	b := s.ScrubString(a)
	if a == b {
		t.Fatalf(`ScrubString(%q) = %q, want a scrambled value`, a, b)
	}
}

func TestTelUS(t *testing.T) {
	mustChange(t, "805-555-1212")
	mustChange(t, "(805) 555-1212")
}
