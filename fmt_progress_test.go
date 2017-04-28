package godog

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/godog/colors"
	"github.com/DATA-DOG/godog/gherkin"
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

    When failing # any.feature:11
	  Error: errored

    When failing # any.feature:24
	  Error: errored


8 scenarios (2 passed, 2 failed, 2 pending, 2 undefined)
23 steps (14 passed, 2 failed, 2 pending, 3 undefined, 2 skipped)
%s

Randomized with seed: %s

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

	var zeroDuration time.Duration
	expected = fmt.Sprintf(expected, zeroDuration.String(), os.Getenv("GODOG_SEED"))
	expected = trimAllLines(expected)

	r.run()

	actual := trimAllLines(buf.String())
	if actual != expected {
		t.Fatalf("expected output does not match: %s", actual)
	}
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
	When step1
	Then step2
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
			s.Step(`^step1$`, func() error { return nil })
			s.Step(`^step2$`, func() error { panic("omg") })
		},
	}

	if !r.run() {
		t.Fatal("the suite should have failed")
	}

	out := buf.String()
	if idx := strings.Index(out, "github.com/DATA-DOG/godog/fmt_progress_test.go:116"); idx == -1 {
		t.Fatal("expected to find panic stacktrace")
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
			s.Step(`^sub2$`, func() Steps { return Steps{"sub-sub", "sub1", "step1"} })
			s.Step(`^step1$`, func() error { return nil })
			s.Step(`^step2$`, func() Steps { return Steps{"sub1", "sub2"} })
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
		features: []*feature{&feature{Feature: feat}},
		initializer: func(s *Suite) {
			s.Step(`^sub1$`, func() error { return nil })
			s.Step(`^sub-sub$`, func() error { return fmt.Errorf("errored") })
			s.Step(`^sub2$`, func() Steps { return Steps{"sub-sub", "sub1", "step1"} })
			s.Step(`^step1$`, func() error { return nil })
			s.Step(`^step2$`, func() Steps { return Steps{"sub1", "sub2"} })
		},
	}

	if !r.run() {
		t.Fatal("the suite should have failed")
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
			s.Step(`^sub2$`, func() []string { return []string{"sub-sub", "sub1", "step1"} })
			s.Step(`^step1$`, func() error { return nil })
			s.Step(`^step2$`, func() []string { return []string{"sub1", "sub2"} })
		},
	}

	if !r.run() {
		t.Fatal("the suite should have failed")
	}
}
