package godog

import "gopkg.in/cucumber/gherkin-go.v3"

// examples is a helper func to cast gherkin.Examples
// or gherkin.BaseExamples if its empty
// @TODO: this should go away with gherkin update
func examples(ex interface{}) (*gherkin.Examples, bool) {
	t, ok := ex.(*gherkin.Examples)
	return t, ok
}
