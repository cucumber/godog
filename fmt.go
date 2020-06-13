package godog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/cucumber/messages-go/v10"

	"github.com/cucumber/godog/colors"
)

type registeredFormatter struct {
	name        string
	description string
	fmt         FormatterFunc
}

var formatters []*registeredFormatter

// FindFmt searches available formatters registered
// and returns FormaterFunc matched by given
// format name or nil otherwise
func FindFmt(name string) FormatterFunc {
	for _, el := range formatters {
		if el.name == name {
			return el.fmt
		}
	}

	return nil
}

// Format registers a feature suite output
// formatter by given name, description and
// FormatterFunc constructor function, to initialize
// formatter with the output recorder.
func Format(name, description string, f FormatterFunc) {
	formatters = append(formatters, &registeredFormatter{
		name:        name,
		fmt:         f,
		description: description,
	})
}

// AvailableFormatters gives a map of all
// formatters registered with their name as key
// and description as value
func AvailableFormatters() map[string]string {
	fmts := make(map[string]string, len(formatters))

	for _, f := range formatters {
		fmts[f.name] = f.description
	}

	return fmts
}

// Formatter is an interface for feature runner
// output summary presentation.
//
// New formatters may be created to represent
// suite results in different ways. These new
// formatters needs to be registered with a
// godog.Format function call
type Formatter interface {
	TestRunStarted()
	Feature(*messages.GherkinDocument, string, []byte)
	Pickle(*messages.Pickle)
	Defined(*messages.Pickle, *messages.Pickle_PickleStep, *StepDefinition)
	Failed(*messages.Pickle, *messages.Pickle_PickleStep, *StepDefinition, error)
	Passed(*messages.Pickle, *messages.Pickle_PickleStep, *StepDefinition)
	Skipped(*messages.Pickle, *messages.Pickle_PickleStep, *StepDefinition)
	Undefined(*messages.Pickle, *messages.Pickle_PickleStep, *StepDefinition)
	Pending(*messages.Pickle, *messages.Pickle_PickleStep, *StepDefinition)
	Summary()
}

// ConcurrentFormatter is an interface for a Concurrent
// version of the Formatter interface.
type ConcurrentFormatter interface {
	Formatter
	Copy(ConcurrentFormatter)
	Sync(ConcurrentFormatter)
}

type storageFormatter interface {
	setStorage(*storage)
}

// FormatterFunc builds a formatter with given
// suite name and io.Writer to record output
type FormatterFunc func(string, io.Writer) Formatter

type stepResultStatus int

const (
	passed stepResultStatus = iota
	failed
	skipped
	undefined
	pending
)

func (st stepResultStatus) clr() colors.ColorFunc {
	switch st {
	case passed:
		return green
	case failed:
		return red
	case skipped:
		return cyan
	default:
		return yellow
	}
}

func (st stepResultStatus) String() string {
	switch st {
	case passed:
		return "passed"
	case failed:
		return "failed"
	case skipped:
		return "skipped"
	case undefined:
		return "undefined"
	case pending:
		return "pending"
	default:
		return "unknown"
	}
}

type pickleStepResult struct {
	Status     stepResultStatus
	finishedAt time.Time
	err        error

	PickleID     string
	PickleStepID string

	def *StepDefinition
}

func newStepResult(pickleID, pickleStepID string, match *StepDefinition) pickleStepResult {
	return pickleStepResult{finishedAt: timeNowFunc(), PickleID: pickleID, PickleStepID: pickleStepID, def: match}
}

func newBaseFmt(suite string, out io.Writer) *basefmt {
	return &basefmt{
		suiteName: suite,
		startedAt: timeNowFunc(),
		indent:    2,
		out:       out,
		lock:      new(sync.Mutex),
	}
}

type basefmt struct {
	suiteName string

	out    io.Writer
	owner  interface{}
	indent int

	storage *storage

	startedAt time.Time

	firstFeature *bool
	lock         *sync.Mutex
}

func (f *basefmt) setStorage(st *storage) {
	f.storage = st
}

func (f *basefmt) TestRunStarted() {
	f.lock.Lock()
	defer f.lock.Unlock()

	firstFeature := true
	f.firstFeature = &firstFeature
}

func (f *basefmt) Pickle(p *messages.Pickle) {}

func (f *basefmt) Defined(*messages.Pickle, *messages.Pickle_PickleStep, *StepDefinition) {}

func (f *basefmt) Feature(ft *messages.GherkinDocument, p string, c []byte) {
	f.lock.Lock()
	defer f.lock.Unlock()

	*f.firstFeature = false
}

func (f *basefmt) Passed(pickle *messages.Pickle, step *messages.Pickle_PickleStep, match *StepDefinition) {
}
func (f *basefmt) Skipped(pickle *messages.Pickle, step *messages.Pickle_PickleStep, match *StepDefinition) {
}
func (f *basefmt) Undefined(pickle *messages.Pickle, step *messages.Pickle_PickleStep, match *StepDefinition) {
}
func (f *basefmt) Failed(pickle *messages.Pickle, step *messages.Pickle_PickleStep, match *StepDefinition, err error) {
}
func (f *basefmt) Pending(pickle *messages.Pickle, step *messages.Pickle_PickleStep, match *StepDefinition) {
}

func (f *basefmt) Summary() {
	var totalSc, passedSc, undefinedSc int
	var totalSt, passedSt, failedSt, skippedSt, pendingSt, undefinedSt int

	pickleResults := f.storage.mustGetPickleResults()
	for _, pr := range pickleResults {
		var prStatus stepResultStatus
		totalSc++

		pickleStepResults := f.storage.mustGetPickleStepResultsByPickleID(pr.PickleID)

		if len(pickleStepResults) == 0 {
			prStatus = undefined
		}

		for _, sr := range pickleStepResults {
			totalSt++

			switch sr.Status {
			case passed:
				prStatus = passed
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
	elapsed := timeNowFunc().Sub(f.startedAt)

	fmt.Fprintln(f.out, "")

	if totalSc == 0 {
		fmt.Fprintln(f.out, "No scenarios")
	} else {
		fmt.Fprintln(f.out, fmt.Sprintf("%d scenarios (%s)", totalSc, strings.Join(scenarios, ", ")))
	}

	if totalSt == 0 {
		fmt.Fprintln(f.out, "No steps")
	} else {
		fmt.Fprintln(f.out, fmt.Sprintf("%d steps (%s)", totalSt, strings.Join(steps, ", ")))
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

	if text := f.snippets(); text != "" {
		fmt.Fprintln(f.out, "")
		fmt.Fprintln(f.out, yellow("You can implement step definitions for undefined steps with these snippets:"))
		fmt.Fprintln(f.out, yellow(text))
	}
}

func (f *basefmt) Sync(cf ConcurrentFormatter) {
	if source, ok := cf.(*basefmt); ok {
		f.lock = source.lock
		f.firstFeature = source.firstFeature
	}
}

func (f *basefmt) Copy(cf ConcurrentFormatter) {}

func (f *basefmt) snippets() string {
	undefinedStepResults := f.storage.mustGetPickleStepResultsByStatus(undefined)
	if len(undefinedStepResults) == 0 {
		return ""
	}

	var index int
	var snips []undefinedSnippet
	// build snippets
	for _, u := range undefinedStepResults {
		pickleStep := f.storage.mustGetPickleStep(u.PickleStepID)

		steps := []string{pickleStep.Text}
		arg := pickleStep.Argument
		if u.def != nil {
			steps = u.def.undefined
			arg = nil
		}
		for _, step := range steps {
			expr := snippetExprCleanup.ReplaceAllString(step, "\\$1")
			expr = snippetNumbers.ReplaceAllString(expr, "(\\d+)")
			expr = snippetExprQuoted.ReplaceAllString(expr, "$1\"([^\"]*)\"$2")
			expr = "^" + strings.TrimSpace(expr) + "$"

			name := snippetNumbers.ReplaceAllString(step, " ")
			name = snippetExprQuoted.ReplaceAllString(name, " ")
			name = strings.TrimSpace(snippetMethodName.ReplaceAllString(name, ""))
			var words []string
			for i, w := range strings.Split(name, " ") {
				switch {
				case i != 0:
					w = strings.Title(w)
				case len(w) > 0:
					w = string(unicode.ToLower(rune(w[0]))) + w[1:]
				}
				words = append(words, w)
			}
			name = strings.Join(words, "")
			if len(name) == 0 {
				index++
				name = fmt.Sprintf("StepDefinitioninition%d", index)
			}

			var found bool
			for _, snip := range snips {
				if snip.Expr == expr {
					found = true
					break
				}
			}
			if !found {
				snips = append(snips, undefinedSnippet{Method: name, Expr: expr, argument: arg})
			}
		}
	}

	sort.Sort(snippetSortByMethod(snips))

	var buf bytes.Buffer
	if err := undefinedSnippetsTpl.Execute(&buf, snips); err != nil {
		panic(err)
	}
	// there may be trailing spaces
	return strings.Replace(buf.String(), " \n", "\n", -1)
}

func isLastStep(pickle *messages.Pickle, step *messages.Pickle_PickleStep) bool {
	return pickle.Steps[len(pickle.Steps)-1].Id == step.Id
}
