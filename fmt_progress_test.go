package godog

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cucumber/gherkin-go/v11"
	"github.com/cucumber/messages-go/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cucumber/godog/colors"
)

var basicGherkinFeature = `
Feature: basic

  Scenario: passing scenario
	When one
	Then two
`

func TestProgressFormatterWhenStepPanics(t *testing.T) {
	const path = "any.feature"

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(basicGherkinFeature), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)

	pickles := gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	r := runner{
		fmt:      progressFunc("progress", w),
		features: []*feature{{GherkinDocument: gd, pickles: pickles}},
		initializer: func(s *Suite) {
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^two$`, func() error { panic("omg") })
		},
	}

	failed := r.concurrent(1, func() Formatter { return progressFunc("progress", w) })
	require.True(t, failed)

	actual := buf.String()
	assert.Contains(t, actual, "godog/fmt_progress_test.go:39")
}

func TestProgressFormatterWithPanicInMultistep(t *testing.T) {
	const path = "any.feature"

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(basicGherkinFeature), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)

	pickles := gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)
	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	r := runner{
		fmt:      progressFunc("progress", w),
		features: []*feature{{GherkinDocument: gd, pickles: pickles}},
		initializer: func(s *Suite) {
			s.Step(`^sub1$`, func() error { return nil })
			s.Step(`^sub-sub$`, func() error { return nil })
			s.Step(`^sub2$`, func() []string { return []string{"sub-sub", "sub1", "one"} })
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^two$`, func() []string { return []string{"sub1", "sub2"} })
		},
	}

	failed := r.concurrent(1, func() Formatter { return progressFunc("progress", w) })
	require.True(t, failed)
}

func TestProgressFormatterMultistepTemplates(t *testing.T) {
	const path = "any.feature"

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(basicGherkinFeature), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)

	pickles := gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	r := runner{
		fmt:      progressFunc("progress", w),
		features: []*feature{{GherkinDocument: gd, pickles: pickles}},
		initializer: func(s *Suite) {
			s.Step(`^sub-sub$`, func() error { return nil })
			s.Step(`^substep$`, func() Steps { return Steps{"sub-sub", `unavailable "John" cost 5`, "one", "three"} })
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^(t)wo$`, func(s string) Steps { return Steps{"undef", "substep"} })
		},
	}

	failed := r.concurrent(1, func() Formatter { return progressFunc("progress", w) })
	require.False(t, failed)

	expected := `.U 2


1 scenarios (1 undefined)
2 steps (1 passed, 1 undefined)
0s

You can implement step definitions for undefined steps with these snippets:

func three() error {
	return godog.ErrPending
}

func unavailableCost(arg1 string, arg2 int) error {
	return godog.ErrPending
}

func undef() error {
	return godog.ErrPending
}

func FeatureContext(s *godog.Suite) {
	s.Step(` + "`^three$`" + `, three)
	s.Step(` + "`^unavailable \"([^\"]*)\" cost (\\d+)$`" + `, unavailableCost)
	s.Step(` + "`^undef$`" + `, undef)
}

`

	actual := buf.String()
	assert.Equal(t, expected, actual)
}

func TestProgressFormatterWhenMultiStepHasArgument(t *testing.T) {
	const path = "any.feature"

	var featureSource = `
Feature: basic

  Scenario: passing scenario
	When one
	Then two:
	"""
	text
	"""
`

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(featureSource), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)

	pickles := gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	r := runner{
		fmt:      progressFunc("progress", w),
		features: []*feature{{GherkinDocument: gd, pickles: pickles}},
		initializer: func(s *Suite) {
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^two:$`, func(doc *messages.PickleStepArgument_PickleDocString) Steps { return Steps{"one"} })
		},
	}

	failed := r.concurrent(1, func() Formatter { return progressFunc("progress", w) })
	require.False(t, failed)
}

func TestProgressFormatterWhenMultiStepHasStepWithArgument(t *testing.T) {
	const path = "any.feature"

	var featureSource = `
Feature: basic

  Scenario: passing scenario
	When one
	Then two`

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(featureSource), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)

	pickles := gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	var subStep = `three:
	"""
	content
	"""`

	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	r := runner{
		fmt:      progressFunc("progress", w),
		features: []*feature{{GherkinDocument: gd, pickles: pickles}},
		initializer: func(s *Suite) {
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^two$`, func() Steps { return Steps{subStep} })
			s.Step(`^three:$`, func(doc *messages.PickleStepArgument_PickleDocString) error { return nil })
		},
	}

	failed := r.concurrent(1, func() Formatter { return progressFunc("progress", w) })
	require.True(t, failed)

	expected := `.F 2


--- Failed steps:

  Scenario: passing scenario # any.feature:4
    Then two # any.feature:6
      Error: nested steps cannot be multiline and have table or content body argument


1 scenarios (1 failed)
2 steps (1 passed, 1 failed)
0s
`

	actual := buf.String()
	assert.Equal(t, expected, actual)
}
