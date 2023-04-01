package nlp

import (
	"bufio"
	"bytes"
)

type DictModel struct {
	Dict map[string]bool
}

func NewDictModel() *DictModel {
	return &DictModel{
		Dict: make(map[string]bool),
	}
}

func (m *DictModel) MarshalText() ([]byte, error) {
	buf := new(bytes.Buffer)
	for k := range m.Dict {
		buf.WriteString(k)
		buf.WriteRune('\n')
	}
	return buf.Bytes(), nil
}

func (m *DictModel) UnmarshalText(b []byte) error {
	m.Dict = make(map[string]bool)
	scanner := bufio.NewScanner(bytes.NewBuffer(b))

	for scanner.Scan() {
		m.Dict[Clean(scanner.Text())] = true
	}

	return nil
}

func (m *DictModel) Recognize(input string) float64 {
	input = Clean(input)
	if _, ok := m.Dict[input]; ok {
		return 1.0
	} else {
		return 0.0
	}
}

func (m *DictModel) Train(input string) {
	input = Clean(input)
	m.Dict[input] = true
}
