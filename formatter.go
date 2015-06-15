package godog

import (
	"fmt"
	"math"
	"reflect"
	"runtime"
	"strings"

	"github.com/DATA-DOG/godog/gherkin"
)

func init() {
	RegisterFormatter("pretty", &pretty{})
}

// Formatter is an interface for feature runner output
type Formatter interface {
	Node(interface{})
	Failed(*gherkin.Step, *stepMatchHandler, error)
	Passed(*gherkin.Step, *stepMatchHandler)
	Skipped(*gherkin.Step)
	Pending(*gherkin.Step)
	Summary()
}

type pretty struct {
	feature        *gherkin.Feature
	scenario       *gherkin.Scenario
	commentPos     int
	doneBackground bool
	background     *gherkin.Background

	// summary
	features []*gherkin.Feature
	failures []*failed
	passes   []*passed
	skips    []*skipped
	pendings []*pending
}

type failed struct {
	feature  *gherkin.Feature
	scenario *gherkin.Scenario
	step     *gherkin.Step
	err      error
}

type passed struct {
	feature  *gherkin.Feature
	scenario *gherkin.Scenario
	step     *gherkin.Step
}

type skipped struct {
	feature  *gherkin.Feature
	scenario *gherkin.Scenario
	step     *gherkin.Step
}

type pending struct {
	feature  *gherkin.Feature
	scenario *gherkin.Scenario
	step     *gherkin.Step
}

func (f *pretty) line(tok *gherkin.Token) string {
	return cl(fmt.Sprintf("# %s:%d", f.feature.Path, tok.Line), black)
}

func (f *pretty) canPrintStep(step *gherkin.Step) bool {
	if f.background == nil {
		return true
	}

	var backgroundStep bool
	for _, s := range f.background.Steps {
		if s == step {
			backgroundStep = true
			break
		}
	}

	if !backgroundStep {
		f.doneBackground = true
		return true
	}

	return !f.doneBackground
}

func (f *pretty) comment(text, comment string) string {
	indent := f.commentPos - len(text) + 1
	return text + strings.Repeat(" ", indent) + comment
}

func (f *pretty) Node(node interface{}) {
	switch t := node.(type) {
	case *gherkin.Feature:
		f.feature = t
		f.doneBackground = false
		f.scenario = nil
		f.background = nil
		f.features = append(f.features, t)
		fmt.Println("\n" + bcl("Feature: ", white) + t.Title)
		fmt.Println(t.Description)
	case *gherkin.Background:
		f.background = t
		fmt.Println("\n" + bcl("Background:", white))
	case *gherkin.Scenario:
		f.scenario = t
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

func (f *pretty) Summary() {

}

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

func (f *pretty) Passed(step *gherkin.Step, match *stepMatchHandler) {
	f.printMatchedStep(step, match, green)
	f.passes = append(f.passes, &passed{
		feature:  f.feature,
		scenario: f.scenario,
		step:     step,
	})
}

func (f *pretty) Skipped(step *gherkin.Step) {
	if f.canPrintStep(step) {
		fmt.Println(cl(step.Token.Text, cyan))
	}
	f.skips = append(f.skips, &skipped{
		feature:  f.feature,
		scenario: f.scenario,
		step:     step,
	})
}

func (f *pretty) Pending(step *gherkin.Step) {
	if f.canPrintStep(step) {
		fmt.Println(cl(step.Token.Text, yellow))
	}
	f.pendings = append(f.pendings, &pending{
		feature:  f.feature,
		scenario: f.scenario,
		step:     step,
	})
}

func (f *pretty) Failed(step *gherkin.Step, match *stepMatchHandler, err error) {
	f.printMatchedStep(step, match, red)
	fmt.Println(strings.Repeat(" ", step.Token.Indent) + bcl(err, red))
	f.failures = append(f.failures, &failed{
		feature:  f.feature,
		scenario: f.scenario,
		step:     step,
		err:      err,
	})
}
