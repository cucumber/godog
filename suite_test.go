package godog

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	format := "progress" // non verbose mode
	concurrency := 4

	var specific bool
	for _, arg := range os.Args[1:] {
		if arg == "-test.v=true" { // go test transforms -v option - verbose mode
			format = "pretty"
			concurrency = 1
			break
		}
		if strings.Index(arg, "-test.run") == 0 {
			specific = true
		}
	}
	var status int
	if !specific {
		status = RunWithOptions("godog", func(s *Suite) {
			GodogContext(s)
		}, Options{
			Format:      format, // pretty format for verbose mode, otherwise - progress
			Paths:       []string{"features"},
			Concurrency: concurrency,           // concurrency for verbose mode is 1
			Randomize:   time.Now().UnixNano(), // randomize scenario execution order
		})
	}

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

// needed in order to use godog cli
func GodogContext(s *Suite) {
	SuiteContext(s)
}
