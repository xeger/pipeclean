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
		if strings.Index(header, `"typ": ""`) >= 0 {
			m := MarkovModel{}

			if err = m.UnmarshalJSON(d); err != nil {
				return nil, err
			}
			return &m, nil
		} else {
			return nil, fmt.Errorf(`nlp.LoadModel: Malformed model JSON (unknown "typ") in "%s"`, name)
		}
	case ".txt":
		m := DictModel{}
		if err = m.UnmarshalText(d); err != nil {
			return nil, err
		}
		return &m, nil
	default:
		return nil, fmt.Errorf(`nlp.LoadModel: Malformed model (unknown extension) of "%s"`, name)
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

	for _, names := range byPrefix {
		switch len(names) {
		case 1: // lonesome JSON or txt file
			m, err := LoadModel(filepath.Join(dirname, names[0]))
			if err != nil {
				return nil, err
			}
			if _, ok := m.(Generator); ok {
				result[m] = m.(Generator)
			} else {
				result[m] = nil
			}
		case 2: // paired JSON and text files
			var jsonFile, textFile string
			for _, name := range names {
				if filepath.Ext(name) == ".json" {
					jsonFile = name
				} else if filepath.Ext(name) == ".txt" {
					textFile = name
				}
			}
			if jsonFile == "" || textFile == "" {
				return nil, fmt.Errorf(`nlp.LoadModels: expected exactly one .txt and one .json file among '%s' and %s`, names[0], names[1])
			}
			m, err := LoadModel(textFile)
			if err != nil {
				return nil, err
			}
			g, err := LoadModel(jsonFile)
			if err != nil {
				return nil, err
			} else if g, ok := g.(Generator); ok {
				result[m] = g
			} else {
				return nil, fmt.Errorf(`nlp.LoadModels: no Generator among '%s' and %s`, names[0], names[1])
			}
		}
	}
	return result, nil
}
