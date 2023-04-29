package scrubbing

import "fmt"

// Percentage is a convenient way to print percentages with a single decimal place.
type Percentage float64

func (p Percentage) String() string {
	return fmt.Sprintf("%.1f%%", p*100)
}

func (p Percentage) MarshalYAML() (interface{}, error) {
	return p.String(), nil
}

// RuleReport describes how a rule was applied to a given input stream.
type RuleReport struct {
	// Defn is the rule definition expressed as a string with compact, human-readable notation.
	Defn string
	// Freq records the frequency (0-100%) with which this rule was applied.
	Freq Percentage
	// Safe determines the frequency of sanitized outputs that did not coincide with any input.
	//   100% => absolutely no overlap (output is perfectly sanitized)
	//     0% => complete overlap (output is effectively NOT sanitized, even if values are transposed; DANGER!)
	Safe Percentage
}

type SummaryReport struct {
	Load Percentage
	Safe Percentage
}

// Record provides statistics about how a policy was applied to a given input stream.
// It can be used to cross-check configuration and policy against real input data to
// make sure that sanitization is effective.
type Report struct {
	// FieldName contains statistics about each field-name rule.
	FieldName []RuleReport
	// Heuristic contains statistics about each heuristic rule.
	Heuristic []RuleReport
	Summary   SummaryReport
}
