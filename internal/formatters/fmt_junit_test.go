package formatters_test

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
)

func Test_JUnitFormatter_StopOnFirstFailure(t *testing.T) {
	featureFile := "formatter-tests/features/stop_on_first_failure.feature"

	// First, verify the normal output (without StopOnFirstFailure)
	var normalBuf bytes.Buffer
	normalOpts := godog.Options{
		Format: "junit",
		Paths:  []string{featureFile},
		Output: &normalBuf,
		Strict: true,
	}

	normalSuite := godog.TestSuite{
		Name: "Normal Run",
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			setupStopOnFailureSteps(sc)
		},
		Options: &normalOpts,
	}
	if status := normalSuite.Run(); status != 1 {
		t.Fatalf("Expected suite to have status 1, but got %d", status)
	}

	// Parse the XML output
	var normalResult JunitPackageSuite
	err := xml.Unmarshal(normalBuf.Bytes(), &normalResult)
	if err != nil {
		t.Fatalf("Failed to parse XML output: %v", err)
	}

	// Now run with StopOnFirstFailure
	var stopBuf bytes.Buffer
	stopOpts := godog.Options{
		Format:        "junit",
		Paths:         []string{featureFile},
		Output:        &stopBuf,
		Strict:        true,
		StopOnFailure: true,
	}

	stopSuite := godog.TestSuite{
		Name: "Stop On First Failure",
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			setupStopOnFailureSteps(sc)
		},
		Options: &stopOpts,
	}
	if status := stopSuite.Run(); status != 1 {
		t.Fatalf("Expected suite to have status 1, but got %d", status)
	}

	// Parse the XML output
	var stopResult JunitPackageSuite
	err = xml.Unmarshal(stopBuf.Bytes(), &stopResult)
	if err != nil {
		t.Fatalf("Failed to parse XML output: %v", err)
	}

	// Verify the second test case is marked as skipped when StopOnFirstFailure is enabled
	if len(stopResult.TestSuites) == 0 || len(stopResult.TestSuites[0].TestCases) < 2 {
		t.Fatal("Expected at least 2 test cases in the results")
	}

	// In a normal run, second test case should not be skipped
	if normalResult.TestSuites[0].TestCases[1].Status == "skipped" {
		t.Errorf("In normal run, second test case should not be skipped")
	}

	// In stop on failure run, second test case should be skipped
	if stopResult.TestSuites[0].TestCases[1].Status != "skipped" {
		t.Errorf("In stop on failure run, second test case should be skipped, but got %s",
			stopResult.TestSuites[0].TestCases[1].Status)
	}
}

// setupStopOnFailureSteps registers the step definitions for the stop-on-failure test
func setupStopOnFailureSteps(sc *godog.ScenarioContext) {
	sc.Step(`^a passing step$`, func() error {
		return nil
	})
	sc.Step(`^a failing step$`, func() error {
		return fmt.Errorf("step failed")
	})
}

// JunitPackageSuite represents the JUnit XML structure for test suites
type JunitPackageSuite struct {
	XMLName    xml.Name          `xml:"testsuites"`
	Name       string            `xml:"name,attr"`
	Tests      int               `xml:"tests,attr"`
	Skipped    int               `xml:"skipped,attr"`
	Failures   int               `xml:"failures,attr"`
	Errors     int               `xml:"errors,attr"`
	Time       string            `xml:"time,attr"`
	TestSuites []*JunitTestSuite `xml:"testsuite"`
}

type JunitTestSuite struct {
	XMLName   xml.Name         `xml:"testsuite"`
	Name      string           `xml:"name,attr"`
	Tests     int              `xml:"tests,attr"`
	Skipped   int              `xml:"skipped,attr"`
	Failures  int              `xml:"failures,attr"`
	Errors    int              `xml:"errors,attr"`
	Time      string           `xml:"time,attr"`
	TestCases []*JunitTestCase `xml:"testcase"`
}

type JunitTestCase struct {
	XMLName xml.Name      `xml:"testcase"`
	Name    string        `xml:"name,attr"`
	Status  string        `xml:"status,attr"`
	Time    string        `xml:"time,attr"`
	Failure *JunitFailure `xml:"failure,omitempty"`
	Error   []*JunitError `xml:"error,omitempty"`
}

type JunitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr,omitempty"`
}

type JunitError struct {
	XMLName xml.Name `xml:"error,omitempty"`
	Message string   `xml:"message,attr"`
	Type    string   `xml:"type,attr"`
}
