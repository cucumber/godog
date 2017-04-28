package godog

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/godog/colors"
	"github.com/DATA-DOG/godog/gherkin"
)

var sampleGherkinFeature = `
Feature: junit formatter

  Background:
    Given passing

  Scenario: passing scenario
    Then passing

  Scenario: failing scenario
    When failing
    Then passing

  Scenario: pending scenario
    When pending
    Then passing

  Scenario: undefined scenario
    When undefined
    Then next undefined

  Scenario Outline: outline
    Given <one>
    When <two>

    Examples:
      | one     | two     |
      | passing | passing |
      | passing | failing |
      | passing | pending |

	Examples:
      | one     | two       |
      | passing | undefined |
`

func TestJUnitFormatterOutput(t *testing.T) {
	feat, err := gherkin.ParseFeature(strings.NewReader(sampleGherkinFeature))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	s := &Suite{
		fmt: junitFunc("junit", w),
		features: []*feature{&feature{
			Path:    "any.feature",
			Feature: feat,
			Content: []byte(sampleGherkinFeature),
		}},
	}

	s.Step(`^passing$`, func() error { return nil })
	s.Step(`^failing$`, func() error { return fmt.Errorf("errored") })
	s.Step(`^pending$`, func() error { return ErrPending })

	var zeroDuration time.Duration
	expected := junitPackageSuite{
		Name:     "junit",
		Tests:    8,
		Skipped:  0,
		Failures: 2,
		Errors:   4,
		Time:     zeroDuration.String(),
		TestSuites: []*junitTestSuite{{
			Name:     "junit formatter",
			Tests:    8,
			Skipped:  0,
			Failures: 2,
			Errors:   4,
			Time:     zeroDuration.String(),
			TestCases: []*junitTestCase{
				{
					Name:   "passing scenario",
					Status: "passed",
					Time:   zeroDuration.String(),
				},
				{
					Name:   "failing scenario",
					Status: "failed",
					Time:   zeroDuration.String(),
					Failure: &junitFailure{
						Message: "Step failing: errored",
					},
					Error: []*junitError{
						{Message: "Step passing", Type: "skipped"},
					},
				},
				{
					Name:   "pending scenario",
					Status: "pending",
					Time:   zeroDuration.String(),
					Error: []*junitError{
						{Message: "Step pending: TODO: write pending definition", Type: "pending"},
						{Message: "Step passing", Type: "skipped"},
					},
				},
				{
					Name:   "undefined scenario",
					Status: "undefined",
					Time:   zeroDuration.String(),
					Error: []*junitError{
						{Message: "Step undefined", Type: "undefined"},
						{Message: "Step next undefined", Type: "undefined"},
					},
				},
				{
					Name:   "outline #1",
					Status: "passed",
					Time:   zeroDuration.String(),
				},
				{
					Name:   "outline #2",
					Status: "failed",
					Time:   zeroDuration.String(),
					Failure: &junitFailure{
						Message: "Step failing: errored",
					},
				},
				{
					Name:   "outline #3",
					Status: "pending",
					Time:   zeroDuration.String(),
					Error: []*junitError{
						{Message: "Step pending: TODO: write pending definition", Type: "pending"},
					},
				},
				{
					Name:   "outline #4",
					Status: "undefined",
					Time:   zeroDuration.String(),
					Error: []*junitError{
						{Message: "Step undefined", Type: "undefined"},
					},
				},
			},
		}},
	}

	s.run()
	s.fmt.Summary()

	var exp bytes.Buffer
	io.WriteString(&exp, xml.Header)

	enc := xml.NewEncoder(&exp)
	enc.Indent("", "  ")
	if err := enc.Encode(expected); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf.String() != exp.String() {
		t.Fatalf("expected output does not match: %s", buf.String())
	}
}
