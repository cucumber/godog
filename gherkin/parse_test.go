package gherkin

import (
	"strings"
	"testing"
)

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
		indent(2, "Scenario: user is able to reset his password"),
	}, "\n")

	p := &parser{
		lx:   newLexer(strings.NewReader(content)),
		path: "usual.feature",
		ast:  newAST(),
	}
	ft, err := p.parseFeature()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	ft.assertTitle("gherkin parser", t)

	ft.AST.assertMatchesTypes([]TokenType{
		TAGS,
		FEATURE,
		TEXT,
		TEXT,
		TEXT,
		NEW_LINE,

		BACKGROUND,
		GIVEN,
		TABLE_ROW,
		NEW_LINE,

		SCENARIO,
		GIVEN,
		AND,
		WHEN,
		THEN,
		NEW_LINE,

		TAGS,
		SCENARIO,
		GIVEN,
		AND,
		WHEN,
		THEN,
		NEW_LINE,

		TAGS,
		SCENARIO,
	}, t)

	ft.assertHasNumScenarios(3, t)

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
}
