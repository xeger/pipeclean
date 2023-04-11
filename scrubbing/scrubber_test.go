package scrubbing_test

import (
	"testing"

	"github.com/xeger/pipeclean/scrubbing"
)

const salt = "github.com/xeger/pipeclean/scrubbing"

func scrub(s string) string {
	return scrubSalted(s, "")
}

func scrubSalted(s, salt string) string {
	sc := scrubbing.NewScrubber(salt, nil, 0.95)
	return sc.ScrubString(s)
}

func TestDeepJSON(t *testing.T) {
	in := `{"email":"joe@foo.com"}`
	exp := `{"email":"jyv@iws.com"}`
	if got := scrub(in); got != exp {
		t.Errorf(`scrub(%q) = %q, want %q`, in, got, exp)
	}
}

func _skip_TestDeepYAML(t *testing.T) {
	in := "email: joe@foo.com\n"
	exp := "email: jyv@iws.com\n"
	if got := scrub(in); got != exp {
		t.Errorf(`scrub(%q) = %q, want %q`, in, got, exp)
	}

	// BUG: YAML metadata & document structure are not preserved
	// TODO: investigate yaml.Node & build a scrubbing/yaml package if needed
	//  - preserves comments
	//  - hopefully has a way to preserve type markers
	in2 := `--- !ruby/hash
email: joe@foo.com
`
	if got2 := scrub(in2); got2 != exp {
		t.Errorf(`scrub(%q) = %q, want %q`, in2, got2, exp)
	}
}

func TestEmail(t *testing.T) {
	cases := map[string]string{
		"joe@foo.com":        "jyv@iws.com",
		"gophers@google.com": "hruhlic@mzovvt.com",
	}
	for in, exp := range cases {
		if got := scrub(in); got != exp {
			t.Errorf(`scrub(%q) = %q, want %q`, in, got, exp)
		}
	}
}

func TestNumerics(t *testing.T) {
	if got := scrub("74"); got != "74" {
		t.Errorf(`scrub(%q) = %q, want unchanged`, "74", got)
	}
}

func TestStreetAddress(t *testing.T) {
	cases := map[string]string{
		"100 Cloverdale Ln": "300 Cloverdale Ln",
		"23846 Maybach Cir": "87624 Maybach Cir",
	}
	for in, exp := range cases {
		if got := scrub(in); got != exp {
			t.Errorf(`scrub(%q) = %q, want %q`, in, got, exp)
		}
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
		if got := scrub(suffix); got != suffix {
			t.Errorf(`scrub(%q) = %q, want unchanged`, suffix, got)
		}
	}
}

func TestTelUS(t *testing.T) {
	cases := map[string]string{
		"805-555-1212":   "705-231-9867",
		"(805) 555-1212": "(902) 418-6892",
	}
	for in, exp := range cases {
		if got := scrub(in); got != exp {
			t.Errorf(`scrub(%q) = %q, want %q`, in, got, exp)
		}
	}
}
