package main

import (
	"fmt"
	"io"
	"math"

	"github.com/cucumber/godog"
)

const (
	PASSED_EMOJI    = "✅"
	SKIPPED_EMOJI   = "➖"
	FAILED_EMOJI    = "❌"
	UNDEFINED_EMOJI = "❓"
	PENDING_EMOJI   = "🚧"
)

func init() {
	godog.Format("emoji", "Progress formatter with emojis", emojiFormatterFunc)
}

func emojiFormatterFunc(suite string, out io.Writer) godog.Formatter {
	return newEmojiFmt(suite, out)
}

func newEmojiFmt(suite string, out io.Writer) *emojiFmt {
	return &emojiFmt{
		Progress: godog.NewProgressfmt(suite, out),
		out:      out,
	}
}

type emojiFmt struct {
	*godog.Progress

	out io.Writer
}

func (f *emojiFmt) TestRunStarted() {}

func (f *emojiFmt) Passed(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.Progress.Basefmt.Passed(scenario, step, match)

	f.Progress.Basefmt.Lock.Lock()
	defer f.Progress.Basefmt.Lock.Unlock()

	f.step(step.Id)
}

func (f *emojiFmt) Skipped(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.Progress.Basefmt.Skipped(scenario, step, match)

	f.Progress.Basefmt.Lock.Lock()
	defer f.Progress.Basefmt.Lock.Unlock()

	f.step(step.Id)
}

func (f *emojiFmt) Undefined(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.Progress.Basefmt.Undefined(scenario, step, match)

	f.Progress.Basefmt.Lock.Lock()
	defer f.Progress.Basefmt.Lock.Unlock()

	f.step(step.Id)
}

func (f *emojiFmt) Failed(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition, err error) {
	f.Progress.Basefmt.Failed(scenario, step, match, err)

	f.Progress.Basefmt.Lock.Lock()
	defer f.Progress.Basefmt.Lock.Unlock()

	f.step(step.Id)
}

func (f *emojiFmt) Pending(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.Progress.Basefmt.Pending(scenario, step, match)

	f.Progress.Basefmt.Lock.Lock()
	defer f.Progress.Basefmt.Lock.Unlock()

	f.step(step.Id)
}

func (f *emojiFmt) Summary() {
	f.printSummaryLegend()
	f.Progress.Summary()
}

func (f *emojiFmt) printSummaryLegend() {
	fmt.Fprint(f.out, "\n\nOutput Legend:\n")
	fmt.Fprint(f.out, fmt.Sprintf("\t%s Passed\n", PASSED_EMOJI))
	fmt.Fprint(f.out, fmt.Sprintf("\t%s Failed\n", FAILED_EMOJI))
	fmt.Fprint(f.out, fmt.Sprintf("\t%s Skipped\n", SKIPPED_EMOJI))
	fmt.Fprint(f.out, fmt.Sprintf("\t%s Undefined\n", UNDEFINED_EMOJI))
	fmt.Fprint(f.out, fmt.Sprintf("\t%s Pending\n", PENDING_EMOJI))
}

func (f *emojiFmt) step(pickleStepID string) {
	pickleStepResult := f.Storage.MustGetPickleStepResult(pickleStepID)

	switch pickleStepResult.Status {
	case godog.Passed:
		fmt.Fprint(f.out, PASSED_EMOJI)
	case godog.Skipped:
		fmt.Fprint(f.out, SKIPPED_EMOJI)
	case godog.Failed:
		fmt.Fprint(f.out, FAILED_EMOJI)
	case godog.Undefined:
		fmt.Fprint(f.out, UNDEFINED_EMOJI)
	case godog.Pending:
		fmt.Fprint(f.out, PENDING_EMOJI)
	}

	*f.Steps++

	if math.Mod(float64(*f.Steps), float64(f.StepsPerRow)) == 0 {
		fmt.Fprintf(f.out, " %d\n", *f.Steps)
	}
}
