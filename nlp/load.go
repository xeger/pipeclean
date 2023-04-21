package nlp

import (
	"fmt"
	"os"
	"path/filepath"
)

func LoadModel(filename string) (Model, error) {
	d, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	ext := mfExt(filename)
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
		return nil, fmt.Errorf(`nlp.LoadModel: Unknown filename extension: %q`, ext)
	}

}

func LoadModels(dirname string) (map[string]Model, error) {
	dir, err := os.ReadDir(dirname)
	if err != nil {
		panic(err.Error())
	}

	result := make(map[string]Model)

	for _, dirent := range dir {
		name := dirent.Name()
		if name[0] == '.' || dirent.IsDir() {
			continue
		}
		model, err := LoadModel(filepath.Join(dirname, name))
		if err != nil {
			return nil, err
		}
		result[mfBase(name)] = model
	}

	return result, nil
}
