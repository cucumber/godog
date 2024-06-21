package interceptor_test

// This example shows how to manipulate the results of test steps in a generic manner.
// Run the sample with : go test -v interceptor_test.go
// Then review the json report and confirm that the statuss report make sense given the scenarios.

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "cucumber", // cucumber json format
}

func TestFeatures(t *testing.T) {
	o := opts
	o.TestingT = t

	status := godog.TestSuite{
		Name:                "intercept",
		Options:             &o,
		ScenarioInitializer: InitializeScenario,
	}.Run()

	if status == 2 {
		t.SkipNow()
	}

	if status != 0 {
		t.Fatalf("zero status code expected, %d received", status)
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {

	ctx.StepContext().Post(func(ctx context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, godog.StepResultStatus, error) {
		if strings.Contains(st.Text, "FLIP_ME") {
			if status == godog.StepFailed {
				fmt.Printf("FLIP_ME to PASSED\n")
				status = godog.StepPassed
				err = nil // knock out any error too
			} else if status == godog.StepPassed {
				fmt.Printf("FLIP_ME to FAILED\n")
				status = godog.StepFailed
				err = fmt.Errorf("FLIP_ME to FAILED")
			} else {
				fmt.Printf("FLIP_ME but stays %v\n", status)
			}
		} else {
			fmt.Printf("NOT FLIP_ME so stays %v\n", status)
		}
		fmt.Printf("FINAL STATUS %v, %v\n", status, err)
		return ctx, status, err
	})

	ctx.Step(`^passing step should be passed$`, func(ctx context.Context) (context.Context, error) {
		// this is a pass and we hope it will be reported as a pass
		return ctx, nil
	})
	ctx.Step(`^failing step should be failed$`, func(ctx context.Context) (context.Context, error) {
		// this is a fail and we hope it will be reported as a fail
		return ctx, fmt.Errorf("intentional failure")
	})
	ctx.Step(`^passing step with the word FLIP_ME should be failed$`, func(ctx context.Context) (context.Context, error) {
		// this is a pass but we hope it will be reported as a fail
		return ctx, nil
	})
	ctx.Step(`^failing step with the word FLIP_ME should be passed$`, func(ctx context.Context) (context.Context, error) {
		// this is a fail but we hope it will be reported as a pass
		return ctx, fmt.Errorf("this failure should be flipped to passed")
	})

}
