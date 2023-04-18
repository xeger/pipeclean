package scrubbing

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/mail"
	"regexp"
	"strconv"
	"strings"

	"github.com/xeger/pipeclean/nlp"
	"github.com/xeger/pipeclean/rand"
	"gopkg.in/yaml.v3"
)

var reBase64 = regexp.MustCompile(`^([A-Za-z0-9+/]{4})*([A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{2}==)?$`)

// Phrase that contains a numeric sequence (i.e. a street address).
var reContainsNum = regexp.MustCompile(`#?[0-9-]{1,7}`)

// Integer in decimal notation with optional leading sign.
var reIntDec = regexp.MustCompile(`[+-]?0|[1-9]\d*`)

// Telephone number.
var reTelUS = regexp.MustCompile(`^\(?\d{3}\)?[ -]?\d{3}-?\d{4}$`)

var reZip = regexp.MustCompile(`^\d{5}(-\d{4})?$`)

type Scrubber struct {
	models  map[string]nlp.Model
	policy  *Policy
	salt    string
	shallow bool
}

func NewScrubber(salt string, models map[string]nlp.Model, policy *Policy) *Scrubber {
	return &Scrubber{
		models: models,
		policy: policy,
		salt:   salt,
	}
}

// EraseString signals to remove a string entirely from the input stream and replace it
// with a format-specific empty value.
//
// It returns true for base64 encoded values since they are opaque and cannot be scrubbed;
// it's safest to remove them from the stream entirely.
func (sc *Scrubber) EraseString(s, field string) bool {
	return reBase64.MatchString(s)
}

// ScrubData recursively scrubs maps and arrays in-place.
func (sc *Scrubber) ScrubData(data any, field string) any {
	switch v := data.(type) {
	case string:
		return sc.ScrubString(v, field)
	case []any:
		for i, e := range v {
			v[i] = sc.ScrubData(e, strconv.Itoa(i))
		}
		return v
	case map[string]any:
		for k, e := range v {
			v[k] = sc.ScrubData(e, k)
		}
		return v
	default:
		return v
	}
}

// ScrubString masks recognized PII in a string, preserving other values.
func (sc *Scrubber) ScrubString(s, field string) string {
	if field != "" {
		if disposition := sc.policy.MatchFieldName(field); disposition != "" {
			switch disposition {
			case "erase":
				return ""
			case "mask":
				return sc.mask(s)
			}
		}
	}

	// Handle deep scrubbing if data is well-formed JSON/YAML.
	if !sc.shallow {
		var data any

		if err := json.Unmarshal([]byte(s), &data); err == nil {
			scrubbed, err := json.Marshal(sc.ScrubData(data, ""))
			if err != nil {
				panic(err)
			}
			return string(scrubbed)
		}

		if err := yaml.Unmarshal([]byte(s), &data); err == nil {
			switch v := data.(type) {
			case []any, map[string]any:
				scrubbed, err := yaml.Marshal(sc.ScrubData(v, ""))
				if err != nil {
					panic(err)
				}
				return string(scrubbed)
			}
		}

		// Empty serialized Ruby YAML hashes.
		if strings.Index(s, "--- !ruby/hash") == 0 {
			return "{}"
		}
	}

	// Match against all models.
	// TODO: fix me after fixing model loading
	/*
		for _, model := range sc.models {
			if model.Recognize(s) >= sc.confidence {
				if generator, ok := model.(nlp.Generator); ok {
					return nlp.ToSameCase(generator.Generate(s), s)
				} else {
					return sc.mask(s)
				}
			}
		}
	*/

	return s
}

// ScrubSubstring performs extra-diligent masking assuming that s is a
// substring of a larger phrase.
func (sc *Scrubber) ScrubSubstring(s, field string) string {
	if reIntDec.MatchString(s) {
		return sc.mask(s)
	}

	return sc.ScrubString(s, field)
}

// Mask scrambles the numeric or alphabetic characters in a string, preserving
// other characters (punctuation, etc) and preserving the length of the string.
//
// Some special-case logic handles the following cases:
//   - email addresses: TLD is left unmasked
func (sc *Scrubber) mask(s string) string {
	if len(s) < 1024 && strings.Index(s, " ") == -1 {
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
