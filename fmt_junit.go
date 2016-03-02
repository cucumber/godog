package godog

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/adrianduke/godog/junit"
	gherkin "github.com/cucumber/gherkin-go"
)

const JUnitResultsEnv = "JUNIT_RESULTS"

func init() {
	Format("junit", "Prints out in junit compatible xml format", &junitFormatter{
		JUnit: make(junit.JUnit, 0),
	})
}

type junitFormatter struct {
	junit.JUnit
}

func (j *junitFormatter) Feature(feature *gherkin.Feature, path string) {
	testSuites := &junit.TestSuites{
		Name:       feature.Name,
		TestSuites: make([]*junit.TestSuite, 0),
	}

	j.JUnit = append(j.JUnit, testSuites)
}

func (j *junitFormatter) Node(node interface{}) {
	testSuite := &junit.TestSuite{
		TestCases: make([]*junit.TestCase, 0),
		Timestamp: time.Now(),
	}

	switch t := node.(type) {
	case *gherkin.ScenarioOutline:
		testSuite.Name = t.Name
	case *gherkin.Scenario:
		testSuite.Name = t.Name
	case *gherkin.Background:
		testSuite.Name = "Background"
	}

	currentSuites := j.JUnit.CurrentSuites()
	currentSuites.TestSuites = append(currentSuites.TestSuites, testSuite)
}

func (j *junitFormatter) Failed(step *gherkin.Step, match *StepDef, err error) {
	testCase := &junit.TestCase{
		Name: step.Text,
	}

	testCase.Failure = &junit.Failure{
		Contents: err.Error(),
	}

	currentSuites := j.JUnit.CurrentSuites()
	currentSuites.Failures++
	currentSuite := currentSuites.CurrentSuite()
	currentSuite.Failures++
	currentSuite.TestCases = append(currentSuite.TestCases, testCase)
}

func (j *junitFormatter) Passed(step *gherkin.Step, match *StepDef) {
	testCase := &junit.TestCase{
		Name: step.Text,
	}

	currentSuites := j.JUnit.CurrentSuites()
	currentSuites.Tests++
	currentSuite := currentSuites.CurrentSuite()
	currentSuite.Tests++
	currentSuite.TestCases = append(currentSuite.TestCases, testCase)
}

func (j *junitFormatter) Skipped(step *gherkin.Step) {
	testCase := &junit.TestCase{
		Name: step.Text,
	}

	currentSuites := j.JUnit.CurrentSuites()
	currentSuite := currentSuites.CurrentSuite()
	currentSuite.Skipped++
	currentSuite.TestCases = append(currentSuite.TestCases, testCase)
}

func (j *junitFormatter) Undefined(step *gherkin.Step) {
	testCase := &junit.TestCase{
		Name: step.Text,
	}

	currentSuites := j.JUnit.CurrentSuites()
	currentSuites.Disabled++
	currentSuite := currentSuites.CurrentSuite()
	currentSuite.Disabled++
	currentSuite.TestCases = append(currentSuite.TestCases, testCase)
}

func (j *junitFormatter) Pending(step *gherkin.Step, match *StepDef) {
	testCase := &junit.TestCase{
		Name: step.Text,
	}

	testCase.Skipped = junit.Skipped{
		Contents: step.Text,
	}

	currentSuites := j.JUnit.CurrentSuites()
	currentSuite := currentSuites.CurrentSuite()
	currentSuite.Skipped++
	currentSuite.TestCases = append(currentSuite.TestCases, testCase)
}

func (j *junitFormatter) Summary() {
	var writer io.Writer
	if outputFilePath := os.Getenv(JUnitResultsEnv); outputFilePath != "" {
		outputFile, err := os.Create(outputFilePath)
		if err != nil {
			panic(err.Error())
		}
		defer outputFile.Close()

		writer = io.Writer(outputFile)
	} else {
		writer = os.Stdout
	}

	enc := xml.NewEncoder(writer)
	enc.Indent("  ", "    ")
	if err := enc.Encode(j.JUnit); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}
