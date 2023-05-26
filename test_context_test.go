package godog

import (
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestScenarioContext_Step(t *testing.T) {
	ctx := ScenarioContext{suite: &suite{}}
	re := regexp.MustCompile(`(?:it is a test)?.{10}x*`)

	type tc struct {
		f func()
		n string
		p interface{}
	}

	for _, c := range []tc{
		{n: "ScenarioContext should accept steps defined with regexp.Regexp",
			f: func() { ctx.Step(re, okEmptyResult) }},
		{n: "ScenarioContext should accept steps defined with bytes slice",
			f: func() { ctx.Step([]byte("(?:it is a test)?.{10}x*"), okEmptyResult) }},
		{n: "ScenarioContext should accept steps handler with error return",
			f: func() { ctx.Step(".*", okEmptyResult) }},
		{n: "ScenarioContext should accept steps handler with error return",
			f: func() { ctx.Step(".*", okErrorResult) }},
		{n: "ScenarioContext should accept steps handler with string slice return",
			f: func() { ctx.Step(".*", okSliceResult) }},
	} {
		t.Run(c.n, func(t *testing.T) {
			assert.NotPanics(t, c.f)
		})
	}

	for _, c := range []tc{
		{n: "ScenarioContext should panic if step expression is neither a string, regex or byte slice",
			p: "expecting expr to be a *regexp.Regexp or a string, got type: int",
			f: func() { ctx.Step(1251, okSliceResult) }},
		{n: "ScenarioContext should panic if step handler is not a function",
			p: "expected handler to be func, but got: int",
			f: func() { ctx.Step(".*", 124) }},
		{n: "ScenarioContext should panic if step handler has more than 2 return values",
			p: "expected handler to return either zero, one or two values, but it has: 3",
			f: func() { ctx.Step(".*", nokLimitCase) }},
		{n: "ScenarioContext should panic if step handler has more than 2 return values (5)",
			p: "expected handler to return either zero, one or two values, but it has: 5",
			f: func() { ctx.Step(".*", nokMore) }},

		{n: "ScenarioContext should panic if step expression is neither a string, regex or byte slice",
			p: "expecting expr to be a *regexp.Regexp or a string, got type: int",
			f: func() { ctx.Step(1251, okSliceResult) }},

		{n: "ScenarioContext should panic if step handler return type is not an error or string slice or void (interface)",
			p: "expected handler to return an error or context.Context, but got: interface",
			f: func() { ctx.Step(".*", nokInvalidReturnInterfaceType) }},
		{n: "ScenarioContext should panic if step handler return type is not an error or string slice or void (slice)",
			p: "expected handler to return []string for multistep, but got: []int",
			f: func() { ctx.Step(".*", nokInvalidReturnSliceType) }},
		{n: "ScenarioContext should panic if step handler return type is not an error or string slice or void (other)",
			p: "expected handler to return an error or []string, but got: chan",
			f: func() { ctx.Step(".*", nokInvalidReturnOtherType) }},
	} {
		t.Run(c.n, func(t *testing.T) {
			assert.PanicsWithValue(t, c.p, c.f)
		})
	}
}

func okEmptyResult()                             {}
func okErrorResult() error                       { return nil }
func okSliceResult() []string                    { return nil }
func nokLimitCase() (string, int, error)         { return "", 0, nil }
func nokMore() (int, int, int, int, error)       { return 0, 0, 0, 0, nil }
func nokInvalidReturnInterfaceType() interface{} { return 0 }
func nokInvalidReturnSliceType() []int           { return nil }
func nokInvalidReturnOtherType() chan int        { return nil }
