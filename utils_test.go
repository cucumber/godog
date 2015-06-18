package godog

import "github.com/DATA-DOG/godog/gherkin"

type testFormatter struct {
	features  []*gherkin.Feature
	scenarios []*gherkin.Scenario

	failed    []*failed
	passed    []*passed
	skipped   []*skipped
	undefined []*undefined
}

func (f *testFormatter) Node(node interface{}) {
	switch t := node.(type) {
	case *gherkin.Feature:
		f.features = append(f.features, t)
	case *gherkin.Scenario:
		f.scenarios = append(f.scenarios, t)
	}
}

func (f *testFormatter) Summary() {}

func (f *testFormatter) Passed(step *gherkin.Step, match *stepMatchHandler) {
	f.passed = append(f.passed, &passed{step: step, handler: match})
}

func (f *testFormatter) Skipped(step *gherkin.Step) {
	f.skipped = append(f.skipped, &skipped{step: step})
}

func (f *testFormatter) Undefined(step *gherkin.Step) {
	f.undefined = append(f.undefined, &undefined{step: step})
}

func (f *testFormatter) Failed(step *gherkin.Step, match *stepMatchHandler, err error) {
	f.failed = append(f.failed, &failed{step: step, handler: match, err: err})
}
