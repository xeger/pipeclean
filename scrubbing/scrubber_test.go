package scrubbing_test

import (
	"testing"

	"github.com/xeger/sqlstream/scrubbing"
)

const salt = "github.com/xeger/sqlstream/scrubbing"

func assertChange(t *testing.T, a string) {
	s := scrubbing.NewScrubber(salt, nil, 0.95)
	b := s.ScrubString(a)
	if a == b {
		t.Errorf(`scrubString(%q) = %q, want a scrambled value`, a, b)
	}
}

func assertNoChange(t *testing.T, a string) {
	s := scrubbing.NewScrubber(salt, nil, 0.95)
	b := s.ScrubString(a)
	if a != b {
		t.Errorf(`scrubString(%q) = %q, want unchanged value`, a, b)
	}
}

func assertEq(t *testing.T, a, expected string) {
	s := scrubbing.NewScrubber(salt, nil, 0.95)
	b := s.ScrubString(a)
	if b != expected {
		t.Errorf(`scrubString(%q) = %q, expected %q`, a, b, expected)
	}
}

func TestEmail(t *testing.T) {
	assertEq(t, "joe@foo.com", "rmy@llb.aby")
	assertEq(t, "gophers@google.com", "gyvbyiy@jofjoz.sea")
}

func TestStreetAddress(t *testing.T) {
	assertEq(t, "100 Cloverdale Ln", "200 Cloverdale Ln")
	assertEq(t, "23846 Maybach Cir", "85295 Maybach Cir")
}

func TestStreetSuffix(t *testing.T) {
	suffixes := []string{
		"Ave",
		"Ave.",
		"Avenue",
		"Blvd",
		"Blvd.",
		"Boulevard",
		"Cir",
		"Cir.",
		"Circle",
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
	assertEq(t, "805-555-1212", "404-788-3692")
	assertEq(t, "(805) 555-1212", "(702) 497-7713")
}
