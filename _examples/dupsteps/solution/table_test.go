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
//  (e.g., `godog.TestSuite` instances), using standard `go` table-driven tests.  See associated README file.
//

// TestSeparateScenarios runs the case(s) defined in the table, then combines the outputs into a single report
func TestSeparateScenarios(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
		steps func(ctx *godog.ScenarioContext)
	}{
		{
			name:  "flat tire",
			paths: []string{"../features/flatTire.feature"},
			steps: func(ctx *godog.ScenarioContext) {
				ctx.Step(`^I ran over a nail and got a flat tire$`, func() {})
				ctx.Step(`^I fixed it$`, func() {})
				ctx.Step(`^I can continue on my way$`, func() {})
			},
		},
		{
			name:  "clogged sink",
			paths: []string{"../features/cloggedDrain.feature"},
			steps: func(ctx *godog.ScenarioContext) {
				ctx.Step(`^I accidentally poured concrete down my drain and clogged the sewer line$`, func() {})
				ctx.Step(`^I fixed it$`, func() {})
				ctx.Step(`^I can once again use my sink$`, func() {})
			},
		},
	}

	outputCollector := MultiWriter{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := defaultOpts

			// test runs only selected features/scenarios
			opts.Paths = test.paths
			opts.Format = "cucumber"
			opts.Output = outputCollector.NewWriter()

			gts := godog.TestSuite{
				Name:                t.Name(),
				ScenarioInitializer: test.steps,
				Options:             &opts,
			}

			assert.Zero(t, gts.Run())
		})
	}

	// then invokes a "combiner" appropriate for the output(s) produced by the test case(s)
	// NOTE: the "combiner" is formatter-specific; this one "knows" to combine "cucumber" reports
	combinedReport, err := CombineCukeReports(outputCollector.GetOutputs())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "combiner error: %s\n", err)
		return
	}

	// route the combined output to where it should go...
	// hmm, it'd be nice to have some CLI options to control this destination...
	fmt.Println(string(combinedReport))
}
