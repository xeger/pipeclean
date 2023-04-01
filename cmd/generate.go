package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xeger/sqlstream/nlp"
)

// Used for flags.
var (
	generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "TODO",
		Long:  `TODO`,
		Run:   generate,
	}
)

func generate(cmd *cobra.Command, args []string) {
	var modelFile string
	if len(args) == 1 {
		modelFile = args[0]
	} else {
		fmt.Fprintln(os.Stderr, "Usage: sqlstream train <sentences|words>")
		os.Exit(1)
	}

	model, err := nlp.LoadModel(modelFile)
	if err != nil {
		panic(err.Error())
	}

	if g, ok := model.(nlp.Generator); ok {
		for i := 0; i < 10; i++ {
			fmt.Println(g.Generate(""))
		}
	} else {
		fmt.Fprintln(os.Stderr, "Model does not support generation.")
		os.Exit(1)
	}
}
