package godog

import (
	"fmt"
	"math"
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
}

type pretty struct {
	feature        *gherkin.Feature
	scenario       *gherkin.Scenario
	doneBackground bool
	background     *gherkin.Background
}

func (f *pretty) line(tok *gherkin.Token) string {
	return cl(fmt.Sprintf("#%s:%d", f.feature.Path, tok.Line), magenta)
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

func (f *pretty) Node(node interface{}) {
	switch t := node.(type) {
	case *gherkin.Feature:
		f.feature = t
		f.doneBackground = false
		f.scenario = nil
		f.background = nil
		fmt.Println("\n"+bcl("Feature: ", white)+t.Title, f.line(t.Token))
		fmt.Println(t.Description)
	case *gherkin.Background:
		f.background = t
		fmt.Println("\n" + bcl("Background:", white))
	case *gherkin.Scenario:
		f.scenario = t
		fmt.Println("\n"+strings.Repeat(" ", t.Token.Indent)+bcl("Scenario: ", white)+t.Title, f.line(t.Token))
	}
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
	fmt.Println(strings.Repeat(" ", step.Token.Indent) + text)
}

func (f *pretty) Passed(step *gherkin.Step, match *stepMatchHandler) {
	f.printMatchedStep(step, match, green)
}

func (f *pretty) Skipped(step *gherkin.Step) {
	if f.canPrintStep(step) {
		fmt.Println(cl(step.Token.Text, cyan))
	}
}

func (f *pretty) Pending(step *gherkin.Step) {
	if f.canPrintStep(step) {
		fmt.Println(cl(step.Token.Text, yellow))
	}
}

func (f *pretty) Failed(step *gherkin.Step, match *stepMatchHandler, err error) {
	f.printMatchedStep(step, match, red)
}
