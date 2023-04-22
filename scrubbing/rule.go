package scrubbing

import (
	"encoding/json"
	"regexp"
)

// FieldNameRule describes a scrubbing policy based on the name of a field.
// and irrespective of its value.
type FieldNameRule struct {
	// In is a field-name matching pattern to test whether this rule applied.
	In *regexp.Regexp
	// Out describes what to do when a value satisfies this rule.
	Out Disposition
}

type fieldNameRuleJSON struct {
	In  string
	Out string
}

func (r *FieldNameRule) MarshalJSON() ([]byte, error) {
	obj := fieldNameRuleJSON{
		In:  r.In.String(),
		Out: string(r.Out),
	}
	return json.Marshal(obj)
}

func (r *FieldNameRule) UnmarshalJSON(b []byte) error {
	var obj fieldNameRuleJSON
	err := json.Unmarshal(b, &obj)
	if err != nil {
		return err
	}

	if in, err := regexp.Compile(obj.In); err != nil {
		return err
	} else {
		r.In = in
	}

	r.Out = Disposition(obj.Out)

	return nil
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
