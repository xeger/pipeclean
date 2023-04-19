package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:       "pipeclean",
		Short:     "PipeClean",
		Long:      `PipeClean Streaming Data Sanitizer.`,
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"json", "mysql"},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&modeFlag, "mode", "m", "", "data format (json, mysql, etc)")
	rootCmd.MarkFlagRequired("mode")
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(recognizeCmd)
	rootCmd.AddCommand(scrubCmd)
	rootCmd.AddCommand(trainCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
