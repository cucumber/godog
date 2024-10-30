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

const fakeFeature = `
		Feature: the feature
			Some feature text 

			Scenario: the scenario
				Some scenario text 
				
				Given some step 
				When other step
				Then final step

			Scenario Outline: the outline
				Given some <value>

				Examples:
					| value |
					| 1     |
					| 2     |
`

const fakeFeatureOther = `
		Feature: the other feature
			Some feature other text 

			Background:
				Given some background step

			Scenario: the other scenario
				Some other scenario text 
				
				Given some other step
				When other other step
				Then final other step
            
			Scenario: the final scenario
				Some other scenario text 
				
				Given some other step
				When other other step
				Then final other step
            `

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

	// FIXME - is thos really desirable - same name but different contents and one gets ignored???
	input := []parser.FeatureContent{
		{Name: "MyCoolDuplicatedFeature", Contents: []byte(fakeFeature)},
		{Name: "MyCoolDuplicatedFeature", Contents: []byte(fakeFeatureOther)},
	}

	featureFromBytes, err := parser.ParseFromBytes("", input)
	require.NoError(t, err)
	require.Len(t, featureFromBytes, 1)
}

func Test_ParseFromBytes_SinglePath(t *testing.T) {
	featureFileName := "godogs.feature"

	baseDir := "base"
	fsys := fstest.MapFS{
		filepath.Join(baseDir, featureFileName): {
			Data: []byte(fakeFeature),
			Mode: fs.FileMode(0644),
		},
	}

	featureFromFile, err := parser.ParseFeatures(fsys, "", []string{baseDir})
	require.NoError(t, err)
	require.Len(t, featureFromFile, 1)

	input := []parser.FeatureContent{
		{Name: filepath.Join(baseDir, featureFileName), Contents: []byte(fakeFeature)},
	}

	featureFromBytes, err := parser.ParseFromBytes("", input)
	require.NoError(t, err)
	require.Len(t, featureFromBytes, 1)

	assert.Equal(t, featureFromFile, featureFromBytes)
}

func Test_ParseFeatures_FromMultiplePaths(t *testing.T) {
	const (
		testFeatureFile = "godogs.feature"
	)

	tests := map[string]struct {
		fsys  fs.FS
		paths []string

		expFeatures  int
		expScenarios int
		expSteps     int
		expError     error
	}{
		"directories with multiple feature files can be parsed": {
			paths: []string{"base/a", "base/b"},
			fsys: fstest.MapFS{
				filepath.Join("base/a", testFeatureFile): {
					Data: []byte(fakeFeature),
				},
				filepath.Join("base/b", testFeatureFile): {
					Data: []byte(fakeFeatureOther),
				},
			},
			expFeatures:  2,
			expScenarios: 5,
			expSteps:     13,
		},
		"path not found errors": {
			fsys:     fstest.MapFS{},
			paths:    []string{"base/a", "base/b"},
			expError: errors.New(`feature path "base/a" is not available`),
		},
		"feature files can be parsed from multiple paths": {
			paths: []string{
				filepath.Join("base/a/", testFeatureFile),
				filepath.Join("base/b/", testFeatureFile),
			},
			fsys: fstest.MapFS{
				filepath.Join("base/a", testFeatureFile): {
					Data: []byte(fakeFeature),
				},
				filepath.Join("base/b", testFeatureFile): {
					Data: []byte(fakeFeatureOther),
				},
			},
			expFeatures:  2,
			expScenarios: 5,
			expSteps:     13,
		},
	}

	for testName, testCase := range tests {

		test := testCase // avoids bug: "loop variable test captured by func literal"
		name := testName // avoids bug: "loop variable test captured by func literal"

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			features, err := parser.ParseFeatures(test.fsys, "", test.paths)
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

			scenarioCount := 0
			stepCount := 0
			for _, feature := range features {
				scenarioCount += len(feature.Pickles)

				for _, pickle := range feature.Pickles {
					stepCount += len(pickle.Steps)
				}
			}

			require.Equal(t, test.expScenarios, scenarioCount, name+" : scenarios")
			require.Equal(t, test.expSteps, stepCount, name+" : steps")

		})
	}
}
