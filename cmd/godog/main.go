package main

import (
	"log"
	"os"
	"os/exec"
	"strings"

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

	cmd := strings.TrimSpace("go run " + builtFile + " " + strings.Join(os.Args[1:], " "))
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		panic(err)
	}
	log.Println("output:", string(out))
}
