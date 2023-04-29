package cmd

import (
	"github.com/spf13/cobra"
	"github.com/xeger/pipeclean/cmd/ui"
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
	rootCmd.PersistentFlags().BoolVarP(&ui.IsVerbose, "verbose", "v", false, "print extra debug output")
	rootCmd.MarkFlagRequired("mode")
	rootCmd.AddCommand(extractCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(learnCmd)
	rootCmd.AddCommand(recognizeCmd)
	rootCmd.AddCommand(scrubCmd)
	rootCmd.AddCommand(trainCmd)
	rootCmd.AddCommand(verifyCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
