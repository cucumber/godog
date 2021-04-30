package internal

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/cucumber/godog/internal/flags"
)

var version bool
var output string

// CreateRootCmd creates the root command.
func CreateRootCmd() cobra.Command {
	rootCmd := cobra.Command{
		Use: "godog",
		Long: `Creates and runs test runner for the given feature files.
Command should be run from the directory of tested package
and contain buildable go source.`,
		Args: cobra.ArbitraryArgs,

		// Deprecated: Use godog build, godog run or godog version.
		// This is to support the legacy direct usage of the root command.
		RunE: func(cmd *cobra.Command, args []string) error {
			if version {
				versionCmdRunFunc(cmd, args)
			}

			if len(output) > 0 {
				buildOutput = output
				buildCmdRunFunc(cmd, args)
			}

			return runCmdRunFunc(cmd, args)
		},
	}

	bindRootCmdFlags(rootCmd.Flags())

	return rootCmd
}

func bindRootCmdFlags(flagSet *flag.FlagSet) {
	flagSet.StringVarP(&output, "output", "o", "", "compiles the test runner to the named file")
	flagSet.BoolVar(&version, "version", false, "show current version")

	flags.BindRunCmdFlags("", flagSet, &opts)

	// Since using the root command directly is deprecated.
	// All flags will be hidden
	flagSet.MarkHidden("output")
	flagSet.MarkHidden("version")
	flagSet.MarkHidden("no-colors")
	flagSet.MarkHidden("concurrency")
	flagSet.MarkHidden("tags")
	flagSet.MarkHidden("format")
	flagSet.MarkHidden("definitions")
	flagSet.MarkHidden("stop-on-failure")
	flagSet.MarkHidden("strict")
	flagSet.MarkHidden("random")
}
