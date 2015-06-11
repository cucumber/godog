package main

import (
	"os"

	"github.com/DATA-DOG/godog"
)

func main() {
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
}
