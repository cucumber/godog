package godog

import (
	"fmt"
	"regexp"
)

type suiteFeature struct {
	suite
}

func (s *suiteFeature) featurePath(args ...Arg) error {
	cfg.featuresPath = args[0].String()
	return nil
}

func (s *suiteFeature) parseFeatures(args ...Arg) (err error) {
	s.features, err = cfg.features()
	return
}

func (s *suiteFeature) numParsed(args ...Arg) (err error) {
	if len(s.features) != int(args[0].Int()) {
		err = fmt.Errorf("expected %d features to be parsed, but have %d", args[0].Int(), len(s.features))
	}
	return
}

func SuiteContext(g Suite) {
	s := &suiteFeature{
		suite: suite{},
	}

	g.Step(
		regexp.MustCompile(`^a feature path "([^"]*)"$`),
		StepHandlerFunc(s.featurePath))
	g.Step(
		regexp.MustCompile(`^I parse features$`),
		StepHandlerFunc(s.parseFeatures))
	g.Step(
		regexp.MustCompile(`^I should have ([\d]+) features? files?$`),
		StepHandlerFunc(s.numParsed))
}
