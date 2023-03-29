package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

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
	model := nlp.NewModel(order, sep)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		model.Train(strings.TrimRight(line, "\r\n\t"))
	}

	marshalled, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		panic(err.Error())
	}
	fmt.Print(string(marshalled))
}
