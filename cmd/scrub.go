package cmd

import (
	"bufio"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/xeger/pipeclean/nlp"
	"github.com/xeger/pipeclean/scrubbing"
	"github.com/xeger/pipeclean/scrubbing/json"
	"github.com/xeger/pipeclean/scrubbing/mysql"
)

// Used for flags.
var (
	parallelism int
	salt        string

	scrubCmd = &cobra.Command{
		Use:   "scrub",
		Short: "Scrub",
		Long:  "Mask or remove sensitive data",
		Run:   scrub,
	}
)

type scrubFunc func(*scrubbing.Scrubber, <-chan string, chan<- string)

func init() {
	scrubCmd.PersistentFlags().IntVar(&parallelism, "parallelism", runtime.NumCPU(), "lines to scrub at once")
	scrubCmd.PersistentFlags().StringVar(&salt, "salt", "", "static diversifier for text-masking PRNG")
}

func loadModels(paths []string) ([]nlp.Model, error) {
	result := make([]nlp.Model, 0)

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
			result = append(result, dirResult...)
		} else {
			model, err := nlp.LoadModel(path)
			if err != nil {
				return nil, err
			}
			result = append(result, model)
		}
	}

	return result, nil
}

func scrub(cmd *cobra.Command, args []string) {
	models, err := loadModels(args)
	if err != nil {
		panic(err.Error())
	}

	switch format {
	case "json":
		scrubJson(models)
	case "mysql":
		scrubMysql(models)
	default:
		panic("unknown format: " + format)
	}
}

func scrubJson(models []nlp.Model) {
	sc := scrubbing.NewScrubber(salt, models, 0.95)
	// TODO: parallelize JSON scrubbing (but not parsing)
	json.Scrub(sc, os.Stdin, os.Stdout)
}

func scrubMysql(models []nlp.Model) {
	N := parallelism

	in := make([]chan string, N)
	out := make([]chan string, N)
	for i := 0; i < N; i++ {
		in[i] = make(chan string)
		out[i] = make(chan string)
		sc := scrubbing.NewScrubber(salt, models, 0.95)
		go mysql.ScrubChan(sc, in[i], out[i])
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
