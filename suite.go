package godog

import (
	"flag"
	"fmt"
	"regexp"

	"github.com/DATA-DOG/godog/gherkin"
)

// Objects implementing the StepHandler interface can be
// registered as step definitions in godog
//
// HandleStep method receives all arguments which
// will be matched according to the regular expression
// which is passed with a step registration.
// The error in return - represents a reason of failure.
//
// Returning signals that the step has finished
// and that the feature runner can move on to the next
// step.
type StepHandler interface {
	HandleStep(args ...interface{}) error
}

// StepHandlerFunc type is an adapter to allow the use of
// ordinary functions as Step handlers.  If f is a function
// with the appropriate signature, StepHandlerFunc(f) is a
// StepHandler object that calls f.
type StepHandlerFunc func(...interface{}) error

// HandleStep calls f(step_arguments...).
func (f StepHandlerFunc) HandleStep(args ...interface{}) error {
	return f(args...)
}

// Suite is an interface which allows various contexts
// to register step definitions and event handlers
type Suite interface {
	Step(exp *regexp.Regexp, h StepHandler)
}

type suite struct {
	steps    map[*regexp.Regexp]StepHandler
	features []*gherkin.Feature
}

// New initializes a suite which supports the Suite
// interface. The instance is passed around to all
// context initialization functions from *_test.go files
func New() *suite {
	if !flag.Parsed() {
		flag.Parse()
	}
	return &suite{
		steps: make(map[*regexp.Regexp]StepHandler),
	}
}

// Step allows to register a StepHandler in Godog
// feature suite, the handler will be applied to all
// steps matching the given regexp
//
// Note that if there are two handlers which may match
// the same step, then the only first matched handler
// will be applied
//
// If none of the StepHandlers are matched, then a pending
// step error will be raised.
func (s *suite) Step(exp *regexp.Regexp, h StepHandler) {
	s.steps[exp] = h
}

// Run - runs a godog feature suite
func (s *suite) Run() {
	var err error
	s.features, err = cfg.features()
	fatal(err)

	fmt.Println("running", cl("godog", cyan)+", num registered steps:", cl(len(s.steps), yellow))
	fmt.Println("have loaded", cl(len(s.features), yellow), "features from path:", cl(cfg.featuresPath, green))
}
