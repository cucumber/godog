package gherkin

import (
	"strings"
	"testing"

	"github.com/l3pp4rd/go-behat/gherkin/lexer"
)

var testFeatureSamples = map[string]string{
	"feature": `Feature: gherkin parser
  in order to run features
  as gherkin lexer
  I need to be able to parse a feature`,
	"only_title": `Feature: gherkin`,
	"empty":      ``,
	"invalid":    `some text`,
	"starts_with_newlines": `

  Feature: gherkin`,
}

func (f *Feature) assertTitle(title string, t *testing.T) {
	if f.Title != title {
		t.Fatalf("expected feature title to be '%s', but got '%s'", title, f.Title)
	}
}

func (f *Feature) assertHasNumScenarios(n int, t *testing.T) {
	if len(f.Scenarios) != n {
		t.Fatalf("expected feature to have '%d' scenarios, but got '%d'", n, len(f.Scenarios))
	}
}

func Test_parse_normal_feature(t *testing.T) {
	p := &parser{
		lx:   lexer.New(strings.NewReader(testFeatureSamples["feature"])),
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
	if len(ft.Description) == 0 {
		t.Fatalf("expected a feature description to be available")
	}

	ft.AST.assertMatchesTypes([]lexer.TokenType{
		lexer.FEATURE,
		lexer.TEXT,
		lexer.TEXT,
		lexer.TEXT,
	}, t)
}

func Test_parse_feature_without_description(t *testing.T) {
	p := &parser{
		lx:   lexer.New(strings.NewReader(testFeatureSamples["only_title"])),
		path: "some.feature",
		ast:  newAST(),
	}
	ft, err := p.parseFeature()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if ft.Title != "gherkin" {
		t.Fatalf("the feature title '%s' was not expected", ft.Title)
	}
	if len(ft.Description) > 0 {
		t.Fatalf("feature description was not expected")
	}

	ft.AST.assertMatchesTypes([]lexer.TokenType{
		lexer.FEATURE,
	}, t)
}

func Test_parse_empty_feature_file(t *testing.T) {
	p := &parser{
		lx:   lexer.New(strings.NewReader(testFeatureSamples["empty"])),
		path: "some.feature",
		ast:  newAST(),
	}
	_, err := p.parseFeature()
	if err != ErrEmpty {
		t.Fatalf("expected an empty file error, but got none")
	}
}

func Test_parse_invalid_feature_with_random_text(t *testing.T) {
	p := &parser{
		lx:   lexer.New(strings.NewReader(testFeatureSamples["invalid"])),
		path: "some.feature",
		ast:  newAST(),
	}
	_, err := p.parseFeature()
	if err == nil {
		t.Fatalf("expected an error but got none")
	}
	p.ast.assertMatchesTypes([]lexer.TokenType{
		lexer.TEXT,
	}, t)
}

func Test_parse_feature_with_newlines(t *testing.T) {
	p := &parser{
		lx:   lexer.New(strings.NewReader(testFeatureSamples["starts_with_newlines"])),
		path: "some.feature",
		ast:  newAST(),
	}
	ft, err := p.parseFeature()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if ft.Title != "gherkin" {
		t.Fatalf("the feature title '%s' was not expected", ft.Title)
	}
	if len(ft.Description) > 0 {
		t.Fatalf("feature description was not expected")
	}

	ft.AST.assertMatchesTypes([]lexer.TokenType{
		lexer.NEW_LINE,
		lexer.NEW_LINE,
		lexer.FEATURE,
	}, t)
}
