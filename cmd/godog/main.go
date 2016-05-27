package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/DATA-DOG/godog"
)

var statusMatch = regexp.MustCompile("^exit status (\\d+)")
var parsedStatus int

func buildAndRun() (int, error) {
	var status int
	// will support Ansi colors for windows
	stdout := createAnsiColorWriter(os.Stdout)
	stderr := createAnsiColorWriter(statusOutputFilter(os.Stderr))

	dir := fmt.Sprintf(filepath.Join("%s", "godog-%d"), os.TempDir(), time.Now().UnixNano())
	err := godog.Build(dir)
	if err != nil {
		return 1, err
	}

	defer os.RemoveAll(dir)

	wd, err := os.Getwd()
	if err != nil {
		return 1, err
	}
	bin := filepath.Join(wd, "godog.test")

	cmdb := exec.Command("go", "test", "-c", "-o", bin)
	cmdb.Dir = dir
	cmdb.Env = os.Environ()
	if details, err := cmdb.CombinedOutput(); err != nil {
		fmt.Fprintln(stderr, string(details))
		return 1, err
	}
	defer os.Remove(bin)

	cmd := exec.Command(bin, os.Args[1:]...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
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
