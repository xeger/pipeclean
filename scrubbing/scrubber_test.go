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

func TestDeepJSON(t *testing.T) {
	json1 := `{"email":"joe@foo.com"}`
	json1s := `{"email":"jyv@iws.com"}`
	assertEq(t, json1, json1s)
}

func TestDeepYAML(t *testing.T) {
	yaml1 := "email: joe@foo.com\n"
	yaml1s := "email: jyv@iws.com\n"
	assertEq(t, yaml1, yaml1s)

	// BUG: YAML metadata & document structure are not preserved
	// TODO: investigate yaml.Node & build a scrubbing/yaml package if needed
	//  - preserves comments
	//  - hopefully has a way to preserve type markers
	yaml2 := `--- !ruby/hash
email: joe@foo.com
`
	assertEq(t, yaml2, yaml1s)
}

func TestEmail(t *testing.T) {
	assertEq(t, "joe@foo.com", "jyv@iws.com")
	assertEq(t, "gophers@google.com", "hruhlic@mzovvt.com")
}

func TestStreetAddress(t *testing.T) {
	assertEq(t, "100 Cloverdale Ln", "300 Cloverdale Ln")
	assertEq(t, "23846 Maybach Cir", "87624 Maybach Cir")
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
	assertEq(t, "805-555-1212", "705-231-9867")
	assertEq(t, "(805) 555-1212", "(902) 418-6892")
}
