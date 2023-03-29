package cmd

import (
	"encoding/json"
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

	var model *nlp.Model = nil
	data, err := os.ReadFile(modelFile)
	if err != nil {
		panic(err.Error())
	}
	err = json.Unmarshal(data, &model)
	if err != nil {
		panic(err.Error())
	}

	for i := 0; i < 10; i++ {
		fmt.Println(model.Generate())
	}
}
