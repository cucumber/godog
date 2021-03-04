package storage

import (
	"reflect"
	"regexp"
	"time"

	"github.com/cucumber/messages-go/v10"
)

// Feature is an object to group together
// the parsed gherkin document and the
// raw content.
type Feature struct {
	*messages.GherkinDocument
	Path    string
	Content []byte
}

// TestRunStartedEvent ...
type TestRunStartedEvent struct {
	StartedAt time.Time
}

// PickleStartedEvent ...
type PickleStartedEvent struct {
	PickleID  string
	StartedAt time.Time
}

// PickleStepResultEvent ...
type PickleStepResultEvent struct {
	PickleStepID string
	Status       PickleStepResultStatus
	FinishedAt   time.Time
	Err          error
}

// PickleStepResultStatus ...
type PickleStepResultStatus int

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

	Args         []interface{}
	HandlerValue reflect.Value

	// multistep related
	Nested    bool
	Undefined []string
}
