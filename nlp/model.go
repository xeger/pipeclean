package nlp

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/mb-14/gomarkov"
)

type Model struct {
	Chain     gomarkov.Chain `json:"chain"`
	Name      string         `json:"name"`
	Separator string         `json:"separator"`
}

func LoadModel(filename string) (*Model, error) {
	d, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	m := Model{}
	if err = json.Unmarshal(d, &m); err != nil {
		return nil, err
	}
	name := filepath.Base(filename)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	m.Name = name
	return &m, nil
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

func (m Model) Recognize(input string) float64 {
	if len(input) < m.Chain.Order {
		return 0
	}
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

func (m Model) Train(input string) {
	tokens := strings.Split(input, m.Separator)
	m.Chain.Add(tokens)
}
