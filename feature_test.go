package godog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	gherkin "github.com/cucumber/gherkin/go/v26"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/cucumber/godog/internal/storage"
	"github.com/cucumber/godog/internal/utils"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/cucumber/godog/internal/formatters"
	"github.com/cucumber/godog/internal/models"
	perr "github.com/pkg/errors"
)

func Test_AllFeaturesRun_AsSubtests(t *testing.T) {
	runOptionalSubtest(t, true)
}

func Test_AllFeaturesRun_NotAsSubtests(t *testing.T) {
	runOptionalSubtest(t, false)
}

// when running as subtests then the trace (and also intelli) will show each scenario distinctly.
// otherwise the telemetry is just one big blob.
func runOptionalSubtest(t *testing.T, subtest bool) {
	const concurrency = 1
	const noRandomFlag = 0
	const format = "progress"

	const expected = `...................................................................... 70
...................................................................... 140
...................................................................... 210
...................................................................... 280
...................................................................... 350
........                                                               358


93 scenarios (93 passed)
358 steps (358 passed)
0s
`
	t.Helper()

	var subtestT *testing.T
	if subtest {
		subtestT = t
	}

	output := new(bytes.Buffer)

	suite := godog.TestSuite{
		Name:                "succeed",
		ScenarioInitializer: InitializeScenarioOuter,
		Options: &godog.Options{
			Strict: true,
			Format: format,
			//Tags:   "@john7 && ~@ignore",
			Tags:        "~@ignore",
			Concurrency: concurrency,
			Paths:       []string{"features"},
			Randomize:   noRandomFlag,
			TestingT:    subtestT, // Optionally - Pass the testing instance to godog so that tests run as subtests
			Output:      godog.NopCloser(output),
			NoColors:    true,
		},
	}

	actualStatus := suite.Run()

	actualOutput, err := io.ReadAll(output)
	require.NoError(t, err)
	println(string(actualOutput))

	assert.Equal(t, godog.ExitSuccess, actualStatus)

	if expected != string(actualOutput) {
		fmt.Printf("Actual output:\n%s\n", string(actualOutput))
	}
	assert.Equal(t, expected, string(actualOutput))
}

func Test_RunsWithStrictAndNonStrictMode(t *testing.T) {
	featureContents := []godog.Feature{
		{
			Name: "Test_RunsWithStrictAndNonStrictMode.feature",
			Contents: []byte(`
Feature: simple undefined feature
  Scenario: simple undefined scenario
    Given simple undefined step
			`),
		},
	}

	// running with strict means it will not ignore faults due to "undefined"
	opts := godog.Options{
		Format:          "progress",
		Output:          godog.NopCloser(ioutil.Discard),
		Strict:          true,
		FeatureContents: featureContents,
	}

	status := godog.TestSuite{
		Name:                "fails",
		ScenarioInitializer: func(_ *godog.ScenarioContext) {},
		Options:             &opts,
	}.Run()

	// should fail in strict mode due to undefined steps
	assert.Equal(t, godog.ExitFailure, status)

	// running with non-strict means it ignores the faults due to "undefined"
	opts.Strict = false
	status = godog.TestSuite{
		Name:                "succeeds",
		ScenarioInitializer: func(_ *godog.ScenarioContext) {},
		Options:             &opts,
	}.Run()

	// should succeed in non-strict mode because undefined is ignored
	assert.Equal(t, godog.ExitSuccess, status)
}

// FIXED ME - NO LONGER DEPENDENT ON HUMONGOUS STEPS AND STILL COMPLETELY VALID !!
func Test_RunsWithFeatureContentsAndPathsOptions(t *testing.T) {

	tempFeatureDir := filepath.Join(os.TempDir(), "features")
	if err := os.MkdirAll(tempFeatureDir, 0755); err != nil {
		t.Fatalf("cannot create temp dir: %v: %v", tempFeatureDir, err)
	}

	simpleFileFeature := `
			Feature: simple content feature
				  Scenario: simple content scenario
					Given simple content step
	`

	featureFile := filepath.Join(tempFeatureDir, "simple.feature")
	if err := os.WriteFile(featureFile, []byte(simpleFileFeature), 0644); err != nil {
		t.Fatalf("cannot write to: %v: %v", featureFile, err)
	}

	simpleContentFeature := []godog.Feature{
		{
			Name: "Test_RunsWithFeatureContentsAndPathsOptions.feature",
			Contents: []byte(`
				Feature: simple file feature
				  Scenario: simple file scenario
					Given simple file step
			`),
		},
	}

	opts := godog.Options{
		Format:          "progress",
		Output:          godog.NopCloser(io.Discard),
		Paths:           []string{tempFeatureDir},
		FeatureContents: simpleContentFeature,
	}
	contentStepCalled := false
	fileStepCalled := false

	suite := godog.TestSuite{
		Name: "succeeds",
		ScenarioInitializer: func(sc *godog.ScenarioContext) {
			sc.Step("^simple content step$", func() {
				contentStepCalled = true
			})
			sc.Step("^simple file step$", func() {
				fileStepCalled = true
			})
		},
		Options: &opts,
	}

	status := suite.Run()

	assert.Equal(t, godog.ExitSuccess, status)
	assert.True(t, contentStepCalled, "step in content was not called")
	assert.True(t, fileStepCalled, "step in file was not called")
}

// This function has to exist to make the CLI part of the build work: go  run ./cmd/godog -f progress
func InitializeScenario(ctx *godog.ScenarioContext) {
	InitializeScenarioOuter(ctx)
}

// InitializeScenario provides steps for godog suite execution and
// can be used for meta-testing of godog features/steps themselves.
//
// Beware, steps or their definitions might change without backward
// compatibility guarantees. A typical user of the godog library should never
// need this, rather it is provided for those developing add-on libraries for godog.
//
// For an example of how to use, see godog's own `features/` and `suite_test.go`.
func InitializeScenarioOuter(ctx *godog.ScenarioContext) {

	//var depth = 1

	tempDir, err := os.MkdirTemp(os.TempDir(), "tests_")
	if err != nil {
		panic(fmt.Errorf("cannot create temp dir: %w", err))
	}

	tc := &godogFeaturesScenarioOuter{
		tempDir: tempDir + "/",
		//scenarioContext: ctx,
		//out1: out,
	}

	//ctx.Before(tc.ResetBeforeEachScenario)

	//ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	//	fmt.Printf("%-2s HOOK BEFORE SCENARIO: %v\n", strings.Repeat(">", depth), sc.Name)
	//	depth++
	//
	//	return ctx, nil
	//})
	//
	//ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
	//	depth--
	//	fmt.Printf("%-2s HOOK AFTER SCENARIO: %v\n", strings.Repeat("<", depth), sc.Name)
	//	if err != nil {
	//		fmt.Printf("%-2s    ERROR: %v\n", strings.Repeat("!", depth), err.Error())
	//	}
	//
	//	return ctx, nil
	//})

	//ctx.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
	//	fmt.Printf("%-2s HOOK BEFORE STEP: %v\n", strings.Repeat(">", depth), st.Text)
	//	depth++
	//	return ctx, nil
	//})
	//
	//ctx.StepContext().After(func(ctx context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
	//	depth--
	//	fmt.Printf("%-2s HOOK AFTER STEP: %v\n", strings.Repeat("<", depth), st.Text)
	//	if err != nil {
	//		fmt.Printf("%-2s ERROR: %v\n", strings.Repeat("!", depth), err.Error())
	//	}
	//	return ctx, nil
	//})

	//ctx.Step(`^(a background step is defined)$`, tc.backgroundStepIsDefined)
	//ctx.Step(`^step '(.*)' should have been executed`, tc.stepShouldHaveBeenExecuted)
	//
	ctx.Step(`^a feature file at "([^"]*)":$`, tc.writeFeatureFile)
	ctx.Step(`^(?:a )?feature path "([^"]*)"$`, tc.featurePath)
	//ctx.Step(`^I parse features$`, tc.parseFeatures)
	//ctx.Step(`^I'm listening to suite events$`, tc.iAmListeningToSuiteEvents) // DOES NOT MAKE SENSE??
	ctx.Step(`^I run feature suite$`, tc.iRunFeatureSuite)
	ctx.Step(`^I run feature suite in Strict mode$`, tc.iRunFeatureSuiteStrict) // FIXME - use this
	ctx.Step(`^I run feature suite with tags "([^"]*)"$`, tc.iRunFeatureSuiteWithTags)
	ctx.Step(`^I run feature suite with formatter "([^"]*)"$`, tc.iRunFeatureSuiteWithFormatter)
	//ctx.Step(`^(?:I )(allow|disable) variable injection`, tc.iSetVariableInjectionTo)
	ctx.Step(`^(?:a )?feature "([^"]*)"(?: file)?:$`, tc.aFeatureFile)
	ctx.Step(`^the suite should have (passed|failed)$`, tc.theSuiteShouldHave)
	//
	//ctx.Step(`^I should have ([\d]+) features? files?:$`, tc.iShouldHaveNumFeatureFiles)
	//ctx.Step(`^I should have ([\d]+) scenarios? registered$`, tc.numScenariosRegistered)
	//ctx.Step(`^there (was|were) ([\d]+) "([^"]*)" events? fired$`, tc.thereWereNumEventsFired)
	//ctx.Step(`^there was event triggered before scenario "([^"]*)"$`, tc.thereWasEventTriggeredBeforeScenario)
	//ctx.Step(`^these events had to be fired for a number of times:$`, tc.theseEventsHadToBeFiredForNumberOfTimes)
	//
	//ctx.Step(`^(?:a )?failing step`, tc.aFailingStep)
	//ctx.Step(`^(.*should not be called)`, tc.aStepThatShouldNotHaveBeenCalled)
	//ctx.Step(`^this step should fail`, tc.aFailingStep)
	ctx.Step(`^the following steps? should be (passed|failed|skipped|undefined|pending):`, tc.followingStepsShouldHave)
	ctx.Step(`^only the following steps? should have run and should be (passed|failed|skipped|undefined|pending):`, tc.onlyFollowingStepsShouldHave)

	ctx.Step(`^the trace should be:$`, tc.theTraceShouldBe)

	ctx.Step(`^the undefined step snippets should be:$`, tc.theUndefinedStepSnippetsShouldBe)
	//
	//// event stream
	ctx.Step(`^the following events should be fired:$`, tc.thereShouldBeEventsFired)
	//
	//// lt
	//ctx.Step(`^savybių aplankas "([^"]*)"$`, tc.featurePath)
	//ctx.Step(`^aš išskaitau savybes$`, tc.parseFeatures)
	//ctx.Step(`^aš turėčiau turėti ([\d]+) savybių failus:$`, tc.iShouldHaveNumFeatureFiles)
	//
	//ctx.Step(`^(?:a )?pending step$`, func() error {
	//	return godog.ErrPending
	//})
	//ctx.Step(`^(?:a )?passing step$`, func() error {
	//	return nil
	//})
	//ctx.Given(`^(?:a )?given step$`, func() error {
	//	return nil
	//})
	//ctx.When(`^(?:a )?when step$`, func() error {
	//	return nil
	//})
	//ctx.Then(`^(?:a )?then step$`, func() error {
	//	return nil
	//})
	ctx.Step(`^the rendered json will be as follows:$`, tc.theRenderedJSONWillBe)
	ctx.Step(`^the rendered events will be as follows:$`, tc.theRenderedEventsWillBe)
	ctx.Step(`^the rendered xml will be as follows:$`, tc.theRenderedXMLWillBe)
	ctx.Step(`^the rendered output will be as follows:$`, tc.theRenderedOutputWillBe)
	//

	//ctx.Step(`^(?:a )?failing multistep$`, func() godog.Steps {
	//	return godog.Steps{"passing step", "failing step"}
	//})
	//
	//ctx.Step(`^(?:a |an )?undefined multistep$`, func() godog.Steps {
	//	return godog.Steps{"passing step", "undefined step", "passing step"}
	//})
	//
	//ctx.Then(`^(?:a |an )?undefined multistep using 'then' function$`, func() godog.Steps {
	//	return godog.Steps{"given step", "undefined step", "then step"}
	//})
	//
	//ctx.Step(`^(?:a )?passing multistep$`, func() godog.Steps {
	//	return godog.Steps{"passing step", "passing step", "passing step"}
	//})
	//
	//ctx.Then(`^(?:a )?passing multistep using 'then' function$`, func() godog.Steps {
	//	return godog.Steps{"given step", "when step", "then step"}
	//})
	//
	//ctx.Step(`^(?:a )?failing nested multistep$`, func() godog.Steps {
	//	return godog.Steps{"passing step", "passing multistep", "failing multistep"}
	//})
	//// Default recovery step
	//ctx.Step(`Ignore.*`, func() error {
	//	return nil
	//})
	//
	ctx.Step(`^call func\(\*godog\.DocString\) with '(.*)':$`, func(str string, docstring *godog.DocString) error {
		if docstring.Content != str {
			return fmt.Errorf("expected %q, got %q", str, docstring.Content)
		}
		return nil
	})
	ctx.Step(`^call func\(string\) with '(.*)':$`, func(str string, docstring string) error {
		if docstring != str {
			return fmt.Errorf("expected %q, got %q", str, docstring)
		}
		return nil
	})
	//
	//ctx.Step(`^passing step without return$`, func() {})
	//
	//ctx.Step(`^having correct context$`, func(ctx context.Context) (context.Context, error) {
	//	if ctx.Value(ctxKey("BeforeScenario")) == nil {
	//		return ctx, errors.New("missing BeforeScenario in context")
	//	}
	//
	//	if ctx.Value(ctxKey("BeforeStep")) == nil {
	//		return ctx, errors.New("missing BeforeStep in context")
	//	}
	//
	//	if ctx.Value(ctxKey("StepState")) == nil {
	//		return ctx, errors.New("missing StepState in context")
	//	}
	//
	//	return context.WithValue(ctx, ctxKey("Step"), true), nil
	//})
	//
	//ctx.Step(`^adding step state to context$`, func(ctx context.Context) context.Context {
	//	return context.WithValue(ctx, ctxKey("StepState"), true)
	//})
	//
	//ctx.Step(`^I return a context from a step$`, tc.iReturnAContextFromAStep)
	//ctx.Step(`^I should see the context in the next step$`, tc.iShouldSeeTheContextInTheNextStep)
	//ctx.Step(`^I can see contexts passed in multisteps$`, func() godog.Steps {
	//	return godog.Steps{
	//		"I return a context from a step",
	//		"I should see the context in the next step",
	//	}
	//})
	//
	//// introduced to test testingT
	ctx.Step(`^testing T (should have|should not have) failed$`, tc.testingTShouldBe)
	//ctx.Step(`^my step (?:fails|skips) the test by calling (FailNow|Fail|SkipNow|Skip) on testing T$`, tc.myStepCallsTFailErrorSkip)
	//ctx.Step(`^my step fails the test by calling (Fatal|Error) on testing T with message "([^"]*)"$`, tc.myStepCallsTErrorFatal)
	//ctx.Step(`^my step fails the test by calling (Fatalf|Errorf) on testing T with message "([^"]*)" and argument "([^"]*)"$`, tc.myStepCallsTErrorfFatalf)
	ctx.Step(`^my step calls Log on testing T with message "([^"]*)"$`, tc.myStepCallsTLog)
	ctx.Step(`^my step calls Logf on testing T with message "([^"]*)" and argument "([^"]*)"$`, tc.myStepCallsTLogf)
	//ctx.Step(`^my step calls testify's assert.Equal with expected "([^"]*)" and actual "([^"]*)"$`, tc.myStepCallsTestifyAssertEqual)
	//ctx.Step(`^my step calls testify's require.Equal with expected "([^"]*)" and actual "([^"]*)"$`, tc.myStepCallsTestifyRequireEqual)
	//ctx.Step(`^my step calls testify's assert.Equal ([0-9]+) times(| with match)$`, tc.myStepCallsTestifyAssertEqualMultipleTimes)
	ctx.Step(`^my step calls godog.Log with message "([^"]*)"$`, tc.myStepCallsDogLog)
	ctx.Step(`^my step calls godog.Logf with message "([^"]*)" and argument "([^"]*)"$`, tc.myStepCallsDogLogf)
	ctx.Step(`^the logged messages should include "([^"]*)"$`, tc.theLoggedMessagesShouldInclude)
	//
	//ctx.StepContext().Before(tc.inject)
}

func InitializeTestSuiteInner(parent *godogFeaturesScenarioOuter) func(ctx *godog.TestSuiteContext) {
	return func(ctx *godog.TestSuiteContext) {

		ctx.BeforeSuite(func() {
			parent.events = append(parent.events, &firedEvent{"BeforeSuite", []interface{}{}})
		})

		ctx.AfterSuite(func() {
			parent.events = append(parent.events, &firedEvent{"AfterSuite", []interface{}{}})
		})
	}
}

func InitializeScenarioInner(parent *godogFeaturesScenarioOuter) func(ctx *godog.ScenarioContext) {

	return func(ctx *godog.ScenarioContext) {

		//var depth = 1

		tc := &godogFeaturesScenarioInner{
			scenarioContext: ctx,
		}

		ctx.Before(tc.ResetBeforeEachScenario)
		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			if tagged(sc.Tags, "@fail_before_scenario") {
				return ctx, fmt.Errorf("failed in before scenario hook")
			}
			return ctx, nil
		})
		ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
			if tagged(sc.Tags, "@fail_after_scenario") {
				return ctx, fmt.Errorf("failed in after scenario hook")
			}
			return ctx, nil
		})

		ctx.Before(func(ctx context.Context, pickle *godog.Scenario) (context.Context, error) {
			parent.events = append(parent.events, &firedEvent{"BeforeScenario", []interface{}{pickle.Name}})

			if ctx.Value(ctxKey("BeforeScenario")) != nil {
				return ctx, errors.New("unexpected BeforeScenario in context (double invocation)")
			}

			return context.WithValue(ctx, ctxKey("BeforeScenario"), pickle.Name), nil
		})

		ctx.After(func(ctx context.Context, pickle *godog.Scenario, err error) (context.Context, error) {
			args := []interface{}{pickle.Name}
			if err != nil {
				args = append(args, err)
			}
			parent.events = append(parent.events, &firedEvent{"AfterScenario", args})

			if ctx.Value(ctxKey("BeforeScenario")) == nil {
				return ctx, errors.New("missing BeforeScenario in context")
			}

			if ctx.Value(ctxKey("AfterStep")) == nil {
				return ctx, errors.New("missing AfterStep in context")
			}

			return context.WithValue(ctx, ctxKey("AfterScenario"), pickle.Name), nil
		})

		ctx.StepContext().Before(func(ctx context.Context, step *godog.Step) (context.Context, error) {
			parent.events = append(parent.events, &firedEvent{"BeforeStep", []interface{}{step.Text}})

			if ctx.Value(ctxKey("BeforeScenario")) == nil {
				return ctx, errors.New("missing BeforeScenario in context")
			}

			// FIXME - THIS IS A SYMPTOM OF THE HOOK ORDERING BUG
			//if ctx.Value(ctxKey("AfterScenario")) != nil {
			//	panic("unexpected premature AfterScenario during AfterStep: " +
			//		ctx.Value(ctxKey("AfterScenario")).(string) +
			//		"\nPreceeding Events...\n  " + strings.Join(parent.events.ToStrings(), "\n  "))
			//}

			return context.WithValue(ctx, ctxKey("BeforeStep"), step.Text), nil
		})

		ctx.StepContext().After(func(ctx context.Context, step *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
			args := []interface{}{step.Text, status}
			if err != nil {
				args = append(args, err)
			}
			parent.events = append(parent.events, &firedEvent{"AfterStep", args})

			if ctx.Value(ctxKey("BeforeScenario")) == nil {
				return ctx, errors.New("missing BeforeScenario in context")
			}

			// FIXME - THIS IS A SYMPTOM OF THE HOOK ORDERING BUG - HACK HACK HACK
			expectPrematureEndOfScenario := status == models.Skipped || status == models.Undefined || step.Text != "with expected \"exp\" and actual \"not\""
			if ctx.Value(ctxKey("AfterScenario")) != nil && !expectPrematureEndOfScenario {
				panic("unexpected premature AfterScenario during AfterStep: " +
					ctx.Value(ctxKey("AfterScenario")).(string) +
					"\nPreceeding Events...\n  " + strings.Join(parent.events.ToStrings(), "\n  "))
			}

			if ctx.Value(ctxKey("BeforeStep")) == nil {
				return ctx, errors.New("missing BeforeStep in context")
			}

			if step.Text == "having correct context" && ctx.Value(ctxKey("Step")) == nil {
				if status != godog.StepSkipped {
					return ctx, fmt.Errorf("unexpected step result status: %s", status)
				}

				return ctx, errors.New("missing Step in context")
			}

			return context.WithValue(ctx, ctxKey("AfterStep"), step.Text), nil
		})

		//
		//ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		//	fmt.Printf("%-2s SCENARIO: %v\n", strings.Repeat(">", depth), sc.Name)
		//	depth++
		//
		//	return ctx, nil
		//})
		//
		//ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		//	depth--
		//	fmt.Printf("%-2s SCENARIO: %v\n", strings.Repeat("<", depth), sc.Name)
		//	if err != nil {
		//		fmt.Printf("%-2s    SCENARIO ERROR: %v\n", strings.Repeat("!", depth), err.Error())
		//	}
		//
		//	return ctx, nil
		//})
		//ctx.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
		//	fmt.Printf("%-2s BEFORE STEP HOOK: %v\n", strings.Repeat(">", depth), st.Text)
		//	depth++
		//	return ctx, nil
		//})
		//
		//ctx.StepContext().After(func(ctx context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
		//	depth--
		//	fmt.Printf("%-2s AFTER STEP HOOK: %v\n", strings.Repeat("<", depth), st.Text)
		//	if err != nil {
		//		fmt.Printf("%-2s STEP ERROR: %v\n", strings.Repeat("!", depth), err.Error())
		//	}
		//	return ctx, nil
		//})
		//
		ctx.Step(`^(a background step is defined)$`, tc.backgroundStepIsDefined)
		ctx.Step(`^step '(.*)' should have been executed`, tc.stepShouldHaveBeenExecuted)
		//
		//ctx.Step(`^(?:a )?feature path "([^"]*)"$`, tc.featurePath)
		//ctx.Step(`^I parse features$`, tc.parseFeatures)
		//ctx.Step(`^I'm listening to suite events$`, tc.iAmListeningToSuiteEvents)
		//ctx.Step(`^I run feature suite$`, tc.iRunFeatureSuite)
		//ctx.Step(`^I run feature suite with tags "([^"]*)"$`, tc.iRunFeatureSuiteWithTags)
		//ctx.Step(`^I run feature suite with formatter "([^"]*)"$`, tc.iRunFeatureSuiteWithFormatter)
		ctx.Step(`^(?:I )(allow|disable) variable injection`, tc.iSetVariableInjectionTo)
		//ctx.Step(`^(?:a )?feature "([^"]*)"(?: file)?:$`, tc.aFeatureFile)
		//ctx.Step(`^the suite should have (passed|failed)$`, tc.theSuiteShouldHave)
		//
		//ctx.Step(`^I should have ([\d]+) features? files?:$`, tc.iShouldHaveNumFeatureFiles)
		//ctx.Step(`^I should have ([\d]+) scenarios? registered$`, tc.numScenariosRegistered)
		//ctx.Step(`^there (was|were) ([\d]+) "([^"]*)" events? fired$`, tc.thereWereNumEventsFired)
		//ctx.Step(`^there was event triggered before scenario "([^"]*)"$`, tc.thereWasEventTriggeredBeforeScenario)
		//ctx.Step(`^these events had to be fired for a number of times:$`, tc.theseEventsHadToBeFiredForNumberOfTimes)
		ctx.Step(`^value2 is twice value1:$`, tc.twiceAsBig)
		//
		ctx.Step(`^(?:a )?failing step`, tc.aFailingStep)
		ctx.Step(`^(.*should not be called)`, tc.aStepThatShouldNotHaveBeenCalled)
		//ctx.Step(`^this step should fail`, tc.aFailingStep)
		//ctx.Step(`^the following steps? should be (passed|failed|skipped|undefined|pending):`, tc.followingStepsShouldHave)
		//ctx.Step(`^the undefined step snippets should be:$`, tc.theUndefinedStepSnippetsShouldBe)
		//
		//// event stream
		//ctx.Step(`^the following events should be fired:$`, tc.thereShouldBeEventsFired)
		//
		//// lt
		//ctx.Step(`^savybių aplankas "([^"]*)"$`, tc.featurePath)
		//ctx.Step(`^aš išskaitau savybes$`, tc.parseFeatures)
		//ctx.Step(`^aš turėčiau turėti ([\d]+) savybių failus:$`, tc.iShouldHaveNumFeatureFiles)
		//
		ctx.Step(`^(?:a )?pending step$`, func() error {
			return godog.ErrPending
		})
		ctx.Step(`^(?:(a|other|second|third|fourth) )?passing step$`, func() error {
			return nil
		})
		ctx.Step(`^(.*passing step that fires an event)$`, func(name string) error {
			parent.events = append(parent.events, &firedEvent{"Step", []interface{}{name}})
			return nil
		})
		ctx.Given(`^(?:a )?given step$`, func() error {
			return nil
		})
		ctx.When(`^(?:a )?when step$`, func() error {
			return nil
		})
		ctx.Then(`^(?:a )?then step$`, func() error {
			return nil
		})
		//
		//// Introduced to test formatter/cucumber.feature
		//ctx.Step(`^the rendered json will be as follows:$`, tc.theRenderedJSONWillBe)
		//
		//// Introduced to test formatter/pretty.feature
		//ctx.Step(`^the rendered output will be as follows:$`, tc.theRenderOutputWillBe)
		//
		//// Introduced to test formatter/junit.feature
		//ctx.Step(`^the rendered xml will be as follows:$`, tc.theRenderedXMLWillBe)
		//
		ctx.Step(`^(?:a )?failing multistep$`, func() godog.Steps {
			return godog.Steps{"passing step", "failing step"}
		})
		//
		ctx.Step(`^(?:a |an )?undefined multistep$`, func() godog.Steps {
			return godog.Steps{"passing step", "undefined step", "passing step"}
		})
		//
		//ctx.Then(`^(?:a |an )?undefined multistep using 'then' function$`, func() godog.Steps {
		//	return godog.Steps{"given step", "undefined step", "then step"}
		//})
		//
		ctx.Step(`^(?:a )?passing multistep$`, func() godog.Steps {
			return godog.Steps{"passing step", "passing step", "passing step"}
		})
		//
		ctx.Then(`^(?:a )?passing multistep using 'then' function$`, func() godog.Steps {
			return godog.Steps{"given step", "when step", "then step"}
		})
		//
		ctx.Step(`^(?:a )?failing nested multistep$`, func() godog.Steps {
			return godog.Steps{"passing step", "passing multistep", "failing multistep"}
		})
		ctx.Step(`IgnoredStep: .*`, func() error {
			return nil
		})
		//
		//ctx.Step(`^call func\(\*godog\.DocString\) with:$`, func(arg *godog.DocString) error {
		//	return nil
		//})
		//ctx.Step(`^call func\(string\) with:$`, func(arg string) error {
		//	return nil
		//})
		//
		//ctx.Step(`^passing step without return$`, func() {})
		//
		//ctx.Step(`^having correct context$`, func(ctx context.Context) (context.Context, error) {
		//	if ctx.Value(ctxKey("BeforeScenario")) == nil {
		//		return ctx, errors.New("missing BeforeScenario in context")
		//	}
		//
		//	if ctx.Value(ctxKey("BeforeStep")) == nil {
		//		return ctx, errors.New("missing BeforeStep in context")
		//	}
		//
		//	if ctx.Value(ctxKey("StepState")) == nil {
		//		return ctx, errors.New("missing StepState in context")
		//	}
		//
		//	return context.WithValue(ctx, ctxKey("Step"), true), nil
		//})
		//
		//ctx.Step(`^adding step state to context$`, func(ctx context.Context) context.Context {
		//	return context.WithValue(ctx, ctxKey("StepState"), true)
		//})
		//
		ctx.Step(`^I return a context from a step$`, tc.iReturnAContextFromAStep)
		ctx.Step(`^I should see the context in the next step$`, tc.iShouldSeeTheContextInTheNextStep)
		//ctx.Step(`^I can see contexts passed in multisteps$`, func() godog.Steps {
		//	return godog.Steps{
		//		"I return a context from a step",
		//		"I should see the context in the next step",
		//	}
		//})
		//
		//// introduced to test testingT
		ctx.Step(`^my step (?:fails|skips) the test by calling (FailNow|Fail|SkipNow|Skip) on testing T$`, tc.myStepCallsTFailErrorSkip)
		ctx.Step(`^my step fails the test by calling (Fatal|Error) on testing T with message "([^"]*)"$`, tc.myStepCallsTErrorFatal)
		ctx.Step(`^my step fails the test by calling (Fatalf|Errorf) on testing T with message "([^"]*)" and argument "([^"]*)"$`, tc.myStepCallsTErrorfFatalf)
		//ctx.Step(`^my step calls Log on testing T with message "([^"]*)"$`, tc.myStepCallsTLog)
		//ctx.Step(`^my step calls Logf on testing T with message "([^"]*)" and argument "([^"]*)"$`, tc.myStepCallsTLogf)
		ctx.Step(`^my step calls testify's assert.Equal with expected "([^"]*)" and actual "([^"]*)"$`, tc.myStepCallsTestifyAssertEqual)
		ctx.Step(`^my step calls testify's require.Equal with expected "([^"]*)" and actual "([^"]*)"$`, tc.myStepCallsTestifyRequireEqual)
		ctx.Step(`^my step calls testify's assert.Equal ([0-9]+) times(| with match)$`, tc.myStepCallsTestifyAssertEqualMultipleTimes)
		//ctx.Step(`^my step calls godog.Log with message "([^"]*)"$`, tc.myStepCallsDogLog)
		//ctx.Step(`^my step calls godog.Logf with message "([^"]*)" and argument "([^"]*)"$`, tc.myStepCallsDogLogf)
		//ctx.Step(`^the logged messages should include "([^"]*)"$`, tc.theLoggedMessagesShouldInclude)
		//
		ctx.StepContext().Before(tc.inject)
	}
}

type ctxKey string

func (tc *godogFeaturesScenarioInner) inject(ctx context.Context, step *godog.Step) (context.Context, error) {
	if !tc.allowInjection {
		return ctx, nil
	}

	step.Text = injectAll(step.Text)

	if step.Argument == nil {
		return ctx, nil
	}

	if table := step.Argument.DataTable; table != nil {
		for i := 0; i < len(table.Rows); i++ {
			for n, cell := range table.Rows[i].Cells {
				table.Rows[i].Cells[n].Value = injectAll(cell.Value)
			}
		}
	}

	if doc := step.Argument.DocString; doc != nil {
		doc.Content = injectAll(doc.Content)
	}

	return ctx, nil
}

func injectAll(src string) string {
	re := regexp.MustCompile(`{{PLACEHOLDER\d}}`)
	out := re.ReplaceAllString(src, "someverylonginjectionsoweacanbesureitsurpasstheinitiallongeststeplenghtanditwillhelptestsmethodsafety")
	return out
}

type firedEvent struct {
	name string
	args []interface{}
}

type firedEvents []*firedEvent

func (f firedEvent) String() string {
	if len(f.args) == 0 {
		return fmt.Sprintf("%s", f.name)
	}

	args := []string{}
	for _, arg := range f.args {
		args = append(args, fmt.Sprintf("[%v]", arg))
	}
	return fmt.Sprintf("%s %v", f.name, strings.Join(args, " "))
}

func (f firedEvents) ToStrings() []string {
	str := []string{}
	for _, ev := range f {
		str = append(str, ev.String())
	}
	return str
}

type godogFeaturesScenarioOuter struct {
	tempDir         string
	paths           []string
	events          firedEvents
	out             *bytes.Buffer
	formatter       *formatters.Base
	failed          bool
	featureContents []godog.Feature
}

type godogFeaturesScenarioInner struct {
	//paths    []string
	features []*models.Feature

	allowInjection bool

	stepsExecuted   []string // ok
	scenarioContext *godog.ScenarioContext
	featureContents []godog.Feature
}

// TODO why is this needed ?
// A new instance is created in the scenario initialiser
func (tc *godogFeaturesScenarioInner) ResetBeforeEachScenario(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
	// reset whole suite with the state
	//tc.out.Reset()
	//tc.paths = []string{}

	tc.features = []*models.Feature{}

	//tc.testedSuite = &godog.suite{}
	//tc.testSuiteContext = godog.TestSuiteContext{}

	// DO NOT RESET tc.scenarioContext = nil as it wipes out the istalled value

	// reset all fired events
	tc.allowInjection = false

	tc.stepsExecuted = []string{}

	return ctx, nil
}

func (tc *godogFeaturesScenarioInner) iSetVariableInjectionTo(state string) error {
	tc.allowInjection = state == "allow"
	return nil
}

var defaultFormatterFunc = formatters.BaseFormatterFunc

func (tc *godogFeaturesScenarioOuter) iRunFeatureSuite() error {
	return tc.runFeatureSuite("", defaultFormatterFunc, false)
}
func (tc *godogFeaturesScenarioOuter) iRunFeatureSuiteStrict() error {
	return tc.runFeatureSuite("", defaultFormatterFunc, true)
}

func (tc *godogFeaturesScenarioOuter) iRunFeatureSuiteWithFormatter(name string) error {
	f := godog.FindFmt(name)
	if f == nil {
		return fmt.Errorf(`formatter "%s" is not available`, name)
	}

	return tc.runFeatureSuite("", f, false)
}

func (tc *godogFeaturesScenarioOuter) iRunFeatureSuiteWithTags(tags string) error {
	return tc.runFeatureSuite(tags, defaultFormatterFunc, false)
}

func (tc *godogFeaturesScenarioOuter) runFeatureSuite(tags string, fmtFunc godog.FormatterFunc, strictMode bool) error {

	tc.out = new(bytes.Buffer)

	formatterFuncInterceptor := func(suiteName string, out io.WriteCloser) godog.Formatter {
		formatter := fmtFunc(suiteName, out)

		base, ok := formatter.(*formatters.Base)
		if ok {
			tc.formatter = base
		}
		return formatter
	}

	features := tc.featureContents

	suite := godog.TestSuite{
		Name:                 "godog",
		TestSuiteInitializer: InitializeTestSuiteInner(tc),
		ScenarioInitializer:  InitializeScenarioInner(tc),
		Options: &godog.Options{
			Formatter:       formatterFuncInterceptor,
			FeatureContents: features,
			Paths:           tc.paths,
			Tags:            tags,
			Strict:          strictMode,
			Output:          colors.Uncolored(godog.NopCloser(io.Writer(tc.out))),
		},
	}

	runResult := suite.Run()
	tc.failed = runResult != godog.ExitSuccess

	return nil
}

func (tc *godogFeaturesScenarioOuter) thereShouldBeEventsFired(doc *godog.DocString) error {
	actual := tc.events.ToStrings()
	expect := strings.Split(strings.TrimSpace(doc.Content), "\n")

	same := utils.SlicesCompare(expect, actual)
	if !same {
		utils.VDiffLists(expect, actual)
		return fmt.Errorf("expected %v events, but got %v", expect, actual)
	}

	return nil
}

func (tc *godogFeaturesScenarioOuter) cleanupSnippet(snip string) string {
	lines := strings.Split(strings.TrimSpace(snip), "\n")
	for i := 0; i < len(lines); i++ {
		lines[i] = strings.TrimSpace(lines[i])
	}

	return strings.Join(lines, "\n")
}

func (tc *godogFeaturesScenarioOuter) theUndefinedStepSnippetsShouldBe(body *godog.DocString) error {

	// fixme john - this should be collected from the output not a formatter call
	f := tc.formatter
	if f == nil {
		return errors.New("formatter has not been set on this test's scenario state, so cannot obtain the Snippets")
	}

	actual := tc.cleanupSnippet(f.Snippets())
	expected := tc.cleanupSnippet(body.Content)

	if actual != expected {
		return fmt.Errorf("snippets do not match actual: %s", f.Snippets())
	}

	return nil
}

type multiContextKey struct{}

func (tc *godogFeaturesScenarioInner) iReturnAContextFromAStep(ctx context.Context) (context.Context, error) {
	return context.WithValue(ctx, multiContextKey{}, "value"), nil
}

func (tc *godogFeaturesScenarioInner) iShouldSeeTheContextInTheNextStep(ctx context.Context) error {
	value, ok := ctx.Value(multiContextKey{}).(string)
	if !ok {
		return errors.New("context does not contain our key")
	}
	if value != "value" {
		return errors.New("context has the wrong value for our key")
	}
	return nil
}

func (tc *godogFeaturesScenarioInner) myStepCallsTFailErrorSkip(ctx context.Context, op string) error {
	switch op {
	case "FailNow":
		godog.T(ctx).FailNow()
	case "Fail":
		godog.T(ctx).Fail()
	case "SkipNow":
		godog.T(ctx).SkipNow()
	case "Skip":
		godog.T(ctx).Skip()
	default:
		return fmt.Errorf("operation %s not supported by myStepCallsTFailErrorSkip", op)
	}
	return nil
}

func (tc *godogFeaturesScenarioOuter) testingTShouldBe(state string) error {
	// FIXME john - canpt detect godog interaction with testing.T unless we switch it to rely on the interfate
	//if !tc.testingT.Failed() && state == "should have" {
	//	return fmt.Errorf("testing.T should have recorded a failure, but none were recorded")
	//}
	//if tc.testingT.Failed() && state == "should not have" {
	//	return fmt.Errorf("testing.T should not have recorded a failure, but errors were recorded")
	//}

	return nil
}

func (tc *godogFeaturesScenarioInner) myStepCallsTErrorFatal(ctx context.Context, op string, message string) error {
	switch op {
	case "Error":
		godog.T(ctx).Error(message)
	case "Fatal":
		godog.T(ctx).Fatal(message)
	default:
		return fmt.Errorf("operation %s not supported by myStepCallsTErrorFatal", op)
	}
	return nil
}

func (tc *godogFeaturesScenarioInner) myStepCallsTErrorfFatalf(ctx context.Context, op string, message string, arg string) error {
	switch op {
	case "Errorf":
		godog.T(ctx).Errorf(message, arg)
	case "Fatalf":
		godog.T(ctx).Fatalf(message, arg)
	default:
		return fmt.Errorf("operation %s not supported by myStepCallsTErrorfFatalf", op)
	}
	return nil
}

func (tc *godogFeaturesScenarioInner) myStepCallsTestifyAssertEqual(ctx context.Context, a string, b string) error {
	assert.Equal(godog.T(ctx), a, b)
	return nil
}

func (tc *godogFeaturesScenarioInner) myStepCallsTestifyAssertEqualMultipleTimes(ctx context.Context, times string, withMatch string) error {
	timesInt, err := strconv.Atoi(times)
	if err != nil {
		return fmt.Errorf("test step has invalid times value %s: %w", times, err)
	}
	for i := 0; i < timesInt; i++ {
		if withMatch == " with match" {
			assert.Equal(godog.T(ctx), fmt.Sprintf("exp%v", i), fmt.Sprintf("exp%v", i))
		} else {
			assert.Equal(godog.T(ctx), "exp", fmt.Sprintf("notexp%v", i))
		}
	}
	return nil
}

func (tc *godogFeaturesScenarioInner) myStepCallsTestifyRequireEqual(ctx context.Context, a string, b string) error {
	require.Equal(godog.T(ctx), a, b)
	return nil
}

func (tc *godogFeaturesScenarioOuter) myStepCallsTLog(ctx context.Context, message string) error {
	godog.T(ctx).Log(message)
	return nil
}

func (tc *godogFeaturesScenarioOuter) myStepCallsTLogf(ctx context.Context, message string, arg string) error {
	godog.T(ctx).Logf(message, arg)
	return nil
}

func (tc *godogFeaturesScenarioOuter) myStepCallsDogLog(ctx context.Context, message string) error {
	godog.Log(ctx, message)
	return nil
}

func (tc *godogFeaturesScenarioOuter) myStepCallsDogLogf(ctx context.Context, message string, arg string) error {
	godog.Logf(ctx, message, arg)
	return nil
}

// theLoggedMessagesShouldInclude asserts that the given message is present in the
// logged messages (i.e. the output of the suite's formatter). If the message is
// not found, it returns an error with the message and the logged messages.
func (tc *godogFeaturesScenarioOuter) theLoggedMessagesShouldInclude(ctx context.Context, message string) error {
	messages := godog.LoggedMessages(ctx)
	for _, m := range messages {
		if strings.Contains(m, message) {
			return nil
		}
	}
	return fmt.Errorf("the message %q was not logged (logged messages: %v)", message, messages)
}

func (tc *godogFeaturesScenarioOuter) followingStepsShouldHave(status string, steps *godog.DocString) error {
	return tc.checkStoredSteps(status, steps, false)
}

func (tc *godogFeaturesScenarioOuter) onlyFollowingStepsShouldHave(status string, steps *godog.DocString) error {
	return tc.checkStoredSteps(status, steps, true)
}

func (tc *godogFeaturesScenarioOuter) checkStoredSteps(status string, steps *godog.DocString, noOtherSteps bool) error {
	var expected = strings.Split(steps.Content, "\n")

	f := tc.formatter
	if f == nil {
		return errors.New("formatter has not been set on this test's scenario state, so cannot obtain the Storage")
	}

	store := f.GetStorage()

	stepStatus, err := models.ToStepResultStatus(status)
	if err != nil {
		return err
	}

	actual := tc.getStepsByStatus(store, stepStatus)

	sort.Strings(actual)
	sort.Strings(expected)

	if len(actual) != len(expected) {
		return perr.Errorf("expected %d %s steps: %q, but got %d %s steps: %q",
			len(expected), status, expected, len(actual), status, actual)
	}

	for i, a := range actual {
		if a != expected[i] {
			return perr.Errorf("%s step %d doesn't match, expected: %s, but got: %s", status, i, expected, actual)
		}
	}

	if noOtherSteps {
		// sort for printing purposes
		allStepResults := tc.getSteps(store)
		sort.Slice(allStepResults, func(i, j int) bool {
			// sort by text then status
			ival := allStepResults[i]
			jval := allStepResults[j]
			if ival.stepText < jval.stepText {
				return false
			}
			return ival.stepResult < jval.stepResult
		})

		if len(allStepResults) != len(expected) {
			return fmt.Errorf("expected only %d steps: %v\nbut got %d steps: %v",
				len(expected), expected, len(allStepResults), allStepResults)
		}
	}

	return nil
}

func (tc *godogFeaturesScenarioOuter) getStepsByStatus(storage *storage.Storage, status models.StepResultStatus) []string {
	actual := []string{}

	for _, st := range storage.MustGetPickleStepResultsByStatus(status) {
		pickleStep := storage.MustGetPickleStep(st.PickleStepID)
		actual = append(actual, pickleStep.Text)
	}
	return actual
}

type stepResult struct {
	stepText   string
	stepResult models.StepResultStatus
}

func (tc *godogFeaturesScenarioOuter) getSteps(storage *storage.Storage) []stepResult {
	results := []stepResult{}

	for _, f := range storage.MustGetFeatures() {
		for _, s := range storage.MustGetPickles(f.Uri) {
			for _, stepRes := range storage.MustGetPickleStepResultsByPickleID(s.Id) {
				step := storage.MustGetPickleStep(stepRes.PickleStepID)

				results = append(results, stepResult{
					stepText:   step.Text,
					stepResult: stepRes.Status,
				})
			}
		}
	}
	return results
}

func (tc *godogFeaturesScenarioOuter) theTraceShouldBe(steps *godog.DocString) error {

	f := tc.formatter
	if f == nil {
		return errors.New("formatter has not been set on this test's scenario state, so cannot obtain the Storage")
	}

	storage := f.GetStorage()

	trace := []string{}

	features := storage.MustGetFeatures()
	for _, feat := range features {
		trace = append(trace, fmt.Sprintf("Feature: %v", feat.Feature.Name))
		scenarios := storage.MustGetPickles(feat.Uri)
		for _, pickle := range scenarios {
			trace = append(trace, fmt.Sprintf("  Scenario: %v", pickle.Name))
			steps := pickle.Steps
			for _, step := range steps {
				result := storage.MustGetPickleStepResult(step.Id)
				trace = append(trace, fmt.Sprintf("    Step: %v : %v", step.Text, result.Status))
				if result.Err != nil {
					trace = append(trace, fmt.Sprintf("    Error: %v", result.Err.Error()))
				}
			}
		}
	}

	expected := steps.Content
	actual := strings.Join(trace, "\n")

	if expected != actual {
		utils.VDiffString(expected, actual)
	}

	return assertExpectedAndActual(assert.Equal, expected, actual, actual)
}

func prt(storage *storage.Storage, psrs models.StepResultStatus) []string {
	var r []string
	for _, p := range storage.MustGetPickleStepResultsByStatus(psrs) {
		pickleStep := storage.MustGetPickleStep(p.PickleStepID)
		r = append(r, fmt.Sprintf("[%s: %s]", p.Status, pickleStep.Text))
	}
	return r
}

//
//func (tc *godogFeaturesScenarioInner) iAmListeningToSuiteEvents() error {
//	return nil
//}

func (tc *godogFeaturesScenarioInner) aFailingStep() error {
	return fmt.Errorf("intentional failure")
}

func (tc *godogFeaturesScenarioInner) aStepThatShouldNotHaveBeenCalled(step string) error {
	return fmt.Errorf("the step '%s' step should have been skipped, but was executed", step)
}

// parse a given feature file body as a feature
func (tc *godogFeaturesScenarioOuter) aFeatureFile(path string, body *godog.DocString) error {
	tc.featureContents = append(tc.featureContents, godog.Feature{
		Name:     path,
		Contents: []byte(body.Content),
	})

	// permit the use of escaped """ docstrings inside this docstring
	contents := strings.ReplaceAll(body.Content, "\\\"", "\"")

	_, err := gherkin.ParseGherkinDocument(strings.NewReader(contents), (&messages.Incrementing{}).NewId)
	//gd.Uri = path
	//
	//pickles := gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)
	//tc.features = append(tc.features, &models.Feature{GherkinDocument: gd, Pickles: pickles})

	return err
}

func (tc *godogFeaturesScenarioInner) backgroundStepIsDefined(stepText string) {
	tc.stepsExecuted = append(tc.stepsExecuted, stepText)
}

func (tc *godogFeaturesScenarioInner) stepShouldHaveBeenExecuted(stepText string) error {
	stepWasExecuted := sliceContains(tc.stepsExecuted, stepText)
	if !stepWasExecuted {
		return fmt.Errorf("step '%s' was not called, found these steps: %v", stepText, tc.stepsExecuted)
	}
	return nil
}

func sliceContains(arr []string, text string) bool {
	for _, s := range arr {
		if s == text {
			return true
		}
	}
	return false
}

func (tc *godogFeaturesScenarioOuter) writeFeatureFile(path string, doc *godog.DocString) error {

	if !strings.HasPrefix(path, "features/") {
		return fmt.Errorf("path must start with features/ but got : %q", path)
	}
	err := os.MkdirAll(tc.tempDir, 0600)
	if err != nil {
		return fmt.Errorf("cannot create temp dir %q: %w", tc.tempDir, err)
	}

	dir := filepath.Join(tc.tempDir, filepath.Dir(path))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create feature dir %q: %v", dir, err)
	}

	featureFile := filepath.Join(tc.tempDir, path)

	if err := os.WriteFile(featureFile, []byte(doc.Content), 0644); err != nil {
		return fmt.Errorf("cannot write to: %v: %v", featureFile, err)
	}

	return nil
}
func (tc *godogFeaturesScenarioOuter) featurePath(path string) {
	tc.paths = append(tc.paths, filepath.Join(tc.tempDir, path))
}

//
//func (tc *godogFeaturesScenarioInner) parseFeatures() error {
//	//	fts, err := parser.ParseFeatures(storage.FS{}, "", tc.paths)
//	//if err != nil {
//	//	return err
//	//}
//	//
//	//tc.features = append(tc.features, fts...)
//
//	return errors.New("parseFeatures shouldn't be used as parsing tests should be done in the parsing model unit tests")
//}

func (tc *godogFeaturesScenarioOuter) theSuiteShouldHave(state string) error {
	if tc.failed && state == "passed" {
		return fmt.Errorf("the feature suite has failed but should have passed")
	}

	if !tc.failed && state == "failed" {
		return fmt.Errorf("the feature suite has passed but should have failed")
	}

	return nil
}

func (tc *godogFeaturesScenarioInner) iShouldHaveNumFeatureFiles(num int, files *godog.DocString) error {
	if len(tc.features) != num {
		return fmt.Errorf("expected %d features to be parsed, but have %d", num, len(tc.features))
	}

	expected := strings.Split(files.Content, "\n")

	var actual []string

	for _, ft := range tc.features {
		actual = append(actual, ft.Uri)
	}

	if len(expected) != len(actual) {
		return fmt.Errorf("expected %d feature paths to be parsed, but have %d", len(expected), len(actual))
	}

	for i := 0; i < len(expected); i++ {
		var matched bool
		split := strings.Split(expected[i], "/")
		exp := filepath.Join(split...)

		for j := 0; j < len(actual); j++ {
			split = strings.Split(actual[j], "/")
			act := filepath.Join(split...)

			if exp == act {
				matched = true
				break
			}
		}

		if !matched {
			return fmt.Errorf(`expected feature path "%s" at position: %d, was not parsed, actual are %+v`, exp, i, actual)
		}
	}

	return nil
}

//
//func (tc *godogFeaturesScenarioInner) numScenariosRegistered(expected int) (err error) {
//	var num int
//	for _, ft := range tc.features {
//		num += len(ft.Pickles)
//	}
//
//	if num != expected {
//		err = fmt.Errorf("expected %d scenarios to be registered, but got %d", expected, num)
//	}
//
//	return
//}
//
//func (tc *godogFeaturesScenarioInner) thereWereNumEventsFired(_ string, expected int, typ string) error {
//
//	var num int
//	for _, event := range tc.events {
//		if event.name == typ {
//			num++
//		}
//	}
//
//	if num != expected {
//		if typ == "BeforeFeature" || typ == "AfterFeature" {
//			return nil
//		}
//
//		return fmt.Errorf("expected %d %s events to be fired, but got %d", expected, typ, num)
//	}
//
//	return nil
//}

//
//func (tc *godogFeaturesScenarioInner) thereWasEventTriggeredBeforeScenario(expected string) error {
//	var found []string
//	for _, event := range tc.events {
//		if event.name != "BeforeScenario" {
//			continue
//		}
//
//		var name string
//		switch t := event.args[0].(type) {
//		case *godog.Scenario:
//			name = t.Name
//		}
//
//		if name == expected {
//			return nil
//		}
//
//		found = append(found, name)
//	}
//
//	if len(found) == 0 {
//		return fmt.Errorf("before scenario event was never triggered or listened")
//	}
//
//	return fmt.Errorf(`expected "%s" scenario, but got these fired %s`, expected, `"`+strings.Join(found, `", "`)+`"`)
//
//}
//
//func (tc *godogFeaturesScenarioInner) theseEventsHadToBeFiredForNumberOfTimes(tbl *godog.Table) error {
//	if len(tbl.Rows[0].Cells) != 2 {
//		return fmt.Errorf("expected two columns for event table row, got: %d", len(tbl.Rows[0].Cells))
//	}
//
//	for _, row := range tbl.Rows {
//		num, err := strconv.ParseInt(row.Cells[1].Value, 10, 0)
//		if err != nil {
//			return err
//		}
//
//		if err := tc.thereWereNumEventsFired("", int(num), row.Cells[0].Value); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}

func (tc *godogFeaturesScenarioInner) twiceAsBig(tbl *godog.Table) error {
	if len(tbl.Rows[0].Cells) != 2 {
		return fmt.Errorf("expected two columns for event table row, got: %d", len(tbl.Rows[0].Cells))
	}

	num1, err := strconv.ParseInt(tbl.Rows[0].Cells[1].Value, 10, 0)
	if err != nil {
		return err
	}
	num2, err := strconv.ParseInt(tbl.Rows[1].Cells[1].Value, 10, 0)
	if err != nil {
		return err
	}
	if num2 != num1*2 {
		return fmt.Errorf("expected %d to be twice as big as %d", num2, num1)
	}

	return nil
}

func (tc *godogFeaturesScenarioOuter) theRenderedJSONWillBe(docstring *godog.DocString) error {
	durationRegex := regexp.MustCompile(`"duration": \d+`)
	locationRegex := regexp.MustCompile(`"location": "(\\u003cautogenerated\\u003e|[\w_]+.go):\d+"`)

	expectedString := docstring.Content
	expectedString = locationRegex.ReplaceAllString(expectedString, `"location": "<autogenerated>:0"`)
	expectedString = durationRegex.ReplaceAllString(expectedString, `"duration": 9999`)

	actualString := tc.out.String()
	actualString = locationRegex.ReplaceAllString(actualString, `"location": "<autogenerated>:0"`)
	actualString = durationRegex.ReplaceAllString(actualString, `"duration": 9999`)

	var expected []interface{}
	if err := json.Unmarshal([]byte(expectedString), &expected); err != nil {
		return perr.Wrapf(err, "unmarshalling expected value: %s", expectedString)
	}

	var actual []interface{}
	if err := json.Unmarshal([]byte(actualString), &actual); err != nil {
		return perr.Wrapf(err, "unmarshalling actual value: %s", actualString)
	}

	err := assertExpectedAndActual(assert.Equal, expected, actual)

	if err != nil {
		err := tc.showJsonComparison(expected, expectedString, actual, actualString)
		if err != nil {
			return err
		}
	}

	return err
}

func (tc *godogFeaturesScenarioOuter) showJsonComparison(expected []interface{}, expectedString string, actual []interface{}, actualString string) error {
	vexpected, err := json.MarshalIndent(&expected, "", "  ")
	if err != nil {
		return perr.Wrapf(err, "marshalling expected value: %s", expectedString)
	}
	vactual, err := json.MarshalIndent(&actual, "", "  ")
	if err != nil {
		return perr.Wrapf(err, "marshalling actual value: %s", actualString)
	}

	utils.VDiffString(string(vexpected), string(vactual))
	return nil
}

func (tc *godogFeaturesScenarioOuter) theRenderedOutputWillBe(docstring *godog.DocString) error {

	durationRegex := regexp.MustCompile(`[\d.]+?(s|ms|µs)`)
	stepHandlerRegex := regexp.MustCompile(`(<autogenerated>|feature_test.go):([\S]+) -> .*`)

	expected := docstring.Content
	expected = durationRegex.ReplaceAllString(expected, "9.99s")
	expected = stepHandlerRegex.ReplaceAllString(expected, "<gofile>:<lineno> -> <gofunc>")
	expected = strings.ReplaceAll(expected, tc.tempDir, "")

	actual := tc.out.String()
	actual = durationRegex.ReplaceAllString(actual, "9.99s")
	actual = stepHandlerRegex.ReplaceAllString(actual, "<gofile>:<lineno> -> <gofunc>")
	actual = strings.ReplaceAll(actual, tc.tempDir, "")

	if actual != expected {
		utils.VDiffString(expected, actual)

		fmt.Printf("Actual:\n%s", actual)
	}
	return assertExpectedAndActual(assert.Equal, expected, actual, actual)
}

func (tc *godogFeaturesScenarioOuter) theRenderedEventsWillBe(docstring *godog.DocString) error {
	timeStampRegex := regexp.MustCompile(`"timestamp":-?\d+`)

	// the file location looks different depending on running vs debugging
	definitionIdDebug := regexp.MustCompile(`"definition_id":"feature_test.go:\d+ -\\u003e [^"]+"`)

	definitionIdRepl := `"definition_id":"feature_test.go:<autogenerated> -\u003e <autogenerated>"`

	expected := docstring.Content
	expected = utils.TrimAllLines(expected)

	actual := tc.out.String()

	actual = definitionIdDebug.ReplaceAllString(actual, definitionIdRepl)
	actual = timeStampRegex.ReplaceAllString(actual, `"timestamp":9999`)

	actualTrimmed := actual
	actual = utils.TrimAllLines(actual)

	if expected != actual {
		utils.VDiffString(expected, actual)
	}
	return assertExpectedAndActual(assert.Equal, expected, actual, actualTrimmed)
}

func (tc *godogFeaturesScenarioOuter) theRenderedXMLWillBe(docstring *godog.DocString) error {
	expectedString := docstring.Content
	actualString := tc.out.String()

	timeRegex := regexp.MustCompile(`time="[\d.]+"`)
	actualString = timeRegex.ReplaceAllString(actualString, `time="9999.9999"`)
	expectedString = timeRegex.ReplaceAllString(expectedString, `time="9999.9999"`)

	var expected formatters.JunitPackageSuite
	if err := xml.Unmarshal([]byte(expectedString), &expected); err != nil {
		return perr.Wrapf(err, "unmarshalling expected value: %s", actualString)
	}

	var actual formatters.JunitPackageSuite
	if err := xml.Unmarshal([]byte(actualString), &actual); err != nil {
		return perr.Wrapf(err, "unmarshalling actual value: %s", actualString)
	}

	return assertExpectedAndActual(assert.Equal, expected, actual)
}

func assertExpectedAndActual(a expectedAndActualAssertion, expected, actual interface{}, msgAndArgs ...interface{}) error {
	var t asserter
	a(&t, expected, actual, msgAndArgs...)

	if t.err != nil {
		return t.err
	}

	return t.err
}

type expectedAndActualAssertion func(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool

type asserter struct {
	err error
}

func (a *asserter) Errorf(format string, args ...interface{}) {
	a.err = fmt.Errorf(format, args...)
}

func tagged(tags []*messages.PickleTag, tagName string) bool {
	for _, tag := range tags {
		if tag.Name == tagName {
			return true
		}
	}
	return false

}
