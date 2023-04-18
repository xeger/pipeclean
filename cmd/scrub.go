package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/xeger/pipeclean/nlp"
	"github.com/xeger/pipeclean/scrubbing"
	scrubjson "github.com/xeger/pipeclean/scrubbing/json"
	"github.com/xeger/pipeclean/scrubbing/mysql"
)

// Used for flags.
var (
	scrubCmd = &cobra.Command{
		Use:   "scrub",
		Short: "Scrub",
		Long:  "Mask or remove sensitive data",
		Run:   scrub,
	}
)

type scrubFunc func(*scrubbing.Scrubber, <-chan string, chan<- string)

func init() {
	scrubCmd.PersistentFlags().Float64VarP(&confidence, "confidence", "c", 0.5, "minimum probability to consider a match")
	scrubCmd.PersistentFlags().StringSliceVarP(&context, "context", "x", []string{}, "extra files to parse for improved accuracy")
	scrubCmd.PersistentFlags().StringVarP(&policy, "policy", "p", "", "policy file (JSON)")
	scrubCmd.PersistentFlags().StringVarP(&salt, "salt", "s", "", "static diversifier for PRNG seed")
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

func scrub(cmd *cobra.Command, args []string) {
	models, err := loadModels(args)
	if err != nil {
		panic(err.Error())
	}

	var pol *scrubbing.Policy
	if policy != "" {
		data, err := ioutil.ReadFile(policy)
		if err != nil {
			panic(err.Error())
		}
		pol = new(scrubbing.Policy)
		err = json.Unmarshal(data, pol)
		if err != nil {
			panic(err.Error())
		}
	} else {
		pol = scrubbing.DefaultPolicy()
	}
	if err := pol.Validate(models); err != nil {
		panic("invalid policy: " + err.Error())
	}

	switch mode {
	case "json":
		scrubJson(models, pol)
	case "mysql":
		scrubMysql(models, pol)
	default:
		// should never happen (cobra should validate)
		panic("unknown mode: " + mode)
	}
}

func scrubJson(models map[string]nlp.Model, pol *scrubbing.Policy) {
	sc := scrubbing.NewScrubber(salt, models, pol)
	// TODO: parallelize JSON scrubbing (but not parsing)
	scrubjson.Scrub(sc, os.Stdin, os.Stdout)
}

func scrubMysql(models map[string]nlp.Model, pol *scrubbing.Policy) {
	// Scan any context provided
	ctx := mysql.NewScrubContext()
	for _, file := range context {
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
		sc := scrubbing.NewScrubber(salt, models, pol)
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
