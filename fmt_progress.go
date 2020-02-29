package godog

import (
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/cucumber/godog/gherkin"
)

func init() {
	Format("progress", "Prints a character per step.", progressFunc)
}

func progressFunc(suite string, out io.Writer) Formatter {
	steps := 0
	return &progress{
		basefmt:     newBaseFmt(suite, out),
		stepsPerRow: 70,
		steps:       &steps,
	}
}

type progress struct {
	*basefmt
	stepsPerRow int
	steps       *int
}

func (f *progress) Summary() {
	left := math.Mod(float64(*f.steps), float64(f.stepsPerRow))
	if left != 0 {
		if *f.steps > f.stepsPerRow {
			fmt.Fprintf(f.out, s(f.stepsPerRow-int(left))+fmt.Sprintf(" %d\n", *f.steps))
		} else {
			fmt.Fprintf(f.out, " %d\n", *f.steps)
		}
	}
	fmt.Fprintln(f.out, "")

	if len(f.failed) > 0 {
		fmt.Fprintln(f.out, "\n--- "+red("Failed steps:")+"\n")
		for _, fail := range f.failed {
			fmt.Fprintln(f.out, s(2)+red(fail.scenarioDesc())+blackb(" # "+fail.scenarioLine()))
			fmt.Fprintln(f.out, s(4)+red(strings.TrimSpace(fail.step.Keyword)+" "+fail.step.Text)+blackb(" # "+fail.line()))
			fmt.Fprintln(f.out, s(6)+red("Error: ")+redb(fmt.Sprintf("%+v", fail.err))+"\n")
		}
	}
	f.basefmt.Summary()
}

func (f *progress) step(res *stepResult) {
	switch res.typ {
	case passed:
		fmt.Fprint(f.out, green("."))
	case skipped:
		fmt.Fprint(f.out, cyan("-"))
	case failed:
		fmt.Fprint(f.out, red("F"))
	case undefined:
		fmt.Fprint(f.out, yellow("U"))
	case pending:
		fmt.Fprint(f.out, yellow("P"))
	}

	*f.steps++

	if math.Mod(float64(*f.steps), float64(f.stepsPerRow)) == 0 {
		fmt.Fprintf(f.out, " %d\n", *f.steps)
	}
}

func (f *progress) Passed(step *gherkin.Step, match *StepDef) {
	f.basefmt.Passed(step, match)

	f.lock.Lock()
	defer f.lock.Unlock()
	f.step(f.passed[len(f.passed)-1])
}

func (f *progress) Skipped(step *gherkin.Step, match *StepDef) {
	f.basefmt.Skipped(step, match)

	f.lock.Lock()
	defer f.lock.Unlock()
	f.step(f.skipped[len(f.skipped)-1])
}

func (f *progress) Undefined(step *gherkin.Step, match *StepDef) {
	f.basefmt.Undefined(step, match)

	f.lock.Lock()
	defer f.lock.Unlock()
	f.step(f.undefined[len(f.undefined)-1])
}

func (f *progress) Failed(step *gherkin.Step, match *StepDef, err error) {
	f.basefmt.Failed(step, match, err)

	f.lock.Lock()
	defer f.lock.Unlock()
	f.step(f.failed[len(f.failed)-1])
}

func (f *progress) Pending(step *gherkin.Step, match *StepDef) {
	f.basefmt.Pending(step, match)

	f.lock.Lock()
	defer f.lock.Unlock()
	f.step(f.pending[len(f.pending)-1])
}

func (f *progress) Sync(cf ConcurrentFormatter) {
	if source, ok := cf.(*progress); ok {
		f.basefmt.Sync(source.basefmt)
		f.steps = source.steps
	}
}

func (f *progress) Copy(cf ConcurrentFormatter) {
	if source, ok := cf.(*progress); ok {
		f.basefmt.Copy(source.basefmt)
	}
}
