package godog

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/DATA-DOG/godog/gherkin"
)

// Regexp is an unified type for regular expression
// it can be either a string or a *regexp.Regexp
type Regexp interface{}

// StepHandler is a func to handle the step
//
// The handler receives all arguments which
// will be matched according to the Regexp
// which is passed with a step registration.
//
// The error in return - represents a reason of failure.
// All consequent scenario steps are skipped.
//
// Returning signals that the step has finished and that
// the feature runner can move on to the next step.
type StepHandler func(...*Arg) error

// ErrUndefined is returned in case if step definition was not found
var ErrUndefined = fmt.Errorf("step is undefined")

// StepDef is a registered step definition
// contains a StepHandler and regexp which
// is used to match a step. Args which
// were matched by last executed step
//
// This structure is passed to the formatter
// when step is matched and is either failed
// or successful
type StepDef struct {
	Args    []*Arg
	Handler StepHandler
	Expr    *regexp.Regexp
}

// Suite is an interface which allows various contexts
// to register steps and event handlers.
//
// When running a test suite, this interface is passed
// to all functions (contexts), which have it as a
// first and only argument.
//
// Note that all event hooks does not catch panic errors
// in order to have a trace information. Only step
// executions are catching panic error since it may
// be a context specific error.
type Suite interface {
	Run()
	Step(expr Regexp, h StepHandler)
	// suite events
	BeforeSuite(f func())
	BeforeScenario(f func(*gherkin.Scenario))
	BeforeStep(f func(*gherkin.Step))
	AfterStep(f func(*gherkin.Step, error))
	AfterScenario(f func(*gherkin.Scenario, error))
	AfterSuite(f func())
}

type suite struct {
	stepHandlers []*StepDef
	features     []*gherkin.Feature
	fmt          Formatter

	failed bool

	// options
	paths         []string
	format        string
	tags          string
	definitions   bool
	stopOnFailure bool
	version       bool

	// suite event handlers
	beforeSuiteHandlers    []func()
	beforeScenarioHandlers []func(*gherkin.Scenario)
	beforeStepHandlers     []func(*gherkin.Step)
	afterStepHandlers      []func(*gherkin.Step, error)
	afterScenarioHandlers  []func(*gherkin.Scenario, error)
	afterSuiteHandlers     []func()
}

// New initializes a Suite. The instance is passed around
// to all context initialization functions from *_test.go files
func New() Suite {
	return &suite{}
}

// Step allows to register a StepHandler in Godog
// feature suite, the handler will be applied to all
// steps matching the given Regexp expr
//
// It will panic if expr is not a valid regular expression
//
// Note that if there are two handlers which may match
// the same step, then the only first matched handler
// will be applied.
//
// If none of the StepHandlers are matched, then
// ErrUndefined error will be returned.
func (s *suite) Step(expr Regexp, h StepHandler) {
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

	s.stepHandlers = append(s.stepHandlers, &StepDef{
		Handler: h,
		Expr:    regex,
	})
}

// BeforeSuite registers a function or method
// to be run once before suite runner.
//
// Use it to prepare the test suite for a spin.
// Connect and prepare database for instance...
func (s *suite) BeforeSuite(f func()) {
	s.beforeSuiteHandlers = append(s.beforeSuiteHandlers, f)
}

// BeforeScenario registers a function or method
// to be run before every scenario.
//
// It is a good practice to restore the default state
// before every scenario so it would be isolated from
// any kind of state.
func (s *suite) BeforeScenario(f func(*gherkin.Scenario)) {
	s.beforeScenarioHandlers = append(s.beforeScenarioHandlers, f)
}

// BeforeStep registers a function or method
// to be run before every scenario
func (s *suite) BeforeStep(f func(*gherkin.Step)) {
	s.beforeStepHandlers = append(s.beforeStepHandlers, f)
}

// AfterStep registers an function or method
// to be run after every scenario
//
// It may be convenient to return a different kind of error
// in order to print more state details which may help
// in case of step failure
//
// In some cases, for example when running a headless
// browser, to take a screenshot after failure.
func (s *suite) AfterStep(f func(*gherkin.Step, error)) {
	s.afterStepHandlers = append(s.afterStepHandlers, f)
}

// AfterScenario registers an function or method
// to be run after every scenario
func (s *suite) AfterScenario(f func(*gherkin.Scenario, error)) {
	s.afterScenarioHandlers = append(s.afterScenarioHandlers, f)
}

// AfterSuite registers a function or method
// to be run once after suite runner
func (s *suite) AfterSuite(f func()) {
	s.afterSuiteHandlers = append(s.afterSuiteHandlers, f)
}

// Run starts the Godog feature suite
func (s *suite) Run() {
	flagSet := flags(s)
	fatal(flagSet.Parse(os.Args[1:]))

	s.paths = flagSet.Args()
	// check the default path
	if len(s.paths) == 0 {
		inf, err := os.Stat("features")
		if err == nil && inf.IsDir() {
			s.paths = []string{"features"}
		}
	}
	// validate formatter
	var names []string
	for _, f := range formatters {
		if f.name == s.format {
			s.fmt = f.fmt
			break
		}
		names = append(names, f.name)
	}

	if s.fmt == nil {
		fatal(fmt.Errorf(`unregistered formatter name: "%s", use one of: %s`, s.format, strings.Join(names, ", ")))
	}

	// check if we need to just show something first
	switch {
	case s.version:
		fmt.Println(cl("Godog", green) + " version is " + cl(Version, yellow))
		return
	case s.definitions:
		s.printStepDefinitions()
		return
	}

	fatal(s.parseFeatures())
	// run a feature suite
	s.run()

	if s.failed {
		os.Exit(1)
	}
}

func (s *suite) run() {
	// run before suite handlers
	for _, f := range s.beforeSuiteHandlers {
		f()
	}
	// run features
	for _, f := range s.features {
		s.runFeature(f)
		if s.failed && s.stopOnFailure {
			// stop on first failure
			break
		}
	}
	// run after suite handlers
	for _, f := range s.afterSuiteHandlers {
		f()
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

	if err = match.Handler(match.Args...); err != nil {
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
		for _, f := range s.beforeStepHandlers {
			f(step)
		}

		err = s.runStep(step)

		// run after step handlers
		for _, f := range s.afterStepHandlers {
			f(step, err)
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
			if s.stopOnFailure {
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
			if s.stopOnFailure {
				return
			}
		}
	}
}

func (s *suite) runScenario(scenario *gherkin.Scenario) (err error) {
	// run before scenario handlers
	for _, f := range s.beforeScenarioHandlers {
		f(scenario)
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
	for _, f := range s.afterScenarioHandlers {
		f(scenario, err)
	}

	return
}

func (s *suite) printStepDefinitions() {
	var longest int
	for _, def := range s.stepHandlers {
		if longest < len(def.Expr.String()) {
			longest = len(def.Expr.String())
		}
	}
	for _, def := range s.stepHandlers {
		location := runtime.FuncForPC(reflect.ValueOf(def.Handler).Pointer()).Name()
		spaces := strings.Repeat(" ", longest-len(def.Expr.String()))
		fmt.Println(cl(def.Expr.String(), yellow)+spaces, cl("# "+location, black))
	}
	if len(s.stepHandlers) == 0 {
		fmt.Println("there were no contexts registered, could not find any step definition..")
	}
}

func (s *suite) parseFeatures() (err error) {
	for _, pat := range s.paths {
		// check if line number is specified
		parts := strings.Split(pat, ":")
		path := parts[0]
		line := -1
		if len(parts) > 1 {
			line, err = strconv.Atoi(parts[1])
			if err != nil {
				return fmt.Errorf("line number should follow after colon path delimiter")
			}
		}
		// parse features
		err = filepath.Walk(path, func(p string, f os.FileInfo, err error) error {
			if err == nil && !f.IsDir() && strings.HasSuffix(p, ".feature") {
				ft, err := gherkin.ParseFile(p)
				switch {
				case err == gherkin.ErrEmpty:
					// its ok, just skip it
				case err != nil:
					return err
				default:
					s.features = append(s.features, ft)
				}
				// filter scenario by line number
				if line != -1 {
					var scenarios []*gherkin.Scenario
					for _, s := range ft.Scenarios {
						if s.Token.Line == line {
							scenarios = append(scenarios, s)
							break
						}
					}
					ft.Scenarios = scenarios
				}
				s.applyTagFilter(ft)
			}
			return err
		})
		// check error
		switch {
		case os.IsNotExist(err):
			return fmt.Errorf(`feature path "%s" is not available`, path)
		case os.IsPermission(err):
			return fmt.Errorf(`feature path "%s" is not accessible`, path)
		case err != nil:
			return err
		}
	}
	return
}

func (s *suite) applyTagFilter(ft *gherkin.Feature) {
	if len(s.tags) == 0 {
		return
	}

	var scenarios []*gherkin.Scenario
	for _, scenario := range ft.Scenarios {
		if s.matchesTags(scenario.Tags) {
			scenarios = append(scenarios, scenario)
		}
	}
	ft.Scenarios = scenarios
}

// based on http://behat.readthedocs.org/en/v2.5/guides/6.cli.html#gherkin-filters
func (s *suite) matchesTags(tags gherkin.Tags) (ok bool) {
	ok = true
	for _, andTags := range strings.Split(s.tags, "&&") {
		var okComma bool
		for _, tag := range strings.Split(andTags, ",") {
			tag = strings.Replace(strings.TrimSpace(tag), "@", "", -1)
			if tag[0] == '~' {
				tag = tag[1:]
				okComma = !tags.Has(gherkin.Tag(tag)) || okComma
			} else {
				okComma = tags.Has(gherkin.Tag(tag)) || okComma
			}
		}
		ok = (false != okComma && ok && okComma) || false
	}
	return
}
