package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/xeger/pipeclean/cmd/ui"
	"github.com/xeger/pipeclean/scrubbing"
	"gopkg.in/yaml.v3"
)

// Used for flags.
var (
	verifyCmd = &cobra.Command{
		Use:   "verify",
		Short: "Verify",
		Long:  "Performs a scrub, discards the output, and prints effectiveness & safety statistics to stdout.",
		Run:   verify,
	}
)

func init() {
	verifyCmd.PersistentFlags().StringVarP(&configFlag, "config", "c", "", "configuration file (JSON)")
	verifyCmd.PersistentFlags().StringSliceVarP(&contextFlag, "context", "x", []string{}, "extra files to parse for improved accuracy")
}

func verify(cmd *cobra.Command, args []string) {
	models, err := loadModels(args)
	if err != nil {
		ui.Fatal(err)
		ui.Exit('>')
	}

	var cfg *Config
	if configFlag != "" {
		cfg, err = NewConfigFile(configFlag)
		if err != nil {
			ui.Fatal(err)
			ui.Exit('>')
		}
	} else {
		cfg = DefaultConfig()
	}
	if errs := cfg.Validate(models); errs != nil {
		ui.Exit('>') // cfg calls ui on its own
	}

	verifier := scrubbing.NewVerifier(cfg.Scrubbing)

	switch modeFlag {
	case "json":
		scrubJson(models, cfg.Scrubbing, verifier)
	case "mysql":
		scrubMysql(models, cfg.Scrubbing, verifier)
	default:
		// should never happen (cobra should validate)
		panic("unknown mode: " + modeFlag)
	}

	report := verifier.Report()
	printable, err := yaml.Marshal(report)
	if err != nil {
		ui.Fatal(err)
	}
	fmt.Println(string(printable))
}
