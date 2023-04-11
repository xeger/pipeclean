package cmd

import (
	"fmt"
	"math/rand"
	"os"

	"github.com/spf13/cobra"
	"github.com/xeger/pipeclean/nlp"
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
		fmt.Fprintln(os.Stderr, "Usage: pipeclean generate <modelFile>")
		os.Exit(1)
	}

	model, err := nlp.LoadModel(modelFile)
	if err != nil {
		panic(err.Error())
	}

	if g, ok := model.(nlp.Generator); ok {
		for i := 0; i < 10; i++ {
			fmt.Println(g.Generate(fmt.Sprintf("%d", rand.Int63())))
		}
	} else {
		fmt.Fprintln(os.Stderr, "Model does not support generation.")
		os.Exit(1)
	}
}
