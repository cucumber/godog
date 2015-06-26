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

	"github.com/cucumber/gherkin-go"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()
var typeOfBytes = reflect.TypeOf([]byte(nil))

type feature struct {
	*gherkin.Feature
	Path string `json:"path"`
}

// ErrUndefined is returned in case if step definition was not found
var ErrUndefined = fmt.Errorf("step is undefined")

// ErrPending should be returned by step definition if
// step implementation is pending
var ErrPending = fmt.Errorf("step implementation is pending")

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
	// Run the test suite
	Run()

	// Registers a step which will execute stepFunc
	// on step expr match
	//
	// expr can be either a string or a *regexp.Regexp
	// stepFunc is a func to handle the step, arguments
	// are set from matched step
	Step(expr interface{}, h interface{})

	// BeforeSuite registers a func to run on initial
	// suite startup
	BeforeSuite(f func())

	// BeforeScenario registers a func to run before
	// every *gherkin.Scenario or *gherkin.ScenarioOutline
	BeforeScenario(f func(interface{}))

	// BeforeStep register a handler before every step
	BeforeStep(f func(*gherkin.Step))

	// AfterStep register a handler after every step
	AfterStep(f func(*gherkin.Step, error))

	// AfterScenario registers a func to run after
	// every *gherkin.Scenario or *gherkin.ScenarioOutline
	AfterScenario(f func(interface{}, error))

	// AfterSuite runs func int the end of tests
	AfterSuite(f func())
}

type suite struct {
	steps    []*StepDef
	features []*feature
	fmt      Formatter

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
	beforeScenarioHandlers []func(interface{})
	beforeStepHandlers     []func(*gherkin.Step)
	afterStepHandlers      []func(*gherkin.Step, error)
	afterScenarioHandlers  []func(interface{}, error)
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
func (s *suite) Step(expr interface{}, stepFunc interface{}) {
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

	v := reflect.ValueOf(stepFunc)
	typ := v.Type()
	if typ.Kind() != reflect.Func {
		panic(fmt.Sprintf("expected handler to be func, but got: %T", stepFunc))
	}
	if typ.NumOut() != 1 {
		panic(fmt.Sprintf("expected handler to return an error, but it has more values in return: %d", typ.NumOut()))
	}
	if typ.Out(0).Kind() != reflect.Interface || !typ.Out(0).Implements(errorInterface) {
		panic(fmt.Sprintf("expected handler to return an error interface, but we have: %s", typ.Out(0).Kind()))
	}
	s.steps = append(s.steps, &StepDef{
		Handler: stepFunc,
		Expr:    regex,
		hv:      v,
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
// to be run before every scenario or scenario outline.
//
// The interface argument may be *gherkin.Scenario
// or *gherkin.ScenarioOutline
//
// It is a good practice to restore the default state
// before every scenario so it would be isolated from
// any kind of state.
func (s *suite) BeforeScenario(f func(interface{})) {
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
// to be run after every scenario or scenario outline
//
// The interface argument may be *gherkin.Scenario
// or *gherkin.ScenarioOutline
func (s *suite) AfterScenario(f func(interface{}, error)) {
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
	for _, h := range s.steps {
		if m := h.Expr.FindStringSubmatch(step.Text); len(m) > 0 {
			var args []interface{}
			for _, m := range m[1:] {
				args = append(args, m)
			}
			if step.Argument != nil {
				args = append(args, step.Argument)
			}
			h.args = args
			return h
		}
	}
	return nil
}

func (s *suite) runStep(step *gherkin.Step, prevStepErr error) (err error) {
	match := s.matchStep(step)
	if match == nil {
		s.fmt.Undefined(step)
		return ErrUndefined
	}

	if prevStepErr != nil {
		s.fmt.Skipped(step)
		return nil
	}

	defer func() {
		if e := recover(); e != nil {
			err, ok := e.(error)
			if !ok {
				err = fmt.Errorf(e.(string))
			}
			s.fmt.Failed(step, match, err)
		}
	}()

	err = match.run()
	switch err {
	case nil:
		s.fmt.Passed(step, match)
	case ErrPending:
		s.fmt.Pending(step, match)
	default:
		s.fmt.Failed(step, match, err)
	}
	return
}

func (s *suite) runSteps(steps []*gherkin.Step) (err error) {
	for _, step := range steps {
		// run before step handlers
		for _, f := range s.beforeStepHandlers {
			f(step)
		}

		stepErr := s.runStep(step, err)
		switch stepErr {
		case ErrUndefined:
			err = stepErr
		case ErrPending:
			err = stepErr
		case nil:
		default:
			err = stepErr
		}

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

func (s *suite) runOutline(outline *gherkin.ScenarioOutline, b *gherkin.Background) (err error) {
	s.fmt.Node(outline)

	for _, example := range outline.Examples {
		s.fmt.Node(example)

		placeholders := example.TableHeader.Cells
		groups := example.TableBody

		for _, group := range groups {
			for _, f := range s.beforeScenarioHandlers {
				f(outline)
			}
			var steps []*gherkin.Step
			for _, outlineStep := range outline.Steps {
				text := outlineStep.Text
				for i, placeholder := range placeholders {
					text = strings.Replace(text, "<"+placeholder.Value+">", group.Cells[i].Value, -1)
				}
				// clone a step
				step := &gherkin.Step{
					Node:     outlineStep.Node,
					Text:     text,
					Keyword:  outlineStep.Keyword,
					Argument: outlineStep.Argument,
				}
				steps = append(steps, step)
			}
			// run background
			if b != nil {
				err = s.runSteps(b.Steps)
			}
			switch err {
			case ErrUndefined:
				s.skipSteps(steps)
			case nil:
				err = s.runSteps(steps)
			default:
				s.skipSteps(steps)
			}

			for _, f := range s.afterScenarioHandlers {
				f(outline, err)
			}

			if s.stopOnFailure && err != ErrUndefined {
				return
			}
		}
	}
	return
}

func (s *suite) runFeature(f *feature) {
	s.fmt.Feature(f.Feature, f.Path)
	for _, scenario := range f.ScenarioDefinitions {
		var err error
		if f.Background != nil {
			s.fmt.Node(f.Background)
		}
		switch t := scenario.(type) {
		case *gherkin.ScenarioOutline:
			err = s.runOutline(t, f.Background)
		case *gherkin.Scenario:
			err = s.runScenario(t, f.Background)
		}
		if err != nil && err != ErrUndefined && err != ErrPending {
			s.failed = true
			if s.stopOnFailure {
				return
			}
		}
	}
}

func (s *suite) runScenario(scenario *gherkin.Scenario, b *gherkin.Background) (err error) {
	// run before scenario handlers
	for _, f := range s.beforeScenarioHandlers {
		f(scenario)
	}

	// background
	if b != nil {
		err = s.runSteps(b.Steps)
	}

	// scenario
	s.fmt.Node(scenario)
	switch err {
	case ErrPending:
		s.skipSteps(scenario.Steps)
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
	for _, def := range s.steps {
		if longest < len(def.Expr.String()) {
			longest = len(def.Expr.String())
		}
	}
	for _, def := range s.steps {
		location := runtime.FuncForPC(def.hv.Pointer()).Name()
		spaces := strings.Repeat(" ", longest-len(def.Expr.String()))
		fmt.Println(cl(def.Expr.String(), yellow)+spaces, cl("# "+location, black))
	}
	if len(s.steps) == 0 {
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
				reader, err := os.Open(p)
				if err != nil {
					return err
				}
				ft, err := gherkin.ParseFeature(reader)
				reader.Close()
				if err != nil {
					return err
				}
				s.features = append(s.features, &feature{Path: p, Feature: ft})
				// filter scenario by line number
				if line != -1 {
					var scenarios []interface{}
					for _, def := range ft.ScenarioDefinitions {
						var ln int
						switch t := def.(type) {
						case *gherkin.Scenario:
							ln = t.Location.Line
						case *gherkin.ScenarioOutline:
							ln = t.Location.Line
						}
						if ln == line {
							scenarios = append(scenarios, def)
							break
						}
					}
					ft.ScenarioDefinitions = scenarios
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

	var scenarios []interface{}
	for _, scenario := range ft.ScenarioDefinitions {
		if s.matchesTags(allTags(ft, scenario)) {
			scenarios = append(scenarios, scenario)
		}
	}
	ft.ScenarioDefinitions = scenarios
}

func allTags(nodes ...interface{}) []string {
	var tags, tmp []string
	for _, node := range nodes {
		var gr []*gherkin.Tag
		switch t := node.(type) {
		case *gherkin.Feature:
			gr = t.Tags
		case *gherkin.ScenarioOutline:
			gr = t.Tags
		case *gherkin.Scenario:
			gr = t.Tags
		case *gherkin.Examples:
			gr = t.Tags
		}

		for _, gtag := range gr {
			tag := strings.TrimSpace(gtag.Name)
			if tag[0] == '@' {
				tag = tag[1:]
			}
			copy(tmp, tags)
			var found bool
			for _, tg := range tmp {
				if tg == tag {
					found = true
					break
				}
			}
			if !found {
				tags = append(tags, tag)
			}
		}
	}
	return tags
}

func hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// based on http://behat.readthedocs.org/en/v2.5/guides/6.cli.html#gherkin-filters
func (s *suite) matchesTags(tags []string) (ok bool) {
	ok = true
	for _, andTags := range strings.Split(s.tags, "&&") {
		var okComma bool
		for _, tag := range strings.Split(andTags, ",") {
			tag = strings.Replace(strings.TrimSpace(tag), "@", "", -1)
			if tag[0] == '~' {
				tag = tag[1:]
				okComma = !hasTag(tags, tag) || okComma
			} else {
				okComma = hasTag(tags, tag) || okComma
			}
		}
		ok = (false != okComma && ok && okComma) || false
	}
	return
}
