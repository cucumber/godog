package gherkin

import (
	"strings"
	"testing"
)

var testLexerSamples = map[string]string{
	"feature": `Feature: gherkin lexer
  in order to run features
  as gherkin lexer
  I need to be able to parse a feature`,

	"background": `Background:`,

	"scenario": "Scenario: tokenize feature file",

	"step_given": `Given a feature file`,

	"step_when": `When I try to read it`,

	"comment": `# an important comment`,

	"step_then": `Then it should give me tokens`,

	"step_given_table": `Given there are users:
      | name | lastname | num |
      | Jack | Sparrow  | 4   |
      | John | Doe      | 79  |`,

	"scenario_outline_with_examples": `Scenario Outline: ls supports kinds of options
	  Given I am in a directory "test"
      And I have a file named "foo"
      And I have a file named "bar"
      When I run "ls" with options "<options>"
	  Then I should see "<result>"

	Examples:
	  | options | result  |
	  | -t      | bar foo |
	  | -tr     | foo bar |`,
}

func Test_feature_read(t *testing.T) {
	l := newLexer(strings.NewReader(testLexerSamples["feature"]))
	tok := l.read()
	if tok.Type != FEATURE {
		t.Fatalf("Expected a 'feature' type, but got: '%s'", tok.Type)
	}
	val := "gherkin lexer"
	if tok.Value != val {
		t.Fatalf("Expected a token value to be '%s', but got: '%s'", val, tok.Value)
	}
	if tok.Line != 1 {
		t.Fatalf("Expected a token line to be '1', but got: '%d'", tok.Line)
	}
	if tok.Indent != 0 {
		t.Fatalf("Expected a token identation to be '0', but got: '%d'", tok.Indent)
	}

	tok = l.read()
	if tok.Type != TEXT {
		t.Fatalf("Expected a 'text' type, but got: '%s'", tok.Type)
	}
	val = "in order to run features"
	if tok.Value != val {
		t.Fatalf("Expected a token value to be '%s', but got: '%s'", val, tok.Value)
	}
	if tok.Line != 2 {
		t.Fatalf("Expected a token line to be '2', but got: '%d'", tok.Line)
	}
	if tok.Indent != 2 {
		t.Fatalf("Expected a token identation to be '2', but got: '%d'", tok.Indent)
	}

	tok = l.read()
	if tok.Type != TEXT {
		t.Fatalf("Expected a 'text' type, but got: '%s'", tok.Type)
	}
	val = "as gherkin lexer"
	if tok.Value != val {
		t.Fatalf("Expected a token value to be '%s', but got: '%s'", val, tok.Value)
	}
	if tok.Line != 3 {
		t.Fatalf("Expected a token line to be '3', but got: '%d'", tok.Line)
	}
	if tok.Indent != 2 {
		t.Fatalf("Expected a token identation to be '2', but got: '%d'", tok.Indent)
	}

	tok = l.read()
	if tok.Type != TEXT {
		t.Fatalf("Expected a 'text' type, but got: '%s'", tok.Type)
	}
	val = "I need to be able to parse a feature"
	if tok.Value != val {
		t.Fatalf("Expected a token value to be '%s', but got: '%s'", val, tok.Value)
	}
	if tok.Line != 4 {
		t.Fatalf("Expected a token line to be '4', but got: '%d'", tok.Line)
	}
	if tok.Indent != 2 {
		t.Fatalf("Expected a token identation to be '2', but got: '%d'", tok.Indent)
	}

	tok = l.read()
	if tok.Type != EOF {
		t.Fatalf("Expected an 'eof' type, but got: '%s'", tok.Type)
	}
}

func Test_minimal_feature(t *testing.T) {
	file := strings.Join([]string{
		testLexerSamples["feature"] + "\n",

		indent(2, testLexerSamples["background"]),
		indent(4, testLexerSamples["step_given"]) + "\n",

		indent(2, testLexerSamples["comment"]),
		indent(2, testLexerSamples["scenario"]),
		indent(4, testLexerSamples["step_given"]),
		indent(4, testLexerSamples["step_when"]),
		indent(4, testLexerSamples["step_then"]),
	}, "\n")
	l := newLexer(strings.NewReader(file))

	var tokens []TokenType
	for tok := l.read(); tok.Type != EOF; tok = l.read() {
		tokens = append(tokens, tok.Type)
	}
	expected := []TokenType{
		FEATURE,
		TEXT,
		TEXT,
		TEXT,
		NEWLINE,

		BACKGROUND,
		GIVEN,
		NEWLINE,

		COMMENT,
		SCENARIO,
		GIVEN,
		WHEN,
		THEN,
	}
	for i := 0; i < len(expected); i++ {
		if expected[i] != tokens[i] {
			t.Fatalf("expected token '%s' at position: %d, is not the same as actual token: '%s'", expected[i], i, tokens[i])
		}
	}
}

func Test_table_row_reading(t *testing.T) {
	file := strings.Join([]string{
		indent(2, testLexerSamples["background"]),
		indent(4, testLexerSamples["step_given_table"]),
		indent(4, testLexerSamples["step_given"]),
	}, "\n")
	l := newLexer(strings.NewReader(file))

	var types []TokenType
	var values []string
	var indents []int
	for tok := l.read(); tok.Type != EOF; tok = l.read() {
		types = append(types, tok.Type)
		values = append(values, tok.Value)
		indents = append(indents, tok.Indent)
	}
	expectedTypes := []TokenType{
		BACKGROUND,
		GIVEN,
		TABLEROW,
		TABLEROW,
		TABLEROW,
		GIVEN,
	}
	expectedIndents := []int{2, 4, 6, 6, 6, 4}
	for i := 0; i < len(expectedTypes); i++ {
		if expectedTypes[i] != types[i] {
			t.Fatalf("expected token type '%s' at position: %d, is not the same as actual: '%s'", expectedTypes[i], i, types[i])
		}
	}
	for i := 0; i < len(expectedIndents); i++ {
		if expectedIndents[i] != indents[i] {
			t.Fatalf("expected token indentation '%d' at position: %d, is not the same as actual: '%d'", expectedIndents[i], i, indents[i])
		}
	}
	if values[2] != "name | lastname | num |" {
		t.Fatalf("table row value '%s' was not expected", values[2])
	}
}

func Test_lexing_of_scenario_outline(t *testing.T) {
	l := newLexer(strings.NewReader(testLexerSamples["scenario_outline_with_examples"]))

	var tokens []TokenType
	for tok := l.read(); tok.Type != EOF; tok = l.read() {
		tokens = append(tokens, tok.Type)
	}
	expected := []TokenType{
		OUTLINE,
		GIVEN,
		AND,
		AND,
		WHEN,
		THEN,
		NEWLINE,

		EXAMPLES,
		TABLEROW,
		TABLEROW,
		TABLEROW,
	}
	for i := 0; i < len(expected); i++ {
		if expected[i] != tokens[i] {
			t.Fatalf("expected token '%s' at position: %d, is not the same as actual token: '%s'", expected[i], i, tokens[i])
		}
	}
}
