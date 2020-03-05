package godog

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/cucumber/gherkin-go/v9"
	"github.com/cucumber/messages-go/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cucumber/godog/colors"
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
	const path = "any.feature"

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(sampleGherkinFeature), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)

	pickles := gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	s := &Suite{
		fmt: junitFunc("junit", w),
		features: []*feature{{
			GherkinDocument: gd,
			pickles:         pickles,
			Path:            path,
			Content:         []byte(sampleGherkinFeature),
		}},
	}

	s.Step(`^passing$`, func() error { return nil })
	s.Step(`^failing$`, func() error { return fmt.Errorf("errored") })
	s.Step(`^pending$`, func() error { return ErrPending })

	const zeroDuration = "0"
	expected := junitPackageSuite{
		Name:     "junit",
		Tests:    8,
		Skipped:  0,
		Failures: 2,
		Errors:   4,
		Time:     zeroDuration,
		TestSuites: []*junitTestSuite{{
			Name:     "junit formatter",
			Tests:    8,
			Skipped:  0,
			Failures: 2,
			Errors:   4,
			Time:     zeroDuration,
			TestCases: []*junitTestCase{
				{
					Name:   "passing scenario",
					Status: "passed",
					Time:   zeroDuration,
				},
				{
					Name:   "failing scenario",
					Status: "failed",
					Time:   zeroDuration,
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
					Time:   zeroDuration,
					Error: []*junitError{
						{Message: "Step pending: TODO: write pending definition", Type: "pending"},
						{Message: "Step passing", Type: "skipped"},
					},
				},
				{
					Name:   "undefined scenario",
					Status: "undefined",
					Time:   zeroDuration,
					Error: []*junitError{
						{Message: "Step undefined", Type: "undefined"},
						{Message: "Step next undefined", Type: "undefined"},
					},
				},
				{
					Name:   "outline #1",
					Status: "passed",
					Time:   zeroDuration,
				},
				{
					Name:   "outline #2",
					Status: "failed",
					Time:   zeroDuration,
					Failure: &junitFailure{
						Message: "Step failing: errored",
					},
				},
				{
					Name:   "outline #3",
					Status: "pending",
					Time:   zeroDuration,
					Error: []*junitError{
						{Message: "Step pending: TODO: write pending definition", Type: "pending"},
					},
				},
				{
					Name:   "outline #4",
					Status: "undefined",
					Time:   zeroDuration,
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
	_, err = io.WriteString(&exp, xml.Header)
	require.NoError(t, err)

	enc := xml.NewEncoder(&exp)
	enc.Indent("", "  ")
	err = enc.Encode(expected)
	require.NoError(t, err)

	assert.Equal(t, exp.String(), buf.String())
}
