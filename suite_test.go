package godog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/DATA-DOG/godog/gherkin"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	status := RunWithOptions("godogs", func(s *Suite) {
		SuiteContext(s)
	}, Options{
		Format:      "progress",
		Paths:       []string{"features"},
		Concurrency: 4,
	})

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func SuiteContext(s *Suite) {
	c := &suiteContext{}

	s.BeforeScenario(c.ResetBeforeEachScenario)

	s.Step(`^a feature path "([^"]*)"$`, c.featurePath)
	s.Step(`^I parse features$`, c.parseFeatures)
	s.Step(`^I'm listening to suite events$`, c.iAmListeningToSuiteEvents)
	s.Step(`^I run feature suite$`, c.iRunFeatureSuite)
	s.Step(`^I run feature suite with formatter "([^"]*)"$`, c.iRunFeatureSuiteWithFormatter)
	s.Step(`^a feature "([^"]*)" file:$`, c.aFeatureFile)
	s.Step(`^the suite should have (passed|failed)$`, c.theSuiteShouldHave)

	s.Step(`^I should have ([\d]+) features? files?:$`, c.iShouldHaveNumFeatureFiles)
	s.Step(`^I should have ([\d]+) scenarios? registered$`, c.numScenariosRegistered)
	s.Step(`^there (was|were) ([\d]+) "([^"]*)" events? fired$`, c.thereWereNumEventsFired)
	s.Step(`^there was event triggered before scenario "([^"]*)"$`, c.thereWasEventTriggeredBeforeScenario)
	s.Step(`^these events had to be fired for a number of times:$`, c.theseEventsHadToBeFiredForNumberOfTimes)

	s.Step(`^a failing step`, c.aFailingStep)
	s.Step(`^this step should fail`, c.aFailingStep)
	s.Step(`^the following steps? should be (passed|failed|skipped|undefined|pending):`, c.followingStepsShouldHave)
	s.Step(`^the undefined step snippets should be:$`, c.theUndefinedStepSnippetsShouldBe)

	// event stream
	s.Step(`^the following events should be fired:$`, c.thereShouldBeEventsFired)

	// lt
	s.Step(`^savybių aplankas "([^"]*)"$`, c.featurePath)
	s.Step(`^aš išskaitau savybes$`, c.parseFeatures)
	s.Step(`^aš turėčiau turėti ([\d]+) savybių failus:$`, c.iShouldHaveNumFeatureFiles)

	s.Step(`^pending step$`, func() error {
		return ErrPending
	})
	s.Step(`^passing step$`, func() error {
		return nil
	})

	// duplicate step to 'a failing step' I added to help test cucumber.feature
	// I needed to have an Scenario Outline where the status was passing or failing
	// I needed the same step def language.
	s.Step(`^failing step$`, c.aFailingStep)

	// Introduced to test formatter/cucumber.feature
	s.Step(`^the rendered json will be as follows:$`, c.theRenderJSONWillBe)

}

type firedEvent struct {
	name string
	args []interface{}
}

type suiteContext struct {
	paths       []string
	testedSuite *Suite
	events      []*firedEvent
	out         bytes.Buffer
}

func (s *suiteContext) ResetBeforeEachScenario(interface{}) {
	// reset whole suite with the state
	s.out.Reset()
	s.paths = []string{}
	s.testedSuite = &Suite{}
	// our tested suite will have the same context registered
	SuiteContext(s.testedSuite)
	// reset all fired events
	s.events = []*firedEvent{}
}

func (s *suiteContext) iRunFeatureSuiteWithFormatter(name string) error {
	f, err := findFmt(name)
	if err != nil {
		return err
	}
	s.testedSuite.fmt = f("godog", &s.out)
	if err := s.parseFeatures(); err != nil {
		return err
	}
	s.testedSuite.run()
	s.testedSuite.fmt.Summary()
	return nil
}

func (s *suiteContext) thereShouldBeEventsFired(doc *gherkin.DocString) error {
	actual := strings.Split(strings.TrimSpace(s.out.String()), "\n")
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

func (s *suiteContext) cleanupSnippet(snip string) string {
	lines := strings.Split(strings.TrimSpace(snip), "\n")
	for i := 0; i < len(lines); i++ {
		lines[i] = strings.TrimSpace(lines[i])
	}
	return strings.Join(lines, "\n")
}

func (s *suiteContext) theUndefinedStepSnippetsShouldBe(body *gherkin.DocString) error {
	f, ok := s.testedSuite.fmt.(*testFormatter)
	if !ok {
		return fmt.Errorf("this step requires testFormatter, but there is: %T", s.testedSuite.fmt)
	}
	actual := s.cleanupSnippet(f.snippets())
	expected := s.cleanupSnippet(body.Content)
	if actual != expected {
		return fmt.Errorf("snippets do not match actual: %s", f.snippets())
	}
	return nil
}

func (s *suiteContext) followingStepsShouldHave(status string, steps *gherkin.DocString) error {
	var expected = strings.Split(steps.Content, "\n")
	var actual, unmatched, matched []string

	f, ok := s.testedSuite.fmt.(*testFormatter)
	if !ok {
		return fmt.Errorf("this step requires testFormatter, but there is: %T", s.testedSuite.fmt)
	}
	switch status {
	case "passed":
		for _, st := range f.passed {
			actual = append(actual, st.step.Text)
		}
	case "failed":
		for _, st := range f.failed {
			actual = append(actual, st.step.Text)
		}
	case "skipped":
		for _, st := range f.skipped {
			actual = append(actual, st.step.Text)
		}
	case "undefined":
		for _, st := range f.undefined {
			actual = append(actual, st.step.Text)
		}
	case "pending":
		for _, st := range f.pending {
			actual = append(actual, st.step.Text)
		}
	default:
		return fmt.Errorf("unexpected step status wanted: %s", status)
	}

	if len(expected) > len(actual) {
		return fmt.Errorf("number of expeted %s steps: %d is less than actual %s steps: %d", status, len(expected), status, len(actual))
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

func (s *suiteContext) iAmListeningToSuiteEvents() error {
	s.testedSuite.BeforeSuite(func() {
		s.events = append(s.events, &firedEvent{"BeforeSuite", []interface{}{}})
	})
	s.testedSuite.AfterSuite(func() {
		s.events = append(s.events, &firedEvent{"AfterSuite", []interface{}{}})
	})
	s.testedSuite.BeforeScenario(func(scenario interface{}) {
		s.events = append(s.events, &firedEvent{"BeforeScenario", []interface{}{scenario}})
	})
	s.testedSuite.AfterScenario(func(scenario interface{}, err error) {
		s.events = append(s.events, &firedEvent{"AfterScenario", []interface{}{scenario, err}})
	})
	s.testedSuite.BeforeStep(func(step *gherkin.Step) {
		s.events = append(s.events, &firedEvent{"BeforeStep", []interface{}{step}})
	})
	s.testedSuite.AfterStep(func(step *gherkin.Step, err error) {
		s.events = append(s.events, &firedEvent{"AfterStep", []interface{}{step, err}})
	})
	return nil
}

func (s *suiteContext) aFailingStep() error {
	return fmt.Errorf("intentional failure")
}

// parse a given feature file body as a feature
func (s *suiteContext) aFeatureFile(name string, body *gherkin.DocString) error {
	ft, err := gherkin.ParseFeature(strings.NewReader(body.Content))
	s.testedSuite.features = append(s.testedSuite.features, &feature{Feature: ft, Path: name})
	return err
}

func (s *suiteContext) featurePath(path string) error {
	s.paths = append(s.paths, path)
	return nil
}

func (s *suiteContext) parseFeatures() error {
	fts, err := parseFeatures("", s.paths)
	if err != nil {
		return err
	}
	s.testedSuite.features = append(s.testedSuite.features, fts...)
	return nil
}

func (s *suiteContext) theSuiteShouldHave(state string) error {
	if s.testedSuite.failed && state == "passed" {
		return fmt.Errorf("the feature suite has failed")
	}
	if !s.testedSuite.failed && state == "failed" {
		return fmt.Errorf("the feature suite has passed")
	}
	return nil
}

func (s *suiteContext) iShouldHaveNumFeatureFiles(num int, files *gherkin.DocString) error {
	if len(s.testedSuite.features) != num {
		return fmt.Errorf("expected %d features to be parsed, but have %d", num, len(s.testedSuite.features))
	}
	expected := strings.Split(files.Content, "\n")
	var actual []string
	for _, ft := range s.testedSuite.features {
		actual = append(actual, ft.Path)
	}
	if len(expected) != len(actual) {
		return fmt.Errorf("expected %d feature paths to be parsed, but have %d", len(expected), len(actual))
	}
	for i := 0; i < len(expected); i++ {
		split := strings.Split(expected[i], "/")
		exp := filepath.Join(split...)
		split = strings.Split(actual[i], "/")
		act := filepath.Join(split...)
		if exp != act {
			return fmt.Errorf(`expected feature path "%s" at position: %d, does not match actual "%s"`, exp, i, act)
		}
	}
	return nil
}

func (s *suiteContext) iRunFeatureSuite() error {
	if err := s.parseFeatures(); err != nil {
		return err
	}
	s.testedSuite.fmt = testFormatterFunc("godog", &s.out)
	s.testedSuite.run()
	s.testedSuite.fmt.Summary()

	return nil
}

func (s *suiteContext) numScenariosRegistered(expected int) (err error) {
	var num int
	for _, ft := range s.testedSuite.features {
		num += len(ft.ScenarioDefinitions)
	}
	if num != expected {
		err = fmt.Errorf("expected %d scenarios to be registered, but got %d", expected, num)
	}
	return
}

func (s *suiteContext) thereWereNumEventsFired(_ string, expected int, typ string) error {
	var num int
	for _, event := range s.events {
		if event.name == typ {
			num++
		}
	}
	if num != expected {
		return fmt.Errorf("expected %d %s events to be fired, but got %d", expected, typ, num)
	}
	return nil
}

func (s *suiteContext) thereWasEventTriggeredBeforeScenario(expected string) error {
	var found []string
	for _, event := range s.events {
		if event.name != "BeforeScenario" {
			continue
		}

		var name string
		switch t := event.args[0].(type) {
		case *gherkin.Scenario:
			name = t.Name
		case *gherkin.ScenarioOutline:
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

func (s *suiteContext) theseEventsHadToBeFiredForNumberOfTimes(tbl *gherkin.DataTable) error {
	if len(tbl.Rows[0].Cells) != 2 {
		return fmt.Errorf("expected two columns for event table row, got: %d", len(tbl.Rows[0].Cells))
	}

	for _, row := range tbl.Rows {
		num, err := strconv.ParseInt(row.Cells[1].Value, 10, 0)
		if err != nil {
			return err
		}
		if err := s.thereWereNumEventsFired("", int(num), row.Cells[0].Value); err != nil {
			return err
		}
	}
	return nil
}

func (s *suiteContext) theRenderJSONWillBe(docstring *gherkin.DocString) error {

	var expected interface{}
	if err := json.Unmarshal([]byte(docstring.Content), &expected); err != nil {
		return err
	}

	var actual interface{}
	if err := json.Unmarshal(s.out.Bytes(), &actual); err != nil {
		return err
	}

	expectedArr := expected.([]interface{})
	actualArr := actual.([]interface{})

	// Created to use in error reporting.
	expectedCompact := &bytes.Buffer{}
	actualCompact := &bytes.Buffer{}
	json.Compact(expectedCompact,[]byte(docstring.Content))
	json.Compact(actualCompact,s.out.Bytes())

	for idx, entry := range expectedArr {

		// Make sure all of the expected are in the actual
		if err := s.mapCompareStructure(entry.(map[string]interface{}), actualArr[idx].(map[string]interface{})); err != nil {
			return fmt.Errorf("err:%v actual result is missing fields: expected:%s actual:%s",err,expectedCompact, actualCompact)
		}

		// Make sure all of actual are in expected
		if err := s.mapCompareStructure(actualArr[idx].(map[string]interface{}),entry.(map[string]interface{})); err != nil {
			return fmt.Errorf("err:%v actual result contains too many fields: expected:%s actual:%s",err,expectedCompact, actualCompact)
		}

		// Make sure the values are correct
		if err := s.mapCompare(entry.(map[string]interface{}), actualArr[idx].(map[string]interface{})); err != nil {
			return fmt.Errorf("err:%v values don't match expected:%s actual:%s",err,expectedCompact, actualCompact)
		}
	}
	return nil
}

/*
  Due to specialize matching logic to ignore exact matches on the "location" and "duration" fields.  It was
  necessary to create this compare function to validate the values of the map.
*/
func (s *suiteContext) mapCompare(expected map[string]interface{}, actual map[string]interface{}) error {

	// Process all keys in the map and handle them based on the type of the field.
	for k, v := range expected {

		if actual[k] == nil {
			return fmt.Errorf("No matching field in actual:[%s] expected value:[%v]", k, v)
		}
		// Process other maps via recursion
		if reflect.TypeOf(v).Kind() == reflect.Map {
			if err := s.mapCompare(v.(map[string]interface{}), actual[k].(map[string]interface{})); err != nil {
				return err
			}
			// This is an array of maps show as a slice
		} else if reflect.TypeOf(v).Kind() == reflect.Slice {
			for i, e := range v.([]interface{}) {
				if err := s.mapCompare(e.(map[string]interface{}), actual[k].([]interface{})[i].(map[string]interface{})); err != nil {
					return err
				}
			}
		// We need special rules to check location so that we are not bound to version of the code.
		} else if k == "location" {

			// location is tricky.  the cucumber value is either a the step def location for passed,failed, and skipped.
			// it is the feature file location for undefined and skipped.
			// I dont have the result context readily available so the expected input will have
			// the context i need contained within its value.
			// FEATURE_PATH myfile.feature:20 or
			// STEP_ID
			t := strings.Split(v.(string)," ")
			if t[0] == "FEATURE_PATH" {
				if actual[k].(string) != t[1]{
					return fmt.Errorf("location has unexpected value [%s] should be [%s]",
						actual[k], t[1])
				}

			} else if t[0] == "STEP_ID" {
				if !strings.Contains(actual[k].(string), "suite_test.go:") {
					return fmt.Errorf("location has unexpected filename [%s] should contain suite_test.go",
						actual[k])
				}

			} else {
				return fmt.Errorf("Bad location value [%v]",v)
			}

		// We need special rules to validate duration too.
		} else if k == "duration" {
			if actual[k].(float64) <= 0 {
				return fmt.Errorf("duration is <= zero: actual:[%v]", actual[k])
			}
		// default numbers in json are coming as float64
		} else if reflect.TypeOf(v).Kind() == reflect.Float64 {
			if v.(float64) != actual[k].(float64) {
				if v.(float64) != actual[k].(float64) {
					return fmt.Errorf("Field:[%s] not matching expected:[%v] actual:[%v]",
						k, v, actual[k])
				}
			}

		} else if reflect.TypeOf(v).Kind() == reflect.String {
			if v.(string) != actual[k].(string) {
				return fmt.Errorf("Field:[%s] not matching expected:[%v] actual:[%v]",
					k, v, actual[k])
			}
		} else {
			return fmt.Errorf("Unexepcted type encountered in json at key:[%s] Type:[%v]", k, reflect.TypeOf(v).Kind())
		}
	}

	return nil
}
/*
  Due to specialize matching logic to ignore exact matches on the "location" and "duration" fields.  It was
  necessary to create this compare function to validate the values of the map.
*/
func (s *suiteContext) mapCompareStructure(expected map[string]interface{}, actual map[string]interface{}) error {

	// Process all keys in the map and handle them based on the type of the field.
	for k, v := range expected {

		if actual[k] == nil {
			return fmt.Errorf("Structure Mismatch: no matching field:[%s] expected value:[%v]",k, v)
		}
		// Process other maps via recursion
		if reflect.TypeOf(v).Kind() == reflect.Map {
			if err := s.mapCompareStructure(v.(map[string]interface{}), actual[k].(map[string]interface{})); err != nil {
				return err
			}
			// This is an array of maps show as a slice
		} else if reflect.TypeOf(v).Kind() == reflect.Slice {
			for i, e := range v.([]interface{}) {
				if err := s.mapCompareStructure(e.(map[string]interface{}), actual[k].([]interface{})[i].(map[string]interface{})); err != nil {
					return err
				}
			}
		}
	}

	return nil
}