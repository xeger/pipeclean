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
