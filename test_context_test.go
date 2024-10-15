package godog

import (
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScenarioContext_Step(t *testing.T) {
	ctx := ScenarioContext{suite: &suite{}}
	re := `(?:it is a test)?.{10}x*`

	type tc struct {
		f func()
		n string
		p interface{}
	}

	for _, c := range []tc{
		{n: "ScenarioContext should accept steps defined with regexp.Regexp",
			f: func() { ctx.Step(regexp.MustCompile(re), okVoidResult) }},
		{n: "ScenarioContext should accept steps defined with bytes slice",
			f: func() { ctx.Step([]byte(re), okVoidResult) }},

		{n: "ScenarioContext should accept steps handler with no return",
			f: func() { ctx.Step(".*", okVoidResult) }},
		{n: "ScenarioContext should accept steps handler with error return",
			f: func() { ctx.Step(".*", okErrorResult) }},
		{n: "ScenarioContext should accept steps handler with godog.Steps return",
			f: func() { ctx.Step(".*", okStepsResult) }},
		{n: "ScenarioContext should accept steps handler with (Context, error) return",
			f: func() { ctx.Step(".*", okContextErrorResult) }},
	} {
		t.Run(c.n, func(t *testing.T) {
			assert.NotPanics(t, c.f)
		})
	}

	for _, c := range []tc{
		{n: "ScenarioContext should panic if step expression is neither a string, regex or byte slice",
			p: "expecting expr to be a *regexp.Regexp or a string or []byte, got type: int",
			f: func() { ctx.Step(1251, okVoidResult) }},
		{n: "ScenarioContext should panic if step handler is not a function",
			p: "expected handler to be func, but got: int",
			f: func() { ctx.Step(".*", 124) }},
		{n: "ScenarioContext should panic if step handler has more than 2 return values",
			p: "expected handler to return either zero, one or two values, but it has: 3",
			f: func() { ctx.Step(".*", nokLimitCase3) }},
		{n: "ScenarioContext should panic if step handler has more than 2 return values (5)",
			p: "expected handler to return either zero, one or two values, but it has: 5",
			f: func() { ctx.Step(".*", nokLimitCase5) }},

		{n: "ScenarioContext should panic if step expression is neither a string, regex or byte slice",
			p: "expecting expr to be a *regexp.Regexp or a string or []byte, got type: int",
			f: func() { ctx.Step(1251, okVoidResult) }},

		{n: "ScenarioContext should panic if step return type is []string",
			p: "expected handler to return one of error or context.Context or godog.Steps or (context.Context, error), but got: []string",
			f: func() { ctx.Step(".*", nokSliceStringResult) }},
		{n: "ScenarioContext should panic if step handler return type is not an error or string slice or void (interface)",
			p: "expected handler to return one of error or context.Context or godog.Steps or (context.Context, error), but got: interface {}",
			f: func() { ctx.Step(".*", nokInvalidReturnInterfaceType) }},
		{n: "ScenarioContext should panic if step handler return type is not an error or string slice or void (slice)",
			p: "expected handler to return one of error or context.Context or godog.Steps or (context.Context, error), but got: []int",
			f: func() { ctx.Step(".*", nokInvalidReturnSliceType) }},
		{n: "ScenarioContext should panic if step handler return type is not an error or string slice or void (other)",
			p: "expected handler to return one of error or context.Context or godog.Steps or (context.Context, error), but got: chan int",
			f: func() { ctx.Step(".*", nokInvalidReturnOtherType) }},
	} {
		t.Run(c.n, func(t *testing.T) {
			assert.PanicsWithValue(t, c.p, c.f)
		})
	}
}

func okVoidResult()                                  {}
func okErrorResult() error                           { return nil }
func okStepsResult() Steps                           { return nil }
func okContextErrorResult() (context.Context, error) { return nil, nil }
func nokSliceStringResult() []string                 { return nil }
func nokLimitCase3() (string, int, error)            { return "", 0, nil }
func nokLimitCase5() (int, int, int, int, error)     { return 0, 0, 0, 0, nil }
func nokInvalidReturnInterfaceType() interface{}     { return 0 }
func nokInvalidReturnSliceType() []int               { return nil }
func nokInvalidReturnOtherType() chan int            { return nil }
