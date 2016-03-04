package godog

import "gopkg.in/cucumber/gherkin-go.v3"

// examples is a helper func to cast gherkin.Examples
// or gherkin.BaseExamples if its empty
func examples(ex interface{}) (*gherkin.Examples, bool) {
	t, ok := ex.(*gherkin.Examples)
	return t, ok
}
