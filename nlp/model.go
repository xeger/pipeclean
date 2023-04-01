package nlp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Model interface {
	Recognize(input string) float64
	Train(input string)
}

func LoadModel(filename string) (Model, error) {
	d, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	m := MarkovModel{}
	if err = json.Unmarshal(d, &m); err != nil {
		return nil, err
	}
	name := filepath.Base(filename)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	m.Name = name
	return &m, nil
}
