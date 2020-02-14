package godog

import (
	"fmt"
	"io"
	"math"
	"strings"
	"sync"

	"github.com/cucumber/messages-go/v9"
)

func init() {
	Format("progress", "Prints a character per step.", progressFunc)
}

func progressFunc(suite string, out io.Writer) Formatter {
	steps := 0
	return &progress{
		basefmt: basefmt{
			started: timeNowFunc(),
			indent:  2,
			out:     out,
		},
		stepsPerRow: 70,
		lock:        new(sync.Mutex),
		steps:       &steps,
	}
}

type progress struct {
	basefmt
	lock        *sync.Mutex
	stepsPerRow int
	steps       *int
}

func (f *progress) Pickle(pickle *messages.Pickle) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.basefmt.Pickle(pickle)
}

func (f *progress) Feature(gd *messages.GherkinDocument, p string, c []byte) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.basefmt.Feature(gd, p, c)
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

	var failedStepsOutput []string
	for _, sr := range f.findStepResults(failed) {
		if sr.status == failed {
			sc := f.findScenario(sr.owner.AstNodeIds[0])
			scenarioDesc := fmt.Sprintf("%s: %s", sc.Keyword, sr.owner.Name)
			scenarioLine := fmt.Sprintf("%s:%d", sr.owner.Uri, sc.Location.Line)

			step := f.findStep(sr.step.AstNodeIds[0])
			stepDesc := strings.TrimSpace(step.Keyword) + " " + sr.step.Text
			stepLine := fmt.Sprintf("%s:%d", sr.owner.Uri, step.Location.Line)

			failedStepsOutput = append(
				failedStepsOutput,
				s(2)+red(scenarioDesc)+blackb(" # "+scenarioLine),
				s(4)+red(stepDesc)+blackb(" # "+stepLine),
				s(6)+red("Error: ")+redb(fmt.Sprintf("%+v", sr.err)),
				"",
			)
		}
	}

	if len(failedStepsOutput) > 0 {
		fmt.Fprintln(f.out, "\n\n--- "+red("Failed steps:")+"\n")
		fmt.Fprint(f.out, strings.Join(failedStepsOutput, "\n"))
	}
	fmt.Fprintln(f.out, "")

	f.basefmt.Summary()
}

func (f *progress) step(res *stepResult) {
	switch res.status {
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

func (f *progress) Passed(pickle *messages.Pickle, step *messages.Pickle_PickleStep, match *StepDefinition) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.basefmt.Passed(pickle, step, match)
	f.step(f.lastStepResult())
}

func (f *progress) Skipped(pickle *messages.Pickle, step *messages.Pickle_PickleStep, match *StepDefinition) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.basefmt.Skipped(pickle, step, match)
	f.step(f.lastStepResult())
}

func (f *progress) Undefined(pickle *messages.Pickle, step *messages.Pickle_PickleStep, match *StepDefinition) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.basefmt.Undefined(pickle, step, match)
	f.step(f.lastStepResult())
}

func (f *progress) Failed(pickle *messages.Pickle, step *messages.Pickle_PickleStep, match *StepDefinition, err error) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.basefmt.Failed(pickle, step, match, err)
	f.step(f.lastStepResult())
}

func (f *progress) Pending(pickle *messages.Pickle, step *messages.Pickle_PickleStep, match *StepDefinition) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.basefmt.Pending(pickle, step, match)
	f.step(f.lastStepResult())
}

func (f *progress) Sync(cf ConcurrentFormatter) {
	if source, ok := cf.(*progress); ok {
		f.lock = source.lock
		f.steps = source.steps
	}
}

func (f *progress) Copy(cf ConcurrentFormatter) {
	if source, ok := cf.(*progress); ok {
		for _, v := range source.features {
			f.features = append(f.features, v)
		}
	}
}
