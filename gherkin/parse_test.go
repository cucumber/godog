package gherkin

import (
	"strings"
	"testing"

	"github.com/l3pp4rd/go-behat/gherkin/lexer"
)

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
		lx:   lexer.New(strings.NewReader(content)),
		path: "usual.feature",
		ast:  newAST(),
	}
	ft, err := p.parseFeature()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	ft.assertTitle("gherkin parser", t)

	ft.AST.assertMatchesTypes([]lexer.TokenType{
		lexer.TAGS,
		lexer.FEATURE,
		lexer.TEXT,
		lexer.TEXT,
		lexer.TEXT,
		lexer.NEW_LINE,

		lexer.BACKGROUND,
		lexer.GIVEN,
		lexer.TABLE_ROW,
		lexer.NEW_LINE,

		lexer.SCENARIO,
		lexer.GIVEN,
		lexer.AND,
		lexer.WHEN,
		lexer.THEN,
		lexer.NEW_LINE,

		lexer.TAGS,
		lexer.SCENARIO,
		lexer.GIVEN,
		lexer.AND,
		lexer.WHEN,
		lexer.THEN,
		lexer.NEW_LINE,

		lexer.TAGS,
		lexer.SCENARIO,
	}, t)
}
