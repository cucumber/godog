package gherkin

import (
	"strings"
	"testing"
)

func (s *Scenario) assertTitle(title string, t *testing.T) {
	if s.Title != title {
		t.Fatalf("expected scenario title to be '%s', but got '%s'", title, s.Title)
	}
}

func (s *Scenario) assertStep(text string, t *testing.T) *Step {
	for _, stp := range s.Steps {
		if stp.Text == text {
			return stp
		}
	}
	t.Fatal("expected scenario '%s' to have step: '%s', but it did not", s.Title, text)
	return nil
}

func (s *Scenario) assertExampleRow(t *testing.T, num int, cols ...string) {
	if s.Examples == nil {
		t.Fatalf("outline scenario '%s' has no examples", s.Title)
	}
	if len(s.Examples.Rows) <= num {
		t.Fatalf("outline scenario '%s' table has no row: %d", s.Title, num)
	}
	if len(s.Examples.Rows[num]) != len(cols) {
		t.Fatalf("outline scenario '%s' table row length, does not match expected: %d", s.Title, len(cols))
	}
	for i, col := range s.Examples.Rows[num] {
		if col != cols[i] {
			t.Fatalf("outline scenario '%s' table row %d, column %d - value '%s', does not match expected: %s", s.Title, num, i, col, cols[i])
		}
	}
}

func Test_parse_scenario_outline(t *testing.T) {

	p := &parser{
		lx:   newLexer(strings.NewReader(testLexerSamples["scenario_outline_with_examples"])),
		path: "usual.feature",
	}
	s, err := p.parseScenario()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	s.assertTitle("ls supports kinds of options", t)

	p.assertMatchesTypes([]TokenType{
		SCENARIO_OUTLINE,
		GIVEN,
		AND,
		AND,
		WHEN,
		THEN,
		NEW_LINE,
		EXAMPLES,
		TABLE_ROW,
		TABLE_ROW,
		TABLE_ROW,
	}, t)

	s.assertStep(`I am in a directory "test"`, t)
	s.assertStep(`I have a file named "foo"`, t)
	s.assertStep(`I have a file named "bar"`, t)
	s.assertStep(`I run "ls" with options "<options>"`, t)
	s.assertStep(`I should see "<result>"`, t)

	s.assertExampleRow(t, 0, "options", "result")
	s.assertExampleRow(t, 1, "-t", "bar foo")
	s.assertExampleRow(t, 2, "-tr", "foo bar")
}
