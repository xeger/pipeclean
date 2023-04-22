package cmd

import (
	"bufio"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/xeger/pipeclean/format/mysql"
	"github.com/xeger/pipeclean/nlp"
	"github.com/xeger/pipeclean/scrubbing"
)

// Used for flags.
var (
	learnCmd = &cobra.Command{
		Use:   "learn",
		Short: "Learn",
		Long:  "Trains models in parallel using data parsed from stdin.",
		Run:   learn,
	}
)

func init() {
	learnCmd.PersistentFlags().BoolVarP(&appendFlag, "append", "r", false, "load existing models before training (default: overwrite)")
	learnCmd.PersistentFlags().StringVarP(&configFlag, "config", "c", "", "configuration file (JSON)")
	learnCmd.PersistentFlags().StringSliceVarP(&contextFlag, "context", "x", []string{}, "extra files to parse for improved accuracy")
}

func learn(cmd *cobra.Command, args []string) {
	var err error

	if len(args) != 1 {
		// TODO better!
		panic("must pass exactly one directory name for model storage")
	}

	models := make(map[string]nlp.Model)
	if appendFlag {
		models, err = loadModels(args)
		if err != nil {
			panic(err.Error())
		}
	}

	var cfg *Config
	if configFlag != "" {
		cfg, err = NewConfigFile(configFlag)
		if err != nil {
			panic("malformed config:" + err.Error())
		}
	} else {
		cfg = DefaultConfig()
	}

	// Initialize any missing models
	for name, md := range cfg.Models.Markov {
		if _, ok := models[name]; !ok {
			models[name] = nlp.NewMarkovModel(md.Order, md.Delim)
		}
	}

	if err = cfg.Validate(models); err != nil {
		os.Exit(int('l'))
	}

	switch modeFlag {
	case "json":
		learnJson(models, cfg.Scrubbing)
	case "mysql":
		learnMysql(models, cfg.Scrubbing)
	default:
		// should never happen (cobra should validate)
		panic("unknown mode: " + modeFlag)
	}

	saveModels(models, args[0])
}

func learnJson(models map[string]nlp.Model, pol *scrubbing.Policy) {
	panic("TODO")
}

func learnMysql(models map[string]nlp.Model, pol *scrubbing.Policy) {
	// Scan any context provided
	ctx := mysql.NewContext()
	for _, file := range contextFlag {
		sql, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err.Error())
		}
		ctx.Scan(string(sql))
	}

	N := runtime.NumCPU()

	in := make([]chan string, N)
	for i := 0; i < N; i++ {
		in[i] = make(chan string)
		go mysql.LearnChan(ctx, models, pol, in[i])
	}
	done := func() {
		for i := 0; i < N; i++ {
			close(in[i])
		}
	}

	reader := bufio.NewReader(os.Stdin)
	l := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		in[l] <- line
	}
	done()
}
