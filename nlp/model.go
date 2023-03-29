package nlp

import (
	"math"
	"strings"

	"github.com/mb-14/gomarkov"
)

type Model struct {
	Chain     gomarkov.Chain `json:"chain"`
	Separator string         `json:"separator"`
}

func NewModel(order int, separator string) *Model {
	return &Model{
		Chain:     *gomarkov.NewChain(order),
		Separator: separator,
	}
}

func (m Model) Generate() string {
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

func (m Model) Recognize(input string, threshold float64) bool {
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
	prob := math.Pow(10, logProb/math.Max(1, float64(len(pairs))))
	return prob >= threshold
}

func (m Model) Train(input string) {
	tokens := strings.Split(input, m.Separator)
	m.Chain.Add(tokens)
}
