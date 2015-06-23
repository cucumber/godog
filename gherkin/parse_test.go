package gherkin

import (
	"strings"
	"testing"
)

func (a *parser) assertMatchesTypes(expected []TokenType, t *testing.T) {
	key := -1
	for _, tok := range a.ast {
		key++
		if len(expected) <= key {
			t.Fatalf("there are more tokens in AST then expected, next is '%s'", tok.Type)
		}
		if expected[key] != tok.Type {
			t.Fatalf("expected ast token '%s', but got '%s' at position: %d", expected[key], tok.Type, key)
		}
	}
	if len(expected)-1 != key {
		t.Fatalf("expected ast length %d, does not match actual: %d", len(expected), key+1)
	}
}

func (s *Scenario) assertHasTag(tag string, t *testing.T) {
	if !s.Tags.Has(Tag(tag)) {
		t.Fatalf("expected scenario '%s' to have '%s' tag, but it did not", s.Title, tag)
	}
}

func (s *Scenario) assertHasNumTags(n int, t *testing.T) {
	if len(s.Tags) != n {
		t.Fatalf("expected scenario '%s' to have '%d' tags, but it has '%d'", s.Title, n, len(s.Tags))
	}
}

func Test_parse_feature_file(t *testing.T) {

	content := strings.Join([]string{
		// feature
		"@global-one @cust",
		testFeatureSamples["feature"] + "\n",
		// background
		indent(2, "Background:"),
		testStepSamples["given_table_hash"] + "\n",
		// scenario - normal without tags
		indent(2, "Scenario: user is able to register"),
		testStepSamples["step_group"] + "\n",
		// scenario - repeated tag, one extra
		indent(2, "@user @cust"),
		indent(2, "Scenario: password is required to login"),
		testStepSamples["step_group_another"] + "\n",
		// scenario - no steps yet
		indent(2, "@todo"), // cust - tag is repeated
		indent(2, "Scenario: user is able to reset his password") + "\n",
		// scenario outline
		testLexerSamples["scenario_outline_with_examples"],
	}, "\n")

	p := &parser{
		lx:   newLexer(strings.NewReader(content)),
		path: "usual.feature",
	}
	ft, err := p.parseFeature()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	ft.assertTitle("gherkin parser", t)

	p.assertMatchesTypes([]TokenType{
		TAGS,
		FEATURE,
		TEXT,
		TEXT,
		TEXT,
		NEWLINE,

		BACKGROUND,
		GIVEN,
		TABLEROW,
		NEWLINE,

		SCENARIO,
		GIVEN,
		AND,
		WHEN,
		THEN,
		NEWLINE,

		TAGS,
		SCENARIO,
		GIVEN,
		AND,
		WHEN,
		THEN,
		NEWLINE,

		TAGS,
		SCENARIO,
		NEWLINE,

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
	}, t)

	ft.assertHasNumScenarios(4, t)

	ft.Scenarios[0].assertHasNumTags(2, t)
	ft.Scenarios[0].assertHasTag("global-one", t)
	ft.Scenarios[0].assertHasTag("cust", t)

	ft.Scenarios[1].assertHasNumTags(3, t)
	ft.Scenarios[1].assertHasTag("global-one", t)
	ft.Scenarios[1].assertHasTag("cust", t)
	ft.Scenarios[1].assertHasTag("user", t)

	ft.Scenarios[2].assertHasNumTags(3, t)
	ft.Scenarios[2].assertHasTag("global-one", t)
	ft.Scenarios[2].assertHasTag("cust", t)
	ft.Scenarios[2].assertHasTag("todo", t)

	ft.Scenarios[3].assertHasNumTags(2, t)
}
