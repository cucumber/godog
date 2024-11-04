package testutils

import (
	"bytes"
	"fmt"
	perr "github.com/pkg/errors"
	"strings"
	"testing"

	gherkin "github.com/cucumber/gherkin/go/v26"
	messages "github.com/cucumber/messages/go/v21"
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

func VDiffString(expected, actual string) {
	list1 := strings.Split(expected, "\n")
	list2 := strings.Split(actual, "\n")

	VDiffLists(list1, list2)
}

func VDiffLists(list1 []string, list2 []string) error {

	var diffs bool
	buf := &bytes.Buffer{}

	// wrapString wraps a string into chunks of the given width.
	wrapString := func(str string, width int) []string {
		return []string{str}
	}
	// Get the length of the longer list
	maxLength := len(list1)
	if len(list2) > maxLength {
		maxLength = len(list2)
	}
	maxlen := 0
	for i := 0; i < len(list1); i++ {
		if len(list1[i]) > maxlen {
			maxlen = len(list1[i])
		}
	}
	for i := 0; i < len(list2); i++ {
		if len(list2[i]) > maxlen {
			maxlen = len(list2[i])
		}
	}

	colWid := maxlen
	fmtTitle := fmt.Sprintf("%2s %%3s: %%-%ds | %%-%ds\n", "", colWid+2, colWid+2)
	fmtData := fmt.Sprintf("%%2s %%3d: %%-%ds | %%-%ds\n", colWid+2, colWid+2)

	fmt.Fprintf(buf, fmtTitle, "#", "expected", "actual")

	for i := 0; i < maxLength; i++ {
		var val1, val2 string

		// Get the value from list1 if it exists
		if i < len(list1) {
			val1 = list1[i]
		} else {
			val1 = "N/A"
		}

		// Get the value from list2 if it exists
		if i < len(list2) {
			val2 = list2[i]
		} else {
			val2 = "N/A"
		}

		// Wrap both strings into slices of strings with fixed width
		wrapped1 := wrapString(val1, colWid)
		wrapped2 := wrapString(val2, colWid)

		// Find the number of wrapped lines needed for the current pair
		maxWrappedLines := len(wrapped1)
		if len(wrapped2) > maxWrappedLines {
			maxWrappedLines = len(wrapped2)
		}

		// Print the wrapped lines with alignment
		for j := 0; j < maxWrappedLines; j++ {
			var line1, line2 string

			// Get the wrapped line or use an empty string if it doesn't exist
			if j < len(wrapped1) {
				line1 = wrapped1[j]
			} else {
				line1 = ""
			}

			if j < len(wrapped2) {
				line2 = wrapped2[j]
			} else {
				line2 = ""
			}

			status := "  "
			// if val1 != val2 {
			if strings.TrimSpace(line1) != strings.TrimSpace(line2) {
				status = "* "
				diffs = true
			}

			delim := "¬"
			// Print the wrapped lines with fixed-width column
			fmt.Fprintf(buf, fmtData, status, i+1, delim+line1+delim, delim+line2+delim)
		}
	}

	if diffs {
		return perr.New(buf.String())
	}
	return nil
}
