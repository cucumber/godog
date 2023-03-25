package parser_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

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

	featureFromBytes, err := parser.ParseFromBytes("", input)
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

	baseDir := filepath.Join(os.TempDir(), t.Name(), "godogs")
	errA := os.MkdirAll(baseDir+"/a", 0755)
	defer os.RemoveAll(baseDir)

	require.Nil(t, errA)

	err := ioutil.WriteFile(filepath.Join(baseDir, featureFileName), []byte(eatGodogContents), 0644)
	require.Nil(t, err)

	featureFromFile, err := parser.ParseFeatures(os.DirFS("./"), "", []string{baseDir})
	require.NoError(t, err)
	require.Len(t, featureFromFile, 1)

	input := []parser.FeatureContent{
		{Name: filepath.Join(baseDir, featureFileName), Contents: []byte(eatGodogContents)},
	}

	featureFromBytes, err := parser.ParseFromBytes("", input)
	require.NoError(t, err)
	require.Len(t, featureFromBytes, 1)

	assert.Equal(t, featureFromFile, featureFromBytes)
}

func Test_ParseFeatures_FromMultiplePaths(t *testing.T) {
	const featureFileName = "godogs.feature"
	const featureFileContents = `Feature: eat godogs
  In order to be happy
  As a hungry gopher
  I need to be able to eat godogs

  Scenario: Eat 5 out of 12
    Given there are 12 godogs
    When I eat 5
		Then there should be 7 remaining`

	baseDir := filepath.Join(os.TempDir(), t.Name(), "godogs")
	errA := os.MkdirAll(baseDir+"/a", 0755)
	errB := os.MkdirAll(baseDir+"/b", 0755)
	defer os.RemoveAll(baseDir)

	require.Nil(t, errA)
	require.Nil(t, errB)

	err := ioutil.WriteFile(filepath.Join(baseDir+"/a", featureFileName), []byte(featureFileContents), 0644)
	require.Nil(t, err)
	err = ioutil.WriteFile(filepath.Join(baseDir+"/b", featureFileName), []byte(featureFileContents), 0644)
	require.Nil(t, err)

	features, err := parser.ParseFeatures(os.DirFS("./"), "", []string{baseDir + "/a", baseDir + "/b"})
	assert.Nil(t, err)
	assert.Len(t, features, 2)

	pickleIDs := map[string]bool{}
	for _, f := range features {
		for _, p := range f.Pickles {
			if pickleIDs[p.Id] {
				assert.Failf(t, "found duplicate pickle ID", "Pickle ID %s was already used", p.Id)
			}

			pickleIDs[p.Id] = true
		}
	}
}
