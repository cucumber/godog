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
		Run: runCmdRunFunc,
	}

	flags.BindRunCmdFlags("", runCmd.Flags(), &opts)

	return runCmd
}

func runCmdRunFunc(cmd *cobra.Command, args []string) {
	osArgs := os.Args[1:]

	if len(osArgs) > 0 && osArgs[0] == "run" {
		osArgs = osArgs[1:]
	}

	status, err := buildAndRunGodog(osArgs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	os.Exit(status)
}

func buildAndRunGodog(args []string) (_ int, err error) {
	bin, err := filepath.Abs(buildOutputDefault)
	if err != nil {
		return 1, err
	}

	if err = builder.Build(bin); err != nil {
		return 1, err
	}

	defer os.Remove(bin)

	return runGodog(bin, args)
}

func runGodog(bin string, args []string) (_ int, err error) {
	cmd := exec.Command(bin, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	if err = cmd.Start(); err != nil {
		return 0, err
	}

	if err = cmd.Wait(); err == nil {
		return 0, nil
	}

	exiterr, ok := err.(*exec.ExitError)
	if !ok {
		return 0, err
	}

	// This works on both Unix and Windows. Although package
	// syscall is generally platform dependent, WaitStatus is
	// defined for both Unix and Windows and in both cases has
	// an ExitStatus() method with the same signature.
	if st, ok := exiterr.Sys().(syscall.WaitStatus); ok {
		return st.ExitStatus(), nil
	}

	return 1, nil
}
