package godog

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/DATA-DOG/godog/gherkin"
)

type firedEvent struct {
	name string
	args []interface{}
}

type suiteFeature struct {
	suite
	events []*firedEvent
}

func (s *suiteFeature) HandleBeforeScenario(scenario *gherkin.Scenario) {
	// reset feature paths
	cfg.paths = []string{}
	// reset event stack
	s.events = []*firedEvent{}
	// reset formatter, which collects all details
	s.fmt = &testFormatter{}
}

func (s *suiteFeature) iAmListeningToSuiteEvents(args ...*Arg) error {
	s.BeforeScenario(BeforeScenarioHandlerFunc(func(scenario *gherkin.Scenario) {
		s.events = append(s.events, &firedEvent{"BeforeScenario", []interface{}{
			scenario,
		}})
	}))
	return nil
}

func (s *suiteFeature) featurePath(args ...*Arg) error {
	cfg.paths = append(cfg.paths, args[0].String())
	return nil
}

func (s *suiteFeature) parseFeatures(args ...*Arg) (err error) {
	s.features, err = cfg.features()
	return
}

func (s *suiteFeature) iShouldHaveNumFeatureFiles(args ...*Arg) error {
	if len(s.features) != args[0].Int() {
		return fmt.Errorf("expected %d features to be parsed, but have %d", args[0].Int(), len(s.features))
	}
	expected := args[1].PyString().Lines
	var actual []string
	for _, ft := range s.features {
		actual = append(actual, ft.Path)
	}
	if len(expected) != len(actual) {
		return fmt.Errorf("expected %d feature paths to be parsed, but have %d", len(expected), len(actual))
	}
	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] {
			return fmt.Errorf(`expected feature path "%s" at position: %d, does not match actual "%s"`, expected[i], i, actual[i])
		}
	}
	return nil
}

func (s *suiteFeature) iRunFeatureSuite(args ...*Arg) error {
	if err := s.parseFeatures(); err != nil {
		return err
	}
	s.run()
	return nil
}

func (s *suiteFeature) numScenariosRegistered(args ...*Arg) (err error) {
	var num int
	for _, ft := range s.features {
		num += len(ft.Scenarios)
	}
	if num != args[0].Int() {
		err = fmt.Errorf("expected %d scenarios to be registered, but got %d", args[0].Int(), num)
	}
	return
}

func (s *suiteFeature) thereWereNumEventsFired(args ...*Arg) error {
	var num int
	for _, event := range s.events {
		if event.name == args[2].String() {
			num++
		}
	}
	if num != args[1].Int() {
		return fmt.Errorf("expected %d %s events to be fired, but got %d", args[1].Int(), args[2].String(), num)
	}
	return nil
}

func (s *suiteFeature) thereWasEventTriggeredBeforeScenario(args ...*Arg) error {
	var found []string
	for _, event := range s.events {
		if event.name != "BeforeScenario" {
			continue
		}

		scenario := event.args[0].(*gherkin.Scenario)
		if scenario.Title == args[0].String() {
			return nil
		}

		found = append(found, scenario.Title)
	}

	if len(found) == 0 {
		return fmt.Errorf("before scenario event was never triggered or listened")
	}

	return fmt.Errorf(`expected "%s" scenario, but got these fired %s`, args[0].String(), `"`+strings.Join(found, `", "`)+`"`)
}

func SuiteContext(g Suite) {
	s := &suiteFeature{
		suite: suite{},
	}

	g.BeforeScenario(s)

	g.Step(
		regexp.MustCompile(`^a feature path "([^"]*)"$`),
		StepHandlerFunc(s.featurePath))
	g.Step(
		regexp.MustCompile(`^I parse features$`),
		StepHandlerFunc(s.parseFeatures))
	g.Step(
		regexp.MustCompile(`^I should have ([\d]+) features? files?:$`),
		StepHandlerFunc(s.iShouldHaveNumFeatureFiles))
	g.Step(
		regexp.MustCompile(`^I should have ([\d]+) scenarios? registered$`),
		StepHandlerFunc(s.numScenariosRegistered))
	g.Step(
		regexp.MustCompile(`^I'm listening to suite events$`),
		StepHandlerFunc(s.iAmListeningToSuiteEvents))
	g.Step(
		regexp.MustCompile(`^I run feature suite$`),
		StepHandlerFunc(s.iRunFeatureSuite))
	g.Step(
		regexp.MustCompile(`^there (was|were) ([\d]+) "([^"]*)" events? fired$`),
		StepHandlerFunc(s.thereWereNumEventsFired))
	g.Step(
		regexp.MustCompile(`^there was event triggered before scenario "([^"]*)"$`),
		StepHandlerFunc(s.thereWasEventTriggeredBeforeScenario))
}
