package cmd

import (
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/xeger/pipeclean/format/mysql"
)

// Used for flags.
var (
	extractCmd = &cobra.Command{
		Use:   "extract",
		Short: "Extract",
		Long:  "Pulls specific fields from inputs; prints values to stdout.",
		Run:   extract,
	}
)

func init() {
	extractCmd.PersistentFlags().StringSliceVarP(&contextFlag, "context", "x", []string{}, "extra files to parse for improved accuracy")
}

func extract(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		// TODO better!
		panic("must pass exactly one directory name for model storage")
	}

	switch modeFlag {
	case "json":
		extractJson(args)
	case "mysql":
		extractMysql(args)
	default:
		// should never happen (cobra should validate)
		panic("unknown mode: " + modeFlag)
	}
}

func extractJson(names []string) {
	panic("TODO")
}

func extractMysql(names []string) {
	ctx := mysql.NewContext()
	for _, file := range contextFlag {
		sql, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err.Error())
		}
		ctx.Scan(string(sql))
	}

	mysql.Extract(ctx, names, os.Stdin, os.Stdout)
}
