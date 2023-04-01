package nlp

type DictModel struct {
	Dict map[string]bool
	Name string
}

func NewDictModel(order int, separator string) *DictModel {
	return &DictModel{
		Dict: make(map[string]bool),
	}
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
