package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "sqlstream",
		Short: "Streaming MySQL Sanitizer",
		Long: `Masks sensitive text in MySQL dumps.
Uses a heuristic rule system, applying language models to avoid depending on
specific schema features such as table or column names. Contains subcommands
for training and testing models.`,
	}
)

func init() {
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(recognizeCmd)
	rootCmd.AddCommand(scrubCmd)
	rootCmd.AddCommand(trainCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
