package cmd

import (
	"encoding/json"
	"os"

	"github.com/xeger/pipeclean/nlp"
	"github.com/xeger/pipeclean/scrubbing"
)

// Shared flag values for all commands.
//
// Helps ensure consistency i.e. if "foo" is a float64 in command A, then
// it must be a float64 in command B, too.
var (
	confidenceFlag  float64
	configFlag      string
	contextFlag     []string
	modeFlag        string
	parallelismFlag int
	saltFlag        string
)

type MarkovDefinition struct {
	// Lookback memory length for state transition table.
	// Higher order uses more memory but (might!) improve generation accuracy.
	Order int
	// Tokenization mode: " " or "".
	Delim string
}

type ModelConfig struct {
	Markov map[string]MarkovDefinition
}

// Config tells pipeclean how to learn from and scrub your data sets.
// It is generally read from a JSON file with a CLI parameter.
//
// There is no overlap between CLI flags and this file. This file
// only changes when data structure changes substantially; but CLI
// flags are much more malleable and vary at the whim of the user
// and the use case.
type Config struct {
	Models    ModelConfig
	Scrubbing *scrubbing.Policy
}

func DefaultConfig() *Config {
	return &Config{
		Models:    ModelConfig{},
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

func (cfg *Config) Validate(models map[string]nlp.Model) error {
	return cfg.Scrubbing.Validate(models)
}

func loadModels(paths []string) (map[string]nlp.Model, error) {
	result := make(map[string]nlp.Model, 0)

	for _, path := range paths {
		fi, err := os.Stat(path)
		if err != nil {
			panic(err.Error())
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
			panic("not implemented: load single file")
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
