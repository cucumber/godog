package formatters

import (
	"io"
	"regexp"

	"github.com/cucumber/messages-go/v16"
)

type registeredFormatter struct {
	name        string
	description string
	fmt         FormatterFunc
}

var registeredFormatters []*registeredFormatter

// FindFmt searches available formatters registered
// and returns FormaterFunc matched by given
// format name or nil otherwise
func FindFmt(name string) FormatterFunc {
	for _, el := range registeredFormatters {
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
	registeredFormatters = append(registeredFormatters, &registeredFormatter{
		name:        name,
		fmt:         f,
		description: description,
	})
}

// AvailableFormatters gives a map of all
// formatters registered with their name as key
// and description as value
func AvailableFormatters() map[string]string {
	fmts := make(map[string]string, len(registeredFormatters))

	for _, f := range registeredFormatters {
		fmts[f.name] = f.description
	}

	return fmts
}

type MetaData struct{}

type Source struct {
	Data      string
	MediaType string
	Uri       string
}

// Formatter is an interface for feature runner
// output summary presentation.
//
// New formatters may be created to represent
// suite results in different ways. These new
// formatters needs to be registered with a
// godog.Format function call
type Formatter interface {

	// parsing phase messages
	MetaData(*MetaData)
	Source(*Source)
	GherkinDocument(*messages.GherkinDocument, string, []byte)
	Pickle(p *messages.Pickle)
	// TestRun()
	// hooks - execution?

	// step matching phase messages
	Defined(*StepDefinition)

	// execution phase messages
	TestRunStarted(*messages.TestRunStarted)
	Feature(*messages.GherkinDocument, string, []byte) // this one is not part of Messages, legacy
	TestCase(*messages.Pickle)
	TestCaseStarted(*messages.TestCaseStarted)
	TestStepStarted(*messages.Pickle, *messages.PickleStep, *StepDefinition)
	Failed(*messages.Pickle, *messages.PickleStep, *StepDefinition, error)
	Passed(*messages.Pickle, *messages.PickleStep, *StepDefinition)
	Skipped(*messages.Pickle, *messages.PickleStep, *StepDefinition)
	Undefined(*messages.Pickle, *messages.PickleStep, *StepDefinition)
	Pending(*messages.Pickle, *messages.PickleStep, *StepDefinition)
	Summary()
}

// FormatterFunc builds a formatter with given
// suite name and io.Writer to record output
type FormatterFunc func(string, io.Writer) Formatter

// StepDefinition is a registered step definition
// contains a StepHandler and regexp which
// is used to match a step. Args which
// were matched by last executed step
//
// This structure is passed to the formatter
// when step is matched and is either failed
// or successful
type StepDefinition struct {
	Expr    *regexp.Regexp
	Handler interface{}
}
