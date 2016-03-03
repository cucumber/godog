package godog

import (
	"encoding/xml"
	"fmt"
	"os"
	"time"

	"gopkg.in/cucumber/gherkin-go.v3"
)

func init() {
	Format("junit", "Prints out in junit compatible xml format", &junitFormatter{
		suites: make([]*junitTestSuites, 0),
	})
}

type junitFormatter struct {
	suites []*junitTestSuites
}

func (j *junitFormatter) Feature(feature *gherkin.Feature, path string) {
	testSuites := &junitTestSuites{
		Name:       feature.Name,
		TestSuites: make([]*junitTestSuite, 0),
	}

	j.suites = append(j.suites, testSuites)
}

func (j *junitFormatter) Node(node interface{}) {
	testSuite := &junitTestSuite{
		TestCases: make([]*TestCase, 0),
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

	currentSuites := j.currentSuites()
	currentSuites.TestSuites = append(currentSuites.TestSuites, testSuite)
}

func (j *junitFormatter) Failed(step *gherkin.Step, match *StepDef, err error) {
	testCase := &TestCase{
		Name: step.Text,
	}

	testCase.Failure = &junitFailure{
		Contents: err.Error(),
	}

	currentSuites := j.currentSuites()
	currentSuites.Failures++
	currentSuite := currentSuites.currentSuite()
	currentSuite.Failures++
	currentSuite.TestCases = append(currentSuite.TestCases, testCase)
}

func (j *junitFormatter) Passed(step *gherkin.Step, match *StepDef) {
	testCase := &TestCase{
		Name: step.Text,
	}

	currentSuites := j.currentSuites()
	currentSuites.Tests++
	currentSuite := currentSuites.currentSuite()
	currentSuite.Tests++
	currentSuite.TestCases = append(currentSuite.TestCases, testCase)
}

func (j *junitFormatter) Skipped(step *gherkin.Step) {
	testCase := &TestCase{
		Name: step.Text,
	}

	currentSuites := j.currentSuites()
	currentSuite := currentSuites.currentSuite()
	currentSuite.Skipped++
	currentSuite.TestCases = append(currentSuite.TestCases, testCase)
}

func (j *junitFormatter) Undefined(step *gherkin.Step) {
	testCase := &TestCase{
		Name: step.Text,
	}

	currentSuites := j.currentSuites()
	currentSuites.Disabled++
	currentSuite := currentSuites.currentSuite()
	currentSuite.Disabled++
	currentSuite.TestCases = append(currentSuite.TestCases, testCase)
}

func (j *junitFormatter) Pending(step *gherkin.Step, match *StepDef) {
	testCase := &TestCase{
		Name: step.Text,
	}

	testCase.Skipped = &junitSkipped{
		Contents: step.Text,
	}

	currentSuites := j.currentSuites()
	currentSuite := currentSuites.currentSuite()
	currentSuite.Skipped++
	currentSuite.TestCases = append(currentSuite.TestCases, testCase)
}

func (j *junitFormatter) Summary() {
	enc := xml.NewEncoder(os.Stdout)
	enc.Indent("  ", "    ")
	if err := enc.Encode(j.suites); err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

type junitFailure struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

type junitError struct {
	Message  string `xml:"message,attr"`
	Type     string `xml:"type,attr"`
	Contents string `xml:",chardata"`
}

type junitProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

type junitSkipped struct {
	Contents string `xml:",chardata"`
}

type SystemErr struct {
	Contents string `xml:",chardata"`
}

type SystemOut struct {
	Contents string `xml:",chardata"`
}

type TestCase struct {
	XMLName    xml.Name      `xml:"testcase"`
	Name       string        `xml:"name,attr"`
	Classname  string        `xml:"classname,attr"`
	Assertions string        `xml:"assertions,attr"`
	Status     string        `xml:"status,attr"`
	Time       string        `xml:"time,attr"`
	Skipped    *junitSkipped `xml:"skipped,omitempty"`
	Failure    *junitFailure `xml:"failure,omitempty"`
	Error      *junitError   `xml:"error,omitempty"`
	SystemOut  *SystemOut    `xml:"system-out,omitempty"`
	SystemErr  *SystemErr    `xml:"system-err,omitempty"`
}

type junitTestSuite struct {
	XMLName    xml.Name         `xml:"testsuite"`
	Name       string           `xml:"name,attr"`
	Tests      int              `xml:"tests,attr"`
	Failures   int              `xml:"failures,attr"`
	Errors     int              `xml:"errors,attr"`
	Disabled   int              `xml:"disabled,attr"`
	Skipped    int              `xml:"skipped,attr"`
	Time       string           `xml:"time,attr"`
	Hostname   string           `xml:"hostname,attr"`
	ID         string           `xml:"id,attr"`
	Package    string           `xml:"package,attr"`
	Timestamp  time.Time        `xml:"timestamp,attr"`
	SystemOut  *SystemOut       `xml:"system-out,omitempty"`
	SystemErr  *SystemErr       `xml:"system-err,omitempty"`
	Properties []*junitProperty `xml:"properties>property,omitempty"`
	TestCases  []*TestCase
}

func (ts *junitTestSuite) currentCase() *TestCase {
	return ts.TestCases[len(ts.TestCases)-1]
}

type junitTestSuites struct {
	XMLName    xml.Name `xml:"testsuites"`
	Name       string   `xml:"name,attr"`
	Tests      int      `xml:"tests,attr"`
	Failures   int      `xml:"failures,attr"`
	Errors     int      `xml:"errors,attr"`
	Disabled   int      `xml:"disabled,attr"`
	Time       string   `xml:"time,attr"`
	TestSuites []*junitTestSuite
}

func (ts *junitTestSuites) currentSuite() *junitTestSuite {
	return ts.TestSuites[len(ts.TestSuites)-1]
}

func (j *junitFormatter) currentSuites() *junitTestSuites {
	return j.suites[len(j.suites)-1]
}
