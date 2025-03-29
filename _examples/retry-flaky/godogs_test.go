package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

func Test_RetryFlaky(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "retry flaky tests",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:     "pretty",
			Paths:      []string{"features/retry.feature"},
			MaxRetries: 2,
		},
	}

	assert.Equal(t, 0, suite.Run())
}

func InitializeScenario(sc *godog.ScenarioContext) {
	sc.Step(`^a step that always passes`, func(ctx context.Context) (context.Context, error) {
		return ctx, nil
	})

	secondTimePass := 0
	sc.Step(`^a step that passes the second time`, func(ctx context.Context) (context.Context, error) {
		secondTimePass++
		if secondTimePass < 2 {
			return ctx, fmt.Errorf("unexpected network connection, %w", godog.ErrRetry)
		}
		return ctx, nil
	})

	thirdTimePass := 0
	sc.Step(`^a step that passes the third time`, func(ctx context.Context) (context.Context, error) {
		thirdTimePass++
		if thirdTimePass < 3 {
			return ctx, fmt.Errorf("unexpected network connection, %w", godog.ErrRetry)
		}
		return ctx, nil
	})

	fifthTimePass := 0
	sc.Step(`^a step that always fails`, func(ctx context.Context) (context.Context, error) {
		fifthTimePass++
		if fifthTimePass < 5 {
			return ctx, fmt.Errorf("must fail, %w", godog.ErrRetry)
		}
		return ctx, nil
	})
}
