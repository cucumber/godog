package godog

import (
	"fmt"
	"strings"

	"github.com/DATA-DOG/godog/gherkin"
)

func init() {
	RegisterFormatter("pretty", &pretty{})
}

// Formatter is an interface for feature runner output
type Formatter interface {
	Node(interface{})
	Failed(*gherkin.Step, error)
	Passed(*gherkin.Step)
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
		fmt.Println("\n"+strings.Repeat(" ", t.Token.Indent)+bcl("Scenario: ", white)+t.Title, f.line(t.Token))
	}
}

func (f *pretty) Passed(step *gherkin.Step) {
	if f.canPrintStep(step) {
		fmt.Println(cl(step.Token.Text, green))
	}
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

func (f *pretty) Failed(step *gherkin.Step, err error) {
	if f.canPrintStep(step) {
		fmt.Println(cl(step.Token.Text, red))
	}
}
