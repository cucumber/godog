package internal

import (
	"fmt"
	"go/build"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cucumber/godog/internal/builder"
)

var buildOutput string
var buildOutputDefault = "godog.test"

// CreateBuildCmd creates the build subcommand.
func CreateBuildCmd() cobra.Command {
	if build.Default.GOOS == "windows" {
		buildOutputDefault += ".exe"
	}

	buildCmd := cobra.Command{
		Use:   "build",
		Short: "Compiles a test runner",
		Long: `Compiles a test runner. Command should be run from the directory of tested
package and contain buildable go source.

The test runner can be executed with the same flags as when using godog run.`,
		Example: `  godog build
  godog build -o ` + buildOutputDefault,
		RunE: buildCmdRunFunc,
	}

	buildCmd.Flags().StringVarP(&buildOutput, "output", "o", buildOutputDefault, `compiles the test runner to the named file
`)

	return buildCmd
}

func buildCmdRunFunc(cmd *cobra.Command, args []string) error {
	bin, err := filepath.Abs(buildOutput)
	if err != nil {
		return fmt.Errorf("could not locate absolute path for path: %s. reason: %v", buildOutput, err)
	}

	if err = builder.Build(bin); err != nil {
		return fmt.Errorf("could not build binary at: %s. reason: %v", buildOutput, err)
	}

	return nil
}
