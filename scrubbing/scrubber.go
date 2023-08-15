package scrubbing

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/xeger/pipeclean/cmd/ui"
	"github.com/xeger/pipeclean/nlp"
	"github.com/xeger/pipeclean/rand"
	"gopkg.in/yaml.v3"
)

var reShortExtension = regexp.MustCompile(`[.][a-z]{2,5}$`)

type Scrubber struct {
	maskAll  bool
	models   map[string]nlp.Model
	policy   *Policy
	salt     string
	shallow  bool
	Verifier *Verifier
}

func NewScrubber(salt string, maskAll bool, policy *Policy, models map[string]nlp.Model) *Scrubber {
	return &Scrubber{
		models:  models,
		maskAll: maskAll,
		policy:  policy,
		salt:    salt,
	}
}

// EraseString signals to remove a string entirely from the input stream and replace it
// with a format-specific empty value.
//
// It returns true for base64 encoded values since they are opaque and cannot be scrubbed;
// it's safest to remove them from the stream entirely.
//
// It records hit statistics if a Verifier is provided, but does not record
// miss statistics under the assumption that the caller will always try
// to call ScrubString() if this returns false.
func (sc *Scrubber) EraseString(s string, names []string) bool {
	if disposition, ruleIndex := sc.policy.MatchFieldName(names); disposition != "" {
		if sc.Verifier != nil {
			sc.Verifier.recordFieldName(s, "", names, ruleIndex, disposition)
		}
		return disposition.Action() == "erase"
	}

	for ruleIndex, rule := range sc.policy.Heuristic {
		model := sc.models[rule.In]
		if model.Recognize(s) >= (1.0 - rule.P) {
			if sc.Verifier != nil {
				sc.Verifier.recordHeuristic(s, "", names, ruleIndex, rule.Out)
			}
			return rule.Out.Action() == "erase"
		}
	}

	// NB: deliberately not recording a pass (we assume caller will try to ScrubString).
	return false
}

// ScrubData recursively scrubs maps and arrays in-place.
// It records no statistics with the Verifier.
func (sc *Scrubber) ScrubData(data any, names []string) any {
	switch v := data.(type) {
	case string:
		return sc.ScrubString(v, names)
	case []any:
		for i, e := range v {
			v[i] = sc.ScrubData(e, []string{strconv.Itoa(i)})
		}
		return v
	case map[string]any:
		for k, e := range v {
			v[k] = sc.ScrubData(e, []string{k})
		}
		return v
	default:
		return v
	}
}

// ScrubString applies rules to sanitize a string, preserving values that do
// not match any rule.
// It records statistics if a Verifier is provided.
func (sc *Scrubber) ScrubString(s string, names []string) string {
	handle := func(disposition Disposition) string {
		switch disposition.Action() {
		case "erase":
			return ""
		case "generate":
			if sc.maskAll {
				return sc.mask(s)
			}
			if model := sc.models[disposition.Parameter()]; model != nil {
				if generator, ok := model.(nlp.Generator); ok {
					return nlp.ToSameCase(generator.Generate(s), s)
				}
			} else {
				// should never happen if Policy has been properly validated
				panic("unknown model name for generate action: " + disposition.Action())
			}
		case "mask":
			return sc.mask(s)
		case "pass":
			return s
		case "replace":
			// TODO
			return sc.replace(s, disposition.Parameter())
		}
		// should never happen if Policy has been properly validated
		ui.ExitBug("unknown policy action: " + disposition.Action())
		return ""
	}

	// First match against field-name rules
	if disposition, ruleIndex := sc.policy.MatchFieldName(names); disposition != "" {
		out := handle(disposition)
		if sc.Verifier != nil {
			sc.Verifier.recordFieldName(s, out, names, ruleIndex, disposition)
		}
		return out
	}

	// Then favor heuristic rules
	for ruleIndex, rule := range sc.policy.Heuristic {
		model := sc.models[rule.In]
		if model.Recognize(s) >= (1.0 - rule.P) {
			out := handle(rule.Out)
			if sc.Verifier != nil {
				sc.Verifier.recordHeuristic(s, out, names, ruleIndex, rule.Out)
			}
			return out
		}
	}

	// Finally, try to recurse into encapsulated structured data
	if !sc.shallow {
		var data any

		if err := json.Unmarshal([]byte(s), &data); err == nil {
			scrubbed, err := json.Marshal(sc.ScrubData(data, nil))
			if err != nil {
				ui.Fatal(err)
			}
			return string(scrubbed)
		}

		if err := yaml.Unmarshal([]byte(s), &data); err == nil {
			switch v := data.(type) {
			case []any, map[string]any:
				scrubbed, err := yaml.Marshal(sc.ScrubData(v, nil))
				if err != nil {
					ui.Fatal(err)
				}
				return string(scrubbed)
			}
		}

		// Empty serialized Ruby YAML hashes.
		if strings.Index(s, "--- !ruby/hash") == 0 {
			return "{}"
		}
	}

	if sc.Verifier != nil {
		sc.Verifier.recordPass(s, names)
	}
	return s
}

// Mask scrambles the numeric or alphabetic characters in a string, preserving
// other characters (punctuation, etc) and preserving the length of the string.
//
// Some special-case logic handles the following cases for short strings < 1KiB:
//   - email addresses: TLD is left unmasked
//   - filenames: extension up to five characters is left unmasked
func (sc *Scrubber) mask(s string) string {
	if len(s) < 1024 {
		// Well-formed email address
		if strings.Index(s, " ") == -1 {
			if a, _ := mail.ParseAddress(s); a != nil {
				at := strings.Index(a.Address, "@")
				local, domain := a.Address[:at], a.Address[at+1:]
				dot := strings.LastIndex(domain, ".")
				if dot > 0 {
					tld := domain[dot+1:]
					prefix := domain[0:dot]
					return fmt.Sprintf("%s@%s.%s", sc.maskWord(local), sc.maskWord(prefix), tld)
				} else {
					return sc.maskWord(domain)
				}
			}
		}

		// Extension (e.g. filename)
		if loc := reShortExtension.FindStringIndex(s); loc != nil {
			prefix, suffix := s[0:loc[0]], s[loc[0]:]
			return fmt.Sprintf("%s%s", sc.maskWord(prefix), suffix)
		}

		if u, err := url.Parse(s); err == nil && u.Scheme != "" {
			if dot := strings.LastIndex(u.Host, "."); dot >= 0 {
				u.Host = sc.maskWord(u.Host[0:dot]) + u.Host[dot:]
			} else {
				u.Host = sc.maskWord(u.Host)
			}
			u.Path = sc.maskWord(u.Path)
			return u.String()
		}
	}

	return sc.maskWord(s)
}

// MaskWord scrambles letters and numbers, preserving case, punctuation, and special characters.
// As a special case, preserves 0 (and thus the distribution of zero to nonzero).
// Always returns the same output for a given input.
func (sc *Scrubber) maskWord(s string) string {
	rand := rand.NewRand(nlp.CleanToken(s))
	h := fnv.New64a()
	if sc.salt != "" {
		h.Write([]byte(sc.salt))
		h.Write([]byte{0})
	}
	h.Write([]byte(s))

	sb := []byte(s)
	for i, b := range sb {
		if b >= 'a' && b <= 'z' {
			sb[i] = 'a' + byte(rand.Uint32()%26)
		} else if b >= 'A' && b <= 'Z' {
			sb[i] = 'A' + byte(rand.Uint32()%26)
		} else if b >= '1' && b <= '9' {
			sb[i] = '1' + byte(rand.Uint32()%9)
		}
	}

	return string(sb)
}

// Replace returns its second parameter, ignoring the first.
func (sc *Scrubber) replace(s string, replacement string) string {
	return replacement
}
