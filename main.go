package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/xeger/sqlstream/nlp"
	"github.com/xeger/sqlstream/scrubbing"
)

var parallelism = flag.Int("p", runtime.NumCPU(), "parallelism level (default: number of CPUs)")

// TODO: use cobra and add subcommands
func main() {
	//mainToScrubSQL()
	mainToPredictWithModel()
}

func mainToGenerateModel() {
	reader := bufio.NewReader(os.Stdin)
	model := nlp.NewModel(4, "")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		model.Train(strings.TrimRight(line, "\r\n\t"))
	}

	for i := 0; i < 10; i++ {
		fmt.Fprintln(os.Stderr, model.Generate())
	}

	marshalled, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		panic(err.Error())
	}
	fmt.Print(string(marshalled))
}

func mainToPredictWithModel() {
	var model *nlp.Model = nil
	data, err := os.ReadFile("cities.markov.json")
	if err != nil {
		panic(err.Error())
	}
	err = json.Unmarshal(data, &model)
	if err != nil {
		panic(err.Error())
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimRight(line, "\r\n\t")
		if model.Recognize(line, 0.50) {
			fmt.Println("✅", line)
		} else {
			fmt.Println("❌", line)
		}
	}
}

func mainToScrubSQL() {
	flag.Parse()
	N := *parallelism

	in := make([]chan string, N)
	out := make([]chan string, N)
	for i := 0; i < N; i++ {
		in[i] = make(chan string)
		out[i] = make(chan string)
		go scrubbing.Scrub(in[i], out[i])
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
