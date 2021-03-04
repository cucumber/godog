package godog

import (
	"context"

	"github.com/cucumber/messages-go/v10"
)

// Scenario represents the executed scenario
type Scenario = messages.Pickle

// Step represents the executed step
type Step = messages.Pickle_PickleStep

// Steps allows to nest steps
// instead of returning an error in step func
// it is possible to return combined steps:
//
//   func multistep(name string) godog.Steps {
//     return godog.Steps{
//       fmt.Sprintf(`an user named "%s"`, name),
//       fmt.Sprintf(`user "%s" is authenticated`, name),
//     }
//   }
//
// These steps will be matched and executed in
// sequential order. The first one which fails
// will result in main step failure.
type Steps []string

// DocString represents the DocString argument made to a step definition
type DocString = messages.PickleStepArgument_PickleDocString

// Table represents the Table argument made to a step definition
type Table = messages.PickleStepArgument_PickleTable

// TestSuiteContext allows various contexts
// to register event handlers.
//
// When running a test suite, the instance of TestSuiteContext
// is passed to all functions (contexts), which
// have it as a first and only argument.
//
// Note that all event hooks does not catch panic errors
// in order to have a trace information
type TestSuiteContext struct{}

// BeforeSuite registers a hook
// to be run once before the suite runner.
//
// Use it to prepare the test suite for a spin.
// Connect and prepare database for instance...
func (ctx *TestSuiteContext) BeforeSuite(fn BeforeSuiteHook)

// BeforeSuiteHook ...
type BeforeSuiteHook func(ctx context.Context) context.Context

// AfterSuite registers a hook
// to be run once after the suite runner
func (ctx *TestSuiteContext) AfterSuite(fn AfterSuiteHook)

// AfterSuiteHook ...
type AfterSuiteHook func(ctx context.Context) context.Context

// ScenarioContext allows various contexts
// to register steps and event handlers.
//
// When running a scenario, the instance of ScenarioContext
// is passed to all functions (contexts), which
// have it as a first and only argument.
//
// Note that all event hooks does not catch panic errors
// in order to have a trace information. Only step
// executions are catching panic error since it may
// be a context specific error.
type ScenarioContext struct{}

// BeforeScenario registers a function or method
// to be run before every scenario.
//
// It is a good practice to restore the default state
// before every scenario so it would be isolated from
// any kind of state.
func (ctx *ScenarioContext) BeforeScenario(fn BeforeScenarioHook)

// BeforeScenarioHook ...
type BeforeScenarioHook func(ctx context.Context, sc Scenario) context.Context

// AfterScenario registers an function or method
// to be run after every scenario.
func (ctx *ScenarioContext) AfterScenario(fn AfterScenarioHook)

// AfterScenarioHook ...
type AfterScenarioHook func(ctx context.Context, sc Scenario, err error) (context.Context, error)

// BeforeStep registers a function or method
// to be run before every step.
func (ctx *ScenarioContext) BeforeStep(fn BeforeStepHook)

// BeforeStepHook ...
type BeforeStepHook func(ctx context.Context, st Step) context.Context

// func stepFuncExample(ctx context.Context, arg1 string) (context.Context, error)

// AfterStep registers an function or method
// to be run after every step.
//
// It may be convenient to return a different kind of error
// in order to print more state details which may help
// in case of step failure
//
// In some cases, for example when running a headless
// browser, to take a screenshot after failure.
func (ctx *ScenarioContext) AfterStep(fn AfterStepHook)

// AfterStepHook ...
type AfterStepHook func(ctx context.Context, st Step, err error) (context.Context, error)

// Step allows to register a *StepDefinition in the
// Godog feature suite, the definition will be applied
// to all steps matching the given Regexp expr.
//
// It will panic if expr is not a valid regular
// expression or stepFunc is not a valid step
// handler.
//
// The expression can be of type: *regexp.Regexp, string or []byte
//
// The stepFunc may accept one or several arguments of type:
// - context.Context
// - int, int8, int16, int32, int64
// - float32, float64
// - string
// - []byte
// - *godog.DocString
// - *godog.Table
//
// The stepFunc need to return either an error or []string for multistep
// and can optionally return context.Context.
//
// Example:
// func stepFunc(ctx context.Context, arg1 int) (context.Context, error)
//
// Note that if there are two definitions which may match
// the same step, then only the first matched handler
// will be applied.
//
// If none of the *StepDefinition is matched, then
// ErrUndefined error will be returned when
// running steps.
func (ctx *ScenarioContext) Step(expr, stepFunc interface{})

// Build creates a test package like go test command at given target path.
// If there are no go files in tested directory, then
// it simply builds a godog executable to scan features.
//
// If there are go test files, it first builds a test
// package with standard go test command.
//
// Finally it generates godog suite executable which
// registers exported godog contexts from the test files
// of tested package.
//
// Returns the path to generated executable
func Build(bin string) error
