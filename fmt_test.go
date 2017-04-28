package godog

import (
	"io"

	"github.com/DATA-DOG/godog/gherkin"
)

type testFormatter struct {
	basefmt
	scenarios []interface{}
}

func testFormatterFunc(suite string, out io.Writer) Formatter {
	return &testFormatter{
		basefmt: basefmt{
			started: timeNowFunc(),
			indent:  2,
			out:     out,
		},
	}
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
