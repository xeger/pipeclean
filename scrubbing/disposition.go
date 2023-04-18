package scrubbing

import "strings"

// Disposition describes how to scrub a piece of data (in the abstract).
// It can be any of the following:
//   - "erase": remove the data entirely from the output
//   - "mask": scramble characters of the data
//   - "generate(modelName)": create dummy replacement data using the given model
type Disposition string

func (d Disposition) String() string {
	return string(d)
}

func (d Disposition) Action() string {
	paren := strings.Index(string(d), "(")
	if paren >= 0 {
		return string(d[:paren])
	}
	return string(d)
}

func (d Disposition) Parameter() string {
	paren := strings.Index(string(d), "(")
	if paren >= 0 {
		return string(d[paren+1 : len(d)-1])
	}
	return ""
}
