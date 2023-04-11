package cmd

import (
	"github.com/spf13/cobra"
)

var (
	format string

	rootCmd = &cobra.Command{
		Use:       "pipeclean",
		Short:     "PipeClean",
		Long:      `PipeClean Streaming Data Sanitizer.`,
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		ValidArgs: []string{"json", "mysql"},
		Run: func(cmd *cobra.Command, args []string) {
			format = args[0]
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&format, "format", "", "input format")
	rootCmd.MarkFlagRequired("format")
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(recognizeCmd)
	rootCmd.AddCommand(scrubCmd)
	rootCmd.AddCommand(trainCmd)
}

func Execute() error {
	return rootCmd.Execute()
}
