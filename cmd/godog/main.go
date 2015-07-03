package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/shiena/ansicolor"
)

var statusMatch = regexp.MustCompile("^exit status (\\d+)$")

func buildAndRun() (status int, err error) {
	// will support Ansi colors for windows
	stdout := ansicolor.NewAnsiColorWriter(os.Stdout)
	buffer := bytes.NewBuffer([]byte(""))
	stderr := ansicolor.NewAnsiColorWriter(buffer)

	builtFile := fmt.Sprintf("%s/%dgodog.go", os.TempDir(), time.Now().UnixNano())

	buf, err := godog.Build()
	if err != nil {
		return
	}

	w, err := os.Create(builtFile)
	if err != nil {
		return
	}
	defer os.Remove(builtFile)

	if _, err = w.Write(buf); err != nil {
		w.Close()
		return
	}
	w.Close()

	cmd := exec.Command("go", append([]string{"run", builtFile}, os.Args[1:]...)...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	defer func() {
		s := strings.TrimSpace(buffer.String())
		if s == "" {
			status = 0
		} else if m := statusMatch.FindStringSubmatch(s); len(m) > 1 {
			status, _ = strconv.Atoi(m[1])
		} else {
			io.Copy(stdout, buffer)
		}
	}()

	return status, cmd.Run()
}

func main() {
	status, err := buildAndRun()
	switch e := err.(type) {
	case nil:
	case *exec.ExitError:
		os.Exit(status)
	default:
		panic(e)
	}
}
