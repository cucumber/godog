package godog

import "github.com/DATA-DOG/godog/gherkin"

// BeforeSuiteHandler can be registered
// in Suite to be executed once before
// running a feature suite
type BeforeSuiteHandler interface {
	HandleBeforeSuite()
}

// BeforeSuiteHandlerFunc is a function implementing
// BeforeSuiteHandler interface
type BeforeSuiteHandlerFunc func()

// HandleBeforeSuite is called once before suite
func (f BeforeSuiteHandlerFunc) HandleBeforeSuite() {
	f()
}

// BeforeScenarioHandler can be registered
// in Suite to be executed before every scenario
// which will be run
type BeforeScenarioHandler interface {
	HandleBeforeScenario(scenario *gherkin.Scenario)
}

// BeforeScenarioHandlerFunc is a function implementing
// BeforeScenarioHandler interface
type BeforeScenarioHandlerFunc func(scenario *gherkin.Scenario)

// HandleBeforeScenario is called with a *gherkin.Scenario argument
// for before every scenario which is run by suite
func (f BeforeScenarioHandlerFunc) HandleBeforeScenario(scenario *gherkin.Scenario) {
	f(scenario)
}

// BeforeStepHandler can be registered
// in Suite to be executed before every step
// which will be run
type BeforeStepHandler interface {
	HandleBeforeStep(step *gherkin.Step)
}

// BeforeStepHandlerFunc is a function implementing
// BeforeStepHandler interface
type BeforeStepHandlerFunc func(step *gherkin.Step)

// HandleBeforeStep is called with a *gherkin.Step argument
// for before every step which is run by suite
func (f BeforeStepHandlerFunc) HandleBeforeStep(step *gherkin.Step) {
	f(step)
}

// AfterStepHandler can be registered
// in Suite to be executed after every step
// which will be run
type AfterStepHandler interface {
	HandleAfterStep(step *gherkin.Step, err error)
}

// AfterStepHandlerFunc is a function implementing
// AfterStepHandler interface
type AfterStepHandlerFunc func(step *gherkin.Step, err error)

// HandleAfterStep is called with a *gherkin.Step argument
// for after every step which is run by suite
func (f AfterStepHandlerFunc) HandleAfterStep(step *gherkin.Step, err error) {
	f(step, err)
}

// AfterScenarioHandler can be registered
// in Suite to be executed after every scenario
// which will be run
type AfterScenarioHandler interface {
	HandleAfterScenario(scenario *gherkin.Scenario, err error)
}

// AfterScenarioHandlerFunc is a function implementing
// AfterScenarioHandler interface
type AfterScenarioHandlerFunc func(scenario *gherkin.Scenario, err error)

// HandleAfterScenario is called with a *gherkin.Scenario argument
// for after every scenario which is run by suite
func (f AfterScenarioHandlerFunc) HandleAfterScenario(scenario *gherkin.Scenario, err error) {
	f(scenario, err)
}

// AfterSuiteHandler can be registered
// in Suite to be executed once after
// running a feature suite
type AfterSuiteHandler interface {
	HandleAfterSuite()
}

// AfterSuiteHandlerFunc is a function implementing
// AfterSuiteHandler interface
type AfterSuiteHandlerFunc func()

// HandleAfterSuite is called once after suite
func (f AfterSuiteHandlerFunc) HandleAfterSuite() {
	f()
}
