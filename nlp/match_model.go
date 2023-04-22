package nlp

import (
	"bufio"
	"bytes"
	"regexp"
)

type MatchDefinition struct{}

type MatchModel struct {
	patterns []*regexp.Regexp
}

func NewMatchModel(patterns []*regexp.Regexp) *MatchModel {
	return &MatchModel{patterns}
}

func (m *MatchModel) MarshalText() ([]byte, error) {
	buf := new(bytes.Buffer)
	for _, p := range m.patterns {
		buf.WriteString(p.String())
		buf.WriteRune('\n')
	}
	return buf.Bytes(), nil
}

func (m *MatchModel) UnmarshalText(b []byte) error {
	sources := make([]string, 0, 4)
	scanner := bufio.NewScanner(bytes.NewBuffer(b))
	for scanner.Scan() {
		sources = append(sources, scanner.Text())
	}
	patterns := make([]*regexp.Regexp, len(sources))
	for i, s := range sources {
		pat, err := regexp.Compile(s)
		if err != nil {
			return err
		}
		patterns[i] = pat
	}
	m.patterns = patterns
	return nil
}

func (m *MatchModel) Recognize(input string) float64 {
	for _, p := range m.patterns {
		if p.MatchString(input) {
			return 1.0
		}
	}
	return 0.0
}

func (m *MatchModel) Train(input string) {
	// TODO: build a training mechanism for regexp!
}
