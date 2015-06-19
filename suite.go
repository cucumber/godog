package godog

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strings"

	"github.com/DATA-DOG/godog/gherkin"
)

// Regexp is an unified type for regular expression
// it can be either a string or a *regexp.Regexp
type Regexp interface{}

// Handler is an unified type for a StepHandler
// interface satisfaction. It may be a function
// or a step handler
type Handler interface{}

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
	HandleStep(args ...*Arg) error
}

// StepHandlerFunc type is an adapter to allow the use of
// ordinary functions as Step handlers.  If f is a function
// with the appropriate signature, StepHandlerFunc(f) is a
// StepHandler object that calls f.
type StepHandlerFunc func(...*Arg) error

// HandleStep calls f(step_arguments...).
func (f StepHandlerFunc) HandleStep(args ...*Arg) error {
	return f(args...)
}

// ErrUndefined is returned in case if step definition was not found
var ErrUndefined = fmt.Errorf("step is undefined")

// StepDef is a registered step definition
// contains a StepHandler, a regexp which
// is used to match a step and Args which
// were matched by last step
type StepDef struct {
	Args    []*Arg
	Handler StepHandler
	Expr    *regexp.Regexp
}

// Suite is an interface which allows various contexts
// to register step definitions and event handlers
type Suite interface {
	Step(expr Regexp, h Handler)
	// suite events
	BeforeSuite(h BeforeSuiteHandler)
	BeforeScenario(h BeforeScenarioHandler)
	BeforeStep(h BeforeStepHandler)
	AfterStep(h AfterStepHandler)
	AfterScenario(h AfterScenarioHandler)
	AfterSuite(h AfterSuiteHandler)
}

type suite struct {
	stepHandlers []*StepDef
	features     []*gherkin.Feature
	fmt          Formatter

	failed bool

	// suite event handlers
	beforeSuiteHandlers    []BeforeSuiteHandler
	beforeScenarioHandlers []BeforeScenarioHandler
	beforeStepHandlers     []BeforeStepHandler
	afterStepHandlers      []AfterStepHandler
	afterScenarioHandlers  []AfterScenarioHandler
	afterSuiteHandlers     []AfterSuiteHandler
}

// New initializes a suite which supports the Suite
// interface. The instance is passed around to all
// context initialization functions from *_test.go files
func New() *suite {
	return &suite{}
}

// Step allows to register a StepHandler in Godog
// feature suite, the handler will be applied to all
// steps matching the given regexp expr
//
// It will panic if expr is not a valid regular expression
// or handler does not satisfy StepHandler interface
//
// Note that if there are two handlers which may match
// the same step, then the only first matched handler
// will be applied
//
// If none of the StepHandlers are matched, then a pending
// step error will be raised.
func (s *suite) Step(expr Regexp, h Handler) {
	var handler StepHandler
	var regex *regexp.Regexp

	switch t := expr.(type) {
	case *regexp.Regexp:
		regex = t
	case string:
		regex = regexp.MustCompile(t)
	case []byte:
		regex = regexp.MustCompile(string(t))
	default:
		panic(fmt.Sprintf("expecting expr to be a *regexp.Regexp or a string, got type: %T", expr))
	}

	switch t := h.(type) {
	case StepHandlerFunc:
		handler = t
	case StepHandler:
		handler = t
	case func(...*Arg) error:
		handler = StepHandlerFunc(t)
	default:
		panic(fmt.Sprintf("expecting handler to satisfy StepHandler interface, got type: %T", h))
	}
	s.stepHandlers = append(s.stepHandlers, &StepDef{
		Handler: handler,
		Expr:    regex,
	})
}

// BeforeSuite registers a BeforeSuiteHandler
// to be run once before suite runner
func (s *suite) BeforeSuite(h BeforeSuiteHandler) {
	s.beforeSuiteHandlers = append(s.beforeSuiteHandlers, h)
}

// BeforeScenario registers a BeforeScenarioHandler
// to be run before every scenario
func (s *suite) BeforeScenario(h BeforeScenarioHandler) {
	s.beforeScenarioHandlers = append(s.beforeScenarioHandlers, h)
}

// BeforeStep registers a BeforeStepHandler
// to be run before every scenario
func (s *suite) BeforeStep(h BeforeStepHandler) {
	s.beforeStepHandlers = append(s.beforeStepHandlers, h)
}

// AfterStep registers an AfterStepHandler
// to be run after every scenario
func (s *suite) AfterStep(h AfterStepHandler) {
	s.afterStepHandlers = append(s.afterStepHandlers, h)
}

// AfterScenario registers an AfterScenarioHandler
// to be run after every scenario
func (s *suite) AfterScenario(h AfterScenarioHandler) {
	s.afterScenarioHandlers = append(s.afterScenarioHandlers, h)
}

// AfterSuite registers a AfterSuiteHandler
// to be run once after suite runner
func (s *suite) AfterSuite(h AfterSuiteHandler) {
	s.afterSuiteHandlers = append(s.afterSuiteHandlers, h)
}

// Run - runs a godog feature suite
func (s *suite) Run() {
	var err error
	if !flag.Parsed() {
		flag.Parse()
	}

	// check if we need to just show something first
	switch {
	case cfg.version:
		fmt.Println(cl("Godog", green) + " version is " + cl(Version, yellow))
		return
	case cfg.definitions:
		s.printStepDefinitions()
		return
	}

	// run a feature suite
	fatal(cfg.validate())
	s.fmt = cfg.formatter()
	s.features, err = cfg.features()
	fatal(err)

	s.run()

	if s.failed {
		os.Exit(1)
	}
}

func (s *suite) run() {
	// run before suite handlers
	for _, h := range s.beforeSuiteHandlers {
		h.HandleBeforeSuite()
	}
	// run features
	for _, f := range s.features {
		s.runFeature(f)
		if s.failed && cfg.stopOnFailure {
			// stop on first failure
			break
		}
	}
	// run after suite handlers
	for _, h := range s.afterSuiteHandlers {
		h.HandleAfterSuite()
	}
	s.fmt.Summary()
}

func (s *suite) matchStep(step *gherkin.Step) *StepDef {
	for _, h := range s.stepHandlers {
		if m := h.Expr.FindStringSubmatch(step.Text); len(m) > 0 {
			var args []*Arg
			for _, a := range m[1:] {
				args = append(args, &Arg{value: a})
			}
			if step.Table != nil {
				args = append(args, &Arg{value: step.Table})
			}
			if step.PyString != nil {
				args = append(args, &Arg{value: step.PyString})
			}
			h.Args = args
			return h
		}
	}
	return nil
}

func (s *suite) runStep(step *gherkin.Step) (err error) {
	match := s.matchStep(step)
	if match == nil {
		s.fmt.Undefined(step)
		return ErrUndefined
	}

	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
			s.fmt.Failed(step, match, err)
		}
	}()

	if err = match.Handler.HandleStep(match.Args...); err != nil {
		s.fmt.Failed(step, match, err)
	} else {
		s.fmt.Passed(step, match)
	}
	return
}

func (s *suite) runSteps(steps []*gherkin.Step) (err error) {
	for _, step := range steps {
		if err != nil {
			s.fmt.Skipped(step)
			continue
		}

		// run before step handlers
		for _, h := range s.beforeStepHandlers {
			h.HandleBeforeStep(step)
		}

		err = s.runStep(step)

		// run after step handlers
		for _, h := range s.afterStepHandlers {
			h.HandleAfterStep(step, err)
		}
	}
	return
}

func (s *suite) skipSteps(steps []*gherkin.Step) {
	for _, step := range steps {
		s.fmt.Skipped(step)
	}
}

func (s *suite) runOutline(scenario *gherkin.Scenario) (err error) {
	placeholders := scenario.Outline.Examples.Rows[0]
	examples := scenario.Outline.Examples.Rows[1:]
	for _, example := range examples {
		var steps []*gherkin.Step
		for _, step := range scenario.Outline.Steps {
			text := step.Text
			for i, placeholder := range placeholders {
				text = strings.Replace(text, "<"+placeholder+">", example[i], -1)
			}
			// clone a step
			cloned := &gherkin.Step{
				Token:      step.Token,
				Text:       text,
				Type:       step.Type,
				PyString:   step.PyString,
				Table:      step.Table,
				Background: step.Background,
				Scenario:   scenario,
			}
			steps = append(steps, cloned)
		}

		// set steps to scenario
		scenario.Steps = steps
		if err = s.runScenario(scenario); err != nil && err != ErrUndefined {
			s.failed = true
			if cfg.stopOnFailure {
				return
			}
		}
	}
	return
}

func (s *suite) runFeature(f *gherkin.Feature) {
	s.fmt.Node(f)
	for _, scenario := range f.Scenarios {
		var err error
		// handle scenario outline differently
		if scenario.Outline != nil {
			err = s.runOutline(scenario)
		} else {
			err = s.runScenario(scenario)
		}
		if err != nil && err != ErrUndefined {
			s.failed = true
			if cfg.stopOnFailure {
				return
			}
		}
	}
}

func (s *suite) runScenario(scenario *gherkin.Scenario) (err error) {
	// run before scenario handlers
	for _, h := range s.beforeScenarioHandlers {
		h.HandleBeforeScenario(scenario)
	}

	// background
	if scenario.Feature.Background != nil {
		s.fmt.Node(scenario.Feature.Background)
		err = s.runSteps(scenario.Feature.Background.Steps)
	}

	// scenario
	s.fmt.Node(scenario)
	switch err {
	case ErrUndefined:
		s.skipSteps(scenario.Steps)
	case nil:
		err = s.runSteps(scenario.Steps)
	default:
		s.skipSteps(scenario.Steps)
	}

	// run after scenario handlers
	for _, h := range s.afterScenarioHandlers {
		h.HandleAfterScenario(scenario, err)
	}

	return
}

func (st *suite) printStepDefinitions() {
	var longest int
	for _, def := range st.stepHandlers {
		if longest < len(def.Expr.String()) {
			longest = len(def.Expr.String())
		}
	}
	for _, def := range st.stepHandlers {
		location := runtime.FuncForPC(reflect.ValueOf(def.Handler).Pointer()).Name()
		fmt.Println(cl(def.Expr.String(), yellow)+s(longest-len(def.Expr.String())), cl("# "+location, black))
	}
	if len(st.stepHandlers) == 0 {
		fmt.Println("there were no contexts registered, could not find any step definition..")
	}
}
