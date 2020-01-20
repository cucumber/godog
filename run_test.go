package godog

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

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

func TestSucceedWithConcurrencyOption(t *testing.T) {
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
