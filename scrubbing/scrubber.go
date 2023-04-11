package scrubbing

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/mail"
	"regexp"
	"strings"

	"github.com/xeger/sqlstream/nlp"
	"github.com/xeger/sqlstream/rand"
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
	confidence float64
	models     []nlp.Model
	salt       string
	shallow    bool
}

func NewScrubber(salt string, models []nlp.Model, confidence float64) *Scrubber {
	return &Scrubber{
		confidence: confidence,
		models:     models,
		salt:       salt,
	}
}

// EraseString signals to remove a string entirely from the input stream and replace it
// with a format-specific empty value.
//
// It returns true for base64 encoded values since they are opaque and cannot be scrubbed;
// it's safest to remove them from the stream entirely.
func (sc *Scrubber) EraseString(s string) bool {
	return reBase64.MatchString(s)
}

// ScrubData recursively scrubs maps and arrays in-place.
func (sc *Scrubber) ScrubData(data any) any {
	switch v := data.(type) {
	case string:
		return sc.ScrubString(v)
	case []any:
		for i, e := range v {
			v[i] = sc.ScrubData(e)
		}
		return v
	case map[string]any:
		for k, e := range v {
			v[k] = sc.ScrubData(e)
		}
		return v
	default:
		return v
	}
}

// ScrubString masks recognized PII in a string, preserving other values.
func (sc *Scrubber) ScrubString(s string) string {
	// Mask well-known numeric formats and abbreviations.
	if reTelUS.MatchString(s) {
		dash := strings.Index(s, "-")
		if dash < 0 {
			return sc.mask(s)
		}
		area, num := s[:dash], s[dash+1:]
		area = sc.mask(area)
		num = sc.mask(num)
		return fmt.Sprintf("%s-%s", area, num)
	} else if reZip.MatchString(s) {
		return sc.mask(s)
	}

	// Mask email addresses w/ consistent local and domain parts.
	if len(s) < 1024 && strings.Index(s, " ") == -1 {
		if a, _ := mail.ParseAddress(s); a != nil {
			at := strings.Index(a.Address, "@")
			local, domain := a.Address[:at], a.Address[at+1:]
			dot := strings.LastIndex(domain, ".")
			if dot > 0 {
				tld := domain[dot+1:]
				prefix := domain[0:dot]
				return fmt.Sprintf("%s@%s.%s", sc.mask(local), sc.mask(prefix), tld)
			} else {
				return sc.mask(domain)
			}
		}
	}

	// Mask each part of short phrases of 2-10 words that contain a numeric component.
	if reContainsNum.MatchString(s) {
		spaces := strings.Count(s, " ")
		if spaces > 1 && spaces < 10 {
			words := strings.Fields(s)
			for i, w := range words {
				words[i] = sc.ScrubSubstring(w)
			}
			return strings.Join(words, " ")
		}
	}

	// Handle deep scrubbing (e.g. JSON/YAML in string).
	if !sc.shallow {
		var data any

		if err := json.Unmarshal([]byte(s), &data); err == nil {
			scrubbed, err := json.Marshal(sc.ScrubData(data))
			if err != nil {
				panic(err)
			}
			return string(scrubbed)
		}

		if err := yaml.Unmarshal([]byte(s), &data); err == nil {
			switch v := data.(type) {
			case []any, map[string]any:
				scrubbed, err := yaml.Marshal(sc.ScrubData(v))
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
	for _, model := range sc.models {
		if model.Recognize(s) >= sc.confidence {
			if generator, ok := model.(nlp.Generator); ok {
				return nlp.ToSameCase(generator.Generate(s), s)
			} else {
				return sc.mask(s)
			}
		}
	}

	return s
}

// ScrubSubstring performs extra-diligent masking assuming that s is a
// substring of a larger phrase.
func (sc *Scrubber) ScrubSubstring(s string) string {
	if reIntDec.MatchString(s) {
		return sc.mask(s)
	}

	return sc.ScrubString(s)
}

// Scrambles letters and numbers; preserves case, punctuation, and special characters.
// As a special case, preserves 0 (and thus the distribution of zero to nonzero).
// Always returns the same output for a given input.
func (sc *Scrubber) mask(s string) string {
	rand := rand.NewRand(nlp.Clean(s))
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
