package godog

import "github.com/DATA-DOG/godog/gherkin"

type testFormatter struct {
	basefmt
	scenarios []interface{}
}

func (f *testFormatter) Node(node interface{}) {
	f.basefmt.Node(node)
	switch t := node.(type) {
	case *gherkin.Scenario:
		f.scenarios = append(f.scenarios, t)
	case *gherkin.ScenarioOutline:
		f.scenarios = append(f.scenarios, t)
	}
}

func (f *testFormatter) Summary() {}
