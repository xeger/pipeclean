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

	header := string(d[0:256])

	name := filepath.Base(filename)
	ext := filepath.Ext(name)
	// nickname = strings.TrimSuffix(name, ext)

	switch ext {
	case ".json":
		if strings.Index(header, markovModelTypeID) >= 0 {
			m := MarkovModel{}

			if err = m.UnmarshalJSON(d); err != nil {
				return nil, err
			}
			return &m, nil
		} else {
			return nil, fmt.Errorf("nlp.LoadModel: Malformed model JSON (unknown type) in %q", name)
		}
	case ".txt":
		m := DictModel{}
		if err = m.UnmarshalText(d); err != nil {
			return nil, err
		}
		return &m, nil
	default:
		return nil, fmt.Errorf(`nlp.LoadModel: Malformed model (unknown extension) of %q`, name)
	}

}

func LoadModels(dirname string) (map[Model]Generator, error) {
	result := make(map[Model]Generator)

	dir, err := os.ReadDir(dirname)
	if err != nil {
		panic(err.Error())
	}

	byPrefix := make(map[string][]string)

	for _, dirent := range dir {
		if dirent.IsDir() {
			continue
		}
		name := dirent.Name()
		ext := filepath.Ext(name)
		prefix := strings.TrimSuffix(name, ext)
		byPrefix[prefix] = append(byPrefix[prefix], name)
	}

	for prefix, names := range byPrefix {
		models := make([]Model, 0, len(names))
		if len(names) > 2 {
			return nil, fmt.Errorf("nlp.LoadModels: Too many models for prefix %q", prefix)
		}

		for _, name := range names {
			m, err := LoadModel(filepath.Join(dirname, name))
			if err != nil {
				return nil, err
			}
			models = append(models, m)
		}

		var model Model
		var generator Generator
		if len(models) > 0 {
			model = models[0]
		}
		for _, m := range models {
			if g, ok := m.(Generator); ok {
				generator = g
			} else {
				model = m
			}
		}

		if model != nil {
			result[model] = generator
		}
	}

	return result, nil
}
