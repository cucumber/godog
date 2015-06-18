package godog

import (
	"fmt"

	"github.com/DATA-DOG/godog/gherkin"
)

// Formatter is an interface for feature runner
// output summary presentation
type Formatter interface {
	Node(interface{})
	Failed(*gherkin.Step, *stepMatchHandler, error)
	Passed(*gherkin.Step, *stepMatchHandler)
	Skipped(*gherkin.Step)
	Undefined(*gherkin.Step)
	Summary()
}

// failed represents a failed step data structure
// with all necessary references
type failed struct {
	step    *gherkin.Step
	handler *stepMatchHandler
	err     error
}

func (f failed) line() string {
	var tok *gherkin.Token
	var ft *gherkin.Feature
	if f.step.Scenario != nil {
		tok = f.step.Scenario.Token
		ft = f.step.Scenario.Feature
	} else {
		tok = f.step.Background.Token
		ft = f.step.Background.Feature
	}
	return fmt.Sprintf("%s:%d", ft.Path, tok.Line)
}

// passed represents a successful step data structure
// with all necessary references
type passed struct {
	step    *gherkin.Step
	handler *stepMatchHandler
}

// skipped represents a skipped step data structure
// with all necessary references
type skipped struct {
	step *gherkin.Step
}

// undefined represents a pending step data structure
// with all necessary references
type undefined struct {
	step *gherkin.Step
}
