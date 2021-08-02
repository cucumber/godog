package godog

import (
	"fmt"
	"regexp"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestScenarioContext_Step(t *testing.T) {
	Convey("When adding steps to ScenarioContext ", t, func() {
		ctx := ScenarioContext{suite: &suite{}}

		Convey("It should accept steps defined with regexp.Regexp", func() {
			re := regexp.MustCompile(`(?:it is a test)?.{10}x*`)
			So(func() { ctx.Step(re, okEmptyResult) }, ShouldNotPanic)
		})

		Convey("It should accept steps defined with bytes slice", func() {
			So(func() { ctx.Step([]byte("(?:it is a test)?.{10}x*"), okEmptyResult) }, ShouldNotPanic)
		})

		Convey("It should accept steps handler with empty return", func() {
			So(func() { ctx.Step(".*", okEmptyResult) }, ShouldNotPanic)
		})

		Convey("It should accept steps handler with error return", func() {
			So(func() { ctx.Step(".*", okErrorResult) }, ShouldNotPanic)
		})

		Convey("It should accept steps handler with string slice return", func() {
			So(func() { ctx.Step(".*", okSliceResult) }, ShouldNotPanic)
		})

		Convey("It should panic if step expression is neither a string, regex or byte slice", func() {
			So(func() { ctx.Step(1251, okSliceResult) }, ShouldPanicWith, fmt.Sprintf("expecting expr to be a *regexp.Regexp or a string, got type: %T", 12))
		})
		Convey("It should panic if step handler", func() {
			Convey("is not a function", func() {
				So(func() { ctx.Step(".*", 124) }, ShouldPanicWith, fmt.Sprintf("expected handler to be func, but got: %T", 12))
			})

			Convey("has more than 1 return value", func() {
				So(func() { ctx.Step(".*", nokLimitCase) }, ShouldPanicWith, fmt.Sprintf("expected handler to return either zero or one value, but it has: 2"))
				So(func() { ctx.Step(".*", nokMore) }, ShouldPanicWith, fmt.Sprintf("expected handler to return either zero or one value, but it has: 5"))
			})

			Convey("return type is not an error or string slice or void", func() {
				So(func() { ctx.Step(".*", nokInvalidReturnInterfaceType) }, ShouldPanicWith, "expected handler to return an error, but got: interface")
				So(func() { ctx.Step(".*", nokInvalidReturnSliceType) }, ShouldPanicWith, "expected handler to return []string for multistep, but got: []int")
				So(func() { ctx.Step(".*", nokInvalidReturnOtherType) }, ShouldPanicWith, "expected handler to return an error or []string, but got: chan")
			})
		})

	})

}

func okEmptyResult()                             {}
func okErrorResult() error                       { return nil }
func okSliceResult() []string                    { return nil }
func nokLimitCase() (int, error)                 { return 0, nil }
func nokMore() (int, int, int, int, error)       { return 0, 0, 0, 0, nil }
func nokInvalidReturnInterfaceType() interface{} { return 0 }
func nokInvalidReturnSliceType() []int           { return nil }
func nokInvalidReturnOtherType() chan int        { return nil }
