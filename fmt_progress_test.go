package godog

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/cucumber/godog/colors"
	"github.com/cucumber/godog/gherkin"
)

func TestProgressFormatterOutput(t *testing.T) {
	feat, err := gherkin.ParseFeature(strings.NewReader(sampleGherkinFeature))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	r := runner{
		fmt: progressFunc("progress", w),
		features: []*feature{&feature{
			Path:    "any.feature",
			Feature: feat,
			Content: []byte(sampleGherkinFeature),
		}},
		initializer: func(s *Suite) {
			s.Step(`^passing$`, func() error { return nil })
			s.Step(`^failing$`, func() error { return fmt.Errorf("errored") })
			s.Step(`^pending$`, func() error { return ErrPending })
		},
	}

	expected := `
...F-.P-.UU.....F..P..U 23


--- Failed steps:

  Scenario: failing scenario # any.feature:10
    When failing # any.feature:11
	  Error: errored

  Scenario Outline: outline # any.feature:22
	When failing # any.feature:24
	  Error: errored


8 scenarios (2 passed, 2 failed, 2 pending, 2 undefined)
23 steps (14 passed, 2 failed, 2 pending, 3 undefined, 2 skipped)
0s

You can implement step definitions for undefined steps with these snippets:

func undefined() error {
	return godog.ErrPending
}

func nextUndefined() error {
	return godog.ErrPending
}

func FeatureContext(s *godog.Suite) {
	s.Step(` + "`^undefined$`" + `, undefined)
	s.Step(` + "`^next undefined$`" + `, nextUndefined)
}`

	expected = trimAllLines(expected)

	r.run()

	actual := trimAllLines(buf.String())

	shouldMatchOutput(expected, actual, t)
}

func trimAllLines(s string) string {
	var lines []string
	for _, ln := range strings.Split(strings.TrimSpace(s), "\n") {
		lines = append(lines, strings.TrimSpace(ln))
	}
	return strings.Join(lines, "\n")
}

var basicGherkinFeature = `
Feature: basic

  Scenario: passing scenario
	When one
	Then two
`

func TestProgressFormatterWhenStepPanics(t *testing.T) {
	feat, err := gherkin.ParseFeature(strings.NewReader(basicGherkinFeature))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	r := runner{
		fmt:      progressFunc("progress", w),
		features: []*feature{&feature{Feature: feat}},
		initializer: func(s *Suite) {
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^two$`, func() error { panic("omg") })
		},
	}

	if !r.run() {
		t.Fatal("the suite should have failed")
	}

	out := buf.String()
	if idx := strings.Index(out, "godog/fmt_progress_test.go:108"); idx == -1 {
		t.Fatalf("expected to find panic stacktrace, actual:\n%s", out)
	}
}

func TestProgressFormatterWithPassingMultisteps(t *testing.T) {
	feat, err := gherkin.ParseFeature(strings.NewReader(basicGherkinFeature))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	r := runner{
		fmt:      progressFunc("progress", w),
		features: []*feature{&feature{Feature: feat}},
		initializer: func(s *Suite) {
			s.Step(`^sub1$`, func() error { return nil })
			s.Step(`^sub-sub$`, func() error { return nil })
			s.Step(`^sub2$`, func() Steps { return Steps{"sub-sub", "sub1", "one"} })
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^two$`, func() Steps { return Steps{"sub1", "sub2"} })
		},
	}

	if r.run() {
		t.Fatal("the suite should have passed")
	}
}

func TestProgressFormatterWithFailingMultisteps(t *testing.T) {
	feat, err := gherkin.ParseFeature(strings.NewReader(basicGherkinFeature))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	r := runner{
		fmt:      progressFunc("progress", w),
		features: []*feature{&feature{Feature: feat, Path: "some.feature"}},
		initializer: func(s *Suite) {
			s.Step(`^sub1$`, func() error { return nil })
			s.Step(`^sub-sub$`, func() error { return fmt.Errorf("errored") })
			s.Step(`^sub2$`, func() Steps { return Steps{"sub-sub", "sub1", "one"} })
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^two$`, func() Steps { return Steps{"sub1", "sub2"} })
		},
	}

	if !r.run() {
		t.Fatal("the suite should have failed")
	}

	expected := `
.F 2


--- Failed steps:

Scenario: passing scenario # some.feature:4
Then two # some.feature:6
Error: sub2: sub-sub: errored


1 scenarios (1 failed)
2 steps (1 passed, 1 failed)
0s
`

	expected = trimAllLines(expected)
	actual := trimAllLines(buf.String())

	shouldMatchOutput(expected, actual, t)
}

func shouldMatchOutput(expected, actual string, t *testing.T) {
	act := []byte(actual)
	exp := []byte(expected)

	if len(act) != len(exp) {
		t.Fatalf("content lengths do not match, expected: %d, actual %d, actual output:\n%s", len(exp), len(act), actual)
	}

	for i := 0; i < len(exp); i++ {
		if act[i] == exp[i] {
			continue
		}

		cpe := make([]byte, len(exp))
		copy(cpe, exp)
		e := append(exp[:i], '^')
		e = append(e, cpe[i:]...)

		cpa := make([]byte, len(act))
		copy(cpa, act)
		a := append(act[:i], '^')
		a = append(a, cpa[i:]...)

		t.Fatalf("expected output does not match:\n%s\n\n%s", string(a), string(e))
	}
}

func TestProgressFormatterWithPanicInMultistep(t *testing.T) {
	feat, err := gherkin.ParseFeature(strings.NewReader(basicGherkinFeature))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	r := runner{
		fmt:      progressFunc("progress", w),
		features: []*feature{&feature{Feature: feat}},
		initializer: func(s *Suite) {
			s.Step(`^sub1$`, func() error { return nil })
			s.Step(`^sub-sub$`, func() error { return nil })
			s.Step(`^sub2$`, func() []string { return []string{"sub-sub", "sub1", "one"} })
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^two$`, func() []string { return []string{"sub1", "sub2"} })
		},
	}

	if !r.run() {
		t.Fatal("the suite should have failed")
	}
}

func TestProgressFormatterMultistepTemplates(t *testing.T) {
	feat, err := gherkin.ParseFeature(strings.NewReader(basicGherkinFeature))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	r := runner{
		fmt:      progressFunc("progress", w),
		features: []*feature{&feature{Feature: feat}},
		initializer: func(s *Suite) {
			s.Step(`^sub-sub$`, func() error { return nil })
			s.Step(`^substep$`, func() Steps { return Steps{"sub-sub", `unavailable "John" cost 5`, "one", "three"} })
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^(t)wo$`, func(s string) Steps { return Steps{"undef", "substep"} })
		},
	}

	if r.run() {
		t.Fatal("the suite should have passed")
	}

	expected := `
.U 2


1 scenarios (1 undefined)
2 steps (1 passed, 1 undefined)
0s

You can implement step definitions for undefined steps with these snippets:

func undef() error {
	return godog.ErrPending
}

func unavailableCost(arg1 string, arg2 int) error {
	return godog.ErrPending
}

func three() error {
	return godog.ErrPending
}

func FeatureContext(s *godog.Suite) {
	s.Step(` + "`^undef$`" + `, undef)
	s.Step(` + "`^unavailable \"([^\"]*)\" cost (\\d+)$`" + `, unavailableCost)
	s.Step(` + "`^three$`" + `, three)
}
`

	expected = trimAllLines(expected)

	actual := trimAllLines(buf.String())
	if actual != expected {
		t.Fatalf("expected output does not match: %s", actual)
	}
}

func TestProgressFormatterWhenMultiStepHasArgument(t *testing.T) {

	var featureSource = `
Feature: basic

  Scenario: passing scenario
	When one
	Then two:
	"""
	text
	"""
`
	feat, err := gherkin.ParseFeature(strings.NewReader(featureSource))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := runner{
		fmt:      progressFunc("progress", ioutil.Discard),
		features: []*feature{&feature{Feature: feat}},
		initializer: func(s *Suite) {
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^two:$`, func(doc *gherkin.DocString) Steps { return Steps{"one"} })
		},
	}

	if r.run() {
		t.Fatal("the suite should have passed")
	}
}

func TestProgressFormatterWhenMultiStepHasStepWithArgument(t *testing.T) {

	var featureSource = `
Feature: basic

  Scenario: passing scenario
	When one
	Then two`

	feat, err := gherkin.ParseFeature(strings.NewReader(featureSource))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var subStep = `three:
	"""
	content
	"""`

	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	r := runner{
		fmt:      progressFunc("progress", w),
		features: []*feature{&feature{Feature: feat}},
		initializer: func(s *Suite) {
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^two$`, func() Steps { return Steps{subStep} })
			s.Step(`^three:$`, func(doc *gherkin.DocString) error { return nil })
		},
	}

	if !r.run() {
		t.Fatal("the suite should have failed")
	}

	expected := `
.F 2


--- Failed steps:

  Scenario: passing scenario # :4
    Then two # :6
      Error: nested steps cannot be multiline and have table or content body argument


1 scenarios (1 failed)
2 steps (1 passed, 1 failed)
0s
`

	expected = trimAllLines(expected)
	actual := trimAllLines(buf.String())
	if actual != expected {
		t.Fatalf("expected output does not match: %s", actual)
	}
}
