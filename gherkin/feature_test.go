package gherkin

import (
	"strings"
	"testing"
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
		lx:   newLexer(strings.NewReader(testFeatureSamples["feature"])),
		path: "some.feature",
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

	p.assertMatchesTypes([]TokenType{
		FEATURE,
		TEXT,
		TEXT,
		TEXT,
	}, t)
}

func Test_parse_feature_without_description(t *testing.T) {
	p := &parser{
		lx:   newLexer(strings.NewReader(testFeatureSamples["only_title"])),
		path: "some.feature",
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

	p.assertMatchesTypes([]TokenType{
		FEATURE,
	}, t)
}

func Test_parse_empty_feature_file(t *testing.T) {
	p := &parser{
		lx:   newLexer(strings.NewReader(testFeatureSamples["empty"])),
		path: "some.feature",
	}
	_, err := p.parseFeature()
	if err != ErrEmpty {
		t.Fatalf("expected an empty file error, but got none")
	}
}

func Test_parse_invalid_feature_with_random_text(t *testing.T) {
	p := &parser{
		lx:   newLexer(strings.NewReader(testFeatureSamples["invalid"])),
		path: "some.feature",
	}
	_, err := p.parseFeature()
	if err == nil {
		t.Fatalf("expected an error but got none")
	}
	p.assertMatchesTypes([]TokenType{
		TEXT,
	}, t)
}

func Test_parse_feature_with_newlines(t *testing.T) {
	p := &parser{
		lx:   newLexer(strings.NewReader(testFeatureSamples["starts_with_newlines"])),
		path: "some.feature",
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

	p.assertMatchesTypes([]TokenType{
		NEW_LINE,
		NEW_LINE,
		FEATURE,
	}, t)
}
