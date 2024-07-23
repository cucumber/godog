package formatters_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cucumber/godog"
)

const fmtOutputTestsFeatureDir = "formatter-tests/features"

var tT *testing.T

func Test_FmtOutput(t *testing.T) {
	tT = t
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

		ctx.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
			att := godog.Attachments(ctx)
			attCount := len(att)
			if attCount > 0 {
				assert.FailNow(tT, fmt.Sprintf("Unexpected Attachments found - should have been empty, found %d\n%+v", attCount, att))
			}

			if st.Text == "a step with multiple attachment calls" {
				ctx = godog.Attach(ctx,
					godog.Attachment{Body: []byte("BeforeStepAttachment"), FileName: "Data Attachment", MediaType: "text/plain"},
				)
			}
			return ctx, nil
		})
		ctx.StepContext().After(func(ctx context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {

			if st.Text == "a step with multiple attachment calls" {
				att := godog.Attachments(ctx)
				attCount := len(att)
				if attCount != 3 {
					assert.FailNow(tT, fmt.Sprintf("Expected 3 Attachments - 1 from the before step and 2 from the step, found %d\n%+v", attCount, att))
				}
				ctx = godog.Attach(ctx,
					godog.Attachment{Body: []byte("AfterStepAttachment"), FileName: "Data Attachment", MediaType: "text/plain"},
				)
			}
			return ctx, nil
		})

		ctx.Step(`^(?:a )?failing step`, failingStepDef)
		ctx.Step(`^(?:a )?pending step$`, pendingStepDef)
		ctx.Step(`^(?:a )?passing step$`, passingStepDef)
		ctx.Step(`^odd (\d+) and even (\d+) number$`, oddEvenStepDef)
		ctx.Step(`^(?:a )?a step with a single attachment call for multiple attachments$`, stepWithSingleAttachmentCall)
		ctx.Step(`^(?:a )?a step with multiple attachment calls$`, stepWithMultipleAttachmentCalls)
	}

	return func(t *testing.T) {
		expectOutputPath := strings.Replace(featureFilePath, "features", fmtName, 1)
		expectOutputPath = strings.TrimSuffix(expectOutputPath, path.Ext(expectOutputPath))
		if _, err := os.Stat(expectOutputPath); err != nil {
			t.Skipf("Couldn't find expected output file %q", expectOutputPath)
		}

		expectedOutput, err := os.ReadFile(expectOutputPath)
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

		// normalise on unix line ending so expected vs actual works cross platform
		expected := normalise(string(expectedOutput))
		actual := normalise(buf.String())
		assert.Equalf(t, expected, actual, "path: %s", expectOutputPath)
	}
}

func normalise(s string) string {

	m := regexp.MustCompile("fmt_output_test.go:[0-9]+")
	normalised := m.ReplaceAllString(s, "fmt_output_test.go:XXX")
	normalised = strings.Replace(normalised, "\r\n", "\n", -1)
	normalised = strings.Replace(normalised, "\\r\\n", "\\n", -1)

	return normalised
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

func stepWithSingleAttachmentCall(ctx context.Context) (context.Context, error) {
	if len(godog.Attachments(ctx)) > 0 {
		assert.FailNow(tT, "Unexpected Attachments found - should have been empty")
	}

	ctx = godog.Attach(ctx,
		godog.Attachment{Body: []byte("TheData1"), FileName: "TheFilename1", MediaType: "text/plain"},
		godog.Attachment{Body: []byte("TheData2"), FileName: "TheFilename2", MediaType: "text/plain"},
	)

	return ctx, nil
}
func stepWithMultipleAttachmentCalls(ctx context.Context) (context.Context, error) {
	if len(godog.Attachments(ctx)) != 1 {
		assert.FailNow(tT, "Expected 1 Attachment that should have been inserted by before step")
	}

	ctx = godog.Attach(ctx,
		godog.Attachment{Body: []byte("TheData1"), FileName: "TheFilename1", MediaType: "text/plain"},
	)
	ctx = godog.Attach(ctx,
		godog.Attachment{Body: []byte("TheData2"), FileName: "TheFilename2", MediaType: "text/plain"},
	)

	return ctx, nil
}
