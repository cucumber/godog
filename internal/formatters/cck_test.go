package formatters_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
)

var cckFeaturesPath = ""

// Directory and file names that begin with "." or "_" are ignored
// by the go tool, as are directories named "testdata".

type BdBd struct {
	buffer *bytes.Buffer
}

func (b *BdBd) Reset() {
	b.buffer = &bytes.Buffer{}
}

func (b *BdBd) Write(p []byte) (int, error) {
	n, err := b.buffer.Write(p)
	return n, err
}

func (b *BdBd) String() string {
	return b.buffer.String()
}

func NewBuffer() *BdBd {
	return &BdBd{
		&bytes.Buffer{},
	}
}

func TestCompatibility(t *testing.T) {

	testCases := []struct {
		path string
	}{
		{"minimal"}, // missing  testCase, source, and meta
	}

	for _, tc := range testCases {

		// var out io.Writer = os.Stdout
		// out := os.Stdout // TODO: Just for now
		buf := NewBuffer()
		// out = buf
		path := fmt.Sprintf("../../_compatibility/%s", tc.path)

		opts := &godog.Options{
			Paths:    []string{path},
			TestingT: t,
			Format:   "message",
			Output:   buf,
		}

		suite := godog.TestSuite{
			ScenarioInitializer:  InitializeScenario,
			TestSuiteInitializer: IntializeTestSuite,
			Options:              opts,
		}

		// fmt.Printf("\nCaptured:\n%s\n", buf.String())

		if suite.Run() != 0 {
			t.Fatalf("running %s returned non-zero status", tc.path)
		}
		fmt.Printf("\nCaptured:\n%s\n------\n", buf.String())
	}
}

func InitializeScenario(sc *godog.ScenarioContext) {

	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// clean the state before every scenario
		if ctx == nil {
			ctx = context.Background()
		}
		return ctx, nil
	})
	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		fmt.Println(sc.Name)
		return ctx, nil
	})
	// Add step definitions here.
	sc.Step(`^I have (\d+) cukes in my belly$`, iHaveCukesInMyBelly)
}

func IntializeTestSuite(sc *godog.TestSuiteContext) {
	sc.BeforeSuite(func() {
		// do any set up here one time before the entire test suite runs, e.g., create a database connection pool
		// that would be too expensive to do before each scenario.
	})
	sc.AfterSuite(func() {
		// do any clean up after the entire test suite is done executiong.
	})
	sc.ScenarioContext().StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
		return ctx, nil
	})
	sc.ScenarioContext().StepContext().After(func(ctx context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
		return ctx, nil
	})
}

func iHaveCukesInMyBelly(arg1 int) error {
	return godog.ErrPending
}
