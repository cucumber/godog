package godog

import (
	"fmt"
	"strings"

	"github.com/DATA-DOG/godog/gherkin"
)

func SuiteContext(s Suite) {
	c := &suiteContext{}

	s.BeforeScenario(c)

	s.Step(`^a feature path "([^"]*)"$`, c.featurePath)
	s.Step(`^I parse features$`, c.parseFeatures)
	s.Step(`^I'm listening to suite events$`, c.iAmListeningToSuiteEvents)
	s.Step(`^I run feature suite$`, c.iRunFeatureSuite)
	s.Step(`^a feature "([^"]*)" file:$`, c.aFeatureFile)
	s.Step(`^the suite should have (passed|failed)$`, c.theSuiteShouldHave)

	s.Step(`^I should have ([\d]+) features? files?:$`, c.iShouldHaveNumFeatureFiles)
	s.Step(`^I should have ([\d]+) scenarios? registered$`, c.numScenariosRegistered)
	s.Step(`^there (was|were) ([\d]+) "([^"]*)" events? fired$`, c.thereWereNumEventsFired)
	s.Step(`^there was event triggered before scenario "([^"]*)"$`, c.thereWasEventTriggeredBeforeScenario)
	s.Step(`^these events had to be fired for a number of times:$`, c.theseEventsHadToBeFiredForNumberOfTimes)

	s.Step(`^a failing step`, c.aFailingStep)
	s.Step(`^this step should fail`, c.aFailingStep)
	s.Step(`^the following steps? should be (passed|failed|skipped|undefined):`, c.followingStepsShouldHave)
}

type firedEvent struct {
	name string
	args []interface{}
}

type suiteContext struct {
	testedSuite *suite
	events      []*firedEvent
	fmt         *testFormatter
}

func (s *suiteContext) HandleBeforeScenario(*gherkin.Scenario) {
	// reset whole suite with the state
	s.fmt = &testFormatter{}
	s.testedSuite = &suite{fmt: s.fmt}
	// our tested suite will have the same context registered
	SuiteContext(s.testedSuite)
	// reset feature paths
	cfg.paths = []string{}
	// reset all fired events
	s.events = []*firedEvent{}
}

func (s *suiteContext) followingStepsShouldHave(args ...*Arg) error {
	var expected []string = args[1].PyString().Lines
	var actual, unmatched []string
	var matched []int

	switch args[0].String() {
	case "passed":
		for _, st := range s.fmt.passed {
			actual = append(actual, st.step.Text)
		}
	case "failed":
		for _, st := range s.fmt.failed {
			actual = append(actual, st.step.Text)
		}
	case "skipped":
		for _, st := range s.fmt.skipped {
			actual = append(actual, st.step.Text)
		}
	case "undefined":
		for _, st := range s.fmt.undefined {
			actual = append(actual, st.step.Text)
		}
	default:
		return fmt.Errorf("unexpected step status wanted: %s", args[0].String())
	}

	if len(expected) > len(actual) {
		return fmt.Errorf("number of expected %s steps: %d is less than actual %s steps: %d", args[0].String(), len(expected), args[0].String(), len(actual))
	}

	for _, a := range actual {
		for i, e := range expected {
			if a == e {
				matched = append(matched, i)
				break
			}
		}
	}

	if len(matched) == len(expected) {
		return nil
	}

	for i, s := range expected {
		var found bool
		for _, m := range matched {
			if i == m {
				found = true
			}
		}
		if !found {
			unmatched = append(unmatched, s)
		}
	}

	return fmt.Errorf("the steps: %s - is not %s", strings.Join(unmatched, ", "), args[0].String())
}

func (s *suiteContext) iAmListeningToSuiteEvents(args ...*Arg) error {
	s.testedSuite.BeforeSuite(BeforeSuiteHandlerFunc(func() {
		s.events = append(s.events, &firedEvent{"BeforeSuite", []interface{}{}})
	}))
	s.testedSuite.AfterSuite(AfterSuiteHandlerFunc(func() {
		s.events = append(s.events, &firedEvent{"AfterSuite", []interface{}{}})
	}))
	s.testedSuite.BeforeScenario(BeforeScenarioHandlerFunc(func(scenario *gherkin.Scenario) {
		s.events = append(s.events, &firedEvent{"BeforeScenario", []interface{}{scenario}})
	}))
	s.testedSuite.AfterScenario(AfterScenarioHandlerFunc(func(scenario *gherkin.Scenario, status Status) {
		s.events = append(s.events, &firedEvent{"AfterScenario", []interface{}{scenario, status}})
	}))
	s.testedSuite.BeforeStep(BeforeStepHandlerFunc(func(step *gherkin.Step) {
		s.events = append(s.events, &firedEvent{"BeforeStep", []interface{}{step}})
	}))
	s.testedSuite.AfterStep(AfterStepHandlerFunc(func(step *gherkin.Step, status Status) {
		s.events = append(s.events, &firedEvent{"AfterStep", []interface{}{step, status}})
	}))
	return nil
}

func (s *suiteContext) aFailingStep(...*Arg) error {
	return fmt.Errorf("intentional failure")
}

// parse a given feature file body as a feature
func (s *suiteContext) aFeatureFile(args ...*Arg) error {
	name := args[0].String()
	body := args[1].PyString().Raw
	feature, err := gherkin.Parse(strings.NewReader(body), name)
	s.testedSuite.features = append(s.testedSuite.features, feature)
	return err
}

func (s *suiteContext) featurePath(args ...*Arg) error {
	cfg.paths = append(cfg.paths, args[0].String())
	return nil
}

func (s *suiteContext) parseFeatures(args ...*Arg) error {
	features, err := cfg.features()
	if err != nil {
		return err
	}
	s.testedSuite.features = append(s.testedSuite.features, features...)
	return nil
}

func (s *suiteContext) theSuiteShouldHave(args ...*Arg) error {
	if s.testedSuite.failed && args[0].String() == "passed" {
		return fmt.Errorf("the feature suite has failed")
	}
	if !s.testedSuite.failed && args[0].String() == "failed" {
		return fmt.Errorf("the feature suite has passed")
	}
	return nil
}

func (s *suiteContext) iShouldHaveNumFeatureFiles(args ...*Arg) error {
	if len(s.testedSuite.features) != args[0].Int() {
		return fmt.Errorf("expected %d features to be parsed, but have %d", args[0].Int(), len(s.testedSuite.features))
	}
	expected := args[1].PyString().Lines
	var actual []string
	for _, ft := range s.testedSuite.features {
		actual = append(actual, ft.Path)
	}
	if len(expected) != len(actual) {
		return fmt.Errorf("expected %d feature paths to be parsed, but have %d", len(expected), len(actual))
	}
	for i := 0; i < len(expected); i++ {
		if expected[i] != actual[i] {
			return fmt.Errorf(`expected feature path "%s" at position: %d, does not match actual "%s"`, expected[i], i, actual[i])
		}
	}
	return nil
}

func (s *suiteContext) iRunFeatureSuite(args ...*Arg) error {
	if err := s.parseFeatures(); err != nil {
		return err
	}
	s.testedSuite.run()
	return nil
}

func (s *suiteContext) numScenariosRegistered(args ...*Arg) (err error) {
	var num int
	for _, ft := range s.testedSuite.features {
		num += len(ft.Scenarios)
	}
	if num != args[0].Int() {
		err = fmt.Errorf("expected %d scenarios to be registered, but got %d", args[0].Int(), num)
	}
	return
}

func (s *suiteContext) thereWereNumEventsFired(args ...*Arg) error {
	var num int
	for _, event := range s.events {
		if event.name == args[2].String() {
			num++
		}
	}
	if num != args[1].Int() {
		return fmt.Errorf("expected %d %s events to be fired, but got %d", args[1].Int(), args[2].String(), num)
	}
	return nil
}

func (s *suiteContext) thereWasEventTriggeredBeforeScenario(args ...*Arg) error {
	var found []string
	for _, event := range s.events {
		if event.name != "BeforeScenario" {
			continue
		}

		scenario := event.args[0].(*gherkin.Scenario)
		if scenario.Title == args[0].String() {
			return nil
		}

		found = append(found, scenario.Title)
	}

	if len(found) == 0 {
		return fmt.Errorf("before scenario event was never triggered or listened")
	}

	return fmt.Errorf(`expected "%s" scenario, but got these fired %s`, args[0].String(), `"`+strings.Join(found, `", "`)+`"`)
}

func (s *suiteContext) theseEventsHadToBeFiredForNumberOfTimes(args ...*Arg) error {
	tbl := args[0].Table()
	if len(tbl.Rows[0]) != 2 {
		return fmt.Errorf("expected two columns for event table row, got: %d", len(tbl.Rows[0]))
	}

	for _, row := range tbl.Rows {
		args := []*Arg{
			StepArgument(""), // ignored
			StepArgument(row[1]),
			StepArgument(row[0]),
		}
		if err := s.thereWereNumEventsFired(args...); err != nil {
			return err
		}
	}
	return nil
}
