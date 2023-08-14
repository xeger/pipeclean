package scrubbing

import (
	"fmt"
	"regexp"

	"github.com/xeger/pipeclean/nlp"
)

// Policy reflects human decisionmaking about which values should be scrubbed
// based on their field name.
type Policy struct {
	// FieldName ensures that certain fields are always scrubbed based on their name.
	// Key: substring of a field name
	// Value: disposition when fields matching this substring are encountered
	FieldName []FieldNameRule `json:"fieldname"`
	// Heuristic applies selected models heuristically to all values to achieve
	// scrubbing based on the type, shape of pattern of the value itself.
	// Key: model name
	// Value: disposition when a value matches the model
	Heuristic []HeuristicRule `json:"heuristic"`
}

// DefaultPolicy returns a Policy with broadly-useful defaults
// that are suitable for a wide variety of use cases.
func DefaultPolicy() *Policy {
	return &Policy{
		FieldName: []FieldNameRule{
			{In: regexp.MustCompile("email"), Out: "mask"},
			{In: regexp.MustCompile("phone"), Out: "mask"},
			{In: regexp.MustCompile("(post(al)?_?code)|zip"), Out: "mask"},
		},
	}
}

// MatchFieldName returns a Disposition for the given field name
// if it matches any of the policy's field-name patterns.
// Otherwise it returns the empty string.
func (p Policy) MatchFieldName(names []string) (Disposition, int) {
	if len(names) > 0 {
		for idx, rule := range p.FieldName {
			for _, n := range names {
				if rule.In.MatchString(n) {
					return rule.Out, idx
				}
			}
		}
	}
	return "", -1
}

// Validate checks that the policy is internally consistent.
func (p Policy) Validate(models map[string]nlp.Model) []error {
	var errs []error

	for i, rule := range p.FieldName {
		switch rule.Out.Action() {
		case "erase", "mask", "pass", "replace":
			continue
		case "generate":
			model := models[rule.Out.Parameter()]
			if model == nil {
				errs = append(errs, fmt.Errorf("unrecognized model %q for fieldname[%d]", rule.Out.Parameter(), i))
			} else if _, ok := model.(nlp.Generator); !ok {
				errs = append(errs, fmt.Errorf("model %q for fieldname[%d] is not a generator", rule.Out.Parameter(), i))
			}
		default:
			errs = append(errs, fmt.Errorf("unknown policy action %q for fieldname[%d]", rule.Out.Action(), i))
		}
	}

	for i, rule := range p.Heuristic {

		modelIn := models[rule.In]
		if modelIn == nil {
			errs = append(errs, fmt.Errorf("unrecognized model %q for heuristic[%d]", rule.In, i))
		}
		switch rule.Out.Action() {
		case "erase", "mask", "replace":
			continue
		case "generate":
			modelOut := models[rule.Out.Parameter()]
			if modelOut == nil {
				errs = append(errs, fmt.Errorf("unrecognized output model %q for heuristic[%d]", rule.Out.Parameter(), i))
			} else if _, ok := modelOut.(nlp.Generator); !ok {
				errs = append(errs, fmt.Errorf("model %q for heuristic[%d] is not a generator", rule.Out.Parameter(), i))
			}
		default:
			errs = append(errs, fmt.Errorf("unknown policy action %q for heuristic[%d]", rule.Out.Action(), i))
		}
	}

	return errs
}
