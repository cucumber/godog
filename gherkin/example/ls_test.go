package main

import (
	"regexp"

	"github.com/DATA-DOG/godog"
)

type lsFeature struct{}

func (s *lsFeature) inDirectory(args ...godog.Arg) error {
	return nil
}

func (s *lsFeature) haveFile(args ...godog.Arg) error {
	return nil
}

func SuiteContext(g godog.Suite) {
	f := &lsFeature{}

	g.Step(
		regexp.MustCompile(`^I am in a directory "([^"]*)"$`),
		godog.StepHandlerFunc(f.inDirectory))
	g.Step(
		regexp.MustCompile(`^I have a file named "([^"]*)"$`),
		godog.StepHandlerFunc(f.haveFile))
}
