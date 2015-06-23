package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/shiena/ansicolor"
)

func main() {
	// will support Ansi colors for windows
	stdout := ansicolor.NewAnsiColorWriter(os.Stdout)

	builtFile := fmt.Sprintf("%s/%dgodog.go", os.TempDir(), time.Now().UnixNano())

	buf, err := godog.Build()
	if err != nil {
		os.Remove(builtFile)
		panic(err)
	}

	w, err := os.Create(builtFile)
	if err != nil {
		os.Remove(builtFile)
		panic(err)
	}
	_, err = w.Write(buf)
	if err != nil {
		os.Remove(builtFile)
		panic(err)
	}
	w.Close()

	cmd := exec.Command("go")
	cmd.Args = append([]string{"go", "run", builtFile}, os.Args[1:]...)
	cmd.Stdout = stdout
	cmd.Stderr = stdout

	err = cmd.Run()
	switch err.(type) {
	case *exec.ExitError:
		// then there is a suite error, we need to provide a
		// way to see the built file and do not remove it here
		os.Exit(1)
	case *exec.Error:
		os.Remove(builtFile)
		panic(err)
	}
}
