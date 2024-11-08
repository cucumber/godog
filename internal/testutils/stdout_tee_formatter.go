package testutils

import (
	"fmt"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/internal/storage"
	messages "github.com/cucumber/messages/go/v21"
	"regexp"
)

// dumps fmt calls to the console and forwards to other formatter
type StdoutTeeFormatter struct {
	Out godog.Formatter
}

func (f StdoutTeeFormatter) SetStorage(s *storage.Storage) {

	type storageFormatter interface {
		SetStorage(*storage.Storage)
	}

	if fmt, ok := f.Out.(storageFormatter); ok {
		fmt.SetStorage(s)
	}
}

func (f StdoutTeeFormatter) TestRunStarted() {
	f.Out.TestRunStarted()
}

func (f StdoutTeeFormatter) Passed(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	fmt.Printf("%9v: %q %q, step: %q, match: %q\n", "passed", scenario.Name, scenario.Uri, step.Text, f.match(match))
	f.Out.Passed(scenario, step, match)
}

func (f StdoutTeeFormatter) Skipped(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	fmt.Printf("%9v: %s %s, step: %s, match: %s\n", "skipped", scenario.Name, scenario.Uri, step.Text, f.match(match))
	f.Out.Skipped(scenario, step, match)
}

func (f StdoutTeeFormatter) Undefined(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	fmt.Printf("%9v: %q %q, step: %q, match: %q\n", "undefined", scenario.Name, scenario.Uri, step.Text, f.match(match))
	f.Out.Undefined(scenario, step, match)
}

func (f StdoutTeeFormatter) Failed(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition, err error) {
	fmt.Printf("%9v: %q %q, step: %q, match: %q, error: %q\n", "failed", scenario.Name, scenario.Uri, step.Text, f.match(match), f.error(err))
	f.Out.Failed(scenario, step, match, err)
}

func (f StdoutTeeFormatter) Pending(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	fmt.Printf("%9v: %q %q, step: %q, match: %q\n", "pending", scenario.Name, scenario.Uri, step.Text, f.match(match))
	f.Out.Pending(scenario, step, match)
}

func (f StdoutTeeFormatter) Ambiguous(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition, err error) {
	fmt.Printf("%9v: %q %q, step: %q, match: %q, error: %q\n", "ambiguous", scenario.Name, scenario.Uri, step.Text, f.match(match), f.error(err))
	f.Out.Ambiguous(scenario, step, match, err)
}

func (f StdoutTeeFormatter) Defined(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	//fmt.Printf("%9v: %q %q, step: %q, match: %q\n", "defined", scenario.Name, scenario.Uri, step.Text, f.match(match))
	f.Out.Defined(scenario, step, match)
}

func (f StdoutTeeFormatter) Feature(doc *messages.GherkinDocument, uri string, content []byte) {
	f.Out.Feature(doc, uri, content)
}

func (f StdoutTeeFormatter) Summary() {
	f.Out.Summary()
}

func (f StdoutTeeFormatter) Pickle(p *messages.Pickle) {
	f.Out.Pickle(p)
}

func (f StdoutTeeFormatter) Close() error {
	return f.Out.Close()
}

func (f StdoutTeeFormatter) error(err error) string {
	if err == nil {
		return "<nil?"
	}
	return err.Error()
}

func (f StdoutTeeFormatter) match(match *godog.StepDefinition) *regexp.Regexp {
	var expr *regexp.Regexp
	if match != nil {
		expr = match.Expr
	}
	return expr
}
