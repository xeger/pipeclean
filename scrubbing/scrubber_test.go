package scrubbing_test

import (
	"testing"

	"github.com/xeger/pipeclean/scrubbing"
)

const salt = "github.com/xeger/pipeclean/scrubbing"

func scrub(s, field string) string {
	return scrubSalted(s, field, "")
}

func scrubSalted(s, field, salt string) string {
	return scrubbing.NewScrubber(salt, nil, scrubbing.DefaultPolicy()).ScrubString(s, []string{field})
}

func TestDeepJSON(t *testing.T) {
	in := `{"email":"joe@foo.com"}`
	exp := `{"email":"jyv@iws.com"}`
	if got := scrub(in, "someJsonField"); got != exp {
		t.Errorf(`scrub(%q) = %q, want %q`, in, got, exp)
	}
}

func _skip_TestDeepYAML(t *testing.T) {
	in := "email: joe@foo.com\n"
	exp := "email: jyv@iws.com\n"
	if got := scrub(in, "someYamlField"); got != exp {
		t.Errorf(`scrub(%q) = %q, want %q`, in, got, exp)
	}

	// BUG: YAML metadata & document structure are not preserved
	// TODO: investigate yaml.Node & build a scrubbing/yaml package if needed
	//  - preserves comments
	//  - hopefully has a way to preserve type markers
	in2 := `--- !ruby/hash
email: joe@foo.com
`
	if got2 := scrub(in2, "someYamlField"); got2 != exp {
		t.Errorf(`scrub(%q) = %q, want %q`, in2, got2, exp)
	}
}

func TestEmail(t *testing.T) {
	cases := map[string]string{
		"joe@foo.com":        "jyv@iws.com",
		"gophers@google.com": "hruhlic@mzovvt.com",
	}
	for in, exp := range cases {
		if got := scrub(in, "email"); got != exp {
			t.Errorf(`scrub(%q) = %q, want %q`, in, got, exp)
		}
	}
}

func TestNumerics(t *testing.T) {
	if got := scrub("74", "someField"); got != "74" {
		t.Errorf(`scrub(%q) = %q, want unchanged`, "74", got)
	}
}

func TestTelUS(t *testing.T) {
	cases := map[string]string{
		"805-555-1212":   "606-245-3192",
		"(805) 555-1212": "(606) 245-3192",
	}
	for in, exp := range cases {
		if got := scrub(in, "phone"); got != exp {
			t.Errorf(`scrub(%q) = %q, want %q`, in, got, exp)
		}
	}
}
