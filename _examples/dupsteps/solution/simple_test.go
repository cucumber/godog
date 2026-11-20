package solution

import (
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/assert"
)

//
//  Demonstration of a unit testing approach for producing a single aggregated report from runs of separate test suites
//  (e.g., `godog.TestSuite` instances), using standard `go` simple test cases.  See associated README file.
//

// the single global var needed to "collect" the output(s) produced by the test(s)
var mw = MultiWriter{}

// TestMain runs the test case(s), then combines the outputs into a single report
func TestMain(m *testing.M) {
	rc := m.Run() // runs the test case(s)

	// then invokes a "combiner" appropriate for the output(s) produced by the test case(s)
	// NOTE: the "combiner" is formatter-specific; this one "knows" to combine "cucumber" reports
	combinedReport, err := CombineCukeReports(mw.GetOutputs())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "combiner error: %s\n", err)
	} else {
		// hmm, it'd be nice to have some CLI options to control this destination...
		fmt.Println(string(combinedReport))
	}

	os.Exit(rc)
}

// one or more test case(s), providing the desired level of step encapsulation

func TestFlatTire(t *testing.T) {
	opts := defaultOpts

	// test runs only selected features/scenarios
	opts.Paths = []string{"../features/flatTire.feature"}
	opts.Output = mw.NewWriter()

	gts := godog.TestSuite{
		Name: t.Name(),
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			ctx.Step(`^I ran over a nail and got a flat tire$`, func() {})
			ctx.Step(`^I fixed it$`, func() {})
			ctx.Step(`^I can continue on my way$`, func() {})
		},
		Options: &opts,
	}

	assert.Zero(t, gts.Run())
}

func TestCloggedDrain(t *testing.T) {
	opts := defaultOpts

	// test runs only selected features/scenarios
	opts.Paths = []string{"../features/cloggedDrain.feature"}
	opts.Output = mw.NewWriter()

	gts := godog.TestSuite{
		Name: t.Name(),
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			ctx.Step(`^I accidentally poured concrete down my drain and clogged the sewer line$`, func() {})
			ctx.Step(`^I fixed it$`, func() {})
			ctx.Step(`^I can once again use my sink$`, func() {})

		},
		Options: &opts,
	}

	assert.Zero(t, gts.Run())
}
