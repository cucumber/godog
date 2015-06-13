package godog

import (
	"flag"
	"fmt"
	"regexp"
	"strconv"

	"github.com/DATA-DOG/godog/gherkin"
)

// Arg is an argument for StepHandler parsed from
// the regexp submatch to handle the step
type Arg string

// Float converts an argument to float64
// or panics if unable to convert it
func (a Arg) Float() float64 {
	v, err := strconv.ParseFloat(string(a), 64)
	if err == nil {
		return v
	}
	panic(fmt.Sprintf(`cannot convert "%s" to float64: %s`, a, err))
}

// Int converts an argument to int64
// or panics if unable to convert it
func (a Arg) Int() int64 {
	v, err := strconv.ParseInt(string(a), 10, 0)
	if err == nil {
		return v
	}
	panic(fmt.Sprintf(`cannot convert "%s" to int64: %s`, a, err))
}

// String converts an argument to string
func (a Arg) String() string {
	return string(a)
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

// Suite is an interface which allows various contexts
// to register step definitions and event handlers
type Suite interface {
	Step(exp *regexp.Regexp, h StepHandler)
}

type suite struct {
	steps    map[*regexp.Regexp]StepHandler
	features []*gherkin.Feature
	fmt      formatter
}

// New initializes a suite which supports the Suite
// interface. The instance is passed around to all
// context initialization functions from *_test.go files
func New() *suite {
	// @TODO: colorize flag help output
	flag.StringVar(&cfg.featuresPath, "features", "features", "Path to feature files")
	flag.StringVar(&cfg.formatterName, "formatter", "pretty", "Formatter name")
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
	s.fmt = cfg.formatter()
	s.features, err = cfg.features()
	fatal(err)

	fmt.Println("running", cl("godog", cyan)+", num registered steps:", cl(len(s.steps), yellow))
	fmt.Println("have loaded", cl(len(s.features), yellow), "features from path:", cl(cfg.featuresPath, green))

	for _, f := range s.features {
		s.runFeature(f)
	}
}

func (s *suite) runStep(step *gherkin.Step) {
	var handler StepHandler
	var args []Arg
	for r, h := range s.steps {
		if m := r.FindStringSubmatch(step.Text); len(m) > 0 {
			handler = h
			for _, a := range m[1:] {
				args = append(args, Arg(a))
			}
			break
		}
	}
	if handler == nil {
		fmt.Println("PENDING")
		return
	}
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("PANIC")
		}
	}()
	if err := handler.HandleStep(args...); err != nil {
		fmt.Println("ERR")
	} else {
		fmt.Println("OK")
	}
}

func (s *suite) runFeature(f *gherkin.Feature) {
	s.fmt.node(f)
	var background bool
	for _, scenario := range f.Scenarios {
		if f.Background != nil {
			if !background {
				s.fmt.node(f.Background)
			}
			for _, step := range f.Background.Steps {
				s.runStep(step)
				if !background {
					s.fmt.node(step)
				}
			}
			background = true
		}
		s.fmt.node(scenario)
		for _, step := range scenario.Steps {
			s.runStep(step)
			s.fmt.node(step)
		}
	}
}
