/* file: $GOPATH/src/assert-godogs/godogs_test.go */
package main

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	messages "github.com/cucumber/messages-go/v10"
	"github.com/stretchr/testify/assert"
)

var opt = godog.Options{Output: colors.Colored(os.Stdout)}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opt)
}

func TestMain(m *testing.M) {
	flag.Parse()
	opt.Paths = flag.Args()

	status := godog.RunWithOptions("godogs", func(s *godog.Suite) {
		FeatureContext(s)
	}, opt)

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func thereAreGodogs(available int) error {
	Godogs = available
	return nil
}

func iEat(num int) error {
	err := assertExpectedAndActual(
		assert.GreaterOrEqual, Godogs, num,
		"You cannot eat %d godogs, there are %d available", num, Godogs,
	)
	if err != nil {
		return err
	}

	Godogs -= num
	return nil
}

func thereShouldBeRemaining(remaining int) error {
	return assertExpectedAndActual(
		assert.Equal, Godogs, remaining,
		"Expected %d godogs to be remaining, but there is %d", remaining, Godogs,
	)
}

func thereShouldBeNoneRemaining() error {
	return assertActual(
		assert.Empty, Godogs,
		"Expected none godogs to be remaining, but there is %d", Godogs,
	)
}

func FeatureContext(s *godog.Suite) {
	s.Step(`^there are (\d+) godogs$`, thereAreGodogs)
	s.Step(`^I eat (\d+)$`, iEat)
	s.Step(`^there should be (\d+) remaining$`, thereShouldBeRemaining)
	s.Step(`^there should be none remaining$`, thereShouldBeNoneRemaining)

	s.BeforeScenario(func(*messages.Pickle) {
		Godogs = 0 // clean the state before every scenario
	})
}

// assertExpectedAndActual is a helper function to allow the step function to call
// assertion functions where you want to compare an expected and an actual value.
func assertExpectedAndActual(a expectedAndActualAssertion, expected, actual interface{}, msgAndArgs ...interface{}) error {
	var t asserter
	a(&t, expected, actual, msgAndArgs...)
	return t.err
}

type expectedAndActualAssertion func(t assert.TestingT, expected, actual interface{}, msgAndArgs ...interface{}) bool

// assertActual is a helper function to allow the step function to call
// assertion functions where you want to compare an actual value to a
// predined state like nil, empty or true/false.
func assertActual(a actualAssertion, actual interface{}, msgAndArgs ...interface{}) error {
	var t asserter
	a(&t, actual, msgAndArgs...)
	return t.err
}

type actualAssertion func(t assert.TestingT, actual interface{}, msgAndArgs ...interface{}) bool

// asserter is used to be able to retrieve the error reported by the called assertion
type asserter struct {
	err error
}

// Errorf is used by the called assertion to report an error
func (a *asserter) Errorf(format string, args ...interface{}) {
	a.err = fmt.Errorf(format, args...)
}
