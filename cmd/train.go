package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xeger/pipeclean/nlp"
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
	var modelType, markovMode, markovSep string
	var markovOrder int

	if len(args) == 1 {
		parts := strings.Split(args[0], ":")
		if len(parts) >= 1 {
			modelType = parts[0]
		}
		if len(parts) >= 2 {
			markovMode = parts[1]
		}
		if len(parts) >= 3 {
			markovOrder, err = strconv.Atoi(parts[2])
			if err != nil {
				markovMode = "ERROR" // cause exit(1) below
			}
		}
	}

	switch modelType {
	case "markov":
		switch markovMode {
		case "sentences":
			markovSep = " "
		case "words":
			markovSep = ""
		default:
			fmt.Fprintln(os.Stderr, "Usage: pipeclean train <modelType>[param1:param2:...]")
			fmt.Fprintln(os.Stderr, "Examples:")
			fmt.Fprintln(os.Stderr, "  pipeclean train dict # dictionary-lookup model")
			fmt.Fprintln(os.Stderr, "  pipeclean train markov:words:5 # markov word model of order 5")
			fmt.Fprintln(os.Stderr, "  pipeclean train markov:sentences:3 # markov sentence model of order 5")
			os.Exit(1)
		}

		reader := bufio.NewReader(os.Stdin)
		model := nlp.NewMarkovModel(markovOrder, markovSep)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			model.Train(line)
		}

		marshalled, err := model.MarshalJSON()
		if err != nil {
			panic(err.Error())
		}
		fmt.Print(string(marshalled))
	case "dict":
		reader := bufio.NewReader(os.Stdin)
		model := nlp.NewDictModel()

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			model.Train(line)
		}

		marshalled, err := model.MarshalText()
		if err != nil {
			panic(err.Error())
		}
		fmt.Print(string(marshalled))
	}
}
