package nlp

import (
	"encoding/json"
	"fmt"
	"regexp"
)

type MatchModel struct {
	patterns []*regexp.Regexp
}

type matchModelJSON struct {
	Type     string   `json:"typ"`
	Patterns []string `json:"pat"`
}

const matchModelTypeID = "github.com/xeger/pipeclean/nlp.MatchModel"

func NewMatchModel(patterns []*regexp.Regexp) *MatchModel {
	return &MatchModel{patterns}
}

func (m *MatchModel) MarshalJSON() ([]byte, error) {
	patterns := make([]string, 0, len(m.patterns))
	for _, p := range m.patterns {
		patterns = append(patterns, p.String())
	}
	obj := matchModelJSON{Type: matchModelTypeID, Patterns: patterns}
	return json.Marshal(obj)
}

func (m *MatchModel) UnmarshalJSON(b []byte) error {
	var obj matchModelJSON
	err := json.Unmarshal(b, &obj)
	if err != nil {
		return err
	}
	if obj.Type != matchModelTypeID {
		return fmt.Errorf("Wrong type; expected %q, got %q", matchModelTypeID, obj.Type)
	}

	m.patterns = make([]*regexp.Regexp, 0, len(obj.Patterns))
	for _, p := range obj.Patterns {
		if r, err := regexp.Compile(p); err != nil {
			return err
		} else {
			m.patterns = append(m.patterns, r)
		}
	}
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
