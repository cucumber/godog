package godog

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/cucumber/gherkin-go/v9"
	"github.com/cucumber/messages-go/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cucumber/godog/colors"
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
	assert.Equal(t, "there were no contexts registered, could not find any step definition..", out)
}

func TestFailsOrPassesBasedOnStrictModeWhenHasPendingSteps(t *testing.T) {
	const path = "any.feature"

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(basicGherkinFeature), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)

	pickles := gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	r := runner{
		fmt:      progressFunc("progress", ioutil.Discard),
		features: []*feature{{GherkinDocument: gd, pickles: pickles}},
		initializer: func(s *Suite) {
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^two$`, func() error { return ErrPending })
		},
	}

	assert.False(t, r.run())

	r.strict = true
	assert.True(t, r.run())
}

func TestFailsOrPassesBasedOnStrictModeWhenHasUndefinedSteps(t *testing.T) {
	const path = "any.feature"

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(basicGherkinFeature), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)

	pickles := gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	r := runner{
		fmt:      progressFunc("progress", ioutil.Discard),
		features: []*feature{{GherkinDocument: gd, pickles: pickles}},
		initializer: func(s *Suite) {
			s.Step(`^one$`, func() error { return nil })
			// two - is undefined
		},
	}

	assert.False(t, r.run())

	r.strict = true
	assert.True(t, r.run())
}

func TestShouldFailOnError(t *testing.T) {
	const path = "any.feature"

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(basicGherkinFeature), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)

	pickles := gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	r := runner{
		fmt:      progressFunc("progress", ioutil.Discard),
		features: []*feature{{GherkinDocument: gd, pickles: pickles}},
		initializer: func(s *Suite) {
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^two$`, func() error { return fmt.Errorf("error") })
		},
	}

	assert.True(t, r.run())
}

func TestFailsWithConcurrencyOptionError(t *testing.T) {
	stderr, closer := bufErrorPipe(t)
	defer closer()
	defer stderr.Close()

	opt := Options{
		Format:      "cucumber",
		Paths:       []string{"features/load:6"},
		Concurrency: 2,
		Output:      ioutil.Discard,
	}

	status := RunWithOptions("fails", func(_ *Suite) {}, opt)
	require.Equal(t, exitOptionError, status)

	closer()

	b, err := ioutil.ReadAll(stderr)
	require.NoError(t, err)

	out := strings.TrimSpace(string(b))
	assert.Equal(t, `format "cucumber" does not support concurrent execution`, out)
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
	require.Equal(t, exitOptionError, status)

	closer()

	b, err := ioutil.ReadAll(stderr)
	require.NoError(t, err)

	out := strings.TrimSpace(string(b))
	assert.Contains(t, out, `unregistered formatter name: "unknown", use one of`)
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
	require.Equal(t, exitOptionError, status)

	closer()

	b, err := ioutil.ReadAll(stderr)
	require.NoError(t, err)

	out := strings.TrimSpace(string(b))
	assert.Equal(t, `feature path "unavailable" is not available`, out)
}

func TestByDefaultRunsFeaturesPath(t *testing.T) {
	opt := Options{
		Format: "progress",
		Output: ioutil.Discard,
		Strict: true,
	}

	status := RunWithOptions("fails", func(_ *Suite) {}, opt)
	// should fail in strict mode due to undefined steps
	assert.Equal(t, exitFailure, status)

	opt.Strict = false
	status = RunWithOptions("succeeds", func(_ *Suite) {}, opt)
	// should succeed in non strict mode due to undefined steps
	assert.Equal(t, exitSuccess, status)
}

func bufErrorPipe(t *testing.T) (io.ReadCloser, func()) {
	stderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)

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

	for _, c := range cases {
		p, ln := extractFeaturePathLine(c.input)
		assert.Equal(t, p, c.path)
		assert.Equal(t, ln, c.line)
	}
}

func TestAllFeaturesRun(t *testing.T) {
	const concurrency = 10
	const format = "progress"

	const expected = `...................................................................... 70
...................................................................... 140
...................................................................... 210
...................................................................... 280
..........................                                             306


79 scenarios (79 passed)
306 steps (306 passed)
0s
`

	actual := testSucceedRun(t, format, concurrency, []string{"features"}, exitSuccess)
	assert.Equalf(t, expected, actual, "[%s]", actual)
}

func TestFormatterConcurrencyRun(t *testing.T) {
	formatters := map[string]int{
		"progress": exitSuccess,
		"junit":    exitSuccess,
		"pretty":   exitSuccess,
		"events":   exitSuccess,
		"cucumber": exitOptionError,
	}

	featurePaths := []string{
		"formatter-tests/features/scenario_outline.feature",
		"formatter-tests/features/single_scenario_with_passing_step.feature",
	}

	const concurrency = 2

	for formatter, expectedRunStatus := range formatters {
		t.Run(
			fmt.Sprintf("%s/concurrency/%d", formatter, concurrency),
			func(t *testing.T) {
				testSucceedRun(t, formatter, concurrency, featurePaths, expectedRunStatus)
			},
		)
	}
}

func testSucceedRun(t *testing.T, format string, concurrency int, featurePaths []string, expectedRunStatus int) string {
	output := new(bytes.Buffer)

	opt := Options{
		Format:      format,
		NoColors:    true,
		Paths:       featurePaths,
		Concurrency: concurrency,
		Output:      output,
	}

	status := RunWithOptions("succeed", func(s *Suite) { SuiteContext(s) }, opt)
	assert.Equal(t, expectedRunStatus, status)

	actual, err := ioutil.ReadAll(output)
	require.NoError(t, err)

	return string(actual)
}
