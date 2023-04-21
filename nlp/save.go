package nlp

import (
	"fmt"
	"os"
	"path/filepath"
)

func SaveModel(m Model, path string, basename string) error {
	var data []byte
	var err error
	var filename string

	switch mt := m.(type) {
	case *MarkovModel:
		data, err = mt.MarshalJSON()
		filename = basename + ".markov.json"
	case *DictModel:
		data, err = mt.MarshalText()
		filename = basename + ".dict.txt"
	case *MatchModel:
		data, err = mt.MarshalText()
		filename = basename + ".match.txt"
	default:
		return fmt.Errorf(`nlp.SaveModel: No serialization strategy for: %T`, m)
	}

	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(path, filename), data, 0644)
}
