package nlp

import (
	"fmt"
)

type CompoundModel struct {
	recognizers []Model
	generator   Generator
}

func NewCompoundModel(underlying []Model) (*CompoundModel, error) {
	m := CompoundModel{recognizers: underlying}
	for _, r := range underlying {
		if g, ok := r.(Generator); ok {
			m.generator = g
			break
		}
	}
	if m.generator == nil {
		return nil, fmt.Errorf("nlp.NewCompoundModel: none of %d underlying models is a Generator", len(underlying))
	}

	return &m, nil
}

func (m *CompoundModel) Generate(seed string) string {
	return m.generator.Generate(seed)
}

func (m *CompoundModel) Recognize(input string) float64 {
	max := 0.0
	for _, r := range m.recognizers {
		prob := r.Recognize(input)
		if prob > max {
			max = prob
		}
		if max >= 1.0 {
			return max
		}
	}

	return max
}

func (m *CompoundModel) Train(input string) {
	for _, r := range m.recognizers {
		r.Train(input)
	}
}
