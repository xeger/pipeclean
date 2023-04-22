package scrubbing_test

import (
	"regexp"
	"testing"

	"github.com/xeger/pipeclean/nlp"
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

func TestHeuristic(t *testing.T) {
	models := map[string]nlp.Model{
		"fruit": nlp.NewMatchModel([]*regexp.Regexp{regexp.MustCompile(`apple|orange`)}),
	}
	pol := &scrubbing.Policy{
		Heuristic: []scrubbing.HeuristicRule{
			{In: "fruit", Out: "erase"},
		},
	}
	tests := map[string]string{
		"apple": "",
		"horse": "horse",
	}
	for in, want := range tests {
		scrubber := scrubbing.NewScrubber(salt, models, pol)
		if got := scrubber.ScrubString(in, nil); got != want {
			t.Errorf(`scrub(%q) = %q, want %q`, in, got, want)
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
