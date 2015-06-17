package godog

import (
	"fmt"
	"regexp"

	"github.com/DATA-DOG/godog/gherkin"
)

type suiteFeature struct {
	suite
	// for hook tests
	befScenarioHook *gherkin.Scenario
}

func (s *suiteFeature) BeforeScenario(scenario *gherkin.Scenario) {
	// reset feature paths
	cfg.paths = []string{}
	// reset hook test references
	s.befScenarioHook = nil
	// reset formatter, which collects all details
	s.fmt = &testFormatter{}
}

func (s *suiteFeature) iHaveBeforeScenarioHook(args ...*Arg) error {
	s.suite.BeforeScenario(BeforeScenarioHandlerFunc(func(scenario *gherkin.Scenario) {
		s.befScenarioHook = scenario
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

func (s *suiteFeature) numParsed(args ...*Arg) (err error) {
	if len(s.features) != args[0].Int() {
		err = fmt.Errorf("expected %d features to be parsed, but have %d", args[0].Int(), len(s.features))
	}
	return
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
		regexp.MustCompile(`^I should have ([\d]+) features? files?$`),
		StepHandlerFunc(s.numParsed))
	g.Step(
		regexp.MustCompile(`^I should have ([\d]+) scenarios? registered$`),
		StepHandlerFunc(s.numScenariosRegistered))
	g.Step(
		regexp.MustCompile(`^I have a before scenario hook$`),
		StepHandlerFunc(s.iHaveBeforeScenarioHook))
}
