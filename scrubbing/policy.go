package scrubbing

import (
	"fmt"
	"strings"

	"github.com/xeger/pipeclean/nlp"
)

// FieldNameRule describes a scrubbing policy based on the name of a field.
// and irrespective of its value.
type FieldNameRule struct {
	// In is a field-name matching pattern to test whether this rule applied.
	In string
	// Out describes what to do when a value satisfies this rule.
	Out Disposition
}

// HeuristicRule describes a scrubbing policy based on a value irrespective
// of its field name.
type HeuristicRule struct {
	// In is the name of a model that will be used to recognize values.
	In string
	// P is the p-value threshold for model recognition for this rule to apply.
	// When matching a value, models output a confidence on the interval [0..1];
	// this is compared to 1.0 - P and if the result is greater, the rule is applied.
	//
	// In other words:
	//  P = 0.0  --> model must be 100% confident (the default value)
	//  P = 0.05 --> model must be 95% confident
	//  and so on

	P float64
	// Out describes what to do when a value satisfies this rule.
	Out Disposition
}

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
			{In: "email", Out: "mask"},
			{In: "phone", Out: "mask"},
			{In: "postcode", Out: "mask"},
			{In: "postal_code", Out: "mask"},
			{In: "postalcode", Out: "mask"},
			{In: "zip", Out: "mask"},
		},
	}
}

// MatchFieldName returns a Disposition for the given field name
// if it matches any of the policy's field-name patterns.
// Otherwise it returns the empty string.
func (p Policy) MatchFieldName(names []string) Disposition {
	for _, rule := range p.FieldName {
		for _, n := range names {
			if strings.Contains(n, rule.In) {
				return rule.Out
			}
		}
	}
	return ""
}

// Validate checks that the policy is internally consistent.
func (p Policy) Validate(models map[string]nlp.Model) []error {
	var errs []error

	for i, rule := range p.FieldName {
		switch rule.Out.Action() {
		case "erase", "mask":
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
		case "erase", "mask":
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
