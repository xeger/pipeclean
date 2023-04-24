package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/xeger/pipeclean/cmd/ui"
	"github.com/xeger/pipeclean/nlp"
	"github.com/xeger/pipeclean/scrubbing"
)

// Shared flag values for all commands.
//
// Helps ensure consistency i.e. if "foo" is a float64 in command A, then
// it must be a float64 in command B, too.
var (
	appendFlag      bool
	confidenceFlag  float64
	configFlag      string
	contextFlag     []string
	markFlag        bool
	modeFlag        string
	parallelismFlag int
	saltFlag        string
)

type ModelConfig struct {
	Dict   *nlp.DictDefinition
	Markov *nlp.MarkovDefinition
	Match  *nlp.MatchDefinition
}

// Validate ensures that the model configuration is valid.
func (mc ModelConfig) Validate() error {
	subs := 0

	if mc.Dict != nil {
		subs++
	}
	if mc.Markov != nil {
		subs++
		if mc.Markov.Order <= 0 {
			return fmt.Errorf(`markov order must be >= 1`)
		}
	}
	if mc.Match != nil {
		subs++
	}

	switch subs {
	case 0:
		return fmt.Errorf(`unknown type`)
	case 1:
		return nil
	default:
		return fmt.Errorf(`ambiguous type`)
	}
}

// Config tells pipeclean how to learn from and scrub your data sets.
// It is generally read from a JSON file with a CLI parameter.
//
// There is no overlap between CLI flags and this file. This file
// only changes when data structure changes substantially; but CLI
// flags are much more malleable and vary at the whim of the user
// and the use case.
type Config struct {
	// Models describes the models used to learn and scrub.
	// Key: model name
	// Value: model configuration
	Models    map[string]ModelConfig
	Scrubbing *scrubbing.Policy
}

func DefaultConfig() *Config {
	return &Config{
		Models:    map[string]ModelConfig{},
		Scrubbing: scrubbing.DefaultPolicy(),
	}
}

// NewConfigFile loads a Config from disk.
func NewConfigFile(filename string) (*Config, error) {
	cfg := DefaultConfig()
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (cfg *Config) Validate(models map[string]nlp.Model) []error {
	var errs []error

	if scrubbingErrors := cfg.Scrubbing.Validate(models); scrubbingErrors != nil {
		h := ui.Fatalf("Invalid scrubbing policy.")
		for _, e := range scrubbingErrors {
			h.Hint(e.Error())
		}
		errs = append(errs, scrubbingErrors...)
	}

	for name, defn := range cfg.Models {
		if err := defn.Validate(); err != nil {
			errs = append(errs, err)
		}
		if m := models[name]; m != nil {
			if defn.Dict != nil {
				if _, ok := m.(*nlp.DictModel); !ok {
					ui.Fatalf("Type mismatch for model %s (declared as Dict; got %T).\n", name, m).Hint("please delete this model and reinitialize it")
					errs = append(errs, nlp.ErrInvalidModel)
				}
			} else if defn.Markov != nil {
				if mt, ok := m.(*nlp.MarkovModel); ok {
					if err := mt.Validate(*defn.Markov); err != nil {
						switch err {
						case nlp.ErrInvalidModel:
							ui.Fatalf("Configuration mismatch for Markov model %s.\n", name).Hint("please delete this model and reinitialize it")
						}
						errs = append(errs, err)
					}
				} else {
					ui.Fatalf("Type mismatch for model %s (declared as Markov; got %T).\n", name, m).Hint("please delete this model and reinitialize it")
					errs = append(errs, nlp.ErrInvalidModel)
				}
			} else if defn.Match != nil {
				if _, ok := m.(*nlp.MatchModel); !ok {
					ui.Fatalf("Type mismatch for model %s (declared as Match; got %T).\n", name, m).Hint("please delete this model and reinitialize it")
					errs = append(errs, nlp.ErrInvalidModel)
				}
			}
		}
	}

	return errs
}

func loadModels(paths []string) (map[string]nlp.Model, error) {
	result := make(map[string]nlp.Model, 0)

	for _, path := range paths {
		fi, err := os.Stat(path)
		if err != nil {
			ui.Fatal(err)
			ui.Exit('>')
		}
		if fi.IsDir() {
			dirResult, err := nlp.LoadModels(path)
			if err != nil {
				return nil, err
			}
			for k, v := range dirResult {
				result[k] = v
			}
		} else {
			ui.ExitNotImplemented("load single model file")
		}
	}

	return result, nil
}

func saveModels(models map[string]nlp.Model, path string) error {
	for name, m := range models {
		err := nlp.SaveModel(m, path, name)
		if err != nil {
			return err
		}
	}

	return nil
}
