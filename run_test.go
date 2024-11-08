package godog

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/cucumber/godog/internal/utils"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	gherkin "github.com/cucumber/gherkin/go/v26"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cucumber/godog/colors"
	"github.com/cucumber/godog/internal/formatters"
	"github.com/cucumber/godog/internal/models"
	"github.com/cucumber/godog/internal/storage"
)

func Test_TestSuite_Run(t *testing.T) {
	for _, tc := range []struct {
		name          string
		body          string
		afterStepCnt  int
		beforeStepCnt int
		log           string
		noStrict      bool
		suitePasses   bool
	}{
		{
			name: "fail_then_pass_fails_scenario", afterStepCnt: 2, beforeStepCnt: 2,
			body: `
					When step fails
					Then step passes`,
			log: `
					>>>> Before suite
					>> Before scenario "test"
					Before step "step fails"
					After step "step fails", error: oops, status: failed
					<< After scenario "test", error: oops
					Before step "step passes"
					After step "step passes", error: <nil>, status: skipped
					<<<< After suite`,
		},
		{
			name: "pending_then_pass_fails_scenario", afterStepCnt: 2, beforeStepCnt: 2,
			body: `
					When step is pending
					Then step passes`,
			log: `
					>>>> Before suite
					>> Before scenario "test"
					Before step "step is pending"
					After step "step is pending", error: step implementation is pending, status: pending
					<< After scenario "test", error: step implementation is pending
					Before step "step passes"
					After step "step passes", error: <nil>, status: skipped
					<<<< After suite`,
		},
		{
			name: "pending_then_pass_no_strict_doesnt_fail_scenario", afterStepCnt: 2, beforeStepCnt: 2, noStrict: true, suitePasses: true,
			body: `
					When step is pending
					Then step passes`,
			log: `
					>>>> Before suite
					>> Before scenario "test"
					Before step "step is pending"
					After step "step is pending", error: step implementation is pending, status: pending
					Before step "step passes"
					After step "step passes", error: <nil>, status: skipped
					<< After scenario "test", error: <nil>
					<<<< After suite`,
		},
		{
			name: "undefined_then_pass_no_strict_doesnt_fail_scenario", afterStepCnt: 2, beforeStepCnt: 2, noStrict: true, suitePasses: true,
			body: `
					When step is undefined
					Then step passes`,
			log: `
					>>>> Before suite
					>> Before scenario "test"
					Before step "step is undefined"
					After step "step is undefined", error: step is undefined, status: undefined
					Before step "step passes"
					After step "step passes", error: <nil>, status: skipped
					<< After scenario "test", error: <nil>
					<<<< After suite`,
		},
		{
			name: "undefined_then_pass_fails_scenario", afterStepCnt: 2, beforeStepCnt: 2,
			body: `
					When step is undefined
					Then step passes`,
			log: `
					>>>> Before suite
					>> Before scenario "test"
					Before step "step is undefined"
					After step "step is undefined", error: step is undefined, status: undefined
					<< After scenario "test", error: step is undefined
					Before step "step passes"
					After step "step passes", error: <nil>, status: skipped
					<<<< After suite`,
		},
		{
			name: "fail_then_undefined_fails_scenario", afterStepCnt: 2, beforeStepCnt: 2,
			body: `
					When step fails
					Then step is undefined`,
			log: `
					>>>> Before suite
					>> Before scenario "test"
					Before step "step fails"
					After step "step fails", error: oops, status: failed
					<< After scenario "test", error: oops
					Before step "step is undefined"
					After step "step is undefined", error: step is undefined, status: undefined
					<<<< After suite`,
		},
		{
			name: "passes", afterStepCnt: 2, beforeStepCnt: 2,
			body: `
					When step passes
					Then step passes`,
			suitePasses: true,
			log: `
					>>>> Before suite
					>> Before scenario "test"
					Before step "step passes"
					<step action>
					After step "step passes", error: <nil>, status: passed
					Before step "step passes"
					<step action>
					After step "step passes", error: <nil>, status: passed
					<< After scenario "test", error: <nil>
					<<<< After suite`,
		},
		{
			name: "skip_does_not_fail_scenario", afterStepCnt: 2, beforeStepCnt: 2,
			body: `
					When step skips scenario
					Then step fails`,
			suitePasses: true,
			log: `
					>>>> Before suite
					>> Before scenario "test"
					Before step "step skips scenario"
					After step "step skips scenario", error: skipped, status: skipped
					Before step "step fails"
					After step "step fails", error: <nil>, status: skipped
					<< After scenario "test", error: <nil>
					<<<< After suite`,
		},
		{
			name: "multistep_passes", afterStepCnt: 6, beforeStepCnt: 6,
			body: `
					When multistep passes
					Then multistep passes`,
			suitePasses: true,
			log: `
					>>>> Before suite
					>> Before scenario "test"
					Before step "multistep passes"
					Before step "step passes"
					<step action>
					After step "step passes", error: <nil>, status: passed
					Before step "step passes"
					<step action>
					After step "step passes", error: <nil>, status: passed
					After step "multistep passes", error: <nil>, status: passed
					Before step "multistep passes"
					Before step "step passes"
					<step action>
					After step "step passes", error: <nil>, status: passed
					Before step "step passes"
					<step action>
					After step "step passes", error: <nil>, status: passed
					After step "multistep passes", error: <nil>, status: passed
					<< After scenario "test", error: <nil>
					<<<< After suite`,
		},
		{
			name: "ambiguous", afterStepCnt: 1, beforeStepCnt: 1,
			body: `
					Then step is ambiguous`,
			log: `
					>>>> Before suite
					>> Before scenario "test"
					Before step "step is ambiguous"
					After step "step is ambiguous", error: ambiguous step definition, step text: step is ambiguous
		        	            		matches:
		        	            			^step is ambiguous$
		        	            			^step is ambiguous$, status: ambiguous
					<< After scenario "test", error: ambiguous step definition, step text: step is ambiguous
		        	            		matches:
		        	            			^step is ambiguous$
		        	            			^step is ambiguous$
					<<<< After suite`,
		},
		{
			name: "ambiguous nested steps", afterStepCnt: 1, beforeStepCnt: 1,
			body: `
				Then multistep has ambiguous`,
			log: `
				>>>> Before suite
				>> Before scenario "test"
				Before step "multistep has ambiguous"
				After step "multistep has ambiguous", error: ambiguous step definition, step text: step is ambiguous
            	            		matches:
            	            			^step is ambiguous$
            	            			^step is ambiguous$, status: ambiguous
				<< After scenario "test", error: ambiguous step definition, step text: step is ambiguous
            	            		matches:
            	            			^step is ambiguous$
            	            			^step is ambiguous$
				<<<< After suite`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			afterScenarioCnt := 0
			beforeScenarioCnt := 0

			afterStepCnt := 0
			beforeStepCnt := 0

			var log string

			suite := TestSuite{
				TestSuiteInitializer: func(suiteContext *TestSuiteContext) {
					suiteContext.BeforeSuite(func() {
						log += fmt.Sprintln(">>>> Before suite")
					})

					suiteContext.AfterSuite(func() {
						log += fmt.Sprintln("<<<< After suite")
					})
				},
				ScenarioInitializer: func(s *ScenarioContext) {
					s.Before(func(ctx context.Context, sc *Scenario) (context.Context, error) {
						log += fmt.Sprintf(">> Before scenario %q\n", sc.Name)
						beforeScenarioCnt++

						return ctx, nil
					})

					s.After(func(ctx context.Context, sc *Scenario, err error) (context.Context, error) {
						log += fmt.Sprintf("<< After scenario %q, error: %v\n", sc.Name, err)
						afterScenarioCnt++

						return ctx, nil
					})

					s.StepContext().Before(func(ctx context.Context, st *Step) (context.Context, error) {
						log += fmt.Sprintf("Before step %q\n", st.Text)
						beforeStepCnt++

						return ctx, nil
					})

					s.StepContext().After(func(ctx context.Context, st *Step, status StepResultStatus, err error) (context.Context, error) {
						log += fmt.Sprintf("After step %q, error: %v, status: %s\n", st.Text, err, status.String())
						afterStepCnt++

						return ctx, nil
					})

					s.Step("^step fails$", func() error {
						return errors.New("oops")
					})

					s.Step("^step skips scenario$", func() error {
						return ErrSkip
					})

					s.Step("^step passes$", func() {
						log += "<step action>\n"
					})

					s.Step("^multistep passes$", func() Steps {
						return Steps{"step passes", "step passes"}
					})

					s.Step("pending", func() error {
						return ErrPending
					})

					s.Step("^step is ambiguous$", func() {
						log += "<step action>\n"
					})
					s.Step("^step is ambiguous$", func() {
						log += "<step action>\n"
					})
					s.Step("^multistep has ambiguous$", func() Steps {
						return Steps{"step is ambiguous"}
					})

				},
				Options: &Options{
					Format:   "pretty",
					Strict:   !tc.noStrict,
					NoColors: true,
					FeatureContents: []Feature{
						{
							Name: tc.name,
							Contents: []byte(utils.TrimAllLines(`
								Feature: test
								Scenario: test
								` + tc.body)),
						},
					},
				},
			}

			suitePasses := suite.Run() == ExitSuccess
			assert.Equal(t, tc.suitePasses, suitePasses)
			assert.Equal(t, 1, afterScenarioCnt)
			assert.Equal(t, 1, beforeScenarioCnt)
			assert.Equal(t, tc.afterStepCnt, afterStepCnt)
			assert.Equal(t, tc.beforeStepCnt, beforeStepCnt)
			assert.Equal(t, utils.TrimAllLines(tc.log), utils.TrimAllLines(log), log)
		})
	}
}

func okStep() error {
	return nil
}

func TestPrintsStepDefinitions(t *testing.T) {
	var buf bytes.Buffer
	w := colors.Uncolored(NopCloser(&buf))
	s := suite{}
	ctx := ScenarioContext{suite: &s}

	steps := []string{
		"^passing step$",
		`^with name "([^"])"`,
	}

	for _, step := range steps {
		ctx.Step(step, okStep)
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
	w := colors.Uncolored(NopCloser(&buf))
	s := &suite{}

	printStepDefinitions(s.steps, w)

	out := strings.TrimSpace(buf.String())
	assert.Equal(t, "there were no contexts registered, could not find any step definition..", out)
}

func Test_FailsOrPassesBasedOnStrictModeWhenHasPendingSteps(t *testing.T) {
	const path = "any.feature"

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(basicGherkinFeature), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)

	gd.Uri = path
	ft := models.Feature{GherkinDocument: gd}
	ft.Pickles = gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	var beforeScenarioFired, afterScenarioFired int

	r := runner{
		fmt:      formatters.ProgressFormatterFunc("progress", NopCloser(ioutil.Discard)),
		features: []*models.Feature{&ft},
		testSuiteInitializer: func(ctx *TestSuiteContext) {
			ctx.ScenarioContext().Before(func(ctx context.Context, sc *Scenario) (context.Context, error) {
				beforeScenarioFired++
				return ctx, nil
			})

			ctx.ScenarioContext().After(func(ctx context.Context, sc *Scenario, err error) (context.Context, error) {
				afterScenarioFired++
				return ctx, nil
			})
		},
		scenarioInitializer: func(ctx *ScenarioContext) {
			ctx.Step(`^one$`, func() error { return nil })
			ctx.Step(`^two$`, func() error { return ErrPending })
		},
		testingT: t,
	}

	r.storage = storage.NewStorage()
	r.storage.MustInsertFeature(&ft)
	for _, pickle := range ft.Pickles {
		r.storage.MustInsertPickle(pickle)
	}

	failed := r.concurrent(1)
	require.False(t, r.testingT.Failed())
	require.False(t, failed)
	assert.Equal(t, 1, beforeScenarioFired)
	assert.Equal(t, 1, afterScenarioFired)

	// avoid t is Failed because this testcase Failed
	r.testingT = nil
	r.strict = true
	failed = r.concurrent(1)
	require.True(t, failed)
	assert.Equal(t, 2, beforeScenarioFired)
	assert.Equal(t, 2, afterScenarioFired)
}

func Test_FailsOrPassesBasedOnStrictModeWhenHasUndefinedSteps(t *testing.T) {
	const path = "any.feature"

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(basicGherkinFeature), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)

	gd.Uri = path
	ft := models.Feature{GherkinDocument: gd}
	ft.Pickles = gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	r := runner{
		fmt:      formatters.ProgressFormatterFunc("progress", NopCloser(ioutil.Discard)),
		features: []*models.Feature{&ft},
		scenarioInitializer: func(ctx *ScenarioContext) {
			ctx.Step(`^one$`, func() error { return nil })
			// two - is undefined
		},
	}

	r.storage = storage.NewStorage()
	r.storage.MustInsertFeature(&ft)
	for _, pickle := range ft.Pickles {
		r.storage.MustInsertPickle(pickle)
	}

	failed := r.concurrent(1)
	require.False(t, failed)

	r.strict = true
	failed = r.concurrent(1)
	require.True(t, failed)
}

func Test_ShouldFailOnError(t *testing.T) {
	const path = "any.feature"

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(basicGherkinFeature), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)

	gd.Uri = path
	ft := models.Feature{GherkinDocument: gd}
	ft.Pickles = gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	r := runner{
		fmt:      formatters.ProgressFormatterFunc("progress", NopCloser(ioutil.Discard)),
		features: []*models.Feature{&ft},
		scenarioInitializer: func(ctx *ScenarioContext) {
			ctx.Step(`^two$`, func() error { return fmt.Errorf("error") })
			ctx.Step(`^one$`, func() error { return nil })
		},
	}

	r.storage = storage.NewStorage()
	r.storage.MustInsertFeature(&ft)
	for _, pickle := range ft.Pickles {
		r.storage.MustInsertPickle(pickle)
	}

	failed := r.concurrent(1)
	require.True(t, failed)
}

func Test_FailsWithUnknownFormatterOptionError(t *testing.T) {
	stderr, closer := bufErrorPipe(t)
	defer closer()
	defer stderr.Close()

	opts := Options{
		Format: "unknown",
		Paths:  []string{"features/load:6"},
		Output: NopCloser(ioutil.Discard),
	}

	status := TestSuite{
		Name:                "fails",
		ScenarioInitializer: func(_ *ScenarioContext) {},
		Options:             &opts,
	}.Run()

	require.Equal(t, ExitOptionError, status)

	closer()

	b, err := ioutil.ReadAll(stderr)
	require.NoError(t, err)

	out := strings.TrimSpace(string(b))
	assert.Contains(t, out, `unregistered formatter name: "unknown", use one of`)
}

func Test_FailsWithOptionErrorWhenLookingForFeaturesInUnavailablePath(t *testing.T) {
	stderr, closer := bufErrorPipe(t)
	defer closer()
	defer stderr.Close()

	opts := Options{
		Format: "progress",
		Paths:  []string{"unavailable"},
		Output: NopCloser(ioutil.Discard),
	}

	status := TestSuite{
		Name:                "fails",
		ScenarioInitializer: func(_ *ScenarioContext) {},
		Options:             &opts,
	}.Run()

	require.Equal(t, ExitOptionError, status)

	closer()

	b, err := ioutil.ReadAll(stderr)
	require.NoError(t, err)

	out := strings.TrimSpace(string(b))
	assert.Equal(t, `feature path "unavailable" is not available`, out)
}

func Test_ByDefaultRunsFeaturesPath(t *testing.T) {
	opts := Options{
		Format: "progress",
		Output: NopCloser(ioutil.Discard),
		Strict: true,
	}

	status := TestSuite{
		Name:                "fails",
		ScenarioInitializer: func(_ *ScenarioContext) {},
		Options:             &opts,
	}.Run()

	// should fail in strict mode due to undefined steps
	assert.Equal(t, ExitFailure, status)

	opts.Strict = false
	status = TestSuite{
		Name:                "succeeds",
		ScenarioInitializer: func(_ *ScenarioContext) {},
		Options:             &opts,
	}.Run()

	// should succeed in non strict mode due to undefined steps
	assert.Equal(t, ExitSuccess, status)
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

const sampleFeature = `
		Feature: scenarios should run in different order if seed is used
		
		  Scenario Outline: Some examples
			# Need enough examples to cause the pseudo-randomness to show up

			Given some step <value>
		
			Examples:
			| value |
			| hello |
			| 1 |
			| 2 |
			| 3 |
			| 4 |
			| 5 |
		`

var sampleFeatures = []Feature{{Name: "test.feature", Contents: []byte(sampleFeature)}}

func Test_RandomizeRun_WithStaticSeed(t *testing.T) {
	const noRandomFlag = 0
	const noConcurrencyFlag = 1
	const formatter = "pretty"

	fmtOutputScenarioInitializer := func(ctx *ScenarioContext) {
		ctx.Step(`^.*`, func() {})
	}

	expectedStatus, expectedOutput := testRun(t,
		fmtOutputScenarioInitializer,
		formatter, noConcurrencyFlag,
		noRandomFlag,
		nil,
		sampleFeatures,
	)

	const staticSeed int64 = 1
	actualStatus, actualOutput := testRun(t,
		fmtOutputScenarioInitializer,
		formatter, noConcurrencyFlag,
		staticSeed,
		nil,
		sampleFeatures,
	)

	actualSeed := parseSeed(actualOutput)
	assert.Equal(t, staticSeed, actualSeed)

	// Removes "Randomized with seed: <seed>" part of the output
	actualOutputSplit := strings.Split(actualOutput, "\n")
	actualOutputSplit = actualOutputSplit[:len(actualOutputSplit)-2]
	actualOutputReduced := strings.Join(actualOutputSplit, "\n")

	assert.Equal(t, expectedStatus, actualStatus)
	assert.NotEqual(t, expectedOutput, actualOutputReduced, "expected the natural and seeded order to be different")

	assertOutput(t, formatter, expectedOutput, actualOutputReduced)
}

func Test_RandomizeRun_RerunWithSeed(t *testing.T) {
	const createRandomSeedFlag = -1
	const noConcurrencyFlag = 1
	const formatter = "pretty"

	fmtOutputScenarioInitializer := func(ctx *ScenarioContext) {
		ctx.Step(`^.*`, func() {})
	}

	expectedStatus, expectedOutput := testRun(t,
		fmtOutputScenarioInitializer,
		formatter, noConcurrencyFlag,
		createRandomSeedFlag,
		nil,
		sampleFeatures,
	)

	expectedSeed := parseSeed(expectedOutput)
	assert.NotZero(t, expectedSeed)

	actualStatus, actualOutput := testRun(t,
		fmtOutputScenarioInitializer,
		formatter, noConcurrencyFlag,
		expectedSeed,
		nil,
		sampleFeatures,
	)

	actualSeed := parseSeed(actualOutput)

	assert.Equal(t, expectedSeed, actualSeed)
	assert.Equal(t, expectedStatus, actualStatus)
	assert.Equal(t, expectedOutput, actualOutput)
}

func Test_FormatOutputRun(t *testing.T) {
	const noRandomFlag = 0
	const noConcurrencyFlag = 1
	const formatter = "junit"

	fmtOutputScenarioInitializer := func(ctx *ScenarioContext) {
		ctx.Step(`^.*$`, func() {})
	}

	//  first collect the output via a memory buffer and use this to verify the file out below
	expectedStatus, expectedOutput := testRun(t,
		fmtOutputScenarioInitializer,
		formatter,
		noConcurrencyFlag,
		noRandomFlag,
		nil,
		sampleFeatures,
	)

	// run again with file output
	dir := filepath.Join(os.TempDir(), t.Name())
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "result.xml")

	actualStatus, actualOutput := testRun(t,
		fmtOutputScenarioInitializer,
		formatter+":"+file, noConcurrencyFlag,
		noRandomFlag,
		nil,
		sampleFeatures,
	)

	result, err := ioutil.ReadFile(file)
	require.NoError(t, err)
	actualOutputFromFile := string(result)

	assert.Equal(t, expectedStatus, actualStatus)
	assert.Empty(t, actualOutput)
	assert.Equal(t, expectedOutput, actualOutputFromFile)
}

func Test_FormatOutputRun_OutputFileError(t *testing.T) {
	const noRandomFlag = 0
	const noConcurrencyFlag = 1
	const formatter = "junit"

	fmtOutputScenarioInitializer := func(ctx *ScenarioContext) {
		ctx.Step(`^.*$`, func() {})
	}

	// locate the output file in a temp dir that we won't actually create
	dir := filepath.Join(os.TempDir(), t.Name())
	file := filepath.Join(dir, "result.xml")

	// !! NOTE !!
	// This test is intended to verify the fact that the library fails to open the file and should
	// ideally verify the user error...
	//   couldn't create file with name: "/tmp/Test_FormatOutputRun_Error/result.xml", error: open /tmp/Test_FormatOutputRun_Error/result.xml: no such file or directory
	// ... however that error gets sent direct to stdout, so there not much we can verify here that's actually a useful test.
	// Ideally, this code would capture the error.
	// Todo - find all the direct stdout/err writes and put them via a Writer that we can mock if needed.
	actualStatus, actualOutput := testRun(t,
		fmtOutputScenarioInitializer,
		formatter+":"+file, noConcurrencyFlag,
		noRandomFlag,
		nil,
		sampleFeatures)

	expectedStatus, expectedOutput := ExitOptionError, ""

	assert.Equal(t, expectedStatus, actualStatus)
	assert.Equal(t, expectedOutput, actualOutput)

	_, err := ioutil.ReadFile(file)
	assert.Error(t, err)
}

// This test runs the tests sequentially and in parallel, and expects the passing and failing tests to be the same
func Test_FormatterConcurrencyRun(t *testing.T) {
	formatters := []string{
		"progress",
		"junit",
		"pretty",
		"events",
		"cucumber",
	}

	featurePaths := []string{"internal/formatters/formatter-tests/features"}

	const concurrency = 100
	const noRandomFlag = 0
	const noConcurrency = 1

	// this is just a few dummy handlers to satisfy the needs of a few scenarios.
	// the real initialiser is in fmt_output_test
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
				expectedStatus, expectedOutput := runWithResults(t,
					fmtOutputScenarioInitializer,
					formatter, noConcurrency,
					noRandomFlag, featurePaths, nil,
				)
				actualStatus, actualOutput := runWithResults(t,
					fmtOutputScenarioInitializer,
					formatter, concurrency,
					noRandomFlag, featurePaths, nil,
				)

				passes := countResultsByStatus(expectedStatus.storage, StepPassed)
				fails := countResultsByStatus(expectedStatus.storage, StepFailed)
				if passes == 0 {
					t.Errorf("for this test to be valid then some scenarios need at least some pass, but got %v passes and %v fails", passes, fails)
				}

				assert.Equal(t, expectedStatus.exitCode, actualStatus.exitCode)
				assertOutput(t, formatter, expectedOutput, actualOutput)
			},
		)
	}
}

func testRun(
	t *testing.T,
	scenarioInitializer func(*ScenarioContext),
	format string,
	concurrency int,
	randomSeed int64,
	featurePaths []string,
	features []Feature,
) (int, string) {
	result, actualOutput := runWithResults(t, scenarioInitializer, format, concurrency, randomSeed, featurePaths, features)
	return result.exitCode, actualOutput
}

func runWithResults(t *testing.T,
	scenarioInitializer func(*ScenarioContext),
	format string, concurrency int, randomSeed int64, featurePaths []string, features []Feature) (RunResult, string) {

	t.Helper()

	opts := Options{
		Format:          format,
		Paths:           featurePaths,
		FeatureContents: features,
		Concurrency:     concurrency,
		Randomize:       randomSeed,
	}

	result, actualOutput := testRunWithOptions(t, opts, scenarioInitializer)
	return result, actualOutput
}

func countResultsByStatus(storage *storage.Storage, status models.StepResultStatus) int {
	actual := []string{}

	for _, st := range storage.MustGetPickleStepResultsByStatus(status) {
		pickleStep := storage.MustGetPickleStep(st.PickleStepID)
		actual = append(actual, pickleStep.Text)
	}
	return len(actual)
}

func testRunWithOptions(
	t *testing.T,
	opts Options,
	scenarioInitializer func(*ScenarioContext),
) (RunResult, string) {
	t.Helper()

	output := new(bytes.Buffer)

	opts.Output = NopCloser(output)
	opts.NoColors = true

	testSuite := TestSuite{
		Name:                "succeed",
		ScenarioInitializer: scenarioInitializer,
		Options:             &opts,
	}

	result := testSuite.RunWithResult()

	actual, err := ioutil.ReadAll(output)
	require.NoError(t, err)

	return result, string(actual)
}

func assertOutput(t *testing.T, formatter string, expected, actual string) {
	switch formatter {
	case "cucumber", "junit", "pretty", "events":
		expectedRows := strings.Split(expected, "\n")
		actualRows := strings.Split(actual, "\n")
		ok := assert.ElementsMatch(t, expectedRows, actualRows)
		if !ok {
			utils.VDiffLists(expectedRows, actualRows)
		}
	case "progress":
		expectedOutput := parseProgressOutput(expected)
		actualOutput := parseProgressOutput(actual)

		assert.Equal(t, expectedOutput.passed, actualOutput.passed)
		assert.Equal(t, expectedOutput.skipped, actualOutput.skipped)
		assert.Equal(t, expectedOutput.failed, actualOutput.failed)
		assert.Equal(t, expectedOutput.undefined, actualOutput.undefined)
		assert.Equal(t, expectedOutput.pending, actualOutput.pending)
		assert.Equal(t, expectedOutput.noOfStepsPerRow, actualOutput.noOfStepsPerRow)
		ok := assert.ElementsMatch(t, expectedOutput.bottomRows, actualOutput.bottomRows)
		if !ok {
			utils.VDiffLists(expectedOutput.bottomRows, actualOutput.bottomRows)
		}
	default:
		panic("unknown formatter: " + formatter)
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

func Test_TestSuite_RetreiveFeatures(t *testing.T) {
	tests := map[string]struct {
		fsys fs.FS

		expFeatures int
	}{
		"standard features": {
			fsys: fstest.MapFS{
				"features/test.feature": {
					Data: []byte(`Feature: test retrieve features
  To test the feature
  I must use this feature

  Scenario: Test function RetrieveFeatures
    Given I create a TestSuite
	When I call TestSuite.RetrieveFeatures
	Then I should have one feature`),
				},
			},
			expFeatures: 1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			features, err := TestSuite{
				Name:    "succeed",
				Options: &Options{FS: test.fsys},
			}.RetrieveFeatures()

			assert.NoError(t, err)
			assert.Equal(t, test.expFeatures, len(features))
		})
	}
}

// do ctx get cancelled across threads?
func Test_ContextShouldBeCancelledAfterScenarioCompletion(t *testing.T) {
	// two scenarios and concurrency is set to 2 so we expect two distinct ctx and cancel events
	numberOfScenarios := 2
	capturedCtx := make(chan context.Context, numberOfScenarios)

	suite := TestSuite{
		ScenarioInitializer: func(scenarioContext *ScenarioContext) {
			scenarioContext.When(`^foo$`, func() {})
			scenarioContext.After(func(ctx context.Context, sc *Scenario, err error) (context.Context, error) {
				// capture the context so the mainline can check it got cancelled
				capturedCtx <- ctx

				return ctx, nil
			})
		},
		Options: &Options{
			Format:      "pretty",
			Concurrency: numberOfScenarios,
			TestingT:    t,
			FeatureContents: []Feature{
				{
					Name: "Scenario Context Cancellation",
					Contents: []byte(`
Feature: dummy
  Scenario: 1: Context should be cancelled by the end of scenario
    When foo

  Scenario: 2: Context should be cancelled by the end of scenario
    When foo
`),
				},
			},
		},
	}

	require.Equal(t, ExitSuccess, suite.Run(), "non-zero status returned, failed to run feature tests")

	// the ctx should have been cancelled by the time godog returns so we should be able to check immediately
	for i := 0; i < numberOfScenarios; i++ {
		select {
		case ctx := <-capturedCtx:
			fmt.Printf("ok: ctx %d found\n", i)
			// now wait for it to have been cancelled - should be immediate
			doneFuture := ctx.Done()
			select {
			case <-doneFuture:
				fmt.Printf("ok: ctx %d cancelled\n", i)
			default:
				assert.Fail(t, "context was not cancelled")
			}

		default:
			assert.Fail(t, fmt.Sprintf("timed out waiting for %d contexts to be captured", numberOfScenarios))
		}

	}
}

// does a newly created child ctx wrapper get cancelled when the scenario completes?
func Test_ChildContextShouldBeCancelledAfterScenarioCompletion(t *testing.T) {
	var childContext context.Context
	suite := TestSuite{
		ScenarioInitializer: func(scenarioContext *ScenarioContext) {
			scenarioContext.When(`^foo$`, func() {})
			scenarioContext.After(func(ctx context.Context, sc *Scenario, err error) (context.Context, error) {
				type ctxKey string
				childContext = context.WithValue(ctx, ctxKey("child"), true)

				return childContext, nil
			})
		},
		Options: &Options{
			Format:         "pretty",
			TestingT:       t,
			DefaultContext: context.Background(),
			FeatureContents: []Feature{
				{
					Name: "Scenario Context Cancellation",
					Contents: []byte(`
Feature: dummy
  Scenario: Context should be cancelled by the end of scenario
    When foo
`),
				},
			},
		},
	}

	require.Equal(t, ExitSuccess, suite.Run(), "non-zero status returned, failed to run feature tests")

	// the ctx should have been cancelled before godog returns so we should be able to check immediately
	select {
	case <-childContext.Done(): // pass
	default:
		assert.Fail(t, "child context was not cancelled")
	}
}
