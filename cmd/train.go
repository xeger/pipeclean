package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/xeger/sqlstream/nlp"
)

// Used for flags.
var (
	trainCmd = &cobra.Command{
		Use:   "train",
		Short: "Build a Markov model from a corpus",
		Long: `Parses words/phrases from stdin, one per line.
Prints a JSON representation of the model to stdout.`,
		Run: train,
	}
)

func train(cmd *cobra.Command, args []string) {
	var err error
	var mode, sep string
	var order int

	if len(args) == 2 {
		mode = args[0]
		order, err = strconv.Atoi(args[1])
		if err != nil {
			mode = "ERROR" // cause exit(1) below
		}
	}

	switch mode {
	case "sentences":
		sep = " "
	case "words":
		sep = ""
	default:
		fmt.Fprintln(os.Stderr, "Usage: sqlstream train <sentences|words> <order>")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	model := nlp.NewMarkovModel(order, sep)
	corpus := make([]string, 0, 65535)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		corpus = append(corpus, nlp.Clean(line))
		model.Train(line)
	}

	miss := 0
	for _, sample := range corpus {
		if c := model.Recognize(sample); c < 0.95 {
			miss++
		}
	}
	hitRate := 1.0 - float64(miss)/float64(len(corpus))
	fmt.Fprintln(os.Stderr, "Trained", len(corpus), "samples with order", order, "and hit rate", hitRate)

	marshalled, err := json.Marshal(model)
	if err != nil {
		panic(err.Error())
	}
	fmt.Print(string(marshalled))
}
