package scrubbing_test

import (
	"regexp"
	"testing"

	"github.com/xeger/pipeclean/nlp"
	"github.com/xeger/pipeclean/scrubbing"
)

const salt = "github.com/xeger/pipeclean/scrubbing"

func scrub(s, field string) string {
	return scrubbing.NewScrubber(salt, false, scrubbing.DefaultPolicy(), nil).ScrubString(s, []string{field})
}

func scrubWithPolicy(s, field string, policy *scrubbing.Policy, models map[string]nlp.Model) string {
	if err := policy.Validate(models); err != nil {
		panic(err)
	}
	return scrubbing.NewScrubber(salt, false, policy, models).ScrubString(s, []string{field})
}

func TestDefaultDeepJSON(t *testing.T) {
	in := `{"email":"joe@foo.com"}`
	exp := `{"email":"jyv@iws.com"}`
	if got := scrub(in, "someJsonField"); got != exp {
		t.Errorf(`scrub(%q) = %q, want %q`, in, got, exp)
	}
}

func TestDefaultEmail(t *testing.T) {
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

func TestDefaultHeuristic(t *testing.T) {
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
		scrubber := scrubbing.NewScrubber(salt, false, pol, models)
		if got := scrubber.ScrubString(in, nil); got != want {
			t.Errorf(`scrub(%q) = %q, want %q`, in, got, want)
		}
	}
}

func TestDefaultNumerics(t *testing.T) {
	if got := scrub("74", "someField"); got != "74" {
		t.Errorf(`scrub(%q) = %q, want unchanged`, "74", got)
	}
}

func TestDefaultTelUS(t *testing.T) {
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

func TestReplacement(t *testing.T) {
	cases := map[scrubbing.Disposition]string{
		"replace({})":   "{}",
		"replace((()))": "(())",
	}

	for out, exp := range cases {
		// asFieldName := &scrubbing.Policy{
		// 	FieldName: []scrubbing.FieldNameRule{
		// 		{In: regexp.MustCompile("foo"), Out: out},
		// 	},
		// }
		// if got := scrubWithPolicy("replace-me", "foo", asFieldName, nil); got != exp {
		// 	t.Errorf(`with FieldNameRule, scrub(%q) = %q, want %q`, out, got, exp)
		// }

		asHeuristic := &scrubbing.Policy{
			Heuristic: []scrubbing.HeuristicRule{
				{In: "bar", Out: out},
			},
		}
		models := map[string]nlp.Model{
			"bar": nlp.NewMatchModel([]*regexp.Regexp{regexp.MustCompile(`^\{\\?"p\\?":`)}),
		}
		if got := scrubWithPolicy(`{\"p\": \"\"}`, "foo", asHeuristic, models); got != exp {
			t.Errorf(`with HeuristicRule, scrub(%q) = %q, want %q`, out, got, exp)
		}
	}
}
