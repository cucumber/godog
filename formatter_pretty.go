package godog

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/DATA-DOG/godog/gherkin"
)

func init() {
	RegisterFormatter("pretty", "Prints every feature with runtime statuses.", &pretty{
		started: time.Now(),
	})
}

// a built in default pretty formatter
type pretty struct {
	feature        *gherkin.Feature
	commentPos     int
	doneBackground bool
	background     *gherkin.Background

	// summary
	started   time.Time
	features  []*gherkin.Feature
	failed    []*failed
	passed    []*passed
	skipped   []*skipped
	undefined []*undefined
}

// a line number representation in feature file
func (f *pretty) line(tok *gherkin.Token) string {
	return cl(fmt.Sprintf("# %s:%d", f.feature.Path, tok.Line), black)
}

// checks whether it should not print a background step once again
func (f *pretty) canPrintStep(step *gherkin.Step) bool {
	if f.background == nil {
		return true
	}

	if step.Background == nil {
		f.doneBackground = true
		return true
	}

	return !f.doneBackground
}

// Node takes a gherkin node for formatting
func (f *pretty) Node(node interface{}) {
	switch t := node.(type) {
	case *gherkin.Feature:
		if f.feature != nil {
			// not a first feature, add a newline
			fmt.Println("")
		}
		f.feature = t
		f.doneBackground = false
		f.background = nil
		f.features = append(f.features, t)
		fmt.Println(bcl("Feature: ", white) + t.Title)
		fmt.Println(t.Description)
	case *gherkin.Background:
		// determine comment position based on step length
		f.commentPos = len(t.Token.Text)
		for _, step := range t.Steps {
			if len(step.Token.Text) > f.commentPos {
				f.commentPos = len(step.Token.Text)
			}
		}
		// do not repeat background
		if !f.doneBackground {
			f.background = t
			fmt.Println("\n" + s(t.Token.Indent) + bcl("Background:", white))
		}
	case *gherkin.Scenario:
		// determine comment position based on step length
		f.commentPos = len(t.Token.Text)
		for _, step := range t.Steps {
			if len(step.Token.Text) > f.commentPos {
				f.commentPos = len(step.Token.Text)
			}
		}
		text := s(t.Token.Indent) + bcl("Scenario: ", white) + t.Title
		text += s(f.commentPos-len(t.Token.Text)+1) + f.line(t.Token)
		fmt.Println("\n" + text)
	}
}

// Summary sumarize the feature formatter output
func (f *pretty) Summary() {
	// failed steps on background are not scenarios
	var failedScenarios []*failed
	for _, fail := range f.failed {
		if fail.step.Scenario != nil {
			failedScenarios = append(failedScenarios, fail)
		}
	}
	if len(failedScenarios) > 0 {
		fmt.Println("\n--- " + cl("Failed scenarios:", red) + "\n")
		for _, fail := range failedScenarios {
			fmt.Println("    " + cl(fail.line(), red))
		}
	}
	var total, passed int
	for _, ft := range f.features {
		total += len(ft.Scenarios)
	}
	passed = total

	var steps, parts, scenarios []string
	nsteps := len(f.passed) + len(f.failed) + len(f.skipped) + len(f.undefined)
	if len(f.passed) > 0 {
		steps = append(steps, cl(fmt.Sprintf("%d passed", len(f.passed)), green))
	}
	if len(f.failed) > 0 {
		passed -= len(f.failed)
		parts = append(parts, cl(fmt.Sprintf("%d failed", len(f.failed)), red))
		steps = append(steps, parts[len(parts)-1])
	}
	if len(f.skipped) > 0 {
		steps = append(steps, cl(fmt.Sprintf("%d skipped", len(f.skipped)), cyan))
	}
	if len(f.undefined) > 0 {
		passed -= len(f.undefined)
		parts = append(parts, cl(fmt.Sprintf("%d undefined", len(f.undefined)), yellow))
		steps = append(steps, parts[len(parts)-1])
	}
	if passed > 0 {
		scenarios = append(scenarios, cl(fmt.Sprintf("%d passed", passed), green))
	}
	scenarios = append(scenarios, parts...)
	elapsed := time.Since(f.started)

	fmt.Println("")
	if total == 0 {
		fmt.Println("No scenarios")
	} else {
		fmt.Println(fmt.Sprintf("%d scenarios (%s)", total, strings.Join(scenarios, ", ")))
	}

	if nsteps == 0 {
		fmt.Println("No steps")
	} else {
		fmt.Println(fmt.Sprintf("%d steps (%s)", nsteps, strings.Join(steps, ", ")))
	}
	fmt.Println(elapsed)
}

func (f *pretty) printStep(stepAction interface{}) {
	var c color
	var step *gherkin.Step
	var h *stepMatchHandler
	var err error
	var suffix, prefix string

	switch typ := stepAction.(type) {
	case *passed:
		step = typ.step
		h = typ.handler
		c = green
	case *failed:
		step = typ.step
		h = typ.handler
		err = typ.err
		c = red
	case *skipped:
		step = typ.step
		c = cyan
	case *undefined:
		step = typ.step
		c = yellow
	default:
		fatal(fmt.Errorf("unexpected step type received: %T", typ))
	}

	if !f.canPrintStep(step) {
		return
	}

	if h != nil {
		if m := (h.expr.FindStringSubmatchIndex(step.Text))[2:]; len(m) > 0 {
			var pos, i int
			for pos, i = 0, 0; i < len(m); i++ {
				if math.Mod(float64(i), 2) == 0 {
					suffix += cl(step.Text[pos:m[i]], c)
				} else {
					suffix += bcl(step.Text[pos:m[i]], c)
				}
				pos = m[i]
			}
			suffix += cl(step.Text[pos:len(step.Text)], c)
		} else {
			suffix = cl(step.Text, c)
		}
		// use reflect to get step handler function name
		name := runtime.FuncForPC(reflect.ValueOf(h.handler).Pointer()).Name()
		suffix += s(f.commentPos-len(step.Token.Text)+1) + cl(fmt.Sprintf("# %s", name), black)
	} else {
		suffix = cl(step.Text, c)
	}

	prefix = s(step.Token.Indent)
	switch step.Token.Type {
	case gherkin.GIVEN:
		prefix += cl("Given", c)
	case gherkin.WHEN:
		prefix += cl("When", c)
	case gherkin.THEN:
		prefix += cl("Then", c)
	case gherkin.AND:
		prefix += cl("And", c)
	case gherkin.BUT:
		prefix += cl("But", c)
	}
	fmt.Println(prefix, suffix)
	if step.PyString != nil {
		fmt.Println(s(step.Token.Indent+2) + cl(`"""`, c))
		fmt.Println(cl(step.PyString.Raw, c))
		fmt.Println(s(step.Token.Indent+2) + cl(`"""`, c))
	}
	if err != nil {
		fmt.Println(s(step.Token.Indent) + bcl(err, red))
	}
}

// Passed is called to represent a passed step
func (f *pretty) Passed(step *gherkin.Step, match *stepMatchHandler) {
	s := &passed{step: step, handler: match}
	f.printStep(s)
	f.passed = append(f.passed, s)
}

// Skipped is called to represent a passed step
func (f *pretty) Skipped(step *gherkin.Step) {
	s := &skipped{step: step}
	f.printStep(s)
	f.skipped = append(f.skipped, s)
}

// Undefined is called to represent a pending step
func (f *pretty) Undefined(step *gherkin.Step) {
	s := &undefined{step: step}
	f.printStep(s)
	f.undefined = append(f.undefined, s)
}

// Failed is called to represent a failed step
func (f *pretty) Failed(step *gherkin.Step, match *stepMatchHandler, err error) {
	s := &failed{step: step, handler: match, err: err}
	f.printStep(s)
	f.failed = append(f.failed, s)
}
