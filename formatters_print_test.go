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

	// inlining steps to have same source code line reference
	suite.Step(`^(?:a )?failing step`, failingStepDef)
	suite.Step(`^(?:a )?pending step$`, pendingStepDef)
	suite.Step(`^(?:a )?passing step$`, passingStepDef)
	suite.Step(`^is <odd> and <even> number$`, oddEvenStepDef)

	pkg := os.Getenv("GODOG_TESTED_PACKAGE")
	os.Setenv("GODOG_TESTED_PACKAGE", "github.com/DATA-DOG/godog")
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
	os.Setenv("GODOG_TESTED_PACKAGE", pkg)
}

func passingStepDef() error { return nil }

func oddEvenStepDef(odd, even int) error { return oddOrEven(odd, even) }

func oddOrEven(odd, even int) error {
	if odd%2 == 0 {
		return fmt.Errorf("%d is not odd", odd)
	}
	if even%2 != 0 {
		return fmt.Errorf("%d is not even", even)
	}
	return nil
}

func pendingStepDef() error { return ErrPending }

func failingStepDef() error { return fmt.Errorf("step failed") }
