package demo

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cucumber/godog"
)

//
//	The tests "pass" by demonstrating the "problem statement" discussed in this `dupsteps`
// 	example; i.e., they are expected to "fail" when the problem is fixed, or can be fixed
//	through a supported `godog` configuration option / enhancement.
//
//  What's being demonstrated is how godog's use of a global list of steps defined across
//  all configured scenarios allows for indeterminate results, based upon the order in
//	which the steps are loaded.  The first "matching" step (text) will be used when
//	evaluating a step for a given scenario, regardless of which scenario it was defined
//	or intended to participate in.
//

// TestFlatTireFirst demonstrates that loading the 'flatTire' step implementations first
// causes the 'cloggedDrain' test to fail
func TestFlatTireFirst(t *testing.T) {
	demonstrateProblemCase(t,
		func(ctx *godog.ScenarioContext) {
			(&flatTire{}).addFlatTireSteps(ctx)
			(&cloggedDrain{}).addCloggedDrainSteps(ctx)
		},
		"drain is clogged",
	)
}

// TestCloggedDrainFirst demonstrates that loading the 'cloggedDrain' step implementations first
// causes the 'flatTire' test to fail
func TestCloggedDrainFirst(t *testing.T) {
	demonstrateProblemCase(t,
		func(ctx *godog.ScenarioContext) {
			(&cloggedDrain{}).addCloggedDrainSteps(ctx)
			(&flatTire{}).addFlatTireSteps(ctx)
		},
		"tire was not fixed",
	)
}

// demonstrateProblemCase sets up the test suite using 'caseInitializer' and demonstrates the expected error result.
func demonstrateProblemCase(t *testing.T, caseInitializer func(ctx *godog.ScenarioContext), expectedError string) {

	var sawExpectedError bool

	opts := defaultOpts
	opts.Format = "pretty"
	opts.Output = &prettyOutputListener{
		wrapped: os.Stdout,
		callback: func(s string) {
			if strings.Contains(s, expectedError) {
				sawExpectedError = true
				fmt.Println("====>")
			}
		},
	}

	suite := godog.TestSuite{
		Name:                t.Name(),
		ScenarioInitializer: caseInitializer,
		Options:             &opts,
	}

	rc := suite.Run()

	// (demonstration of) expected error
	assert.NotZero(t, rc)

	// demonstrate that the expected error message was seen in the godog output
	assert.True(t, sawExpectedError)
}

// Implementation of the steps for the "Clogged Drain" scenario

type cloggedDrain struct {
	drainIsClogged bool
}

func (cd *cloggedDrain) addCloggedDrainSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I accidentally poured concrete down my drain and clogged the sewer line$`, cd.clogSewerLine)
	ctx.Step(`^I fixed it$`, cd.iFixedIt)
	ctx.Step(`^I can once again use my sink$`, cd.useTheSink)
}

func (cd *cloggedDrain) clogSewerLine() error {
	cd.drainIsClogged = true

	return nil
}

func (cd *cloggedDrain) iFixedIt() error {
	cd.drainIsClogged = false

	return nil
}

func (cd *cloggedDrain) useTheSink() error {
	if cd.drainIsClogged {
		return fmt.Errorf("drain is clogged")
	}

	return nil
}

// Implementation of the steps for the "Flat Tire" scenario

type flatTire struct {
	tireIsFlat bool
}

func (ft *flatTire) addFlatTireSteps(ctx *godog.ScenarioContext) {
	ctx.Step(`^I ran over a nail and got a flat tire$`, ft.gotFlatTire)
	ctx.Step(`^I fixed it$`, ft.iFixedIt)
	ctx.Step(`^I can continue on my way$`, ft.continueOnMyWay)
}

func (ft *flatTire) gotFlatTire() error {
	ft.tireIsFlat = true

	return nil
}

func (ft *flatTire) iFixedIt() error {
	ft.tireIsFlat = false

	return nil
}

func (ft *flatTire) continueOnMyWay() error {
	if ft.tireIsFlat {
		return fmt.Errorf("tire was not fixed")
	}

	return nil
}

// standard godog global environment initialization sequence...

var defaultOpts = godog.Options{
	Strict: true,
	Paths:  []string{"../features"},
}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &defaultOpts)
}

// a godog "pretty" output listener used to detect the expected godog error

type prettyOutputListener struct {
	wrapped  io.Writer
	callback func(string)
	buf      []byte
}

func (lw *prettyOutputListener) Write(p []byte) (n int, err error) {
	lw.buf = append(lw.buf, p...)

	for {
		idx := bytes.IndexByte(lw.buf, '\n')
		if idx == -1 {
			break
		}

		line := string(lw.buf[:idx])

		lw.callback(line)

		if _, err := lw.wrapped.Write(lw.buf[:idx+1]); err != nil {
			return len(p), err
		}

		lw.buf = lw.buf[idx+1:]
	}

	return len(p), nil
}
