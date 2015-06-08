package gherkin

import (
	"strings"
	"testing"

	"github.com/l3pp4rd/go-behat/gherkin/lexer"
)

var testFeatureSamples = map[string]string{
	"full": `Feature: gherkin parser
  in order to run features
  as gherkin lexer
  I need to be able to parse a feature`,
	"only_title": `Feature: gherkin`,
}

func Test_normal_feature_parsing(t *testing.T) {
	p := &parser{
		lx:   lexer.New(strings.NewReader(testFeatureSamples["full"])),
		path: "some.feature",
		ast:  newAST(),
	}
	ft, err := p.parseFeature()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if ft.Title != "gherkin parser" {
		t.Fatalf("the feature title '%s' was not expected", ft.Title)
	}
}
