package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var location string
	switch {
	case os.Args[1] != "":
		location = os.Args[1]
	default:
		location = "."
	}
	if err := ls(location, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func ls(path string, w io.Writer) error {
	return filepath.Walk(path, func(p string, f os.FileInfo, err error) error {
		switch {
		case f.IsDir() && f.Name() != "." && f.Name() != ".." && filepath.Base(path) != f.Name():
			w.Write([]byte(f.Name() + "\n"))
			return filepath.SkipDir
		case !f.IsDir():
			w.Write([]byte(f.Name() + "\n"))
		}
		return err
	})
}
