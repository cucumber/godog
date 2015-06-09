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
	"empty":      ``,
	"invalid":    `some text`,
	"starts_with_newlines": `

  Feature: gherkin`,
}

func (a *AST) assertMatchesTypes(expected []lexer.TokenType, t *testing.T) {
	key := -1
	for item := a.head; item != nil; item = item.next {
		key += 1
		if expected[key] != item.value.Type {
			t.Fatalf("expected ast token '%s', but got '%s' at position: %d", expected[key], item.value.Type, key)
		}
	}
	if len(expected)-1 != key {
		t.Fatalf("expected ast length %d, does not match actual: %d", len(expected), key+1)
	}
}

func Test_parse_normal_feature(t *testing.T) {
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
	if len(ft.Description) == 0 {
		t.Fatalf("expected a feature description to be available")
	}

	ft.AST.assertMatchesTypes([]lexer.TokenType{
		lexer.FEATURE,
		lexer.TEXT,
		lexer.TEXT,
		lexer.TEXT,
		lexer.EOF,
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
		lexer.EOF,
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
	p.ast.assertMatchesTypes([]lexer.TokenType{
		lexer.EOF,
	}, t)
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
		lexer.EOF,
	}, t)
}
