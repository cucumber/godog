package formatters_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/internal/utils"

	messages "github.com/cucumber/messages/go/v21"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const inputFeatureFilesDir = "formatter-tests/features"
const expectedOutputFilesDir = "formatter-tests"

var tT *testing.T

// Set this to true to generate output files for new tests or to refresh old ones.
// You MUST then manually verify the generated file before committing.
// Leave this as false for normal testing.
const regenerateOutputs = false

// Set to true to generate output containing ansi colour sequences as opposed to tags like '<red>'.
// this useful for confirming what the end user would see.
// Leave this as false for normal testing.
const generateRawOutput = false

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
			featureFilePath := fmt.Sprintf("%s/%s", inputFeatureFilesDir, featureFile)

			testSuite := fmtOutputTest(fmtName, featureFilePath)
			t.Run(testName, testSuite)
		}
	}

	os.Setenv("GODOG_TESTED_PACKAGE", pkg)
}

func listFmtOutputTestsFeatureFiles() (featureFiles []string, err error) {

	err = filepath.Walk(inputFeatureFilesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			featureFiles = append(featureFiles, info.Name())
			return nil
		}

		return nil
	})

	return
}

func fmtOutputTest(fmtName, featureFilePath string) func(*testing.T) {
	fmtOutputScenarioInitializer := func(ctx *godog.ScenarioContext) {
		stepIndex := 0
		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			if tagged(sc.Tags, "@fail_before_scenario") {
				return ctx, fmt.Errorf("failed in before scenario hook")
			}

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
			if tagged(sc.Tags, "@fail_after_scenario") {
				return ctx, fmt.Errorf("failed in after scenario hook")
			}

			if strings.Contains(sc.Name, "attachment") {
				att := godog.Attachments(ctx)
				attCount := len(att)
				if attCount != 4 {
					assert.FailNow(tT, "Unexpected attachments: "+sc.Name, "expected 4, found %d", attCount)
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
		ctx.Step(`^ambiguous step.*$`, ambiguousStepDef)
		ctx.Step(`^ambiguous step$`, ambiguousStepDef)
		ctx.Step(`^odd (\d+) and even (\d+) number$`, oddEvenStepDef)
		ctx.Step(`^(?:a )?a step with a single attachment call for multiple attachments$`, stepWithSingleAttachmentCall)
		ctx.Step(`^(?:a )?a step with multiple attachment calls$`, stepWithMultipleAttachmentCalls)
	}

	return func(t *testing.T) {
		expectOutputPath := strings.Replace(featureFilePath, inputFeatureFilesDir, expectedOutputFilesDir+"/"+fmtName, 1)
		expectOutputPath = strings.TrimSuffix(expectOutputPath, ".feature")
		fmt.Printf("fmt_output_test for format %10s : feature file %v\n", fmtName, featureFilePath)

		//goland:noinspection
		if !regenerateOutputs {
			if _, err := os.Stat(expectOutputPath); err != nil {
				// the test author needs to write an "expected output" file for any formats they want the test feature to be verified against
				t.Skipf("Skipping test for feature %q for format %q, because no 'expected output' file %q", featureFilePath, fmtName, expectOutputPath)
			}
		}

		var buf bytes.Buffer

		opts := godog.Options{
			Format: fmtName,
			Paths:  []string{featureFilePath},
			Output: godog.NopCloser(io.Writer(&buf)),
			Strict: true,
		}

		godog.TestSuite{
			Name:                fmtName,
			ScenarioInitializer: fmtOutputScenarioInitializer,
			Options:             &opts,
		}.Run()

		out := translateAnsiEscapeToTags(t, buf)

		//goland:noinspection
		if regenerateOutputs {
			data := []byte(out)
			if generateRawOutput {
				data = buf.Bytes()
			}
			err := os.WriteFile(expectOutputPath, data, 0644)
			require.NoError(t, err)
		} else {
			// normalise on unix line ending so expected vs actual works cross-platform
			actual := normalise(out)

			expectedOutput, err := os.ReadFile(expectOutputPath)
			require.NoError(t, err)

			expected := normalise(string(expectedOutput))

			assert.Equalf(t, expected, actual, "path: %s", expectOutputPath)

			// display as a side by side listing as the output of the assert is all one line with embedded newlines and useless
			if expected != actual {
				fmt.Printf("Error: fmt: %s, path: %s\n", fmtName, expectOutputPath)
				utils.VDiffString(expected, actual)
			}
		}
	}
}

func translateAnsiEscapeToTags(t *testing.T, buf bytes.Buffer) string {
	var bufTagged bytes.Buffer

	outTagged := &tagColorWriter{w: &bufTagged}

	_, err := outTagged.Write(buf.Bytes())
	require.NoError(t, err)

	colourTagged := bufTagged.String()
	return colourTagged
}

func tagged(tags []*messages.PickleTag, tagName string) bool {
	for _, tag := range tags {
		if tag.Name == tagName {
			return true
		}
	}
	return false

}

func normalise(s string) string {

	m := regexp.MustCompile("fmt_output_test.go:[0-9]+")
	normalised := m.ReplaceAllString(s, "fmt_output_test.go:XXX")
	normalised = strings.Replace(normalised, "\r\n", "\n", -1)
	normalised = strings.Replace(normalised, "\\r\\n", "\\n", -1)

	return normalised
}

func passingStepDef() error {
	return nil
}

func ambiguousStepDef() error {
	return nil
}

func oddEvenStepDef(odd, even int) error {
	return oddOrEven(odd, even)
}

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
