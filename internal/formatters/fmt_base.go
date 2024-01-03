package formatters

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	messages "github.com/cucumber/messages/go/v21"

	"github.com/cucumber/godog/colors"
	"github.com/cucumber/godog/formatters"
	"github.com/cucumber/godog/internal/models"
	"github.com/cucumber/godog/internal/snippets"
	"github.com/cucumber/godog/internal/storage"
	"github.com/cucumber/godog/internal/utils"
)

// BaseFormatterFunc implements the FormatterFunc for the base formatter.
func BaseFormatterFunc(suite string, out io.Writer, snippetFunc string) formatters.Formatter {
	return NewBase(suite, out, snippetFunc)
}

// NewBase creates a new base formatter.
func NewBase(suite string, out io.Writer, snippetFunc string) *Base {
	return &Base{
		snippetFunc: snippets.Find(snippetFunc),
		suiteName:   suite,
		indent:      2,
		out:         out,
		Lock:        new(sync.Mutex),
	}
}

// Base is a base formatter.
type Base struct {
	suiteName   string
	out         io.Writer
	indent      int
	snippetFunc snippets.Func

	Storage *storage.Storage
	Lock    *sync.Mutex
}

// SetStorage assigns gherkin data storage.
func (f *Base) SetStorage(st *storage.Storage) {
	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.Storage = st
}

// TestRunStarted is triggered on test start.
func (f *Base) TestRunStarted() {}

// Feature receives gherkin document.
func (f *Base) Feature(*messages.GherkinDocument, string, []byte) {}

// Pickle receives scenario.
func (f *Base) Pickle(*messages.Pickle) {}

// Defined receives step definition.
func (f *Base) Defined(*messages.Pickle, *messages.PickleStep, *formatters.StepDefinition) {
}

// Passed captures passed step.
func (f *Base) Passed(*messages.Pickle, *messages.PickleStep, *formatters.StepDefinition) {}

// Skipped captures skipped step.
func (f *Base) Skipped(*messages.Pickle, *messages.PickleStep, *formatters.StepDefinition) {
}

// Undefined captures undefined step.
func (f *Base) Undefined(*messages.Pickle, *messages.PickleStep, *formatters.StepDefinition) {
}

// Failed captures failed step.
func (f *Base) Failed(*messages.Pickle, *messages.PickleStep, *formatters.StepDefinition, error) {
}

// Pending captures pending step.
func (f *Base) Pending(*messages.Pickle, *messages.PickleStep, *formatters.StepDefinition) {
}

// Summary renders summary information.
func (f *Base) Summary() {
	var totalSc, passedSc, undefinedSc int
	var totalSt, passedSt, failedSt, skippedSt, pendingSt, undefinedSt int

	pickleResults := f.Storage.MustGetPickleResults()
	for _, pr := range pickleResults {
		var prStatus models.StepResultStatus
		totalSc++

		pickleStepResults := f.Storage.MustGetPickleStepResultsByPickleID(pr.PickleID)

		if len(pickleStepResults) == 0 {
			prStatus = undefined
		}

		for _, sr := range pickleStepResults {
			totalSt++

			switch sr.Status {
			case passed:
				passedSt++
			case failed:
				prStatus = failed
				failedSt++
			case skipped:
				skippedSt++
			case undefined:
				prStatus = undefined
				undefinedSt++
			case pending:
				prStatus = pending
				pendingSt++
			}
		}

		if prStatus == passed {
			passedSc++
		} else if prStatus == undefined {
			undefinedSc++
		}
	}

	var steps, parts, scenarios []string
	if passedSt > 0 {
		steps = append(steps, green(fmt.Sprintf("%d passed", passedSt)))
	}
	if failedSt > 0 {
		parts = append(parts, red(fmt.Sprintf("%d failed", failedSt)))
		steps = append(steps, red(fmt.Sprintf("%d failed", failedSt)))
	}
	if pendingSt > 0 {
		parts = append(parts, yellow(fmt.Sprintf("%d pending", pendingSt)))
		steps = append(steps, yellow(fmt.Sprintf("%d pending", pendingSt)))
	}
	if undefinedSt > 0 {
		parts = append(parts, yellow(fmt.Sprintf("%d undefined", undefinedSc)))
		steps = append(steps, yellow(fmt.Sprintf("%d undefined", undefinedSt)))
	} else if undefinedSc > 0 {
		// there may be some scenarios without steps
		parts = append(parts, yellow(fmt.Sprintf("%d undefined", undefinedSc)))
	}
	if skippedSt > 0 {
		steps = append(steps, cyan(fmt.Sprintf("%d skipped", skippedSt)))
	}
	if passedSc > 0 {
		scenarios = append(scenarios, green(fmt.Sprintf("%d passed", passedSc)))
	}
	scenarios = append(scenarios, parts...)

	testRunStartedAt := f.Storage.MustGetTestRunStarted().StartedAt
	elapsed := utils.TimeNowFunc().Sub(testRunStartedAt)

	fmt.Fprintln(f.out, "")

	if totalSc == 0 {
		fmt.Fprintln(f.out, "No scenarios")
	} else {
		fmt.Fprintf(f.out, "%d scenarios (%s)\n", totalSc, strings.Join(scenarios, ", "))
	}

	if totalSt == 0 {
		fmt.Fprintln(f.out, "No steps")
	} else {
		fmt.Fprintf(f.out, "%d steps (%s)\n", totalSt, strings.Join(steps, ", "))
	}

	elapsedString := elapsed.String()
	if elapsed.Nanoseconds() == 0 {
		// go 1.5 and 1.6 prints 0 instead of 0s, if duration is zero.
		elapsedString = "0s"
	}
	fmt.Fprintln(f.out, elapsedString)

	// prints used randomization seed
	seed, err := strconv.ParseInt(os.Getenv("GODOG_SEED"), 10, 64)
	if err == nil && seed != 0 {
		fmt.Fprintln(f.out, "")
		fmt.Fprintln(f.out, "Randomized with seed:", colors.Yellow(seed))
	}
	if text := f.Snippets(); text != "" {
		fmt.Fprintln(f.out, "")
		fmt.Fprintln(f.out, yellow("You can implement step definitions for undefined steps with these snippets:"))
		fmt.Fprintln(f.out, yellow(text))
	}
}

// Snippets returns code suggestions for undefined steps.
func (f *Base) Snippets() string {
	return f.snippetFunc(f.Storage)
}
