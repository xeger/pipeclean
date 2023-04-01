package scrubbing

import (
	"testing"
)

func assertChange(t *testing.T, a string) {
	s := NewScrubber(nil, 0.95)
	b := s.scrubString(a)
	if a == b {
		t.Fatalf(`scrubString(%q) = %q, want a scrambled value`, a, b)
	}
}

func assertNoChange(t *testing.T, a string) {
	s := NewScrubber(nil, 0.95)
	b := s.scrubString(a)
	if a != b {
		t.Fatalf(`scrubString(%q) = %q, want unchanged value`, a, b)
	}
}

func TestStreetSuffix(t *testing.T) {
	suffixes := []string{
		"Ave",
		"Ave.",
		"Avenue",
		"Blvd",
		"Blvd.",
		"Boulevard",
		"Dr",
		"Dr.",
		"Drive",
		"Wy",
		"Wy.",
		"Way",
	}
	for _, suffix := range suffixes {
		assertNoChange(t, suffix)
	}
}

func TestTelUS(t *testing.T) {
	assertChange(t, "805-555-1212")
	assertChange(t, "(805) 555-1212")
}
