package nlp

import (
	"math"
	"strings"

	"github.com/mb-14/gomarkov"
)

type MarkovModel struct {
	Chain     gomarkov.Chain `json:"chain"`
	Name      string
	Separator string `json:"separator"`
}

func NewMarkovModel(order int, separator string) *MarkovModel {
	return &MarkovModel{
		Chain:     *gomarkov.NewChain(order),
		Separator: separator,
	}
}

// TODO: respect seed (need to improve gomarkov library)
func (m *MarkovModel) Generate(seed string) string {
	// seed = Clean(seed)
	order := m.Chain.Order
	state := make(gomarkov.NGram, 0)
	for i := 0; i < order; i++ {
		state = append(state, gomarkov.StartToken)
	}
	for state[len(state)-1] != gomarkov.EndToken {
		next, _ := m.Chain.Generate(state[(len(state) - order):])
		state = append(state, next)
	}
	return strings.Join(state[order:len(state)-1], m.Separator)
}

func (m *MarkovModel) Recognize(input string) float64 {
	input = Clean(input)
	tokens := strings.Split(input, m.Separator)
	logProb := float64(0)
	pairs := gomarkov.MakePairs(tokens, m.Chain.Order)
	for _, pair := range pairs {
		prob, _ := m.Chain.TransitionProbability(pair.NextState, pair.CurrentState)
		if prob > 0 {
			logProb += math.Log10(prob)
		} else {
			logProb += math.Log10(0.05)
		}
	}
	return math.Pow(10, logProb/math.Max(1, float64(len(pairs))))
}

func (m *MarkovModel) Train(input string) {
	input = Clean(input)
	tokens := strings.Split(input, m.Separator)
	m.Chain.Add(tokens)
}
