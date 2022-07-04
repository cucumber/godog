package main

import (
	"fmt"
	"io"
	"math"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v16"
)

const (
	passedEmoji    = "✅"
	skippedEmoji   = "➖"
	failedEmoji    = "❌"
	undefinedEmoji = "❓"
	pendingEmoji   = "🚧"
)

func init() {
	godog.Format("emoji", "Progress formatter with emojis", emojiFormatterFunc)
}

func emojiFormatterFunc(suite string, out io.Writer) godog.Formatter {
	return newEmojiFmt(suite, out)
}

func newEmojiFmt(suite string, out io.Writer) *emojiFmt {
	return &emojiFmt{
		ProgressFmt: godog.NewProgressFmt(suite, out),
		out:         out,
	}
}

type emojiFmt struct {
	*godog.ProgressFmt

	out io.Writer
}

func (f *emojiFmt) TestRunStarted(t *messages.TestRunStarted) {}

func (f *emojiFmt) Passed(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.ProgressFmt.Base.Passed(scenario, step, match)

	f.ProgressFmt.Base.Lock.Lock()
	defer f.ProgressFmt.Base.Lock.Unlock()

	f.step(step.Id)
}

func (f *emojiFmt) Skipped(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.ProgressFmt.Base.Skipped(scenario, step, match)

	f.ProgressFmt.Base.Lock.Lock()
	defer f.ProgressFmt.Base.Lock.Unlock()

	f.step(step.Id)
}

func (f *emojiFmt) Undefined(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.ProgressFmt.Base.Undefined(scenario, step, match)

	f.ProgressFmt.Base.Lock.Lock()
	defer f.ProgressFmt.Base.Lock.Unlock()

	f.step(step.Id)
}

func (f *emojiFmt) Failed(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition, err error) {
	f.ProgressFmt.Base.Failed(scenario, step, match, err)

	f.ProgressFmt.Base.Lock.Lock()
	defer f.ProgressFmt.Base.Lock.Unlock()

	f.step(step.Id)
}

func (f *emojiFmt) Pending(scenario *godog.Scenario, step *godog.Step, match *godog.StepDefinition) {
	f.ProgressFmt.Base.Pending(scenario, step, match)

	f.ProgressFmt.Base.Lock.Lock()
	defer f.ProgressFmt.Base.Lock.Unlock()

	f.step(step.Id)
}

func (f *emojiFmt) Summary() {
	f.printSummaryLegend()
	f.ProgressFmt.Summary()
}

func (f *emojiFmt) printSummaryLegend() {
	fmt.Fprint(f.out, "\n\nOutput Legend:\n")
	fmt.Fprint(f.out, fmt.Sprintf("\t%s Passed\n", passedEmoji))
	fmt.Fprint(f.out, fmt.Sprintf("\t%s Failed\n", failedEmoji))
	fmt.Fprint(f.out, fmt.Sprintf("\t%s Skipped\n", skippedEmoji))
	fmt.Fprint(f.out, fmt.Sprintf("\t%s Undefined\n", undefinedEmoji))
	fmt.Fprint(f.out, fmt.Sprintf("\t%s Pending\n", pendingEmoji))
}

func (f *emojiFmt) step(pickleStepID string) {
	pickleStepResult := f.Storage.MustGetPickleStepResult(pickleStepID)

	switch pickleStepResult.Status {
	case godog.StepPassed:
		fmt.Fprint(f.out, fmt.Sprintf(" %s", passedEmoji))
	case godog.StepSkipped:
		fmt.Fprint(f.out, fmt.Sprintf(" %s", skippedEmoji))
	case godog.StepFailed:
		fmt.Fprint(f.out, fmt.Sprintf(" %s", failedEmoji))
	case godog.StepUndefined:
		fmt.Fprint(f.out, fmt.Sprintf(" %s", undefinedEmoji))
	case godog.StepPending:
		fmt.Fprint(f.out, fmt.Sprintf(" %s", pendingEmoji))
	}

	*f.Steps++

	if math.Mod(float64(*f.Steps), float64(f.StepsPerRow)) == 0 {
		fmt.Fprintf(f.out, " %d\n", *f.Steps)
	}
}
