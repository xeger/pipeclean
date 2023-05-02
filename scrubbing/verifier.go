package scrubbing

import (
	"sync"

	"github.com/xeger/pipeclean/rand"
)

type Verifier struct {
	mx sync.Mutex

	policy *Policy

	// For each FieldName rule index, keeps a record of input and output string hashes
	// when scrubbing occured. Helps ensure that sanitized output is dissimilar enough
	// from input when a generator is in use.
	fieldNameInOut map[int]map[int64]int64
	// For each FieldName rule index, keeps a record of which actual field names
	// were processed by that rule.
	fieldNameFields map[int]map[string]bool

	// For each Heuristic rule index, keeps a record of input and output string hashes
	// when scrubbing occured. Helps ensure that sanitized output is dissimilar enough
	// from input when a generator is in use.
	heuristicInOut map[int]map[int64]int64
	// For each Heuristic rule index, keeps a record of which actual field names
	// were processed by that rule.
	heuristicFields map[int]map[string]bool

	// Keeps a record of which input strings were passed to output without scrubbing.
	passIn     map[int64]bool
	passFields map[string]bool
}

func (v *Verifier) recordFieldName(in, out string, names []string, ruleIndex int, disposition Disposition) {
	v.mx.Lock()
	defer v.mx.Unlock()

	inHash := rand.Hash(in)
	outHash := rand.Hash(out)

	inOut := v.fieldNameInOut[ruleIndex]
	if inOut == nil {
		inOut = make(map[int64]int64)
		v.fieldNameInOut[ruleIndex] = inOut
	}
	inOut[inHash] = outHash

	fields := v.fieldNameFields[ruleIndex]
	if fields == nil {
		fields = make(map[string]bool)
		v.fieldNameFields[ruleIndex] = fields
	}
	for _, name := range names {
		fields[name] = true
	}
}

func (v *Verifier) recordHeuristic(in, out string, names []string, ruleIndex int, disposition Disposition) {
	v.mx.Lock()
	defer v.mx.Unlock()

	inHash := rand.Hash(in)
	outHash := rand.Hash(out)

	inOut := v.heuristicInOut[ruleIndex]
	if inOut == nil {
		inOut = make(map[int64]int64)
		v.heuristicInOut[ruleIndex] = inOut
	}
	inOut[inHash] = outHash

	fields := v.heuristicFields[ruleIndex]
	if fields == nil {
		fields = make(map[string]bool)
		v.heuristicFields[ruleIndex] = fields
	}
	for _, name := range names {
		fields[name] = true
	}
}

func (v *Verifier) recordPass(in string, names []string) {
	v.mx.Lock()
	defer v.mx.Unlock()

	if in == "" {
		return // the zero string does not contribute to statistics
	}

	v.passIn[rand.Hash(in)] = true
	for _, name := range names {
		v.passFields[name] = true
	}
}

// NewVerifier creates a Scrubber linked to a Verifier.
// After all scrubbing is complete, call the Verifier to produce statistics.
func NewVerifier(pol *Policy) *Verifier {
	verifier := &Verifier{
		policy:          pol,
		fieldNameInOut:  make(map[int]map[int64]int64),
		fieldNameFields: make(map[int]map[string]bool),
		heuristicInOut:  make(map[int]map[int64]int64),
		heuristicFields: make(map[int]map[string]bool),
		passIn:          make(map[int64]bool),
		passFields:      make(map[string]bool),
	}

	return verifier
}

// Report produces a YAML-printable summary of the Verifier's findings.
func (v *Verifier) Report() *Report {
	r := &Report{
		FieldName: make([]RuleReport, len(v.policy.FieldName)),
		Heuristic: make([]RuleReport, len(v.policy.Heuristic)),
	}

	// Count the number of distinct input strings seen, categorizing by passed or scrubbed.
	distinctScrubbed := 0
	for _, inOut := range v.fieldNameInOut {
		distinctScrubbed += len(inOut)
	}
	for _, inOut := range v.heuristicInOut {
		distinctScrubbed += len(inOut)
	}
	distinctPassed := len(v.passIn)

	for i, rule := range v.policy.FieldName {
		r.FieldName[i].Defn = rule.String()

		inOut := v.fieldNameInOut[i]
		if inOut == nil {
			continue
		}

		r.FieldName[i].Freq = Percentage(float64(len(inOut)) / float64(distinctScrubbed+distinctPassed))

		overlap := 0
		out := map[int64]bool{}
		for _, outHash := range inOut {
			out[outHash] = true
		}
		for inHash := range inOut {
			if out[inHash] {
				overlap++
			}
		}
		r.FieldName[i].Safe = Percentage(1.0 - float64(overlap)/float64(len(inOut)))
		r.Summary.Safe += r.FieldName[i].Safe
	}

	for i, rule := range v.policy.Heuristic {
		r.Heuristic[i].Defn = rule.String()

		inOut := v.heuristicInOut[i]
		if inOut == nil {
			continue
		}

		r.Heuristic[i].Freq = Percentage(float64(len(inOut)) / float64(distinctScrubbed+distinctPassed))

		overlap := 0
		out := map[int64]bool{}
		for _, outHash := range inOut {
			out[outHash] = true
		}
		for inHash := range inOut {
			if out[inHash] {
				overlap++
			}
		}
		r.Heuristic[i].Safe = Percentage(1.0 - float64(overlap)/float64(len(inOut)))
		r.Summary.Safe += r.Heuristic[i].Safe
	}

	r.Summary.Load = Percentage(float64(distinctScrubbed) / float64(distinctScrubbed+distinctPassed))
	r.Summary.Safe /= Percentage(len(r.FieldName) + len(r.Heuristic))

	return r
}
