package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/xeger/pipeclean/nlp"
	"github.com/xeger/pipeclean/scrubbing"
	"github.com/xeger/pipeclean/scrubbing/mysql"
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
	learnCmd.PersistentFlags().StringVarP(&configFlag, "config", "c", "", "configuration file (JSON)")
	learnCmd.PersistentFlags().StringSliceVarP(&contextFlag, "context", "x", []string{}, "extra files to parse for improved accuracy")
}

func learn(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		// TODO better!
		panic("must pass exactly one directory name for model storage")
	}

	models, err := loadModels(args)
	if err != nil {
		panic(err.Error())
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
	if err := cfg.Validate(models); err != nil {
		panic("invalid config: " + err.Error())
	}

	switch modeFlag {
	case "json":
		scrubJson(models, cfg.Scrubbing)
	case "mysql":
		scrubMysql(models, cfg.Scrubbing)
	default:
		// should never happen (cobra should validate)
		panic("unknown mode: " + modeFlag)
	}
}

func learnJson(models map[string]nlp.Model, pol *scrubbing.Policy) {
	panic("TODO")
}

func learnMysql(models map[string]nlp.Model, pol *scrubbing.Policy) {
	// Scan any context provided
	ctx := mysql.NewScrubContext()
	for _, file := range contextFlag {
		sql, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err.Error())
		}
		ctx.Scan(string(sql))
	}

	N := runtime.NumCPU()

	in := make([]chan string, N)
	out := make([]chan string, N)
	for i := 0; i < N; i++ {
		in[i] = make(chan string)
		out[i] = make(chan string)
		sc := scrubbing.NewScrubber(saltFlag, models, pol)
		go mysql.ScrubChan(ctx, sc, in[i], out[i])
	}
	drain := func(to int) {
		for i := 0; i < to; i++ {
			fmt.Print(<-out[i])
		}
	}
	done := func() {
		for i := 0; i < N; i++ {
			close(in[i])
			close(out[i])
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
		l = (l + 1) % N
		if l == 0 {
			drain(N)
		}
	}
	drain(l)
	done()
}
