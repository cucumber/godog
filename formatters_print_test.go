package godog

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestPrintingFormatters(t *testing.T) {
	features, err := parseFeatures("", []string{"formatter-tests"})
	if err != nil {
		t.Fatalf("failed to parse formatter features: %v", err)
	}

	var buf bytes.Buffer
	out := &tagColorWriter{w: &buf}

	suite := &Suite{
		features: features,
	}

	suite.Step(`^(?:a )?failing step`, func() error {
		return fmt.Errorf("step failed")
	})
	suite.Step(`^this step should fail`, func() error {
		return fmt.Errorf("step failed")
	})
	suite.Step(`^(?:a )?pending step$`, func() error {
		return ErrPending
	})
	suite.Step(`^(?:a )?passing step$`, func() error {
		return nil
	})

	for _, feat := range features {
		for name := range AvailableFormatters() {
			expectOutputPath := strings.Replace(feat.Path, "features", name, 1)
			expectOutputPath = strings.TrimRight(expectOutputPath, ".feature")
			if _, err := os.Stat(expectOutputPath); err != nil {
				continue
			}

			buf.Reset()                          // flush the output
			suite.fmt = FindFmt(name)(name, out) // prepare formatter
			suite.features = []*feature{feat}    // set the feature

			expectedOutput, err := ioutil.ReadFile(expectOutputPath)
			if err != nil {
				t.Fatal(err)
			}

			suite.run()
			suite.fmt.Summary()

			expected := string(expectedOutput)
			actual := buf.String()

			if actual != expected {
				t.Fatalf("%s does not match to:\n%s", expectOutputPath, actual)
			}
		}
	}
}
