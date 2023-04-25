package nlp

import (
	"encoding/json"
	"math"
	"strings"

	"github.com/xeger/gomarkov"
	"github.com/xeger/pipeclean/rand"
)

type MarkovDefinition struct {
	// Lookback memory length for state transition table.
	// Higher order uses more memory but (might!) improve generation accuracy.
	Order int
	// Tokenization mode: " " or "".
	Delim string
}

type MarkovModel struct {
	chain     gomarkov.Chain
	separator string
	stats     modelStats
}

type markovModelJSON struct {
	Separator string         `json:"separator"`
	Chain     gomarkov.Chain `json:"chain"`
	Stats     modelStats     `json:"stats"`
}

func NewMarkovModel(order int, separator string) *MarkovModel {
	return &MarkovModel{
		chain:     *gomarkov.NewChain(order),
		separator: separator,
		stats: modelStats{
			FreqN: make(map[int]int),
		},
	}
}

func (m *MarkovModel) MarshalJSON() ([]byte, error) {
	obj := markovModelJSON{Separator: m.separator, Chain: m.chain, Stats: m.stats}
	return json.Marshal(obj)
}

func (m *MarkovModel) UnmarshalJSON(b []byte) error {
	var obj markovModelJSON
	err := json.Unmarshal(b, &obj)
	if err != nil {
		return err
	}

	m.chain = obj.Chain
	m.separator = obj.Separator
	m.stats = obj.Stats
	m.stats.Derive()
	return nil
}

// Generate derives a random string deterministically from the seed.
// The length is guaranteed to be between the min and max lengths seen during training.
func (m *MarkovModel) Generate(seed string) string {
	seed = Clean(seed)
	rand := rand.NewRand(seed)

	order := m.chain.Order
	state := make(gomarkov.NGram, 0)
	for i := 0; i < order; i++ {
		state = append(state, gomarkov.StartToken)
	}
	for state[len(state)-1] != gomarkov.EndToken && len(state) < m.stats.MaxN+order {
		next, err := m.chain.GenerateDeterministic(state[(len(state)-order):], rand)
		if err != nil {
			panic("MarkovModel.Generate: " + err.Error())
		}
		if next != gomarkov.EndToken || len(state) >= m.stats.MinN {
			state = append(state, next)
		}
	}

	// Handle empty models or erroneous output
	start := order
	end := len(state) - 1
	if len(state) <= order {
		start = len(state) - 1
	}

	return strings.Join(state[start:end], m.separator)
}

func (m *MarkovModel) Recognize(input string) float64 {
	if len(input) < m.chain.Order {
		return 0.0
	}
	input = Clean(input)
	tokens := strings.Split(input, m.separator)
	logProb := float64(0)
	pairs := gomarkov.MakePairs(tokens, m.chain.Order)
	for _, pair := range pairs {
		prob, _ := m.chain.TransitionProbability(pair.NextState, pair.CurrentState)
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
	tokens := strings.Split(input, m.separator)
	m.chain.Add(tokens)
	m.stats.Add(input)
}

func (m *MarkovModel) Validate(md MarkovDefinition) error {
	if m.chain.Order != md.Order {
		return ErrInvalidModel
	}
	if m.separator != md.Delim {
		return ErrInvalidModel
	}
	return nil
}
