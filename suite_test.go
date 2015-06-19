package godog

import (
	"fmt"
	"strings"

	"github.com/DATA-DOG/godog/gherkin"
)

type firedEvent struct {
	name string
	args []interface{}
}

type suiteFeature struct {
	testedSuite *suite
	events      []*firedEvent
}

func (s *suiteFeature) HandleBeforeScenario(*gherkin.Scenario) {
	// reset whole suite with the state
	s.testedSuite = &suite{fmt: &testFormatter{}}
	// our tested suite will have the same context registered
	SuiteContext(s.testedSuite)
	// reset feature paths
	cfg.paths = []string{}
	// reset all fired events
	s.events = []*firedEvent{}
}

func (s *suiteFeature) iAmListeningToSuiteEvents(args ...*Arg) error {
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

func (s *suiteFeature) aFailingStep(...*Arg) error {
	return fmt.Errorf("intentional failure")
}

// parse a given feature file body as a feature
func (s *suiteFeature) aFeatureFile(args ...*Arg) error {
	name := args[0].String()
	body := args[1].PyString().Raw
	feature, err := gherkin.Parse(strings.NewReader(body), name)
	s.testedSuite.features = append(s.testedSuite.features, feature)
	return err
}

func (s *suiteFeature) featurePath(args ...*Arg) error {
	cfg.paths = append(cfg.paths, args[0].String())
	return nil
}

func (s *suiteFeature) parseFeatures(args ...*Arg) (err error) {
	s.testedSuite.features, err = cfg.features()
	return
}

func (s *suiteFeature) theSuitePassedSuccessfully(...*Arg) error {
	if s.testedSuite.failed {
		return fmt.Errorf("the feature suite has failed")
	}
	return nil
}

func (s *suiteFeature) iShouldHaveNumFeatureFiles(args ...*Arg) error {
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

func (s *suiteFeature) iRunFeatureSuite(args ...*Arg) error {
	if err := s.parseFeatures(); err != nil {
		return err
	}
	s.testedSuite.run()
	return nil
}

func (s *suiteFeature) numScenariosRegistered(args ...*Arg) (err error) {
	var num int
	for _, ft := range s.testedSuite.features {
		num += len(ft.Scenarios)
	}
	if num != args[0].Int() {
		err = fmt.Errorf("expected %d scenarios to be registered, but got %d", args[0].Int(), num)
	}
	return
}

func (s *suiteFeature) thereWereNumEventsFired(args ...*Arg) error {
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

func (s *suiteFeature) thereWasEventTriggeredBeforeScenario(args ...*Arg) error {
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

func (s *suiteFeature) theseEventsHadToBeFiredForNumberOfTimes(args ...*Arg) error {
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

func SuiteContext(g Suite) {
	s := &suiteFeature{}

	g.BeforeScenario(s)

	g.Step(`^a feature path "([^"]*)"$`, s.featurePath)
	g.Step(`^I parse features$`, s.parseFeatures)
	g.Step(`^I'm listening to suite events$`, s.iAmListeningToSuiteEvents)
	g.Step(`^I run feature suite$`, s.iRunFeatureSuite)
	g.Step(`^a feature "([^"]*)" file:$`, s.aFeatureFile)
	g.Step(`^the suite should have passed successfully$`, s.theSuitePassedSuccessfully)

	g.Step(`^I should have ([\d]+) features? files?:$`, s.iShouldHaveNumFeatureFiles)
	g.Step(`^I should have ([\d]+) scenarios? registered$`, s.numScenariosRegistered)
	g.Step(`^there (was|were) ([\d]+) "([^"]*)" events? fired$`, s.thereWereNumEventsFired)
	g.Step(`^there was event triggered before scenario "([^"]*)"$`, s.thereWasEventTriggeredBeforeScenario)
	g.Step(`^these events had to be fired for a number of times:$`, s.theseEventsHadToBeFiredForNumberOfTimes)

	g.Step(`^a failing step`, s.aFailingStep)
	g.Step(`^this step should fail`, s.aFailingStep)
}
