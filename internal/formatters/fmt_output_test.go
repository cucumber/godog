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
		stepIndex := 0
		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			if strings.Contains(sc.Name, "attachment") {
				att := godog.Attachments(ctx)
				attCount := len(att)
				if attCount != 0 {
					assert.FailNowf(tT, "Unexpected attachments: "+sc.Name, "should have been empty, found %d", attCount)
				}

				ctx = godog.Attach(ctx,
					godog.Attachment{Body: []byte("BeforeScenarioAttachment"), FileName: "Before Scenario Attachment 1", MediaType: "text/plain"},
				)
			}
			return ctx, nil
		})

		ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {

			if strings.Contains(sc.Name, "attachment") {
				att := godog.Attachments(ctx)
				attCount := len(att)
				if attCount != 4 {
					assert.FailNow(tT, "Unexpected attachements: "+sc.Name, "expected 4, found %d", attCount)
				}
				ctx = godog.Attach(ctx,
					godog.Attachment{Body: []byte("AfterScenarioAttachment"), FileName: "After Scenario Attachment 2", MediaType: "text/plain"},
				)
			}
			return ctx, nil
		})

		ctx.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
			stepIndex++

			if strings.Contains(st.Text, "attachment") {
				att := godog.Attachments(ctx)
				attCount := len(att)

				// 1 for before scenario ONLY if this is the 1st step
				expectedAttCount := 0
				if stepIndex == 1 {
					expectedAttCount = 1
				}

				if attCount != expectedAttCount {
					assert.FailNow(tT, "Unexpected attachments: "+st.Text, "expected 1, found %d\n%+v", attCount, att)
				}
				ctx = godog.Attach(ctx,
					godog.Attachment{Body: []byte("BeforeStepAttachment"), FileName: "Before Step Attachment 3", MediaType: "text/plain"},
				)
			}
			return ctx, nil
		})
		ctx.StepContext().After(func(ctx context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {

			if strings.Contains(st.Text, "attachment") {
				att := godog.Attachments(ctx)
				attCount := len(att)

				// 1 for before scenario ONLY if this is the 1st step
				// 1 for before before step
				// 2 from from step
				expectedAttCount := 3
				if stepIndex == 1 {
					expectedAttCount = 4
				}

				if attCount != expectedAttCount {
					// 1 from before scenario, 1 from before step, 1 from step
					assert.FailNow(tT, "Unexpected attachments: "+st.Text, "expected 4, found %d", attCount)
				}
				ctx = godog.Attach(ctx,
					godog.Attachment{Body: []byte("AfterStepAttachment"), FileName: "After Step Attachment 4", MediaType: "text/plain"},
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
			// the test author needs to write an "expected output" file for any formats they want the test feature to be verified against
			t.Skipf("Skipping test for feature '%v' for format '%v', because no 'expected output' file %q", featureFilePath, fmtName, expectOutputPath)
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
	aCount := len(godog.Attachments(ctx))
	if aCount != 2 {
		// 1 from before scenario, 1 from before step
		assert.FailNowf(tT, "Unexpected Attachments found", "should have been 2, but found %v", aCount)
	}

	ctx = godog.Attach(ctx,
		godog.Attachment{Body: []byte("TheData1"), FileName: "TheFilename1", MediaType: "text/plain"},
		godog.Attachment{Body: []byte("TheData2"), FileName: "TheFilename2", MediaType: "text/plain"},
	)

	return ctx, nil
}
func stepWithMultipleAttachmentCalls(ctx context.Context) (context.Context, error) {
	aCount := len(godog.Attachments(ctx))
	if aCount != 1 {
		assert.FailNowf(tT, "Unexpected Attachments found", "Expected 1 Attachment, but found %v", aCount)
	}

	ctx = godog.Attach(ctx,
		godog.Attachment{Body: []byte("TheData1"), FileName: "TheFilename3", MediaType: "text/plain"},
	)
	ctx = godog.Attach(ctx,
		godog.Attachment{Body: []byte("TheData2"), FileName: "TheFilename4", MediaType: "text/plain"},
	)

	return ctx, nil
}
