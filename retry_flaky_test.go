package godog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RetryFlakySteps(t *testing.T) {
	output := new(bytes.Buffer)

	featureContents := []Feature{
		{
			Name: "RetryFlaky",
			Contents: []byte(`
Feature: retry flaky steps
  Scenario: Test cases that pass aren't retried
    Given a step that always passes

  Scenario: Test cases that fail are retried if within the limit
    Given a step that passes the second time

  Scenario: Test cases that fail will continue to retry up to the limit
    Given a step that passes the third time

  Scenario: Test cases won't retry after failing more than the limit
    Given a step that always fails

  Scenario: Test cases won't retry when the status is UNDEFINED
    Given a non-existent step
`),
		},
	}

	opts := Options{
		NoColors:        true,
		Format:          "pretty",
		Output:          output,
		FeatureContents: featureContents,
		MaxRetries:      3,
	}

	status := TestSuite{
		Name: "retry flaky",
		ScenarioInitializer: func(ctx *ScenarioContext) {
			ctx.Step(`^a step that always passes`, func(ctx context.Context) (context.Context, error) {
				return ctx, nil
			})

			secondTimePass := 0
			ctx.Step(`^a step that passes the second time`, func(ctx context.Context) (context.Context, error) {
				secondTimePass++
				if secondTimePass < 2 {
					return ctx, fmt.Errorf("unexpected network connection, %w", ErrRetry)
				}
				return ctx, nil
			})

			thirdTimePass := 0
			ctx.Step(`^a step that passes the third time`, func(ctx context.Context) (context.Context, error) {
				thirdTimePass++
				if thirdTimePass < 3 {
					return ctx, fmt.Errorf("unexpected network connection, %w", ErrRetry)
				}
				return ctx, nil
			})

			fifthTimePass := 0
			ctx.Step(`^a step that always fails`, func(ctx context.Context) (context.Context, error) {
				fifthTimePass++
				if fifthTimePass < 5 {
					return ctx, fmt.Errorf("must fail, %w", ErrRetry)
				}
				return ctx, nil
			})
		},
		Options: &opts,
	}.Run()

	const expected = `Feature: retry flaky steps

  Scenario: Test cases that pass aren't retried # RetryFlaky:3
    Given a step that always passes             # retry_flaky_test.go:51 -> github.com/cucumber/godog.Test_RetryFlakySteps.func1.1

  Scenario: Test cases that fail are retried if within the limit # RetryFlaky:6
    Given a step that passes the second time                     # retry_flaky_test.go:56 -> github.com/cucumber/godog.Test_RetryFlakySteps.func1.2

  Scenario: Test cases that fail will continue to retry up to the limit # RetryFlaky:9
    Given a step that passes the third time                             # retry_flaky_test.go:65 -> github.com/cucumber/godog.Test_RetryFlakySteps.func1.3

  Scenario: Test cases won't retry after failing more than the limit # RetryFlaky:12
    Given a step that always fails                                   # retry_flaky_test.go:74 -> github.com/cucumber/godog.Test_RetryFlakySteps.func1.4
    must fail, retry step

  Scenario: Test cases won't retry when the status is UNDEFINED # RetryFlaky:15
    Given a non-existent step

--- Failed steps:

  Scenario: Test cases won't retry after failing more than the limit # RetryFlaky:12
    Given a step that always fails # RetryFlaky:13
      Error: must fail, retry step


5 scenarios (3 passed, 1 failed, 1 undefined)
5 steps (3 passed, 1 failed, 1 undefined)
0s

You can implement step definitions for undefined steps with these snippets:

func aNonexistentStep() error {
	return godog.ErrPending
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step(` + "`^a non-existent step$`" + `, aNonexistentStep)
}

`

	actualOutput, err := io.ReadAll(output)
	require.NoError(t, err)

	assert.Equal(t, exitFailure, status)
	assert.Equal(t, expected, string(actualOutput))
}
