package godog

import (
	"fmt"

	"github.com/cucumber/gherkin-go"
)

type registeredFormatter struct {
	name        string
	fmt         Formatter
	description string
}

var formatters []*registeredFormatter

// RegisterFormatter registers a feature suite output
// Formatter as the name and descriptiongiven.
// Formatter is used to represent suite output
func RegisterFormatter(name, description string, f Formatter) {
	formatters = append(formatters, &registeredFormatter{
		name:        name,
		fmt:         f,
		description: description,
	})
}

// Formatter is an interface for feature runner
// output summary presentation.
//
// New formatters may be created to represent
// suite results in different ways. These new
// formatters needs to be registered with a
// RegisterFormatter function call
type Formatter interface {
	Feature(*gherkin.Feature, string)
	Node(interface{})
	Failed(*gherkin.Step, *StepDef, error)
	Passed(*gherkin.Step, *StepDef)
	Skipped(*gherkin.Step)
	Undefined(*gherkin.Step)
	Pending(*gherkin.Step, *StepDef)
	Summary()
}

// failed represents a failed step data structure
// with all necessary references
type failed struct {
	feature *feature
	owner   interface{}
	step    *gherkin.Step
	def     *StepDef
	err     error
}

func (f failed) line() string {
	return fmt.Sprintf("%s:%d", f.feature.Path, f.step.Location.Line)
}

// passed represents a successful step data structure
// with all necessary references
type passed struct {
	feature *feature
	owner   interface{}
	step    *gherkin.Step
	def     *StepDef
}

// skipped represents a skipped step data structure
// with all necessary references
type skipped struct {
	feature *feature
	owner   interface{}
	step    *gherkin.Step
}

// undefined represents an undefined step data structure
// with all necessary references
type undefined struct {
	feature *feature
	owner   interface{}
	step    *gherkin.Step
}

// pending represents a pending step data structure
// with all necessary references
type pending struct {
	feature *feature
	owner   interface{}
	step    *gherkin.Step
	def     *StepDef
}
