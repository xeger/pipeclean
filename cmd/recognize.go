package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xeger/pipeclean/cmd/ui"
	"github.com/xeger/pipeclean/nlp"
)

// Used for flags.
var (
	recognizeCmd = &cobra.Command{
		Use:   "recognize",
		Short: "Recognize",
		Long: `Parses words/phrases from stdin, one per line.
Prints input lines that match the model.`,
		Run: recognize,
	}
)

func init() {
	recognizeCmd.PersistentFlags().Float64VarP(&confidenceFlag, "confidence", "c", 0.5, "minimum probability to consider a match")
}

func recognize(cmd *cobra.Command, args []string) {
	var modelFile string
	if len(args) == 1 {
		modelFile = args[0]
	} else {
		ui.Fatalf("Usage: pipeclean recognize <modelFile>")
		ui.Exit('-')
	}

	model, err := nlp.LoadModel(modelFile)
	if err != nil {
		ui.Fatal(err)
		ui.Exit('>')
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimRight(line, "\r\n\t")
		if model.Recognize(line) >= confidenceFlag {
			fmt.Println(line)
		}
	}
}
