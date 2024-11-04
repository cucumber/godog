package formatters_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/internal/flags"
	"github.com/stretchr/testify/assert"
)

func TestBase_Summary(t *testing.T) {
	var features []flags.Feature

	features = append(features,
		flags.Feature{Name: "f1", Contents: []byte(`
Feature: f1

Scenario: f1s1
When step passed f1s1:1
Then step failed f1s1:2
		`)},
		flags.Feature{Name: "f2", Contents: []byte(`
Feature: f2

Scenario: f2s1
When step passed f2s1:1
Then step passed f2s1:2

Scenario: f2s2
When step failed f2s2:1
Then step passed f2s2:2

Scenario: f2s3
When step passed f2s3:1
Then step skipped f2s3:2
And step passed f2s3:3
And step failed f2s3:4

Scenario: f2s4
When step passed f2s4:1
Then step is undefined f2s4:2
And step passed f2s4:3
`)},
	)

	out := bytes.NewBuffer(nil)
	suite := godog.TestSuite{
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
				if err != nil {
					_, _ = out.WriteString(fmt.Sprintf("scenario %q ended with error %q\n", sc.Name, err.Error()))
				} else {
					_, _ = out.WriteString(fmt.Sprintf("scenario %q passed\n", sc.Name))
				}

				return ctx, nil
			})
			sc.StepContext().After(func(ctx context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
				_, _ = out.WriteString(fmt.Sprintf("step %q finished with status %s\n", st.Text, status.String()))
				return ctx, nil
			})
			sc.Step("failed (.+)", func(s string) error {
				_, _ = out.WriteString(fmt.Sprintf("\nstep invoked: %q, failed\n", s))
				return errors.New("failed")
			})
			sc.Step("skipped (.+)", func(s string) error {
				_, _ = out.WriteString(fmt.Sprintf("\nstep invoked: %q, skipped\n", s))
				return godog.ErrSkip
			})
			sc.Step("passed (.+)", func(s string) {
				_, _ = out.WriteString(fmt.Sprintf("\nstep invoked: %q, passed\n", s))
			})
		},
		Options: &godog.Options{
			Output:          out,
			NoColors:        true,
			Strict:          true,
			Format:          "progress",
			FeatureContents: features,
		},
	}

	assert.Equal(t, 1, suite.Run())
	//TODO - test hard to interpret/maintain as it's mixing output from the formatter and stuff from user defined step glue
	// and then asserting exact interleaving of the merged output, but this exact sequencing isn't a functional requirement
	assert.Equal(t, `
step invoked: "f1s1:1", passed
step "step passed f1s1:1" finished with status passed
.
step invoked: "f1s1:2", failed
step "step failed f1s1:2" finished with status failed
scenario "f1s1" ended with error "failed"
F
step invoked: "f2s1:1", passed
step "step passed f2s1:1" finished with status passed
.
step invoked: "f2s1:2", passed
step "step passed f2s1:2" finished with status passed
scenario "f2s1" passed
.
step invoked: "f2s2:1", failed
step "step failed f2s2:1" finished with status failed
scenario "f2s2" ended with error "failed"
Fstep "step passed f2s2:2" finished with status skipped
-
step invoked: "f2s3:1", passed
step "step passed f2s3:1" finished with status passed
.
step invoked: "f2s3:2", skipped
step "step skipped f2s3:2" finished with status skipped
-step "step passed f2s3:3" finished with status skipped
-step "step failed f2s3:4" finished with status skipped
scenario "f2s3" passed
-
step invoked: "f2s4:1", passed
step "step passed f2s4:1" finished with status passed
.step "step is undefined f2s4:2" finished with status undefined
scenario "f2s4" ended with error "step is undefined"
Ustep "step passed f2s4:3" finished with status skipped
- 13


--- Failed steps:

  Scenario: f1s1 # f1:4
    Then step failed f1s1:2 # f1:6
      Error: failed

  Scenario: f2s2 # f2:8
    When step failed f2s2:1 # f2:9
      Error: failed


5 scenarios (2 passed, 2 failed, 1 undefined)
13 steps (5 passed, 2 failed, 1 undefined, 5 skipped)
0s

You can implement step definitions for undefined steps with these snippets:

func stepIsUndefinedFS(arg1, arg2, arg3 int) error {
	return godog.ErrPending
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(`+"`"+`^step is undefined f(\d+)s(\d+):(\d+)$`+"`"+`, stepIsUndefinedFS)
}

`, out.String())
}
