package gherkin

import (
	"strings"
	"testing"

	"github.com/l3pp4rd/go-behat/gherkin/lexer"
)

var testStepSamples = map[string]string{
	"given": indent(4, `Given I'm a step`),

	"given_table_hash": `Given there are users:
  | name | John Doe |`,

	"given_table": `Given there are users:
  | name | lastname |
  | John | Doe      |
  | Jane | Doe      |`,

	"then_pystring": `Then there should be text:
  """
    Some text
    And more
  """`,

	"when_pystring_empty": `When I do request with body:
  """
  """`,

	"when_pystring_unclosed": `When I do request with body:
  """
  {"json": "data"}
  ""`,

	"step_group": `Given there are conditions
  And there are more conditions
  When I do something
  Then something should happen`,

	"step_group_another": `Given an admin user "John Doe"
  And user "John Doe" belongs to user group "editors"
  When I do something
  Then I expect the result`,
}

func (s *Step) assertType(typ StepType, t *testing.T) {
	if s.Type != typ {
		t.Fatalf("expected step '%s' type to be '%s', but got '%s'", s.Text, typ, s.Type)
	}
}

func (s *Step) assertText(text string, t *testing.T) {
	if s.Text != text {
		t.Fatalf("expected step text to be '%s', but got '%s'", text, s.Text)
	}
}

func (s *Step) assertPyString(text string, t *testing.T) {
	if s.PyString == nil {
		t.Fatalf("step '%s %s' has no pystring", s.Type, s.Text)
	}
	if s.PyString.Body != text {
		t.Fatalf("expected step pystring body to be '%s', but got '%s'", text, s.PyString.Body)
	}
}

func (s *Step) assertTableRow(t *testing.T, num int, cols ...string) {
	if s.Table == nil {
		t.Fatalf("step '%s %s' has no table", s.Type, s.Text)
	}
	if len(s.Table.rows) <= num {
		t.Fatalf("step '%s %s' table has no row: %d", s.Type, s.Text, num)
	}
	if len(s.Table.rows[num]) != len(cols) {
		t.Fatalf("step '%s %s' table row length, does not match expected: %d", s.Type, s.Text, len(cols))
	}
	for i, col := range s.Table.rows[num] {
		if col != cols[i] {
			t.Fatalf("step '%s %s' table row %d, column %d - value '%s', does not match expected: %s", s.Type, s.Text, num, i, col, cols[i])
		}
	}
}

func Test_parse_basic_given_step(t *testing.T) {
	p := &parser{
		lx:   lexer.New(strings.NewReader(testStepSamples["given"])),
		path: "some.feature",
		ast:  newAST(),
	}
	steps, err := p.parseSteps()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(steps) != 1 {
		t.Fatalf("expected one step to be parsed")
	}

	steps[0].assertType(Given, t)
	steps[0].assertText("I'm a step", t)

	p.next() // step over to eof
	p.ast.assertMatchesTypes([]lexer.TokenType{
		lexer.GIVEN,
		lexer.EOF,
	}, t)
}

func Test_parse_hash_table_given_step(t *testing.T) {
	p := &parser{
		lx:   lexer.New(strings.NewReader(testStepSamples["given_table_hash"])),
		path: "some.feature",
		ast:  newAST(),
	}
	steps, err := p.parseSteps()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(steps) != 1 {
		t.Fatalf("expected one step to be parsed")
	}

	steps[0].assertType(Given, t)
	steps[0].assertText("there are users:", t)
	steps[0].assertTableRow(t, 0, "name", "John Doe")

	p.next() // step over to eof
	p.ast.assertMatchesTypes([]lexer.TokenType{
		lexer.GIVEN,
		lexer.TABLE_ROW,
		lexer.EOF,
	}, t)
}

func Test_parse_table_given_step(t *testing.T) {
	p := &parser{
		lx:   lexer.New(strings.NewReader(testStepSamples["given_table"])),
		path: "some.feature",
		ast:  newAST(),
	}
	steps, err := p.parseSteps()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(steps) != 1 {
		t.Fatalf("expected one step to be parsed")
	}

	steps[0].assertType(Given, t)
	steps[0].assertText("there are users:", t)
	steps[0].assertTableRow(t, 0, "name", "lastname")
	steps[0].assertTableRow(t, 1, "John", "Doe")
	steps[0].assertTableRow(t, 2, "Jane", "Doe")

	p.next() // step over to eof
	p.ast.assertMatchesTypes([]lexer.TokenType{
		lexer.GIVEN,
		lexer.TABLE_ROW,
		lexer.TABLE_ROW,
		lexer.TABLE_ROW,
		lexer.EOF,
	}, t)
}

func Test_parse_pystring_step(t *testing.T) {
	p := &parser{
		lx:   lexer.New(strings.NewReader(testStepSamples["then_pystring"])),
		path: "some.feature",
		ast:  newAST(),
	}
	steps, err := p.parseSteps()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(steps) != 1 {
		t.Fatalf("expected one step to be parsed")
	}

	steps[0].assertType(Then, t)
	steps[0].assertText("there should be text:", t)
	steps[0].assertPyString(strings.Join([]string{
		indent(4, "Some text"),
		indent(4, "And more"),
	}, "\n"), t)

	p.next() // step over to eof
	p.ast.assertMatchesTypes([]lexer.TokenType{
		lexer.THEN,
		lexer.PYSTRING,
		lexer.TEXT,
		lexer.AND, // we do not care what we parse inside PYSTRING even if its whole behat feature text
		lexer.PYSTRING,
		lexer.EOF,
	}, t)
}

func Test_parse_empty_pystring_step(t *testing.T) {
	p := &parser{
		lx:   lexer.New(strings.NewReader(testStepSamples["when_pystring_empty"])),
		path: "some.feature",
		ast:  newAST(),
	}
	steps, err := p.parseSteps()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(steps) != 1 {
		t.Fatalf("expected one step to be parsed")
	}

	steps[0].assertType(When, t)
	steps[0].assertText("I do request with body:", t)
	steps[0].assertPyString("", t)

	p.next() // step over to eof
	p.ast.assertMatchesTypes([]lexer.TokenType{
		lexer.WHEN,
		lexer.PYSTRING,
		lexer.PYSTRING,
		lexer.EOF,
	}, t)
}

func Test_parse_unclosed_pystring_step(t *testing.T) {
	p := &parser{
		lx:   lexer.New(strings.NewReader(testStepSamples["when_pystring_unclosed"])),
		path: "some.feature",
		ast:  newAST(),
	}
	_, err := p.parseSteps()
	if err == nil {
		t.Fatalf("expected an error, but got none")
	}
	p.ast.assertMatchesTypes([]lexer.TokenType{
		lexer.WHEN,
		lexer.PYSTRING,
		lexer.TEXT,
		lexer.TEXT,
		lexer.EOF,
	}, t)
}

func Test_parse_step_group(t *testing.T) {
	p := &parser{
		lx:   lexer.New(strings.NewReader(testStepSamples["step_group"])),
		path: "some.feature",
		ast:  newAST(),
	}
	steps, err := p.parseSteps()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(steps) != 4 {
		t.Fatalf("expected four steps to be parsed, but got: %d", len(steps))
	}

	steps[0].assertType(Given, t)
	steps[0].assertText("there are conditions", t)
	steps[1].assertType(Given, t)
	steps[1].assertText("there are more conditions", t)
	steps[2].assertType(When, t)
	steps[2].assertText("I do something", t)
	steps[3].assertType(Then, t)
	steps[3].assertText("something should happen", t)

	p.next() // step over to eof
	p.ast.assertMatchesTypes([]lexer.TokenType{
		lexer.GIVEN,
		lexer.AND,
		lexer.WHEN,
		lexer.THEN,
		lexer.EOF,
	}, t)
}

func Test_parse_another_step_group(t *testing.T) {
	p := &parser{
		lx:   lexer.New(strings.NewReader(testStepSamples["step_group_another"])),
		path: "some.feature",
		ast:  newAST(),
	}
	steps, err := p.parseSteps()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if len(steps) != 4 {
		t.Fatalf("expected four steps to be parsed, but got: %d", len(steps))
	}

	steps[0].assertType(Given, t)
	steps[0].assertText(`an admin user "John Doe"`, t)
	steps[1].assertType(Given, t)
	steps[1].assertText(`user "John Doe" belongs to user group "editors"`, t)
	steps[2].assertType(When, t)
	steps[2].assertText("I do something", t)
	steps[3].assertType(Then, t)
	steps[3].assertText("I expect the result", t)

	p.next() // step over to eof
	p.ast.assertMatchesTypes([]lexer.TokenType{
		lexer.GIVEN,
		lexer.AND,
		lexer.WHEN,
		lexer.THEN,
		lexer.EOF,
	}, t)
}
