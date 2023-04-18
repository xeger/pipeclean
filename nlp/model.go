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

// MfBase returns the file's base name with no extensions.
func mfBase(filename string) string {
	base := filepath.Base(filename)
	dot := strings.Index(base, ".")
	if dot < 0 {
		return base
	}
	return base[:dot]
}

// MfExt returns the last two extensions of the file's base name.
func mfExt(filename string) string {
	base := filepath.Base(filename)
	suffix2 := filepath.Ext(base)
	suffix1 := filepath.Ext(strings.TrimSuffix(base, suffix2))
	return strings.Join([]string{suffix1, suffix2}, "")
}

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
