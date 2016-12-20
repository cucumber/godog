package main

import (
	"fmt"
	"go/build"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"syscall"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/colors"
)

var statusMatch = regexp.MustCompile("^exit status (\\d+)")
var parsedStatus int

func buildAndRun() (int, error) {
	var status int

	bin, err := filepath.Abs("godog.test")
	if err != nil {
		return 1, err
	}
	if build.Default.GOOS == "windows" {
		bin += ".exe"
	}
	if err = godog.Build(bin); err != nil {
		return 1, err
	}
	defer os.Remove(bin)

	cmd := exec.Command(bin, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	if err = cmd.Start(); err != nil {
		return status, err
	}

	if err = cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0
			status = 1

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if st, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				status = st.ExitStatus()
			}
			return status, nil
		}
		return status, err
	}
	return status, nil
}

func main() {
	var vers bool
	var output string

	opt := godog.Options{Output: colors.Colored(os.Stdout)}
	flagSet := godog.FlagSet(&opt)
	flagSet.BoolVar(&vers, "version", false, "Show current version.")
	flagSet.StringVar(&output, "o", "", "Build and output test runner executable to given target path.")
	flagSet.StringVar(&output, "output", "", "Build and output test runner executable to given target path.")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(output) > 0 {
		bin, err := filepath.Abs(output)
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not locate absolute path for:", output, err)
			os.Exit(1)
		}
		if err = godog.Build(bin); err != nil {
			fmt.Fprintln(os.Stderr, "could not build binary at:", output, err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if vers {
		fmt.Fprintln(os.Stdout, "Godog version is:", godog.Version)
		os.Exit(0) // should it be 0?
	}

	status, err := buildAndRun()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// it might be a case, that status might not be resolved
	// in some OSes. this is attempt to parse it from stderr
	if parsedStatus > status {
		status = parsedStatus
	}
	os.Exit(status)
}

func statusOutputFilter(w io.Writer) io.Writer {
	return writerFunc(func(b []byte) (int, error) {
		if m := statusMatch.FindStringSubmatch(string(b)); len(m) > 1 {
			parsedStatus, _ = strconv.Atoi(m[1])
			// skip status stderr output
			return len(b), nil
		}
		return w.Write(b)
	})
}

type writerFunc func([]byte) (int, error)

func (w writerFunc) Write(b []byte) (int, error) {
	return w(b)
}
