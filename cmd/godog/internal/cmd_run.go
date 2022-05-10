package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/cucumber/godog/internal/builder"
	"github.com/cucumber/godog/internal/flags"
)

var opts flags.Options

// CreateRunCmd creates the run subcommand.
func CreateRunCmd() cobra.Command {
	runCmd := cobra.Command{
		Use:   "run [features]",
		Short: "Compiles and runs a test runner",
		Long: `Compiles and runs test runner for the given feature files.
Command should be run from the directory of tested package and contain
buildable go source.`,
		Example: `  godog run
  godog run <feature>
  godog run <feature> <feature>

  Optional feature(s) to run:
    dir (features/)
    feature (*.feature)
    scenario at specific line (*.feature:10)
  If no feature arguments are supplied, godog will use "features/" by default.`,
		RunE:         runCmdRunFunc,
		SilenceUsage: true,
	}

	flags.BindRunCmdFlags("", runCmd.Flags(), &opts)

	return runCmd
}

func runCmdRunFunc(cmd *cobra.Command, args []string) error {

	osArgs := os.Args[1:]

	if len(osArgs) > 0 && osArgs[0] == "run" {
		osArgs = osArgs[1:]
	}

	if err := buildAndRunGodog(osArgs); err != nil {
		return err
	}

	return nil
}

func buildAndRunGodog(args []string) (err error) {
	bin, err := filepath.Abs(buildOutputDefault)
	if err != nil {
		return err
	}

	if err = builder.Build(bin); err != nil {
		return err
	}

	defer os.Remove(bin)

	return runGodog(bin, args)
}

func runGodog(bin string, args []string) (err error) {
	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	if err = cmd.Start(); err != nil {
		return err
	}

	if err = cmd.Wait(); err == nil {
		return nil
	}

	exiterr, ok := err.(*exec.ExitError)
	if !ok {
		return err
	}

	st, ok := exiterr.Sys().(syscall.WaitStatus)
	if !ok {
		return fmt.Errorf("failed to convert error to syscall wait status. original error: %w", exiterr)
	}

	// This works on both Unix and Windows. Although package
	// syscall is generally platform dependent, WaitStatus is
	// defined for both Unix and Windows and in both cases has
	// an ExitStatus() method with the same signature.
	if st.ExitStatus() > 0 {
		return err
	}

	return nil
}
