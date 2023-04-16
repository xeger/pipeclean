package nlp

import (
	"fmt"
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

	name := filepath.Base(filename)
	suffix2 := filepath.Ext(name)
	suffix1 := filepath.Ext(strings.TrimSuffix(name, suffix2))
	ext := strings.Join([]string{suffix1, suffix2}, "")

	// nickname = strings.TrimSuffix(name, ext)

	switch ext {
	case ".markov.json":
		m := MarkovModel{}

		if err = m.UnmarshalJSON(d); err != nil {
			return nil, err
		}
		return &m, nil
	case ".dict.txt":
		m := DictModel{}
		if err = m.UnmarshalText(d); err != nil {
			return nil, err
		}
		return &m, nil
	case ".match.txt":
		m := MatchModel{}
		if err = m.UnmarshalText(d); err != nil {
			return nil, err
		}
		return &m, nil
	default:
		return nil, fmt.Errorf(`nlp.LoadModel: Unknown filename extension: %q`, name)
	}

}

func LoadModels(dirname string) ([]Model, error) {
	dir, err := os.ReadDir(dirname)
	if err != nil {
		panic(err.Error())
	}

	result := make([]Model, 0, len(dir))

	byPrefix := make(map[string][]string)
	for _, dirent := range dir {
		name := dirent.Name()
		if name[0] == '.' || dirent.IsDir() {
			continue
		}
		ext := filepath.Ext(name)
		prefix := strings.TrimSuffix(name, ext)
		byPrefix[prefix] = append(byPrefix[prefix], name)
	}

	for prefix, names := range byPrefix {
		underlying := make([]Model, 0, len(names))
		if len(names) > 2 {
			return nil, fmt.Errorf("nlp.LoadModels: Too many models for prefix %q", prefix)
		}

		for _, name := range names {
			m, err := LoadModel(filepath.Join(dirname, name))
			if err != nil {
				return nil, err
			}
			underlying = append(underlying, m)
		}

		if len(underlying) == 1 {
			result = append(result, underlying...)
		} else {
			compound, err := NewCompoundModel(underlying)
			if err != nil {
				return nil, err
			}
			result = append(result, compound)
		}
	}

	return result, nil
}
