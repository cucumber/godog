package parser_test

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cucumber/godog/internal/parser"
)

func Test_FeatureFilePathParser(t *testing.T) {
	type Case struct {
		input string
		path  string
		line  int
	}

	cases := []Case{
		{"/home/test.feature", "/home/test.feature", -1},
		{"/home/test.feature:21", "/home/test.feature", 21},
		{"test.feature", "test.feature", -1},
		{"test.feature:2", "test.feature", 2},
		{"", "", -1},
		{"/c:/home/test.feature", "/c:/home/test.feature", -1},
		{"/c:/home/test.feature:3", "/c:/home/test.feature", 3},
		{"D:\\home\\test.feature:3", "D:\\home\\test.feature", 3},
	}

	for _, c := range cases {
		p, ln := parser.ExtractFeaturePathLine(c.input)
		assert.Equal(t, p, c.path)
		assert.Equal(t, ln, c.line)
	}
}

func Test_ParseFromBytes_FromMultipleFeatures_DuplicateNames(t *testing.T) {
	eatGodogContents := `
Feature: eat godogs
  In order to be happy
  As a hungry gopher
  I need to be able to eat godogs

  Scenario: Eat 5 out of 12
    Given there are 12 godogs
    When I eat 5
    Then there should be 7 remaining`
	input := []parser.FeatureContent{
		{Name: "MyCoolDuplicatedFeature", Contents: []byte(eatGodogContents)},
		{Name: "MyCoolDuplicatedFeature", Contents: []byte(eatGodogContents)},
	}

	featureFromBytes, err := parser.ParseFromBytes("", "", input)
	require.NoError(t, err)
	require.Len(t, featureFromBytes, 1)
}

func Test_ParseFromBytes_FromMultipleFeatures(t *testing.T) {
	featureFileName := "godogs.feature"
	eatGodogContents := `
Feature: eat godogs
  In order to be happy
  As a hungry gopher
  I need to be able to eat godogs

  Scenario: Eat 5 out of 12
    Given there are 12 godogs
    When I eat 5
    Then there should be 7 remaining`

	baseDir := "base"
	fsys := fstest.MapFS{
		filepath.Join(baseDir, featureFileName): {
			Data: []byte(eatGodogContents),
			Mode: fs.FileMode(0644),
		},
	}

	featureFromFile, err := parser.ParseFeatures(fsys, "", "", []string{baseDir})
	require.NoError(t, err)
	require.Len(t, featureFromFile, 1)

	input := []parser.FeatureContent{
		{Name: filepath.Join(baseDir, featureFileName), Contents: []byte(eatGodogContents)},
	}

	featureFromBytes, err := parser.ParseFromBytes("", "", input)
	require.NoError(t, err)
	require.Len(t, featureFromBytes, 1)

	assert.Equal(t, featureFromFile, featureFromBytes)
}

func Test_ParseFeatures_FromMultiplePaths(t *testing.T) {
	const (
		defaultFeatureFile     = "godogs.feature"
		defaultFeatureContents = `Feature: eat godogs
  In order to be happy
  As a hungry gopher
  I need to be able to eat godogs

  Scenario: Eat 5 out of 12
    Given there are 12 godogs
    When I eat 5
		Then there should be 7 remaining`
	)

	tests := map[string]struct {
		fsys  fs.FS
		paths []string

		expFeatures int
		expError    error
	}{
		"feature directories can be parsed": {
			paths: []string{"base/a", "base/b"},
			fsys: fstest.MapFS{
				filepath.Join("base/a", defaultFeatureFile): {
					Data: []byte(defaultFeatureContents),
				},
				filepath.Join("base/b", defaultFeatureFile): {
					Data: []byte(defaultFeatureContents),
				},
			},
			expFeatures: 2,
		},
		"path not found errors": {
			fsys:     fstest.MapFS{},
			paths:    []string{"base/a", "base/b"},
			expError: errors.New(`feature path "base/a" is not available`),
		},
		"feature files can be parsed": {
			paths: []string{
				filepath.Join("base/a/", defaultFeatureFile),
				filepath.Join("base/b/", defaultFeatureFile),
			},
			fsys: fstest.MapFS{
				filepath.Join("base/a", defaultFeatureFile): {
					Data: []byte(defaultFeatureContents),
				},
				filepath.Join("base/b", defaultFeatureFile): {
					Data: []byte(defaultFeatureContents),
				},
			},
			expFeatures: 2,
		},
	}

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			features, err := parser.ParseFeatures(test.fsys, "", "", test.paths)
			if test.expError != nil {
				require.Error(t, err)
				require.EqualError(t, err, test.expError.Error())
				return
			}

			assert.Nil(t, err)
			assert.Len(t, features, test.expFeatures)

			pickleIDs := map[string]bool{}
			for _, f := range features {
				for _, p := range f.Pickles {
					if pickleIDs[p.Id] {
						assert.Failf(t, "found duplicate pickle ID", "Pickle ID %s was already used", p.Id)
					}

					pickleIDs[p.Id] = true
				}
			}
		})
	}
}

func Test_ParseFeatures_Localisation(t *testing.T) {
	tests := map[string]struct {
		dialect  string
		contents string
	}{
		"english": {
			dialect: "en",
			contents: `
Feature: dummy
  Rule: dummy
    Background: dummy
      Given dummy
      When dummy
      Then dummy
    Scenario: dummy
      Given dummy
      When dummy
      Then dummy
      And dummy
      But dummy
    Example: dummy
      Given dummy
      When dummy
      Then dummy
    Scenario Outline: dummy
      Given dummy
      When dummy
      Then dummy
      `,
		},
		"afrikaans": {
			dialect: "af",
			contents: `
Funksie: dummy
  Regel: dummy
    Agtergrond: dummy
      Gegewe dummy
      Wanneer dummy
      Dan dummy
    Voorbeeld: dummy
      Gegewe dummy
      Wanneer dummy
      Dan dummy
      En dummy
      Maar dummy
    Voorbeelde: dummy
      Gegewe dummy
      Wanneer dummy
      Dan dummy
    Situasie Uiteensetting: dummy
      Gegewe dummy
      Wanneer dummy
      Dan dummy
      `,
		},
		"arabic": {
			dialect: "ar",
			contents: `
خاصية: dummy
  Rule: dummy
    الخلفية: dummy
      بفرض  dummy
      متى  dummy
      اذاً  dummy
    مثال: dummy
      بفرض  dummy
      متى  dummy
      اذاً  dummy
      و dummy
      لكن dummy
    امثلة: dummy
      بفرض  dummy
      متى  dummy
      اذاً  dummy
    سيناريو مخطط: dummy
      بفرض  dummy
      متى  dummy
      اذاً  dummy
      `,
		},
		"chinese simplified": {
			dialect: "zh-CN",
			contents: `
功能: dummy
  规则: dummy
    背景: dummy
      假如 dummy
      当 dummy
      那么 dummy
    场景: dummy
      假如 dummy
      当 dummy
      那么 dummy
      而且 dummy
      但是 dummy
    例子: dummy
      假如 dummy
      当 dummy
      那么 dummy
    场景大纲: dummy
      假如 dummy
      当 dummy
      那么 dummy
      `,
		},
	}

	featureFileName := "godogs.feature"
	baseDir := "base"

	for name, test := range tests {
		test := test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fsys := fstest.MapFS{
				filepath.Join(baseDir, featureFileName): {
					Data: []byte(test.contents),
					Mode: fs.FileMode(0o644),
				},
			}

			featureTestDialect, err := parser.ParseFeatures(fsys, "", test.dialect, []string{baseDir})
			require.NoError(t, err)
			require.Len(t, featureTestDialect, 1)
		})
	}
}
