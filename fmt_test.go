package godog

import (
	"io"
	"testing"

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

func TestShouldFindFormatter(t *testing.T) {
	cases := map[string]bool{
		"progress": true, // true means should be available
		"unknown":  false,
		"junit":    true,
		"cucumber": true,
		"pretty":   true,
		"custom":   true, // is available for test purposes only
		"undef":    false,
	}

	for name, shouldFind := range cases {
		actual := findFmt(name)
		if actual == nil && shouldFind {
			t.Fatalf("expected %s formatter should be available", name)
		}
		if actual != nil && !shouldFind {
			t.Fatalf("expected %s formatter should not be available", name)
		}
	}
}
