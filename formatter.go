package godog

import (
	"fmt"

	"github.com/DATA-DOG/godog/gherkin"
)

type formatter interface {
	node(interface{})
}

type pretty struct{}

func (f *pretty) node(node interface{}) {
	switch t := node.(type) {
	case *gherkin.Feature:
		fmt.Println(bcl("Feature: ", white) + t.Title)
		fmt.Println(t.Description + "\n")
	case *gherkin.Background:
		fmt.Println(bcl("Background:", white))
	case *gherkin.Scenario:
		fmt.Println(bcl("Scenario: ", white) + t.Title)
	case *gherkin.Step:
		fmt.Println(bcl(t.Token.Text, green))
	}
}
