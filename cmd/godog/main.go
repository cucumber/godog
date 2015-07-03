package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/shiena/ansicolor"
)

func buildAndRun() error {
	// will support Ansi colors for windows
	stdout := ansicolor.NewAnsiColorWriter(os.Stdout)

	builtFile := fmt.Sprintf("%s/%dgodog.go", os.TempDir(), time.Now().UnixNano())
	// @TODO: then there is a suite error or panic, it may
	// be interesting to see the built file. But we
	// even cannot determine the status of exit error
	// so leaving it for the future

	buf, err := godog.Build()
	if err != nil {
		return err
	}
	w, err := os.Create(builtFile)
	if err != nil {
		return err
	}
	defer os.Remove(builtFile)
	if _, err = w.Write(buf); err != nil {
		w.Close()
		return err
	}
	w.Close()

	cmd := exec.Command("go", append([]string{"run", builtFile}, os.Args[1:]...)...)
	cmd.Stdout = stdout
	cmd.Stderr = stdout

	return cmd.Run()
}

func main() {
	switch err := buildAndRun().(type) {
	case nil:
	case *exec.ExitError:
		os.Exit(1)
	default:
		panic(err)
	}
}
