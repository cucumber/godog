package customformatter

import (
	"io"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v16"
)

func init() {
	godog.Format("custom", "Custom formatter", customFormatterFunc)
}

func customFormatterFunc(suite string, out io.Writer) godog.Formatter {
	return &customFmt{suiteName: suite, out: out}
}

type customFmt struct {
	suiteName string
	out       io.Writer
}

func (f *customFmt) TestRunStarted()                                                   {}
func (f *customFmt) Feature(*messages.GherkinDocument, string, []byte)                 {}
func (f *customFmt) Pickle(*godog.Scenario)                                            {}
func (f *customFmt) Defined(*godog.Scenario, *godog.Step, *godog.StepDefinition)       {}
func (f *customFmt) Passed(*godog.Scenario, *godog.Step, *godog.StepDefinition)        {}
func (f *customFmt) Skipped(*godog.Scenario, *godog.Step, *godog.StepDefinition)       {}
func (f *customFmt) Undefined(*godog.Scenario, *godog.Step, *godog.StepDefinition)     {}
func (f *customFmt) Failed(*godog.Scenario, *godog.Step, *godog.StepDefinition, error) {}
func (f *customFmt) Pending(*godog.Scenario, *godog.Step, *godog.StepDefinition)       {}
func (f *customFmt) Summary()                                                          {}
