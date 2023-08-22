package godogs_test

// This example shows how to set up test suite runner with Go subtests and godog command line parameters.
// Sample commands:
// * run all scenarios from default directory (features): go test -test.run "^TestFeatures/"
// * run all scenarios and list subtest names: go test -test.v -test.run "^TestFeatures/"
// * run all scenarios from one feature file: go test -test.v -godog.paths features/nodogs.feature -test.run "^TestFeatures/"
// * run all scenarios from multiple feature files: go test -test.v -godog.paths features/nodogs.feature,features/godogs.feature -test.run "^TestFeatures/"
// * run single scenario as a subtest: go test -test.v -test.run "^TestFeatures/Eat_5_out_of_12$"
// * show usage help: go test -godog.help
// * show usage help if there were other test files in directory: go test -godog.help godogs_test.go
// * run scenarios with multiple formatters: go test -test.v -godog.format cucumber:cuc.json,pretty -test.run "^TestFeatures/"

import (
	"context"
	"flag"
	"fmt"
	"github.com/cucumber/godog/_examples/godogs"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var opts = godog.Options{
	Output:      colors.Colored(os.Stdout),
	Concurrency: 4,
}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opts)
}

func TestFeatures(t *testing.T) {
	o := opts
	o.TestingT = t

	status := godog.TestSuite{
		Name:                 "godogs",
		Options:              &o,
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  InitializeScenario,
	}.Run()

	if status == 2 {
		t.SkipNow()
	}

	if status != 0 {
		t.Fatalf("zero status code expected, %d received", status)
	}
}

type godogsCtxKey struct{}

func godogsToContext(ctx context.Context, g godogs.Godogs) context.Context {
	return context.WithValue(ctx, godogsCtxKey{}, &g)
}

func godogsFromContext(ctx context.Context) *godogs.Godogs {
	g, _ := ctx.Value(godogsCtxKey{}).(*godogs.Godogs)

	return g
}

// Concurrent execution of scenarios may lead to race conditions on shared resources.
// Use context to maintain data separation and avoid data races.

// Step definition can optionally receive context as a first argument.

func thereAreGodogs(ctx context.Context, available int) {
	godogsFromContext(ctx).Add(available)
}

// Step definition can return error, context, context and error, or nothing.

func iEat(ctx context.Context, num int) error {
	return godogsFromContext(ctx).Eat(num)
}

func thereShouldBeRemaining(ctx context.Context, remaining int) error {
	available := godogsFromContext(ctx).Available()
	if available != remaining {
		return fmt.Errorf("expected %d godogs to be remaining, but there is %d", remaining, available)
	}

	return nil
}

func thereShouldBeNoneRemaining(ctx context.Context) error {
	return thereShouldBeRemaining(ctx, 0)
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() { fmt.Println("Get the party started!") })
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// Add initial godogs to context.
		return godogsToContext(ctx, 0), nil
	})

	ctx.Step(`^there are (\d+) godogs$`, thereAreGodogs)
	ctx.Step(`^I eat (\d+)$`, iEat)
	ctx.Step(`^there should be (\d+) remaining$`, thereShouldBeRemaining)
	ctx.Step(`^there should be none remaining$`, thereShouldBeNoneRemaining)
}
