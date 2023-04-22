package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/xeger/pipeclean/cmd/ui"
	scrubjson "github.com/xeger/pipeclean/format/json"
	"github.com/xeger/pipeclean/format/mysql"
	"github.com/xeger/pipeclean/nlp"
	"github.com/xeger/pipeclean/scrubbing"
)

// Used for flags.
var (
	scrubCmd = &cobra.Command{
		Use:   "scrub",
		Short: "Scrub",
		Long:  "Masks sensitive data from stdin. Prints results to stdout.",
		Run:   scrub,
	}
)

type scrubFunc func(*scrubbing.Scrubber, <-chan string, chan<- string)

func init() {
	scrubCmd.PersistentFlags().StringVarP(&configFlag, "config", "c", "", "configuration file (JSON)")
	scrubCmd.PersistentFlags().StringSliceVarP(&contextFlag, "context", "x", []string{}, "extra files to parse for improved accuracy")
	scrubCmd.PersistentFlags().StringVarP(&saltFlag, "salt", "s", "", "static diversifier for PRNG seed")
}

func scrub(cmd *cobra.Command, args []string) {
	models, err := loadModels(args)
	if err != nil {
		ui.Fatal(err)
		ui.Exit('>')
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

func scrubJson(models map[string]nlp.Model, pol *scrubbing.Policy) {
	sc := scrubbing.NewScrubber(saltFlag, models, pol)
	// TODO: parallelize JSON scrubbing (but not parsing)
	scrubjson.Scrub(sc, os.Stdin, os.Stdout)
}

func scrubMysql(models map[string]nlp.Model, pol *scrubbing.Policy) {
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
