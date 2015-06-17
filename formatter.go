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

// Formatter is an interface for feature runner
// output summary presentation
type Formatter interface {
	Node(interface{})
	Failed(*gherkin.Step, *stepMatchHandler, error)
	Passed(*gherkin.Step, *stepMatchHandler)
	Skipped(*gherkin.Step)
	Undefined(*gherkin.Step)
	Summary()
}

// general pretty formatter structure
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

// failed represents a failed step data structure
// with all necessary references
type failed struct {
	step *gherkin.Step
	err  error
}

func (f failed) line() string {
	var tok *gherkin.Token
	var ft *gherkin.Feature
	if f.step.Scenario != nil {
		tok = f.step.Scenario.Token
		ft = f.step.Scenario.Feature
	} else {
		tok = f.step.Background.Token
		ft = f.step.Background.Feature
	}
	return fmt.Sprintf("%s:%d", ft.Path, tok.Line)
}

// passed represents a successful step data structure
// with all necessary references
type passed struct {
	step *gherkin.Step
}

// skipped represents a skipped step data structure
// with all necessary references
type skipped struct {
	step *gherkin.Step
}

// undefined represents a pending step data structure
// with all necessary references
type undefined struct {
	step *gherkin.Step
}

// a line number representation in feature file
func (f *pretty) line(tok *gherkin.Token) string {
	return cl(fmt.Sprintf("# %s:%d", f.feature.Path, tok.Line), black)
}

// checks whether it should not print a background step once again
func (f *pretty) canPrintStep(step *gherkin.Step) bool {
	if f.background != nil {
		return step.Background != nil
	}
	return true
}

// Node takes a gherkin node for formatting
func (f *pretty) Node(node interface{}) {
	switch t := node.(type) {
	case *gherkin.Feature:
		f.feature = t
		f.doneBackground = false
		f.background = nil
		f.features = append(f.features, t)
		fmt.Println(bcl("Feature: ", white) + t.Title)
		fmt.Println(t.Description)
	case *gherkin.Background:
		f.background = t
		fmt.Println("\n" + bcl("Background:", white))
	case *gherkin.Scenario:
		f.commentPos = len(t.Token.Text)
		for _, step := range t.Steps {
			if len(step.Token.Text) > f.commentPos {
				f.commentPos = len(step.Token.Text)
			}
		}
		text := strings.Repeat(" ", t.Token.Indent) + bcl("Scenario: ", white) + t.Title
		text += strings.Repeat(" ", f.commentPos-len(t.Token.Text)+1) + f.line(t.Token)
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

// prints a single matched step
func (f *pretty) printMatchedStep(step *gherkin.Step, match *stepMatchHandler, c color) {
	if !f.canPrintStep(step) {
		return
	}
	var text string
	if m := (match.expr.FindStringSubmatchIndex(step.Text))[2:]; len(m) > 0 {
		var pos, i int
		for pos, i = 0, 0; i < len(m); i++ {
			if math.Mod(float64(i), 2) == 0 {
				text += cl(step.Text[pos:m[i]], c)
			} else {
				text += bcl(step.Text[pos:m[i]], c)
			}
			pos = m[i]
		}
		text += cl(step.Text[pos:len(step.Text)], c)
	} else {
		text = cl(step.Text, c)
	}

	// use reflect to get step handler function name
	name := runtime.FuncForPC(reflect.ValueOf(match.handler).Pointer()).Name()

	switch step.Token.Type {
	case gherkin.GIVEN:
		text = cl("Given", c) + " " + text
	case gherkin.WHEN:
		text = cl("When", c) + " " + text
	case gherkin.THEN:
		text = cl("Then", c) + " " + text
	case gherkin.AND:
		text = cl("And", c) + " " + text
	case gherkin.BUT:
		text = cl("But", c) + " " + text
	}
	text = strings.Repeat(" ", step.Token.Indent) + text
	text += strings.Repeat(" ", f.commentPos-len(step.Token.Text)+1) + cl(fmt.Sprintf("# %s", name), black)
	fmt.Println(text)
}

// Passed is called to represent a passed step
func (f *pretty) Passed(step *gherkin.Step, match *stepMatchHandler) {
	f.printMatchedStep(step, match, green)
	f.passed = append(f.passed, &passed{step})
}

// Skipped is called to represent a passed step
func (f *pretty) Skipped(step *gherkin.Step) {
	if f.canPrintStep(step) {
		fmt.Println(cl(step.Token.Text, cyan))
	}
	f.skipped = append(f.skipped, &skipped{step})
}

// Undefined is called to represent a pending step
func (f *pretty) Undefined(step *gherkin.Step) {
	if f.canPrintStep(step) {
		fmt.Println(cl(step.Token.Text, yellow))
	}
	f.undefined = append(f.undefined, &undefined{step})
}

// Failed is called to represent a failed step
func (f *pretty) Failed(step *gherkin.Step, match *stepMatchHandler, err error) {
	f.printMatchedStep(step, match, red)
	fmt.Println(strings.Repeat(" ", step.Token.Indent) + bcl(err, red))
	f.failed = append(f.failed, &failed{step, err})
}
