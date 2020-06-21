package godog

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/cucumber/gherkin-go/v11"
	"github.com/cucumber/messages-go/v10"
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

	printStepDefinitions(s.steps, w)

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

	printStepDefinitions(s.steps, w)

	out := strings.TrimSpace(buf.String())
	assert.Equal(t, "there were no contexts registered, could not find any step definition..", out)
}

func TestFailsOrPassesBasedOnStrictModeWhenHasPendingSteps(t *testing.T) {
	const path = "any.feature"

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(basicGherkinFeature), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)

	gd.Uri = path
	ft := feature{GherkinDocument: gd}
	ft.pickles = gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	r := runner{
		fmt:      progressFunc("progress", ioutil.Discard),
		features: []*feature{&ft},
		initializer: func(s *Suite) {
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^two$`, func() error { return ErrPending })
		},
	}

	r.storage = newStorage()
	r.storage.mustInsertFeature(&ft)
	for _, pickle := range ft.pickles {
		r.storage.mustInsertPickle(pickle)
	}

	failed := r.concurrent(1, func() Formatter { return progressFunc("progress", ioutil.Discard) })
	require.False(t, failed)

	r.strict = true
	failed = r.concurrent(1, func() Formatter { return progressFunc("progress", ioutil.Discard) })
	require.True(t, failed)
}

func TestFailsOrPassesBasedOnStrictModeWhenHasUndefinedSteps(t *testing.T) {
	const path = "any.feature"

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(basicGherkinFeature), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)

	gd.Uri = path
	ft := feature{GherkinDocument: gd}
	ft.pickles = gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	r := runner{
		fmt:      progressFunc("progress", ioutil.Discard),
		features: []*feature{&ft},
		initializer: func(s *Suite) {
			s.Step(`^one$`, func() error { return nil })
			// two - is undefined
		},
	}

	r.storage = newStorage()
	r.storage.mustInsertFeature(&ft)
	for _, pickle := range ft.pickles {
		r.storage.mustInsertPickle(pickle)
	}

	failed := r.concurrent(1, func() Formatter { return progressFunc("progress", ioutil.Discard) })
	require.False(t, failed)

	r.strict = true
	failed = r.concurrent(1, func() Formatter { return progressFunc("progress", ioutil.Discard) })
	require.True(t, failed)
}

func TestShouldFailOnError(t *testing.T) {
	const path = "any.feature"

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(basicGherkinFeature), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)

	gd.Uri = path
	ft := feature{GherkinDocument: gd}
	ft.pickles = gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	r := runner{
		fmt:      progressFunc("progress", ioutil.Discard),
		features: []*feature{&ft},
		initializer: func(s *Suite) {
			s.Step(`^one$`, func() error { return nil })
			s.Step(`^two$`, func() error { return fmt.Errorf("error") })
		},
	}

	r.storage = newStorage()
	r.storage.mustInsertFeature(&ft)
	for _, pickle := range ft.pickles {
		r.storage.mustInsertPickle(pickle)
	}

	failed := r.concurrent(1, func() Formatter { return progressFunc("progress", ioutil.Discard) })
	require.True(t, failed)
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

func Test_RandomizeRun(t *testing.T) {
	const noRandomFlag = 0
	const createRandomSeedFlag = -1
	const noConcurrencyFlag = 1
	const formatter = "pretty"
	const featurePath = "formatter-tests/features/with_few_empty_scenarios.feature"

	fmtOutputScenarioInitializer := func(ctx *ScenarioContext) {
		ctx.Step(`^(?:a )?failing step`, failingStepDef)
		ctx.Step(`^(?:a )?pending step$`, pendingStepDef)
		ctx.Step(`^(?:a )?passing step$`, passingStepDef)
		ctx.Step(`^odd (\d+) and even (\d+) number$`, oddEvenStepDef)
	}

	expectedStatus, expectedOutput := testRun(t,
		fmtOutputScenarioInitializer,
		formatter, noConcurrencyFlag,
		noRandomFlag, []string{featurePath},
	)

	actualStatus, actualOutput := testRun(t,
		fmtOutputScenarioInitializer,
		formatter, noConcurrencyFlag,
		createRandomSeedFlag, []string{featurePath},
	)

	expectedSeed := parseSeed(actualOutput)
	assert.NotZero(t, expectedSeed)

	// Removes "Randomized with seed: <seed>" part of the output
	actualOutputSplit := strings.Split(actualOutput, "\n")
	actualOutputSplit = actualOutputSplit[:len(actualOutputSplit)-2]
	actualOutputReduced := strings.Join(actualOutputSplit, "\n")

	assert.NotEqual(t, expectedOutput, actualOutputReduced)
	assertOutput(t, formatter, expectedOutput, actualOutputReduced)

	expectedStatus, expectedOutput = actualStatus, actualOutput

	actualStatus, actualOutput = testRun(t,
		fmtOutputScenarioInitializer,
		formatter, noConcurrencyFlag,
		expectedSeed, []string{featurePath},
	)

	actualSeed := parseSeed(actualOutput)

	assert.Equal(t, expectedSeed, actualSeed)
	assert.Equal(t, expectedStatus, actualStatus)
	assert.Equal(t, expectedOutput, actualOutput)
}

func Test_AllFeaturesRun(t *testing.T) {
	const concurrency = 100
	const noRandomFlag = 0
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

	fmtOutputSuiteInitializer := func(s *Suite) { SuiteContext(s) }
	fmtOutputScenarioInitializer := InitializeScenario

	actualStatus, actualOutput := testRunWithOptions(t,
		fmtOutputSuiteInitializer,
		format, concurrency, []string{"features"},
	)

	assert.Equal(t, exitSuccess, actualStatus)
	assert.Equal(t, expected, actualOutput)

	actualStatus, actualOutput = testRun(t,
		fmtOutputScenarioInitializer,
		format, concurrency,
		noRandomFlag, []string{"features"},
	)

	assert.Equal(t, exitSuccess, actualStatus)
	assert.Equal(t, expected, actualOutput)
}

func TestFormatterConcurrencyRun(t *testing.T) {
	formatters := []string{
		"progress",
		"junit",
		"pretty",
		"events",
		"cucumber",
	}

	featurePaths := []string{"formatter-tests/features"}

	const concurrency = 100
	const noRandomFlag = 0
	const noConcurrency = 1

	fmtOutputSuiteInitializer := func(s *Suite) {
		s.Step(`^(?:a )?failing step`, failingStepDef)
		s.Step(`^(?:a )?pending step$`, pendingStepDef)
		s.Step(`^(?:a )?passing step$`, passingStepDef)
		s.Step(`^odd (\d+) and even (\d+) number$`, oddEvenStepDef)
	}

	fmtOutputScenarioInitializer := func(ctx *ScenarioContext) {
		ctx.Step(`^(?:a )?failing step`, failingStepDef)
		ctx.Step(`^(?:a )?pending step$`, pendingStepDef)
		ctx.Step(`^(?:a )?passing step$`, passingStepDef)
		ctx.Step(`^odd (\d+) and even (\d+) number$`, oddEvenStepDef)
	}

	for _, formatter := range formatters {
		t.Run(
			fmt.Sprintf("%s/concurrency/%d", formatter, concurrency),
			func(t *testing.T) {
				expectedStatus, expectedOutput := testRunWithOptions(t,
					fmtOutputSuiteInitializer,
					formatter, noConcurrency, featurePaths,
				)
				actualStatus, actualOutput := testRunWithOptions(t,
					fmtOutputSuiteInitializer,
					formatter, concurrency, featurePaths,
				)

				assert.Equal(t, expectedStatus, actualStatus)
				assertOutput(t, formatter, expectedOutput, actualOutput)

				expectedStatus, expectedOutput = testRun(t,
					fmtOutputScenarioInitializer,
					formatter, noConcurrency,
					noRandomFlag, featurePaths,
				)
				actualStatus, actualOutput = testRun(t,
					fmtOutputScenarioInitializer,
					formatter, concurrency,
					noRandomFlag, featurePaths,
				)

				assert.Equal(t, expectedStatus, actualStatus)
				assertOutput(t, formatter, expectedOutput, actualOutput)
			},
		)
	}
}

func testRunWithOptions(t *testing.T, initializer func(*Suite), format string, concurrency int, featurePaths []string) (int, string) {
	output := new(bytes.Buffer)

	opts := Options{
		Format:      format,
		NoColors:    true,
		Paths:       featurePaths,
		Concurrency: concurrency,
		Output:      output,
	}

	status := RunWithOptions("succeed", initializer, opts)

	actual, err := ioutil.ReadAll(output)
	require.NoError(t, err)

	return status, string(actual)
}

func testRun(
	t *testing.T,
	scenarioInitializer func(*ScenarioContext),
	format string,
	concurrency int,
	randomSeed int64,
	featurePaths []string,
) (int, string) {
	output := new(bytes.Buffer)

	opts := Options{
		Format:      format,
		NoColors:    true,
		Paths:       featurePaths,
		Concurrency: concurrency,
		Randomize:   randomSeed,
		Output:      output,
	}

	status := TestSuite{
		Name:                "succeed",
		ScenarioInitializer: scenarioInitializer,
		Options:             &opts,
	}.Run()

	actual, err := ioutil.ReadAll(output)
	require.NoError(t, err)

	return status, string(actual)
}

func assertOutput(t *testing.T, formatter string, expected, actual string) {
	switch formatter {
	case "cucumber", "junit", "pretty", "events":
		expectedRows := strings.Split(expected, "\n")
		actualRows := strings.Split(actual, "\n")
		assert.ElementsMatch(t, expectedRows, actualRows)
	case "progress":
		expectedOutput := parseProgressOutput(expected)
		actualOutput := parseProgressOutput(actual)

		assert.Equal(t, expectedOutput.passed, actualOutput.passed)
		assert.Equal(t, expectedOutput.skipped, actualOutput.skipped)
		assert.Equal(t, expectedOutput.failed, actualOutput.failed)
		assert.Equal(t, expectedOutput.undefined, actualOutput.undefined)
		assert.Equal(t, expectedOutput.pending, actualOutput.pending)
		assert.Equal(t, expectedOutput.noOfStepsPerRow, actualOutput.noOfStepsPerRow)
		assert.ElementsMatch(t, expectedOutput.bottomRows, actualOutput.bottomRows)
	}
}

func parseProgressOutput(output string) (parsed progressOutput) {
	mainParts := strings.Split(output, "\n\n\n")

	topRows := strings.Split(mainParts[0], "\n")
	parsed.bottomRows = strings.Split(mainParts[1], "\n")

	parsed.noOfStepsPerRow = make([]string, len(topRows))
	for idx, row := range topRows {
		rowParts := strings.Split(row, " ")
		stepResults := strings.Split(rowParts[0], "")
		parsed.noOfStepsPerRow[idx] = rowParts[1]

		for _, stepResult := range stepResults {
			switch stepResult {
			case ".":
				parsed.passed++
			case "-":
				parsed.skipped++
			case "F":
				parsed.failed++
			case "U":
				parsed.undefined++
			case "P":
				parsed.pending++
			}
		}
	}

	return parsed
}

type progressOutput struct {
	passed          int
	skipped         int
	failed          int
	undefined       int
	pending         int
	noOfStepsPerRow []string
	bottomRows      []string
}

func passingStepDef() error { return nil }

func oddEvenStepDef(odd, even int) error { return oddOrEven(odd, even) }

func oddOrEven(odd, even int) error {
	if odd%2 == 0 {
		return fmt.Errorf("%d is not odd", odd)
	}
	if even%2 != 0 {
		return fmt.Errorf("%d is not even", even)
	}
	return nil
}

func pendingStepDef() error { return ErrPending }

func failingStepDef() error { return fmt.Errorf("step failed") }

func parseSeed(str string) (seed int64) {
	re := regexp.MustCompile(`Randomized with seed: (\d*)`)
	match := re.FindStringSubmatch(str)

	if len(match) > 0 {
		var err error
		if seed, err = strconv.ParseInt(match[1], 10, 64); err != nil {
			seed = 0
		}
	}

	return
}
