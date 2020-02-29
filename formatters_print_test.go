package godog

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintingFormatters(t *testing.T) {
	features, err := parseFeatures("", []string{"formatter-tests"})
	require.NoError(t, err)

	var buf bytes.Buffer
	out := &tagColorWriter{w: &buf}

	suite := &Suite{
		features: features,
	}

	// inlining steps to have same source code line reference
	suite.Step(`^(?:a )?failing step`, failingStepDef)
	suite.Step(`^(?:a )?pending step$`, pendingStepDef)
	suite.Step(`^(?:a )?passing step$`, passingStepDef)
	suite.Step(`^odd (\d+) and even (\d+) number$`, oddEvenStepDef)

	pkg := os.Getenv("GODOG_TESTED_PACKAGE")
	os.Setenv("GODOG_TESTED_PACKAGE", "github.com/cucumber/godog")
	for _, feat := range features {
		for name := range AvailableFormatters() {
			expectOutputPath := strings.Replace(feat.Path, "features", name, 1)
			expectOutputPath = strings.TrimSuffix(expectOutputPath, path.Ext(expectOutputPath))
			if _, err := os.Stat(expectOutputPath); err != nil {
				continue
			}

			buf.Reset()                          // flush the output
			suite.fmt = FindFmt(name)(name, out) // prepare formatter
			suite.features = []*feature{feat}    // set the feature

			expectedOutput, err := ioutil.ReadFile(expectOutputPath)
			require.NoError(t, err)

			suite.run()
			suite.fmt.Summary()

			expected := string(expectedOutput)
			expected = trimAllLines(expected)

			actual := buf.String()
			actual = trimAllLines(actual)

			assert.Equalf(t, expected, actual, "expected: [%s], actual: [%s]", expected, actual)
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
