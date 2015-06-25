package godog

import "github.com/cucumber/gherkin-go"

type testFormatter struct {
	owner     interface{}
	features  []*feature
	scenarios []interface{}

	failed    []*failed
	passed    []*passed
	skipped   []*skipped
	undefined []*undefined
}

func (f *testFormatter) Feature(ft *gherkin.Feature, p string) {
	f.features = append(f.features, &feature{Path: p, Feature: ft})
}

func (f *testFormatter) Node(node interface{}) {
	switch t := node.(type) {
	case *gherkin.Scenario:
		f.scenarios = append(f.scenarios, t)
		f.owner = t
	case *gherkin.ScenarioOutline:
		f.scenarios = append(f.scenarios, t)
		f.owner = t
	case *gherkin.Background:
		f.owner = t
	}
}

func (f *testFormatter) Summary() {}

func (f *testFormatter) Passed(step *gherkin.Step, match *StepDef) {
	f.passed = append(f.passed, &passed{owner: f.owner, feature: f.features[len(f.features)-1], step: step, def: match})
}

func (f *testFormatter) Skipped(step *gherkin.Step) {
	f.skipped = append(f.skipped, &skipped{owner: f.owner, feature: f.features[len(f.features)-1], step: step})
}

func (f *testFormatter) Undefined(step *gherkin.Step) {
	f.undefined = append(f.undefined, &undefined{owner: f.owner, feature: f.features[len(f.features)-1], step: step})
}

func (f *testFormatter) Failed(step *gherkin.Step, match *StepDef, err error) {
	f.failed = append(f.failed, &failed{owner: f.owner, feature: f.features[len(f.features)-1], step: step, def: match, err: err})
}
