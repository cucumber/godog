package internal

import (
	"fmt"
	"go/build"
	"os"
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
		Run: buildCmdRunFunc,
	}

	buildCmd.Flags().StringVarP(&buildOutput, "output", "o", buildOutputDefault, "compiles the test runner to the named file")

	return buildCmd
}

func buildCmdRunFunc(cmd *cobra.Command, args []string) {
	bin, err := filepath.Abs(buildOutput)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not locate absolute path for:", buildOutput, err)
		os.Exit(1)
	}

	if err = builder.Build(bin); err != nil {
		fmt.Fprintln(os.Stderr, "could not build binary at:", buildOutput, err)
		os.Exit(1)
	}

	os.Exit(0)
}
