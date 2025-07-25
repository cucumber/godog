package internal

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/cucumber/godog"
)

// CreateVersionCmd creates the version subcommand.
func CreateVersionCmd() cobra.Command {
	versionCmd := cobra.Command{
		Use:     "version",
		Short:   "Show current version",
		Run:     versionCmdRunFunc,
		Version: godog.Version,
	}

	return versionCmd
}

func versionCmdRunFunc(cmd *cobra.Command, args []string) {
	if _, err := fmt.Fprintln(os.Stdout, "Godog version is:", godog.Version); err != nil {
		log.Fatalf("failed to print Godog version: %v", err)
	}
}
