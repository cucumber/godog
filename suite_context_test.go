package godog

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cucumber/gherkin-go/v11"
	"github.com/cucumber/messages-go/v10"
	"github.com/stretchr/testify/assert"

	"github.com/cucumber/godog/colors"
)

func InitializeScenario(ctx *ScenarioContext) {
	tc := &godogFeaturesScenario{}

	ctx.BeforeScenario(tc.ResetBeforeEachScenario)

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

	ctx.Step(`^(?:a )?passing multistep$`, func() Steps {
		return Steps{"passing step", "passing step", "passing step"}
	})

	ctx.Step(`^(?:a )?failing nested multistep$`, func() Steps {
		return Steps{"passing step", "passing multistep", "failing multistep"}
	})
	// Default recovery step
	ctx.Step(`Ignore.*`, func() error {
		return nil
	})

	ctx.BeforeStep(tc.inject)
}

func (tc *godogFeaturesScenario) inject(step *Step) {
	if !tc.allowInjection {
		return
	}

	step.Text = injectAll(step.Text)

	if table := step.Argument.GetDataTable(); table != nil {
		for i := 0; i < len(table.Rows); i++ {
			for n, cell := range table.Rows[i].Cells {
				table.Rows[i].Cells[n].Value = injectAll(cell.Value)
			}
		}
	}

	if doc := step.Argument.GetDocString(); doc != nil {
		doc.Content = injectAll(doc.Content)
	}
}

type godogFeaturesScenario struct {
	paths          []string
	testedSuite    *Suite
	events         []*firedEvent
	out            bytes.Buffer
	allowInjection bool
}

func (tc *godogFeaturesScenario) ResetBeforeEachScenario(*Scenario) {
	// reset whole suite with the state
	tc.out.Reset()
	tc.paths = []string{}

	tc.testedSuite = &Suite{scenarioInitializer: InitializeScenario}

	// reset all fired events
	tc.events = []*firedEvent{}
	tc.allowInjection = false
}

func (tc *godogFeaturesScenario) iSetVariableInjectionTo(to string) error {
	tc.allowInjection = to == "allow"
	return nil
}

func (tc *godogFeaturesScenario) iRunFeatureSuiteWithTags(tags string) error {
	if err := tc.parseFeatures(); err != nil {
		return err
	}

	for _, feat := range tc.testedSuite.features {
		feat.pickles = applyTagFilter(tags, feat.pickles)
	}

	tc.testedSuite.storage = newStorage()
	for _, feat := range tc.testedSuite.features {
		tc.testedSuite.storage.mustInsertFeature(feat)

		for _, pickle := range feat.pickles {
			tc.testedSuite.storage.mustInsertPickle(pickle)
		}
	}

	fmt := newBaseFmt("godog", &tc.out)
	fmt.setStorage(tc.testedSuite.storage)
	tc.testedSuite.fmt = fmt

	testRunStarted := testRunStarted{StartedAt: timeNowFunc()}
	tc.testedSuite.storage.mustInsertTestRunStarted(testRunStarted)

	tc.testedSuite.fmt.TestRunStarted()
	tc.testedSuite.run()
	tc.testedSuite.fmt.Summary()

	return nil
}

func (tc *godogFeaturesScenario) iRunFeatureSuiteWithFormatter(name string) error {
	if err := tc.parseFeatures(); err != nil {
		return err
	}

	f := FindFmt(name)
	if f == nil {
		return fmt.Errorf(`formatter "%s" is not available`, name)
	}

	tc.testedSuite.storage = newStorage()
	for _, feat := range tc.testedSuite.features {
		tc.testedSuite.storage.mustInsertFeature(feat)

		for _, pickle := range feat.pickles {
			tc.testedSuite.storage.mustInsertPickle(pickle)
		}
	}

	tc.testedSuite.fmt = f("godog", colors.Uncolored(&tc.out))
	if fmt, ok := tc.testedSuite.fmt.(storageFormatter); ok {
		fmt.setStorage(tc.testedSuite.storage)
	}

	testRunStarted := testRunStarted{StartedAt: timeNowFunc()}
	tc.testedSuite.storage.mustInsertTestRunStarted(testRunStarted)

	tc.testedSuite.fmt.TestRunStarted()
	tc.testedSuite.run()
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
	f, ok := tc.testedSuite.fmt.(*basefmt)
	if !ok {
		return fmt.Errorf("this step requires *basefmt, but there is: %T", tc.testedSuite.fmt)
	}

	actual := tc.cleanupSnippet(f.snippets())
	expected := tc.cleanupSnippet(body.Content)

	if actual != expected {
		return fmt.Errorf("snippets do not match actual: %s", f.snippets())
	}

	return nil
}

func (tc *godogFeaturesScenario) followingStepsShouldHave(status string, steps *DocString) error {
	var expected = strings.Split(steps.Content, "\n")
	var actual, unmatched, matched []string

	f, ok := tc.testedSuite.fmt.(*basefmt)
	if !ok {
		return fmt.Errorf("this step requires *basefmt, but there is: %T", tc.testedSuite.fmt)
	}

	switch status {
	case "passed":
		for _, st := range f.storage.mustGetPickleStepResultsByStatus(passed) {
			pickleStep := f.storage.mustGetPickleStep(st.PickleStepID)
			actual = append(actual, pickleStep.Text)
		}
	case "failed":
		for _, st := range f.storage.mustGetPickleStepResultsByStatus(failed) {
			pickleStep := f.storage.mustGetPickleStep(st.PickleStepID)
			actual = append(actual, pickleStep.Text)
		}
	case "skipped":
		for _, st := range f.storage.mustGetPickleStepResultsByStatus(skipped) {
			pickleStep := f.storage.mustGetPickleStep(st.PickleStepID)
			actual = append(actual, pickleStep.Text)
		}
	case "undefined":
		for _, st := range f.storage.mustGetPickleStepResultsByStatus(undefined) {
			pickleStep := f.storage.mustGetPickleStep(st.PickleStepID)
			actual = append(actual, pickleStep.Text)
		}
	case "pending":
		for _, st := range f.storage.mustGetPickleStepResultsByStatus(pending) {
			pickleStep := f.storage.mustGetPickleStep(st.PickleStepID)
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
	tc.testedSuite.BeforeSuite(func() {
		tc.events = append(tc.events, &firedEvent{"BeforeSuite", []interface{}{}})
	})

	tc.testedSuite.AfterSuite(func() {
		tc.events = append(tc.events, &firedEvent{"AfterSuite", []interface{}{}})
	})

	tc.testedSuite.BeforeFeature(func(ft *messages.GherkinDocument) {
		tc.events = append(tc.events, &firedEvent{"BeforeFeature", []interface{}{ft}})
	})

	tc.testedSuite.AfterFeature(func(ft *messages.GherkinDocument) {
		tc.events = append(tc.events, &firedEvent{"AfterFeature", []interface{}{ft}})
	})

	tc.testedSuite.BeforeScenario(func(pickle *Scenario) {
		tc.events = append(tc.events, &firedEvent{"BeforeScenario", []interface{}{pickle}})
	})

	tc.testedSuite.AfterScenario(func(pickle *Scenario, err error) {
		tc.events = append(tc.events, &firedEvent{"AfterScenario", []interface{}{pickle, err}})
	})

	tc.testedSuite.BeforeStep(func(step *Step) {
		tc.events = append(tc.events, &firedEvent{"BeforeStep", []interface{}{step}})
	})

	tc.testedSuite.AfterStep(func(step *Step, err error) {
		tc.events = append(tc.events, &firedEvent{"AfterStep", []interface{}{step, err}})
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
	tc.testedSuite.features = append(tc.testedSuite.features, &feature{GherkinDocument: gd, pickles: pickles})

	return err
}

func (tc *godogFeaturesScenario) featurePath(path string) error {
	tc.paths = append(tc.paths, path)
	return nil
}

func (tc *godogFeaturesScenario) parseFeatures() error {
	fts, err := parseFeatures("", tc.paths)
	if err != nil {
		return err
	}

	tc.testedSuite.features = append(tc.testedSuite.features, fts...)

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
	if len(tc.testedSuite.features) != num {
		return fmt.Errorf("expected %d features to be parsed, but have %d", num, len(tc.testedSuite.features))
	}

	expected := strings.Split(files.Content, "\n")

	var actual []string

	for _, ft := range tc.testedSuite.features {
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
	for _, ft := range tc.testedSuite.features {
		num += len(ft.pickles)
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
	actualSuiteCtxReg := regexp.MustCompile(`suite_context_test.go:\d+`)

	expectedString := docstring.Content
	expectedString = expectedSuiteCtxReg.ReplaceAllString(expectedString, `suite_context_test.go:0`)

	actualString := tc.out.String()
	actualString = actualSuiteCtxReg.ReplaceAllString(actualString, `suite_context_test.go:0`)

	var expected []cukeFeatureJSON
	if err := json.Unmarshal([]byte(expectedString), &expected); err != nil {
		return err
	}

	var actual []cukeFeatureJSON
	if err := json.Unmarshal([]byte(actualString), &actual); err != nil {
		return err
	}

	return assertExpectedAndActual(assert.Equal, expected, actual)
}

func (tc *godogFeaturesScenario) theRenderOutputWillBe(docstring *DocString) error {
	expectedSuiteCtxReg := regexp.MustCompile(`suite_context.go:\d+`)
	actualSuiteCtxReg := regexp.MustCompile(`suite_context_test.go:\d+`)

	expectedSuiteCtxFuncReg := regexp.MustCompile(`SuiteContext.func(\d+)`)
	actualSuiteCtxFuncReg := regexp.MustCompile(`github.com/cucumber/godog.InitializeScenario.func(\d+)`)

	suiteCtxPtrReg := regexp.MustCompile(`\*suiteContext`)

	expected := docstring.Content
	expected = trimAllLines(expected)
	expected = expectedSuiteCtxReg.ReplaceAllString(expected, "suite_context_test.go:0")
	expected = expectedSuiteCtxFuncReg.ReplaceAllString(expected, "InitializeScenario.func$1")
	expected = suiteCtxPtrReg.ReplaceAllString(expected, "*godogFeaturesScenario")

	actual := tc.out.String()
	actual = trimAllLines(actual)
	actual = actualSuiteCtxReg.ReplaceAllString(actual, "suite_context_test.go:0")
	actual = actualSuiteCtxFuncReg.ReplaceAllString(actual, "InitializeScenario.func$1")

	expectedRows := strings.Split(expected, "\n")
	actualRows := strings.Split(actual, "\n")

	return assertExpectedAndActual(assert.ElementsMatch, expectedRows, actualRows)
}

func (tc *godogFeaturesScenario) theRenderXMLWillBe(docstring *DocString) error {
	expectedString := docstring.Content
	actualString := tc.out.String()

	var expected junitPackageSuite
	if err := xml.Unmarshal([]byte(expectedString), &expected); err != nil {
		return err
	}

	var actual junitPackageSuite
	if err := xml.Unmarshal([]byte(actualString), &actual); err != nil {
		return err
	}

	return assertExpectedAndActual(assert.Equal, expected, actual)
}
