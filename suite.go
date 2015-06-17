package godog

import (
	"flag"
	"fmt"
	"regexp"

	"github.com/DATA-DOG/godog/gherkin"
)

type BeforeScenarioHandler interface {
	BeforeScenario(scenario *gherkin.Scenario)
}

type BeforeScenarioHandlerFunc func(scenario *gherkin.Scenario)

func (f BeforeScenarioHandlerFunc) BeforeScenario(scenario *gherkin.Scenario) {
	f(scenario)
}

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
	HandleStep(args ...Arg) error
}

// StepHandlerFunc type is an adapter to allow the use of
// ordinary functions as Step handlers.  If f is a function
// with the appropriate signature, StepHandlerFunc(f) is a
// StepHandler object that calls f.
type StepHandlerFunc func(...Arg) error

// HandleStep calls f(step_arguments...).
func (f StepHandlerFunc) HandleStep(args ...Arg) error {
	return f(args...)
}

var errPending = fmt.Errorf("pending step")

type stepMatchHandler struct {
	handler StepHandler
	expr    *regexp.Regexp
}

// Suite is an interface which allows various contexts
// to register step definitions and event handlers
type Suite interface {
	Step(expr *regexp.Regexp, h StepHandler)
	BeforeScenario(h BeforeScenarioHandler)
}

type suite struct {
	beforeScenarioHandlers []BeforeScenarioHandler
	stepHandlers           []*stepMatchHandler
	features               []*gherkin.Feature
	fmt                    Formatter
}

// New initializes a suite which supports the Suite
// interface. The instance is passed around to all
// context initialization functions from *_test.go files
func New() *suite {
	return &suite{}
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
func (s *suite) Step(expr *regexp.Regexp, h StepHandler) {
	s.stepHandlers = append(s.stepHandlers, &stepMatchHandler{
		handler: h,
		expr:    expr,
	})
}

func (s *suite) BeforeScenario(h BeforeScenarioHandler) {
	s.beforeScenarioHandlers = append(s.beforeScenarioHandlers, h)
}

// Run - runs a godog feature suite
func (s *suite) Run() {
	var err error
	if !flag.Parsed() {
		flag.Parse()
	}
	fatal(cfg.validate())

	s.fmt = cfg.formatter()
	s.features, err = cfg.features()
	fatal(err)

	for _, f := range s.features {
		s.runFeature(f)
	}
	s.fmt.Summary()
}

func (s *suite) runStep(step *gherkin.Step) (err error) {
	var match *stepMatchHandler
	var args []Arg
	for _, h := range s.stepHandlers {
		if m := h.expr.FindStringSubmatch(step.Text); len(m) > 0 {
			match = h
			for _, a := range m[1:] {
				args = append(args, Arg(a))
			}
			break
		}
	}
	if match == nil {
		s.fmt.Undefined(step)
		return errPending
	}

	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
			s.fmt.Failed(step, match, err)
		}
	}()

	if err = match.handler.HandleStep(args...); err != nil {
		s.fmt.Failed(step, match, err)
	} else {
		s.fmt.Passed(step, match)
	}
	return
}

func (s *suite) runSteps(steps []*gherkin.Step) bool {
	var failed bool
	for _, step := range steps {
		if failed {
			s.fmt.Skipped(step)
			continue
		}
		if err := s.runStep(step); err != nil {
			failed = true
		}
	}
	return failed
}

func (s *suite) skipSteps(steps []*gherkin.Step) {
	for _, step := range steps {
		s.fmt.Skipped(step)
	}
}

func (s *suite) runFeature(f *gherkin.Feature) {
	s.fmt.Node(f)
	var failed bool
	for _, scenario := range f.Scenarios {
		// run before scenario handlers
		for _, h := range s.beforeScenarioHandlers {
			h.BeforeScenario(scenario)
		}
		// background
		if f.Background != nil && !failed {
			s.fmt.Node(f.Background)
			failed = s.runSteps(f.Background.Steps)
		}

		// scenario
		s.fmt.Node(scenario)
		if failed {
			s.skipSteps(scenario.Steps)
		} else {
			s.runSteps(scenario.Steps)
		}
	}
}
