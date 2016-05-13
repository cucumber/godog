package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/shiena/ansicolor"
)

var statusMatch = regexp.MustCompile("^exit status (\\d+)")
var parsedStatus int

func buildAndRun() (int, error) {
	var status int
	// will support Ansi colors for windows
	stdout := ansicolor.NewAnsiColorWriter(os.Stdout)
	stderr := ansicolor.NewAnsiColorWriter(statusOutputFilter(os.Stderr))

	builtFile := fmt.Sprintf("%s/%dgodog.go", os.TempDir(), time.Now().UnixNano())

	buf, err := godog.Build()
	if err != nil {
		return status, err
	}

	w, err := os.Create(builtFile)
	if err != nil {
		return status, err
	}
	defer os.Remove(builtFile)

	if _, err = w.Write(buf); err != nil {
		w.Close()
		return status, err
	}
	w.Close()

	cmd := exec.Command("go", append([]string{"run", builtFile}, os.Args[1:]...)...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

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
		panic(err)
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
