package godog

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	gherkin "github.com/cucumber/gherkin/go/v26"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cucumber/godog/colors"
	"github.com/cucumber/godog/internal/formatters"
	"github.com/cucumber/godog/internal/models"
	"github.com/cucumber/godog/internal/parser"
	"github.com/cucumber/godog/internal/storage"
	"github.com/cucumber/godog/internal/tags"
	"github.com/cucumber/godog/internal/utils"
)

// InitializeScenario provides steps for godog suite execution and
// can be used for meta-testing of godog features/steps themselves.
//
// Beware, steps or their definitions might change without backward
// compatibility guarantees. A typical user of the godog library should never
// need this, rather it is provided for those developing add-on libraries for godog.
//
// For an example of how to use, see godog's own `features/` and `suite_test.go`.
func InitializeScenario(ctx *ScenarioContext) {
	tc := &godogFeaturesScenario{}

	ctx.Before(tc.ResetBeforeEachScenario)

	ctx.Step(`^(?:a )?feature path "([^"]*)"$`, tc.featurePath)
	ctx.Step(`^I parse features$`, tc.parseFeatures)
	ctx.Step(`^I'm listening to suite events$`, tc.iAmListeningToSuiteEvents)
	ctx.Step(`^I run feature suite$`, tc.iRunFeatureSuite)
	ctx.Step(`^I run feature suite with tags "([^"]*)"$`, tc.iRunFeatureSuiteWithTags)
	ctx.Step(`^I run feature suite with formatter "([^"]*)"$`, tc.iRunFeatureSuiteWithFormatter)
	ctx.Step(`^(?:I )(allow|disable) variable injection`, tc.iSetVariableInjectionTo)
	ctx.Step(`^(?:a )?feature "([^"]*)"(?: file)?:$`, tc.aFeatureFile)
	ctx.Step(`^the suite should have (passed|failed)$`, tc.theSuiteShouldHave)

	ctx.Step(`^I should have ([\d]+) features? files?:$`, tc.iShouldHaveNumFeatureFiles)
	ctx.Step(`^I should have ([\d]+) scenarios? registered$`, tc.numScenariosRegistered)
	ctx.Step(`^there (was|were) ([\d]+) "([^"]*)" events? fired$`, tc.thereWereNumEventsFired)
	ctx.Step(`^there was event triggered before scenario "([^"]*)"$`, tc.thereWasEventTriggeredBeforeScenario)
	ctx.Step(`^these events had to be fired for a number of times:$`, tc.theseEventsHadToBeFiredForNumberOfTimes)

	ctx.Step(`^(?:a )?failing step`, tc.aFailingStep)
	ctx.Step(`^this step should fail`, tc.aFailingStep)
	ctx.Step(`^the following steps? should be (passed|failed|skipped|undefined|pending):`, tc.followingStepsShouldHave)
	ctx.Step(`^the undefined step snippets should be:$`, tc.theUndefinedStepSnippetsShouldBe)

	// event stream
	ctx.Step(`^the following events should be fired:$`, tc.thereShouldBeEventsFired)

	// lt
	ctx.Step(`^savybių aplankas "([^"]*)"$`, tc.featurePath)
	ctx.Step(`^aš išskaitau savybes$`, tc.parseFeatures)
	ctx.Step(`^aš turėčiau turėti ([\d]+) savybių failus:$`, tc.iShouldHaveNumFeatureFiles)

	ctx.Step(`^(?:a )?pending step$`, func() error {
		return ErrPending
	})
	ctx.Step(`^(?:a )?passing step$`, func() error {
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

	// Introduced to test formatter/cucumber.feature
	ctx.Step(`^the rendered json will be as follows:$`, tc.theRenderJSONWillBe)

	// Introduced to test formatter/pretty.feature
	ctx.Step(`^the rendered output will be as follows:$`, tc.theRenderOutputWillBe)

	// Introduced to test formatter/junit.feature
	ctx.Step(`^the rendered xml will be as follows:$`, tc.theRenderXMLWillBe)

	ctx.Step(`^(?:a )?failing multistep$`, func() Steps {
		return Steps{"passing step", "failing step"}
	})

	ctx.Step(`^(?:a |an )?undefined multistep$`, func() Steps {
		return Steps{"passing step", "undefined step", "passing step"}
	})

	ctx.Then(`^(?:a |an )?undefined multistep using 'then' function$`, func() Steps {
		return Steps{"given step", "undefined step", "then step"}
	})

	ctx.Step(`^(?:a )?passing multistep$`, func() Steps {
		return Steps{"passing step", "passing step", "passing step"}
	})

	ctx.Then(`^(?:a )?passing multistep using 'then' function$`, func() Steps {
		return Steps{"given step", "when step", "then step"}
	})

	ctx.Step(`^(?:a )?failing nested multistep$`, func() Steps {
		return Steps{"passing step", "passing multistep", "failing multistep"}
	})
	// Default recovery step
	ctx.Step(`Ignore.*`, func() error {
		return nil
	})

	ctx.Step(`^call func\(\*godog\.DocString\) with:$`, func(arg *DocString) error {
		return nil
	})
	ctx.Step(`^call func\(string\) with:$`, func(arg string) error {
		return nil
	})

	ctx.Step(`^passing step without return$`, func() {})

	ctx.Step(`^having correct context$`, func(ctx context.Context) (context.Context, error) {
		if ctx.Value(ctxKey("BeforeScenario")) == nil {
			return ctx, errors.New("missing BeforeScenario in context")
		}

		if ctx.Value(ctxKey("BeforeStep")) == nil {
			return ctx, errors.New("missing BeforeStep in context")
		}

		if ctx.Value(ctxKey("StepState")) == nil {
			return ctx, errors.New("missing StepState in context")
		}

		return context.WithValue(ctx, ctxKey("Step"), true), nil
	})

	ctx.Step(`^adding step state to context$`, func(ctx context.Context) context.Context {
		return context.WithValue(ctx, ctxKey("StepState"), true)
	})

	ctx.Step(`^I return a context from a step$`, tc.iReturnAContextFromAStep)
	ctx.Step(`^I should see the context in the next step$`, tc.iShouldSeeTheContextInTheNextStep)
	ctx.Step(`^I can see contexts passed in multisteps$`, func() Steps {
		return Steps{
			"I return a context from a step",
			"I should see the context in the next step",
		}
	})

	// introduced to test testingT
	ctx.Step(`^my step (?:fails|skips) the test by calling (FailNow|Fail|SkipNow|Skip) on testing T$`, tc.myStepCallsTFailErrorSkip)
	ctx.Step(`^my step fails the test by calling (Fatal|Error) on testing T with message "([^"]*)"$`, tc.myStepCallsTErrorFatal)
	ctx.Step(`^my step fails the test by calling (Fatalf|Errorf) on testing T with message "([^"]*)" and argument "([^"]*)"$`, tc.myStepCallsTErrorfFatalf)
	ctx.Step(`^my step calls Log on testing T with message "([^"]*)"$`, tc.myStepCallsTLog)
	ctx.Step(`^my step calls Logf on testing T with message "([^"]*)" and argument "([^"]*)"$`, tc.myStepCallsTLogf)
	ctx.Step(`^my step calls testify's assert.Equal with expected "([^"]*)" and actual "([^"]*)"$`, tc.myStepCallsTestifyAssertEqual)
	ctx.Step(`^my step calls testify's require.Equal with expected "([^"]*)" and actual "([^"]*)"$`, tc.myStepCallsTestifyRequireEqual)
	ctx.Step(`^my step calls testify's assert.Equal ([0-9]+) times(| with match)$`, tc.myStepCallsTestifyAssertEqualMultipleTimes)
	ctx.Step(`^my step calls godog.Log with message "([^"]*)"$`, tc.myStepCallsDogLog)
	ctx.Step(`^my step calls godog.Logf with message "([^"]*)" and argument "([^"]*)"$`, tc.myStepCallsDogLogf)
	ctx.Step(`^the logged messages should include "([^"]*)"$`, tc.theLoggedMessagesShouldInclude)

	ctx.StepContext().Before(tc.inject)
}

type ctxKey string

func (tc *godogFeaturesScenario) inject(ctx context.Context, step *Step) (context.Context, error) {
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
	re := regexp.MustCompile(`{{[^{}]+}}`)
	return re.ReplaceAllStringFunc(
		src,
		func(key string) string {
			injectRegex := regexp.MustCompile(`^{{.+}}$`)

			if injectRegex.MatchString(key) {
				return "someverylonginjectionsoweacanbesureitsurpasstheinitiallongeststeplenghtanditwillhelptestsmethodsafety"
			}

			return key
		},
	)
}

type firedEvent struct {
	name string
	args []interface{}
}

type godogFeaturesScenario struct {
	paths            []string
	features         []*models.Feature
	testedSuite      *suite
	testSuiteContext TestSuiteContext
	events           []*firedEvent
	out              bytes.Buffer
	allowInjection   bool
}

func (tc *godogFeaturesScenario) ResetBeforeEachScenario(ctx context.Context, sc *Scenario) (context.Context, error) {
	// reset whole suite with the state
	tc.out.Reset()
	tc.paths = []string{}

	tc.features = []*models.Feature{}
	tc.testedSuite = &suite{}
	tc.testSuiteContext = TestSuiteContext{}

	// reset all fired events
	tc.events = []*firedEvent{}
	tc.allowInjection = false

	return ctx, nil
}

func (tc *godogFeaturesScenario) iSetVariableInjectionTo(to string) error {
	tc.allowInjection = to == "allow"
	return nil
}

func (tc *godogFeaturesScenario) iRunFeatureSuiteWithTags(tags string) error {
	return tc.iRunFeatureSuiteWithTagsAndFormatter(tags, formatters.BaseFormatterFunc)
}

func (tc *godogFeaturesScenario) iRunFeatureSuiteWithFormatter(name string) error {
	f := FindFmt(name)
	if f == nil {
		return fmt.Errorf(`formatter "%s" is not available`, name)
	}

	return tc.iRunFeatureSuiteWithTagsAndFormatter("", f)
}

func (tc *godogFeaturesScenario) iRunFeatureSuiteWithTagsAndFormatter(filter string, fmtFunc FormatterFunc) error {
	if err := tc.parseFeatures(); err != nil {
		return err
	}

	for _, feat := range tc.features {
		feat.Pickles = tags.ApplyTagFilter(filter, feat.Pickles)
	}

	tc.testedSuite.storage = storage.NewStorage()
	for _, feat := range tc.features {
		tc.testedSuite.storage.MustInsertFeature(feat)

		for _, pickle := range feat.Pickles {
			tc.testedSuite.storage.MustInsertPickle(pickle)
		}
	}

	tc.testedSuite.fmt = fmtFunc("godog", colors.Uncolored(&tc.out))
	if fmt, ok := tc.testedSuite.fmt.(storageFormatter); ok {
		fmt.SetStorage(tc.testedSuite.storage)
	}

	testRunStarted := models.TestRunStarted{StartedAt: utils.TimeNowFunc()}
	tc.testedSuite.storage.MustInsertTestRunStarted(testRunStarted)
	tc.testedSuite.fmt.TestRunStarted()

	for _, f := range tc.testSuiteContext.beforeSuiteHandlers {
		f()
	}

	for _, ft := range tc.features {
		tc.testedSuite.fmt.Feature(ft.GherkinDocument, ft.Uri, ft.Content)

		for _, pickle := range ft.Pickles {
			if tc.testedSuite.stopOnFailure && tc.testedSuite.failed {
				continue
			}

			sc := ScenarioContext{suite: tc.testedSuite}
			InitializeScenario(&sc)

			err := tc.testedSuite.runPickle(pickle)
			if tc.testedSuite.shouldFail(err) {
				tc.testedSuite.failed = true
			}
		}
	}

	for _, f := range tc.testSuiteContext.afterSuiteHandlers {
		f()
	}

	tc.testedSuite.fmt.Summary()

	return nil
}

func (tc *godogFeaturesScenario) thereShouldBeEventsFired(doc *DocString) error {
	actual := strings.Split(strings.TrimSpace(tc.out.String()), "\n")
	expect := strings.Split(strings.TrimSpace(doc.Content), "\n")

	if len(expect) != len(actual) {
		return fmt.Errorf("expected %d events, but got %d", len(expect), len(actual))
	}

	type ev struct {
		Event string
	}

	for i, event := range actual {
		exp := strings.TrimSpace(expect[i])
		var act ev

		if err := json.Unmarshal([]byte(event), &act); err != nil {
			return fmt.Errorf("failed to read event data: %v", err)
		}

		if act.Event != exp {
			return fmt.Errorf(`expected event: "%s" at position: %d, but actual was "%s"`, exp, i, act.Event)
		}
	}

	return nil
}

func (tc *godogFeaturesScenario) cleanupSnippet(snip string) string {
	lines := strings.Split(strings.TrimSpace(snip), "\n")
	for i := 0; i < len(lines); i++ {
		lines[i] = strings.TrimSpace(lines[i])
	}

	return strings.Join(lines, "\n")
}

func (tc *godogFeaturesScenario) theUndefinedStepSnippetsShouldBe(body *DocString) error {
	f, ok := tc.testedSuite.fmt.(*formatters.Base)
	if !ok {
		return fmt.Errorf("this step requires *formatters.Base, but there is: %T", tc.testedSuite.fmt)
	}

	actual := tc.cleanupSnippet(f.Snippets())
	expected := tc.cleanupSnippet(body.Content)

	if actual != expected {
		return fmt.Errorf("snippets do not match actual: %s", f.Snippets())
	}

	return nil
}

type multiContextKey struct{}

func (tc *godogFeaturesScenario) iReturnAContextFromAStep(ctx context.Context) (context.Context, error) {
	return context.WithValue(ctx, multiContextKey{}, "value"), nil
}

func (tc *godogFeaturesScenario) iShouldSeeTheContextInTheNextStep(ctx context.Context) error {
	value, ok := ctx.Value(multiContextKey{}).(string)
	if !ok {
		return errors.New("context does not contain our key")
	}
	if value != "value" {
		return errors.New("context has the wrong value for our key")
	}
	return nil
}

func (tc *godogFeaturesScenario) myStepCallsTFailErrorSkip(ctx context.Context, op string) error {
	switch op {
	case "FailNow":
		T(ctx).FailNow()
	case "Fail":
		T(ctx).Fail()
	case "SkipNow":
		T(ctx).SkipNow()
	case "Skip":
		T(ctx).Skip()
	default:
		return fmt.Errorf("operation %s not supported by iCallTFailErrorSkip", op)
	}
	return nil
}

func (tc *godogFeaturesScenario) myStepCallsTErrorFatal(ctx context.Context, op string, message string) error {
	switch op {
	case "Error":
		T(ctx).Error(message)
	case "Fatal":
		T(ctx).Fatal(message)
	default:
		return fmt.Errorf("operation %s not supported by iCallTErrorFatal", op)
	}
	return nil
}

func (tc *godogFeaturesScenario) myStepCallsTErrorfFatalf(ctx context.Context, op string, message string, arg string) error {
	switch op {
	case "Errorf":
		T(ctx).Errorf(message, arg)
	case "Fatalf":
		T(ctx).Fatalf(message, arg)
	default:
		return fmt.Errorf("operation %s not supported by iCallTErrorfFatalf", op)
	}
	return nil
}

func (tc *godogFeaturesScenario) myStepCallsTestifyAssertEqual(ctx context.Context, a string, b string) error {
	assert.Equal(T(ctx), a, b)
	return nil
}

func (tc *godogFeaturesScenario) myStepCallsTestifyAssertEqualMultipleTimes(ctx context.Context, times string, withMatch string) error {
	timesInt, err := strconv.Atoi(times)
	if err != nil {
		return fmt.Errorf("test step has invalid times value %s: %w", times, err)
	}
	for i := 0; i < timesInt; i++ {
		if withMatch == " with match" {
			assert.Equal(T(ctx), fmt.Sprintf("exp%v", i), fmt.Sprintf("exp%v", i))
		} else {
			assert.Equal(T(ctx), "exp", fmt.Sprintf("notexp%v", i))
		}
	}
	return nil
}

func (tc *godogFeaturesScenario) myStepCallsTestifyRequireEqual(ctx context.Context, a string, b string) error {
	require.Equal(T(ctx), a, b)
	return nil
}

func (tc *godogFeaturesScenario) myStepCallsTLog(ctx context.Context, message string) error {
	T(ctx).Log(message)
	return nil
}

func (tc *godogFeaturesScenario) myStepCallsTLogf(ctx context.Context, message string, arg string) error {
	T(ctx).Logf(message, arg)
	return nil
}

func (tc *godogFeaturesScenario) myStepCallsDogLog(ctx context.Context, message string) error {
	Log(ctx, message)
	return nil
}

func (tc *godogFeaturesScenario) myStepCallsDogLogf(ctx context.Context, message string, arg string) error {
	Logf(ctx, message, arg)
	return nil
}

func (tc *godogFeaturesScenario) theLoggedMessagesShouldInclude(ctx context.Context, message string) error {
	messages := LoggedMessages(ctx)
	for _, m := range messages {
		if strings.Contains(m, message) {
			return nil
		}
	}
	return fmt.Errorf("the message %q was not logged (logged messages: %v)", message, messages)
}

func (tc *godogFeaturesScenario) followingStepsShouldHave(status string, steps *DocString) error {
	expected := strings.Split(steps.Content, "\n")
	var actual, unmatched, matched []string

	storage := tc.testedSuite.storage

	switch status {
	case "passed":
		for _, st := range storage.MustGetPickleStepResultsByStatus(models.Passed) {
			pickleStep := storage.MustGetPickleStep(st.PickleStepID)
			actual = append(actual, pickleStep.Text)
		}
	case "failed":
		for _, st := range storage.MustGetPickleStepResultsByStatus(models.Failed) {
			pickleStep := storage.MustGetPickleStep(st.PickleStepID)
			actual = append(actual, pickleStep.Text)
		}
	case "skipped":
		for _, st := range storage.MustGetPickleStepResultsByStatus(models.Skipped) {
			pickleStep := storage.MustGetPickleStep(st.PickleStepID)
			actual = append(actual, pickleStep.Text)
		}
	case "undefined":
		for _, st := range storage.MustGetPickleStepResultsByStatus(models.Undefined) {
			pickleStep := storage.MustGetPickleStep(st.PickleStepID)
			actual = append(actual, pickleStep.Text)
		}
	case "pending":
		for _, st := range storage.MustGetPickleStepResultsByStatus(models.Pending) {
			pickleStep := storage.MustGetPickleStep(st.PickleStepID)
			actual = append(actual, pickleStep.Text)
		}
	default:
		return fmt.Errorf("unexpected step status wanted: %s", status)
	}

	if len(expected) > len(actual) {
		return fmt.Errorf("number of expected %s steps: %d is less than actual %s steps: %d", status, len(expected), status, len(actual))
	}

	for _, a := range actual {
		for _, e := range expected {
			if a == e {
				matched = append(matched, e)
				break
			}
		}
	}

	if len(matched) >= len(expected) {
		return nil
	}

	for _, s := range expected {
		var found bool
		for _, m := range matched {
			if s == m {
				found = true
				break
			}
		}

		if !found {
			unmatched = append(unmatched, s)
		}
	}

	return fmt.Errorf("the steps: %s - are not %s", strings.Join(unmatched, ", "), status)
}

func (tc *godogFeaturesScenario) iAmListeningToSuiteEvents() error {
	tc.testSuiteContext.BeforeSuite(func() {
		tc.events = append(tc.events, &firedEvent{"BeforeSuite", []interface{}{}})
	})

	tc.testSuiteContext.AfterSuite(func() {
		tc.events = append(tc.events, &firedEvent{"AfterSuite", []interface{}{}})
	})

	scenarioContext := ScenarioContext{suite: tc.testedSuite}

	scenarioContext.Before(func(ctx context.Context, pickle *Scenario) (context.Context, error) {
		tc.events = append(tc.events, &firedEvent{"BeforeScenario", []interface{}{pickle}})

		if ctx.Value(ctxKey("BeforeScenario")) != nil {
			return ctx, errors.New("unexpected BeforeScenario in context (double invocation)")
		}

		return context.WithValue(ctx, ctxKey("BeforeScenario"), pickle.Name), nil
	})

	scenarioContext.Before(func(ctx context.Context, sc *Scenario) (context.Context, error) {
		if sc.Name == "failing before and after scenario" || sc.Name == "failing before scenario" {
			return context.WithValue(ctx, ctxKey("AfterStep"), sc.Name), errors.New("failed in before scenario hook")
		}

		return ctx, nil
	})

	scenarioContext.After(func(ctx context.Context, sc *Scenario, err error) (context.Context, error) {
		if sc.Name == "failing before and after scenario" || sc.Name == "failing after scenario" {
			return ctx, errors.New("failed in after scenario hook")
		}

		return ctx, nil
	})

	scenarioContext.After(func(ctx context.Context, pickle *Scenario, err error) (context.Context, error) {
		tc.events = append(tc.events, &firedEvent{"AfterScenario", []interface{}{pickle, err}})

		if ctx.Value(ctxKey("BeforeScenario")) == nil {
			return ctx, errors.New("missing BeforeScenario in context")
		}

		if ctx.Value(ctxKey("AfterStep")) == nil {
			return ctx, errors.New("missing AfterStep in context")
		}

		return context.WithValue(ctx, ctxKey("AfterScenario"), pickle.Name), nil
	})

	scenarioContext.StepContext().Before(func(ctx context.Context, step *Step) (context.Context, error) {
		tc.events = append(tc.events, &firedEvent{"BeforeStep", []interface{}{step}})

		if ctx.Value(ctxKey("BeforeScenario")) == nil {
			return ctx, errors.New("missing BeforeScenario in context")
		}

		return context.WithValue(ctx, ctxKey("BeforeStep"), step.Text), nil
	})

	scenarioContext.StepContext().After(func(ctx context.Context, step *Step, status StepResultStatus, err error) (context.Context, error) {
		tc.events = append(tc.events, &firedEvent{"AfterStep", []interface{}{step, err}})

		if ctx.Value(ctxKey("BeforeScenario")) == nil {
			return ctx, errors.New("missing BeforeScenario in context")
		}

		if ctx.Value(ctxKey("AfterScenario")) != nil && status != models.Skipped {
			panic("unexpected premature AfterScenario during AfterStep: " + ctx.Value(ctxKey("AfterScenario")).(string))
		}

		if ctx.Value(ctxKey("BeforeStep")) == nil {
			return ctx, errors.New("missing BeforeStep in context")
		}

		if step.Text == "having correct context" && ctx.Value(ctxKey("Step")) == nil {
			if status != StepSkipped {
				return ctx, fmt.Errorf("unexpected step result status: %s", status)
			}

			return ctx, errors.New("missing Step in context")
		}

		return context.WithValue(ctx, ctxKey("AfterStep"), step.Text), nil
	})

	return nil
}

func (tc *godogFeaturesScenario) aFailingStep() error {
	return fmt.Errorf("intentional failure")
}

// parse a given feature file body as a feature
func (tc *godogFeaturesScenario) aFeatureFile(path string, body *DocString) error {
	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(body.Content), (&messages.Incrementing{}).NewId)
	gd.Uri = path

	pickles := gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)
	tc.features = append(tc.features, &models.Feature{GherkinDocument: gd, Pickles: pickles})

	return err
}

func (tc *godogFeaturesScenario) featurePath(path string) {
	tc.paths = append(tc.paths, path)
}

func (tc *godogFeaturesScenario) parseFeatures() error {
	fts, err := parser.ParseFeatures(storage.FS{}, "", "", tc.paths)
	if err != nil {
		return err
	}

	tc.features = append(tc.features, fts...)

	return nil
}

func (tc *godogFeaturesScenario) theSuiteShouldHave(state string) error {
	if tc.testedSuite.failed && state == "passed" {
		return fmt.Errorf("the feature suite has failed")
	}

	if !tc.testedSuite.failed && state == "failed" {
		return fmt.Errorf("the feature suite has passed")
	}

	return nil
}

func (tc *godogFeaturesScenario) iShouldHaveNumFeatureFiles(num int, files *DocString) error {
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

func (tc *godogFeaturesScenario) iRunFeatureSuite() error {
	return tc.iRunFeatureSuiteWithTags("")
}

func (tc *godogFeaturesScenario) numScenariosRegistered(expected int) (err error) {
	var num int
	for _, ft := range tc.features {
		num += len(ft.Pickles)
	}

	if num != expected {
		err = fmt.Errorf("expected %d scenarios to be registered, but got %d", expected, num)
	}

	return
}

func (tc *godogFeaturesScenario) thereWereNumEventsFired(_ string, expected int, typ string) error {
	var num int
	for _, event := range tc.events {
		if event.name == typ {
			num++
		}
	}

	if num != expected {
		if typ == "BeforeFeature" || typ == "AfterFeature" {
			return nil
		}

		return fmt.Errorf("expected %d %s events to be fired, but got %d", expected, typ, num)
	}

	return nil
}

func (tc *godogFeaturesScenario) thereWasEventTriggeredBeforeScenario(expected string) error {
	var found []string
	for _, event := range tc.events {
		if event.name != "BeforeScenario" {
			continue
		}

		var name string
		switch t := event.args[0].(type) {
		case *Scenario:
			name = t.Name
		}

		if name == expected {
			return nil
		}

		found = append(found, name)
	}

	if len(found) == 0 {
		return fmt.Errorf("before scenario event was never triggered or listened")
	}

	return fmt.Errorf(`expected "%s" scenario, but got these fired %s`, expected, `"`+strings.Join(found, `", "`)+`"`)
}

func (tc *godogFeaturesScenario) theseEventsHadToBeFiredForNumberOfTimes(tbl *Table) error {
	if len(tbl.Rows[0].Cells) != 2 {
		return fmt.Errorf("expected two columns for event table row, got: %d", len(tbl.Rows[0].Cells))
	}

	for _, row := range tbl.Rows {
		num, err := strconv.ParseInt(row.Cells[1].Value, 10, 0)
		if err != nil {
			return err
		}

		if err := tc.thereWereNumEventsFired("", int(num), row.Cells[0].Value); err != nil {
			return err
		}
	}

	return nil
}

func (tc *godogFeaturesScenario) theRenderJSONWillBe(docstring *DocString) error {
	expectedSuiteCtxReg := regexp.MustCompile(`suite_context.go:\d+`)
	actualSuiteCtxReg := regexp.MustCompile(`(suite_context_test\.go|\\u003cautogenerated\\u003e):\d+`)

	expectedString := docstring.Content
	expectedString = expectedSuiteCtxReg.ReplaceAllString(expectedString, `<autogenerated>:0`)

	actualString := tc.out.String()
	actualString = actualSuiteCtxReg.ReplaceAllString(actualString, `<autogenerated>:0`)

	var expected []formatters.CukeFeatureJSON
	if err := json.Unmarshal([]byte(expectedString), &expected); err != nil {
		return err
	}

	var actual []formatters.CukeFeatureJSON
	if err := json.Unmarshal([]byte(actualString), &actual); err != nil {
		return err
	}

	return assertExpectedAndActual(assert.Equal, expected, actual)
}

func (tc *godogFeaturesScenario) theRenderOutputWillBe(docstring *DocString) error {
	expectedSuiteCtxReg := regexp.MustCompile(`(suite_context\.go|suite_context_test\.go):\d+`)
	actualSuiteCtxReg := regexp.MustCompile(`(suite_context_test\.go|\<autogenerated\>):\d+`)

	expectedSuiteCtxFuncReg := regexp.MustCompile(`SuiteContext.func(\d+)`)
	actualSuiteCtxFuncReg := regexp.MustCompile(`github.com/cucumber/godog.InitializeScenario.func(\d+)`)

	suiteCtxPtrReg := regexp.MustCompile(`\*suiteContext`)

	expected := docstring.Content
	expected = trimAllLines(expected)
	expected = expectedSuiteCtxReg.ReplaceAllString(expected, "<autogenerated>:0")
	expected = expectedSuiteCtxFuncReg.ReplaceAllString(expected, "InitializeScenario.func$1")
	expected = suiteCtxPtrReg.ReplaceAllString(expected, "*godogFeaturesScenario")

	actual := tc.out.String()
	actual = actualSuiteCtxReg.ReplaceAllString(actual, "<autogenerated>:0")
	actual = actualSuiteCtxFuncReg.ReplaceAllString(actual, "InitializeScenario.func$1")
	actualTrimmed := actual
	actual = trimAllLines(actual)

	return assertExpectedAndActual(assert.Equal, expected, actual, actualTrimmed)
}

func (tc *godogFeaturesScenario) theRenderXMLWillBe(docstring *DocString) error {
	expectedString := docstring.Content
	actualString := tc.out.String()

	var expected formatters.JunitPackageSuite
	if err := xml.Unmarshal([]byte(expectedString), &expected); err != nil {
		return err
	}

	var actual formatters.JunitPackageSuite
	if err := xml.Unmarshal([]byte(actualString), &actual); err != nil {
		return err
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

func trimAllLines(s string) string {
	var lines []string
	for _, ln := range strings.Split(strings.TrimSpace(s), "\n") {
		lines = append(lines, strings.TrimSpace(ln))
	}
	return strings.Join(lines, "\n")
}

func TestScenarioContext_After_cancelled(t *testing.T) {
	ctxDone := make(chan struct{})
	suite := TestSuite{
		ScenarioInitializer: func(scenarioContext *ScenarioContext) {
			scenarioContext.When(`^foo$`, func() {})
			scenarioContext.After(func(ctx context.Context, sc *Scenario, err error) (context.Context, error) {
				go func() {
					<-ctx.Done()
					close(ctxDone)
				}()

				return ctx, nil
			})
		},
		Options: &Options{
			Format:   "pretty",
			TestingT: t,
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

	require.Equal(t, 0, suite.Run(), "non-zero status returned, failed to run feature tests")

	select {
	case <-ctxDone:
		return
	case <-time.After(5 * time.Second):
		assert.Fail(t, "failed to wait for context cancellation")
	}
}

func TestTestSuite_Run(t *testing.T) {
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
					When something unknown happens
					Then step passes`,
			log: `
					>>>> Before suite
					>> Before scenario "test"
					Before step "something unknown happens"
					After step "something unknown happens", error: step is undefined: something unknown happens, status: undefined
					Before step "step passes"
					After step "step passes", error: <nil>, status: skipped
					<< After scenario "test", error: <nil>
					<<<< After suite`,
		},
		{
			name: "undefined_then_pass_fails_scenario", afterStepCnt: 2, beforeStepCnt: 2,
			body: `
					When something unknown happens
					Then step passes`,
			log: `
					>>>> Before suite
					>> Before scenario "test"
					Before step "something unknown happens"
					After step "something unknown happens", error: step is undefined: something unknown happens, status: undefined
					<< After scenario "test", error: step is undefined: something unknown happens
					Before step "step passes"
					After step "step passes", error: <nil>, status: skipped
					<<<< After suite`,
		},
		{
			name: "fail_then_undefined_fails_scenario", afterStepCnt: 2, beforeStepCnt: 2,
			body: `
					When step fails
					Then something unknown happens`,
			log: `
					>>>> Before suite
					>> Before scenario "test"
					Before step "step fails"
					After step "step fails", error: oops, status: failed
					<< After scenario "test", error: oops
					Before step "something unknown happens"
					After step "something unknown happens", error: step is undefined: something unknown happens, status: undefined
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
							Contents: []byte(trimAllLines(`
								Feature: test
								Scenario: test
								` + tc.body)),
						},
					},
				},
			}

			suitePasses := suite.Run() == 0
			assert.Equal(t, tc.suitePasses, suitePasses)
			assert.Equal(t, 1, afterScenarioCnt)
			assert.Equal(t, 1, beforeScenarioCnt)
			assert.Equal(t, tc.afterStepCnt, afterStepCnt)
			assert.Equal(t, tc.beforeStepCnt, beforeStepCnt)
			assert.Equal(t, trimAllLines(tc.log), trimAllLines(log), log)
		})
	}
}
