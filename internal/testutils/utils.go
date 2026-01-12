package testutils

import (
	"strings"
	"testing"

	gherkin "github.com/cucumber/gherkin/go/v26"
	messages "github.com/cucumber/messages/go/v31"
	"github.com/stretchr/testify/require"

	"github.com/cucumber/godog/internal/models"
)

// BuildTestFeature creates a feature for testing purpose.
//
// The created feature includes:
//   - a background
//   - one normal scenario with three steps
//   - one outline scenario with one example and three steps
func BuildTestFeature(t *testing.T) models.Feature {
	newIDFunc := (&messages.Incrementing{}).NewId

	gherkinDocument, err := gherkin.ParseGherkinDocument(strings.NewReader(featureContent), newIDFunc)
	require.NoError(t, err)

	path := t.Name()
	gherkinDocument.Uri = path
	pickles := gherkin.Pickles(*gherkinDocument, path, newIDFunc)

	ft := models.Feature{GherkinDocument: gherkinDocument, Pickles: pickles, Content: []byte(featureContent)}
	require.Len(t, ft.Pickles, 2)

	require.Len(t, ft.Pickles[0].AstNodeIds, 1)
	require.Len(t, ft.Pickles[0].Steps, 3)

	require.Len(t, ft.Pickles[1].AstNodeIds, 2)
	require.Len(t, ft.Pickles[1].Steps, 3)

	return ft
}

const featureContent = `Feature: eat godogs
In order to be happy
As a hungry gopher
I need to be able to eat godogs

Background:
  Given there are <begin> godogs

Scenario: Eat 5 out of 12
  When I eat 5
  Then there should be 7 remaining

Scenario Outline: Eat <dec> out of <beginning>
  When I eat <dec>
  Then there should be <remain> remaining

  Examples:
	| begin | dec | remain |
	| 12    | 5   | 7      |`

// BuildTestFeature creates a feature with rules for testing purpose.
//
// The created feature includes:
//   - a background
//   - one normal scenario with three steps
//   - one outline scenario with one example and three steps
func BuildTestFeatureWithRules(t *testing.T) models.Feature {
	newIDFunc := (&messages.Incrementing{}).NewId

	gherkinDocument, err := gherkin.ParseGherkinDocument(strings.NewReader(featureWithRuleContent), newIDFunc)
	require.NoError(t, err)

	path := t.Name()
	gherkinDocument.Uri = path
	pickles := gherkin.Pickles(*gherkinDocument, path, newIDFunc)

	ft := models.Feature{GherkinDocument: gherkinDocument, Pickles: pickles, Content: []byte(featureWithRuleContent)}
	require.Len(t, ft.Pickles, 2)

	require.Len(t, ft.Pickles[0].AstNodeIds, 1)
	require.Len(t, ft.Pickles[0].Steps, 3)

	require.Len(t, ft.Pickles[1].AstNodeIds, 2)
	require.Len(t, ft.Pickles[1].Steps, 3)

	return ft
}

const featureWithRuleContent = `Feature: eat godogs
In order to be happy
As a hungry gopher
I need to be able to eat godogs

Rule: eating godogs

Background:
  Given there are <begin> godogs

Scenario: Eat 5 out of 12
  When I eat 5
  Then there should be 7 remaining

Scenario Outline: Eat <dec> out of <beginning>
  When I eat <dec>
  Then there should be <remain> remaining

  Examples:
	| begin | dec | remain |
	| 12    | 5   | 7      |`
