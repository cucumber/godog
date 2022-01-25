package internal

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cucumber/godog"
)

// CreateVersionCmd creates the version subcommand.
func CreateVersionCmd() cobra.Command {
	versionCmd := cobra.Command{
		Use:     "version",
		Short:   "Show current version",
		Version: godog.Version,
	}

	return versionCmd
}

func versionCmdRunFunc(cmd *cobra.Command, args []string) {
	fmt.Fprintln(os.Stdout, "Godog version is:", godog.Version)
}
