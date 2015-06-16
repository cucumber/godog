package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/shiena/ansicolor"
)

func main() {
	// will support Ansi colors for windows
	stdout := ansicolor.NewAnsiColorWriter(os.Stdout)

	builtFile := fmt.Sprintf("%s/%dgodog.go", os.TempDir(), time.Now().UnixNano())
	defer os.Remove(builtFile) // comment out for debug

	buf, err := godog.Build()
	if err != nil {
		panic(err)
	}

	w, err := os.Create(builtFile)
	if err != nil {
		panic(err)
	}
	_, err = w.Write(buf)
	if err != nil {
		panic(err)
	}
	w.Close()

	c := strings.TrimSpace("go run " + builtFile + " " + strings.Join(os.Args[1:], " "))
	// @TODO: support for windows
	cmd := exec.Command("sh", "-c", c)
	cmd.Stdout = stdout

	err = cmd.Run()
	switch err.(type) {
	case *exec.ExitError:
		os.Exit(1)
	case *exec.Error:
		panic(err)
	}
}
