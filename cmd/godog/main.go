package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/DATA-DOG/godog"
	"github.com/shiena/ansicolor"
)

func main() {
	// will support Ansi colors for windows
	stdout := ansicolor.NewAnsiColorWriter(os.Stdout)
	stderr := ansicolor.NewAnsiColorWriter(os.Stdout)

	builtFile := os.TempDir() + "/godog_build.go"
	if err := os.Remove(builtFile); err != nil && !os.IsNotExist(err) {
		panic(err)
	}

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
	cmd := exec.Command("sh", "-c", c)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}
