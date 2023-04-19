package scrubbing

import (
	"fmt"
	"strings"

	"github.com/xeger/pipeclean/nlp"
)

// Policy reflects human decisionmaking about which values should be scrubbed
// based on their field name.
type Policy struct {
	// FieldName ensures that certain fields are always scrubbed based on their name.
	FieldName map[string]Disposition `json:"fieldname"`
}

// DefaultPolicy returns a Policy with broadly-useful defaults
// that are suitable for a wide variety of use cases.
func DefaultPolicy() *Policy {
	return &Policy{
		FieldName: map[string]Disposition{
			"email":      "mask",
			"phone":      "mask",
			"postcode":   "mask",
			"postalcode": "mask",
			"zip":        "mask",
		},
	}
}

// MatchFieldName returns a disposition ("erase" or "mask") for the given
// field name if it matches any of the policy's field-name patterns.
// Otherwise it returns the empty string.
func (p Policy) MatchFieldName(fieldName string) Disposition {
	for k, v := range p.FieldName {
		if strings.Contains(fieldName, k) {
			return v
		}
	}
	return ""
}

// Validate checks that the policy is internally consistent.
func (p Policy) Validate(models map[string]nlp.Model) error {
	for fn, d := range p.FieldName {
		switch d.Action() {
		case "erase", "mask":
			continue
		case "generate":
			model := models[d.Parameter()]
			if model == nil {
				return fmt.Errorf("unrecognized model %q for fieldname %q", d.Parameter(), fn)
			} else if _, ok := model.(nlp.Generator); !ok {
				return fmt.Errorf("model %q for fieldname %q is not a generator", d.Parameter(), fn)
			}
		default:
			return fmt.Errorf("unknown policy action %q for fieldname %q", d.Action(), fn)
		}
	}

	return nil
}
