package internal

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"log"

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
		RunE: runRootCmd,
	}

	err := bindRootCmdFlags(rootCmd.Flags())
	if err != nil {
		log.Println(err)
	}

	return rootCmd
}

func runRootCmd(cmd *cobra.Command, args []string) error {
	if version {
		versionCmdRunFunc(cmd, args)
		return nil
	}

	if len(output) > 0 {
		buildOutput = output
		if err := buildCmdRunFunc(cmd, args); err != nil {
			return err
		}
	}

	return runCmdRunFunc(cmd, args)
}

func bindRootCmdFlags(flagSet *pflag.FlagSet) error {
	flagSet.StringVarP(&output, "output", "o", "", "compiles the test runner to the named file")
	flagSet.BoolVar(&version, "version", false, "show current version")

	flags.BindRunCmdFlags("", flagSet, &opts)

	// Since using the root command directly is deprecated.
	// All flags will be hidden
	for _, name := range []string{
		"output",
		"version",
		"no-colors",
		"concurrency",
		"tags",
		"format",
		"definitions",
		"stop-on-failure",
		"strict",
		"random",
	} {
		if err := flagSet.MarkHidden(name); err != nil {
			return fmt.Errorf("failed to hide deprecated root command flag %q: %w", name, err)
		}
	}
	return nil
}
