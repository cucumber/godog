package godog

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/godog/colors"
	"github.com/DATA-DOG/godog/gherkin"
)

func okStep() error {
	return nil
}

func TestPrintsStepDefinitions(t *testing.T) {
	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	s := &Suite{}

	steps := []string{
		"^passing step$",
		`^with name "([^"])"`,
	}

	for _, step := range steps {
		s.Step(step, okStep)
	}
	s.printStepDefinitions(w)

	out := buf.String()
	ref := `okStep`
	for i, def := range strings.Split(strings.TrimSpace(out), "\n") {
		if idx := strings.Index(def, steps[i]); idx == -1 {
			t.Fatalf(`step "%s" was not found in output`, steps[i])
		}
		if idx := strings.Index(def, ref); idx == -1 {
			t.Fatalf(`step definition reference "%s" was not found in output: "%s"`, ref, def)
		}
	}
}

func TestPrintsNoStepDefinitionsIfNoneFound(t *testing.T) {
	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	s := &Suite{}
	s.printStepDefinitions(w)

	out := strings.TrimSpace(buf.String())
	if out != "there were no contexts registered, could not find any step definition.." {
		t.Fatalf("expected output does not match to: %s", out)
	}
}

func TestFailsOrPassesBasedOnStrictModeWhenHasPendingSteps(t *testing.T) {
	feat, err := gherkin.ParseFeature(strings.NewReader(basicGherkinFeature))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := runner{
		fmt:      progressFunc("progress", ioutil.Discard),
		features: []*feature{&feature{Feature: feat}},
		initializer: func(s *Suite) {
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^two$`, func() error { return ErrPending })
		},
	}

	if r.run() {
		t.Fatal("the suite should have passed")
	}

	r.strict = true
	if !r.run() {
		t.Fatal("the suite should have failed")
	}
}

func TestFailsOrPassesBasedOnStrictModeWhenHasUndefinedSteps(t *testing.T) {
	feat, err := gherkin.ParseFeature(strings.NewReader(basicGherkinFeature))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := runner{
		fmt:      progressFunc("progress", ioutil.Discard),
		features: []*feature{&feature{Feature: feat}},
		initializer: func(s *Suite) {
			s.Step(`^one$`, func() error { return nil })
			// two - is undefined
		},
	}

	if r.run() {
		t.Fatal("the suite should have passed")
	}

	r.strict = true
	if !r.run() {
		t.Fatal("the suite should have failed")
	}
}

func TestShouldFailOnError(t *testing.T) {
	feat, err := gherkin.ParseFeature(strings.NewReader(basicGherkinFeature))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r := runner{
		fmt:      progressFunc("progress", ioutil.Discard),
		features: []*feature{&feature{Feature: feat}},
		initializer: func(s *Suite) {
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^two$`, func() error { return fmt.Errorf("error") })
		},
	}

	if !r.run() {
		t.Fatal("the suite should have failed")
	}
}

func TestFailsWithConcurrencyOptionError(t *testing.T) {
	stderr, closer := bufErrorPipe(t)
	defer closer()
	defer stderr.Close()

	opt := Options{
		Format:      "pretty",
		Paths:       []string{"features/load:6"},
		Concurrency: 2,
		Output:      ioutil.Discard,
	}

	status := RunWithOptions("fails", func(_ *Suite) {}, opt)
	if status != exitOptionError {
		t.Fatalf("expected exit status to be 2, but was: %d", status)
	}
	closer()

	b, err := ioutil.ReadAll(stderr)
	if err != nil {
		t.Fatal(err)
	}

	out := strings.TrimSpace(string(b))
	if out != `format "pretty" does not support concurrent execution` {
		t.Fatalf("unexpected error output: \"%s\"", out)
	}
}

func TestFailsWithUnknownFormatterOptionError(t *testing.T) {
	stderr, closer := bufErrorPipe(t)
	defer closer()
	defer stderr.Close()

	opt := Options{
		Format: "unknown",
		Paths:  []string{"features/load:6"},
		Output: ioutil.Discard,
	}

	status := RunWithOptions("fails", func(_ *Suite) {}, opt)
	if status != exitOptionError {
		t.Fatalf("expected exit status to be 2, but was: %d", status)
	}
	closer()

	b, err := ioutil.ReadAll(stderr)
	if err != nil {
		t.Fatal(err)
	}

	out := strings.TrimSpace(string(b))
	if !strings.Contains(out, `unregistered formatter name: "unknown", use one of`) {
		t.Fatalf("unexpected error output: %q", out)
	}
}

func TestFailsWithOptionErrorWhenLookingForFeaturesInUnavailablePath(t *testing.T) {
	stderr, closer := bufErrorPipe(t)
	defer closer()
	defer stderr.Close()

	opt := Options{
		Format: "progress",
		Paths:  []string{"unavailable"},
		Output: ioutil.Discard,
	}

	status := RunWithOptions("fails", func(_ *Suite) {}, opt)
	if status != exitOptionError {
		t.Fatalf("expected exit status to be 2, but was: %d", status)
	}
	closer()

	b, err := ioutil.ReadAll(stderr)
	if err != nil {
		t.Fatal(err)
	}

	out := strings.TrimSpace(string(b))
	if out != `feature path "unavailable" is not available` {
		t.Fatalf("unexpected error output: \"%s\"", out)
	}
}

func TestByDefaultRunsFeaturesPath(t *testing.T) {
	opt := Options{
		Format: "progress",
		Output: ioutil.Discard,
		Strict: true,
	}

	status := RunWithOptions("fails", func(_ *Suite) {}, opt)
	// should fail in strict mode due to undefined steps
	if status != exitFailure {
		t.Fatalf("expected exit status to be 1, but was: %d", status)
	}

	opt.Strict = false
	status = RunWithOptions("succeeds", func(_ *Suite) {}, opt)
	// should succeed in non strict mode due to undefined steps
	if status != exitSuccess {
		t.Fatalf("expected exit status to be 0, but was: %d", status)
	}
}

func bufErrorPipe(t *testing.T) (io.ReadCloser, func()) {
	stderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stderr = w
	return r, func() {
		w.Close()
		os.Stderr = stderr
	}
}

func TestFeatureFilePathParser(t *testing.T) {

	type Case struct {
		input string
		path  string
		line  int
	}

	cases := []Case{
		{"/home/test.feature", "/home/test.feature", -1},
		{"/home/test.feature:21", "/home/test.feature", 21},
		{"test.feature", "test.feature", -1},
		{"test.feature:2", "test.feature", 2},
		{"", "", -1},
		{"/c:/home/test.feature", "/c:/home/test.feature", -1},
		{"/c:/home/test.feature:3", "/c:/home/test.feature", 3},
		{"D:\\home\\test.feature:3", "D:\\home\\test.feature", 3},
	}

	for i, c := range cases {
		p, ln := extractFeaturePathLine(c.input)
		if p != c.path {
			t.Fatalf(`result path "%s" != "%s" at %d`, p, c.path, i)
		}
		if ln != c.line {
			t.Fatalf(`result line "%d" != "%d" at %d`, ln, c.line, i)
		}
	}
}

func TestSucceedWithProgressAndConcurrencyOption(t *testing.T) {
	output := new(bytes.Buffer)

	opt := Options{
		Format:      "progress",
		NoColors:    true,
		Paths:       []string{"features"},
		Concurrency: 2,
		Output:      output,
	}

	expectedOutput := `...................................................................... 70
...................................................................... 140
...................................................................... 210
.......................................                                249


60 scenarios (60 passed)
249 steps (249 passed)
0s`

	status := RunWithOptions("succeed", func(s *Suite) { SuiteContext(s) }, opt)
	if status != exitSuccess {
		t.Fatalf("expected exit status to be 0, but was: %d", status)
	}

	b, err := ioutil.ReadAll(output)
	if err != nil {
		t.Fatal(err)
	}

	out := strings.TrimSpace(string(b))
	if out != expectedOutput {
		t.Fatalf("unexpected output: \"%s\"", out)
	}
}

func TestSucceedWithJunitAndConcurrencyOption(t *testing.T) {
	output := new(bytes.Buffer)

	opt := Options{
		Format:      "junit",
		NoColors:    true,
		Paths:       []string{"features"},
		Concurrency: 2,
		Output:      output,
	}

	zeroSecondsString := (0 * time.Second).String()

	expectedOutput := `<?xml version="1.0" encoding="UTF-8"?>
<testsuites name="succeed" tests="60" skipped="0" failures="0" errors="0" time="` + zeroSecondsString + `">
  <testsuite name="cucumber json formatter" tests="9" skipped="0" failures="0" errors="0" time="` + zeroSecondsString + `">
    <testcase name="Support of Feature Plus Scenario Node" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="Support of Feature Plus Scenario Node With Tags" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="Support of Feature Plus Scenario Outline" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="Support of Feature Plus Scenario Outline With Tags" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="Support of Feature Plus Scenario With Steps" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="Support of Feature Plus Scenario Outline With Steps" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="Support of Comments" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="Support of Docstrings" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="Support of Undefined, Pending and Skipped status" status="passed" time="` + zeroSecondsString + `"></testcase>
  </testsuite>
  <testsuite name="event stream formatter" tests="3" skipped="0" failures="0" errors="0" time="` + zeroSecondsString + `">
    <testcase name="should fire only suite events without any scenario" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should process simple scenario" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should process outline scenario" status="passed" time="` + zeroSecondsString + `"></testcase>
  </testsuite>
  <testsuite name="load features" tests="6" skipped="0" failures="0" errors="0" time="` + zeroSecondsString + `">
    <testcase name="load features within path" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="load a specific feature file" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="loaded feature should have a number of scenarios #1" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="loaded feature should have a number of scenarios #2" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="loaded feature should have a number of scenarios #3" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="load a number of feature files" status="passed" time="` + zeroSecondsString + `"></testcase>
  </testsuite>
  <testsuite name="run background" tests="3" skipped="0" failures="0" errors="0" time="` + zeroSecondsString + `">
    <testcase name="should run background steps" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should skip all consequent steps on failure" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should continue undefined steps" status="passed" time="` + zeroSecondsString + `"></testcase>
  </testsuite>
  <testsuite name="run features" tests="11" skipped="0" failures="0" errors="0" time="` + zeroSecondsString + `">
    <testcase name="should run a normal feature" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should skip steps after failure" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should skip all scenarios if background fails" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should skip steps after undefined" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should match undefined steps in a row" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should skip steps on pending" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should handle pending step" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should mark undefined steps after pending" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should fail suite if undefined steps follow after the failure" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should fail suite and skip pending step after failed step" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should fail suite and skip next step after failed step" status="passed" time="` + zeroSecondsString + `"></testcase>
  </testsuite>
  <testsuite name="run features with nested steps" tests="6" skipped="0" failures="0" errors="0" time="` + zeroSecondsString + `">
    <testcase name="should run passing multistep successfully" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should fail multistep" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should fail nested multistep" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should skip steps after undefined multistep" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should match undefined steps in a row" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should mark undefined steps after pending" status="passed" time="` + zeroSecondsString + `"></testcase>
  </testsuite>
  <testsuite name="run outline" tests="6" skipped="0" failures="0" errors="0" time="` + zeroSecondsString + `">
    <testcase name="should run a normal outline" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should continue through examples on failure" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should skip examples on background failure" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should translate step table body" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should translate step doc string argument #1" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should translate step doc string argument #2" status="passed" time="` + zeroSecondsString + `"></testcase>
  </testsuite>
  <testsuite name="suite events" tests="6" skipped="0" failures="0" errors="0" time="` + zeroSecondsString + `">
    <testcase name="triggers before scenario event" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="triggers appropriate events for a single scenario" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="triggers appropriate events whole feature" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="triggers appropriate events for two feature files" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should not trigger events on empty feature" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should not trigger events on empty scenarios" status="passed" time="` + zeroSecondsString + `"></testcase>
  </testsuite>
  <testsuite name="tag filters" tests="4" skipped="0" failures="0" errors="0" time="` + zeroSecondsString + `">
    <testcase name="should filter outline examples by tags" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should filter scenarios by X tag" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should filter scenarios by X tag not having Y" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should filter scenarios having Y and Z tags" status="passed" time="` + zeroSecondsString + `"></testcase>
  </testsuite>
  <testsuite name="undefined step snippets" tests="5" skipped="0" failures="0" errors="0" time="` + zeroSecondsString + `">
    <testcase name="should generate snippets" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should generate snippets with more arguments" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should handle escaped symbols" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should handle string argument followed by comma" status="passed" time="` + zeroSecondsString + `"></testcase>
    <testcase name="should handle arguments in the beggining or end of the step" status="passed" time="` + zeroSecondsString + `"></testcase>
  </testsuite>
  <testsuite name="užkrauti savybes" tests="1" skipped="0" failures="0" errors="0" time="` + zeroSecondsString + `">
    <testcase name="savybių užkrovimas iš aplanko" status="passed" time="` + zeroSecondsString + `"></testcase>
  </testsuite>
</testsuites>`

	status := RunWithOptions("succeed", func(s *Suite) { SuiteContext(s) }, opt)
	if status != exitSuccess {
		t.Fatalf("expected exit status to be 0, but was: %d", status)
	}

	b, err := ioutil.ReadAll(output)
	if err != nil {
		t.Fatal(err)
	}

	out := strings.TrimSpace(string(b))
	if out != expectedOutput {
		t.Fatalf("unexpected output: \"%s\"", out)
	}
}
