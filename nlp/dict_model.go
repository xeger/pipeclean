package nlp

import (
	"bufio"
	"bytes"
)

type DictModel struct {
	dict map[string]bool
}

func NewDictModel() *DictModel {
	return &DictModel{
		dict: make(map[string]bool),
	}
}

func (m *DictModel) MarshalText() ([]byte, error) {
	buf := new(bytes.Buffer)
	for k := range m.dict {
		buf.WriteString(k)
		buf.WriteRune('\n')
	}
	return buf.Bytes(), nil
}

func (m *DictModel) UnmarshalText(b []byte) error {
	m.dict = make(map[string]bool)
	scanner := bufio.NewScanner(bytes.NewBuffer(b))

	for scanner.Scan() {
		m.dict[Clean(scanner.Text())] = true
	}

	return nil
}

func (m *DictModel) Recognize(input string) float64 {
	input = Clean(input)
	if _, ok := m.dict[input]; ok {
		return 1.0
	} else {
		return 0.0
	}
}

func (m *DictModel) Train(input string) {
	input = Clean(input)
	m.dict[input] = true
}
