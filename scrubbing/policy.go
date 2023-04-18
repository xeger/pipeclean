package scrubbing

import "strings"

// HeuristicPolicy reflects human decisionmaking about which values should be
// scrubbed based on their content (and not on any surrounding context such as
// field name).
type HeuristicPolicy struct {
	ModelName  string  `json:"modelname"`
	Confidence float64 `json:"confidence"`
}

// Policy reflects human decisionmaking about which values should be scrubbed
// based on their field name.
type Policy struct {
	// FieldName ensures that certain fields are always scrubbed based on their name.
	//     Keys: field-name substring to match e.g. "email", "smtp_addr"
	//   Values: how to scrub matching fields; "erase" or "mask"
	FieldName map[string]string `json:"fieldname"`
	/// Heuristic allows values to be scrubbed regardless of where they appear in
	/// the input stream and irrespective of their field name. It names an NLP model
	/// used to recognize values and a confidence threshhold for recognition.
	///
	Heuristic []HeuristicPolicy `json:"heuristic"`
}

// DefaultPolicy returns a Policy with broadly-useful defaults
// that are suitable for a wide variety of use cases.
func DefaultPolicy() *Policy {
	return &Policy{
		FieldName: map[string]string{
			"email":      "mask",
			"phone":      "mask",
			"postcode":   "mask",
			"postalcode": "mask",
			"zip":        "mask",
		},
		Heuristic: []HeuristicPolicy{},
	}
}

// MatchFieldName returns a disposition ("erase" or "mask") for the given
// field name if it matches any of the policy's field-name patterns.
// Otherwise it returns the empty string.
func (p Policy) MatchFieldName(fieldName string) string {
	for k, v := range p.FieldName {
		if strings.Contains(fieldName, k) {
			return v
		}
	}
	return ""
}
