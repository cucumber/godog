package formatters_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cucumber/godog"
)

const fmtOutputTestsFeatureDir = "formatter-tests/features"

func Test_FmtOutput(t *testing.T) {
	pkg := os.Getenv("GODOG_TESTED_PACKAGE")
	os.Setenv("GODOG_TESTED_PACKAGE", "github.com/cucumber/godog")

	featureFiles, err := listFmtOutputTestsFeatureFiles()
	require.Nil(t, err)

	formatters := []string{"cucumber", "events", "junit", "pretty", "progress", "junit,pretty"}

	for _, fmtName := range formatters {
		for _, featureFile := range featureFiles {
			testName := fmt.Sprintf("%s/%s", fmtName, featureFile)
			featureFilePath := fmt.Sprintf("%s/%s", fmtOutputTestsFeatureDir, featureFile)
			t.Run(testName, fmtOutputTest(fmtName, testName, featureFilePath))
		}
	}

	os.Setenv("GODOG_TESTED_PACKAGE", pkg)
}

func listFmtOutputTestsFeatureFiles() (featureFiles []string, err error) {
	err = filepath.Walk(fmtOutputTestsFeatureDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			featureFiles = append(featureFiles, info.Name())
			return nil
		}

		if info.Name() == "features" {
			return nil
		}

		return filepath.SkipDir
	})

	return
}

func fmtOutputTest(fmtName, testName, featureFilePath string) func(*testing.T) {
	fmtOutputScenarioInitializer := func(ctx *godog.ScenarioContext) {
		ctx.Step(`^(?:a )?failing step`, failingStepDef)
		ctx.Step(`^(?:a )?pending step$`, pendingStepDef)
		ctx.Step(`^(?:a )?passing step$`, passingStepDef)
		ctx.Step(`^odd (\d+) and even (\d+) number$`, oddEvenStepDef)
	}

	return func(t *testing.T) {
		expectOutputPath := strings.Replace(featureFilePath, "features", fmtName, 1)
		expectOutputPath = strings.TrimSuffix(expectOutputPath, path.Ext(expectOutputPath))
		if _, err := os.Stat(expectOutputPath); err != nil {
			t.Skipf("Couldn't find expected output file %q", expectOutputPath)
		}

		expectedOutput, err := ioutil.ReadFile(expectOutputPath)
		require.NoError(t, err)

		var buf bytes.Buffer
		out := &tagColorWriter{w: &buf}

		opts := godog.Options{
			Format: fmtName,
			Paths:  []string{featureFilePath},
			Output: out,
		}

		godog.TestSuite{
			Name:                fmtName,
			ScenarioInitializer: fmtOutputScenarioInitializer,
			Options:             &opts,
		}.Run()

		expected := string(expectedOutput)
		actual := buf.String()
		assert.Equalf(t, expected, actual, "path: %s", expectOutputPath)
	}
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

func pendingStepDef() error { return godog.ErrPending }

func failingStepDef() error { return fmt.Errorf("step failed") }
