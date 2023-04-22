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
	// Key: substring of a field name
	// Value: disposition when fields matching this substring are encountered
	FieldName map[string]Disposition `json:"fieldname"`
	// Heuristic applies selected models heuristically to all values to achieve
	// scrubbing based on the type, shape of pattern of the value itself.
	// Key: model name
	// Value: disposition when a value matches the model
	Heuristic map[string]Disposition `json:"heuristic"`
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

// MatchFieldName returns a Disposition for the given field name
// if it matches any of the policy's field-name patterns.
// Otherwise it returns the empty string.
func (p Policy) MatchFieldName(names []string) Disposition {
	for k, v := range p.FieldName {
		for _, n := range names {
			if strings.Contains(n, k) {
				return v
			}
		}
	}
	return ""
}

// Validate checks that the policy is internally consistent.
func (p Policy) Validate(models map[string]nlp.Model) []error {
	var errs []error

	for fn, d := range p.FieldName {
		switch d.Action() {
		case "erase", "mask":
			continue
		case "generate":
			model := models[d.Parameter()]
			if model == nil {
				errs = append(errs, fmt.Errorf("unrecognized model %q for fieldname %q", d.Parameter(), fn))
			} else if _, ok := model.(nlp.Generator); !ok {
				errs = append(errs, fmt.Errorf("model %q for fieldname %q is not a generator", d.Parameter(), fn))
			}
		default:
			errs = append(errs, fmt.Errorf("unknown policy action %q for fieldname %q", d.Action(), fn))
		}
	}

	for mn, d := range p.Heuristic {
		modelIn := models[mn]
		if modelIn == nil {
			errs = append(errs, fmt.Errorf("unrecognized model %q for heuristic %q", mn, d))
		}
		switch d.Action() {
		case "erase", "mask":
			continue
		case "generate":
			modelOut := models[d.Parameter()]
			if modelOut == nil {
				errs = append(errs, fmt.Errorf("unrecognized output model %q for heuristic %q", d.Parameter(), mn))
			} else if _, ok := modelOut.(nlp.Generator); !ok {
				errs = append(errs, fmt.Errorf("model %q for heuristic %q is not a generator", d.Parameter(), mn))
			}
		default:
			errs = append(errs, fmt.Errorf("unknown policy action %q for heuristic %q", d.Action(), mn))
		}
	}

	return errs
}
