package storage_test

import (
	"testing"
	"time"

	"github.com/cucumber/godog/internal/messages"
	"github.com/stretchr/testify/assert"

	"github.com/cucumber/godog/internal/models"
	"github.com/cucumber/godog/internal/storage"
	"github.com/cucumber/godog/internal/testutils"
)

func Test_MustGetPickle(t *testing.T) {
	s := storage.NewStorage()
	ft := testutils.BuildTestFeature(t)

	expected := ft.Pickles[0]
	s.MustInsertPickle(expected)

	actual := s.MustGetPickle(expected.Id)
	assert.Equal(t, expected, actual)
}

func Test_MustGetPickles(t *testing.T) {
	s := storage.NewStorage()
	ft := testutils.BuildTestFeature(t)

	expected := ft.Pickles
	for _, pickle := range expected {
		s.MustInsertPickle(pickle)
	}

	actual := s.MustGetPickles(ft.Uri)
	assert.Equal(t, expected, actual)
}

func Test_MustGetPickleStep(t *testing.T) {
	s := storage.NewStorage()
	ft := testutils.BuildTestFeature(t)

	for _, pickle := range ft.Pickles {
		s.MustInsertPickle(pickle)
	}

	for _, pickle := range ft.Pickles {
		for _, expected := range pickle.Steps {
			actual := s.MustGetPickleStep(expected.Id)
			assert.Equal(t, expected, actual)
		}
	}
}

func Test_MustGetTestRunStarted(t *testing.T) {
	s := storage.NewStorage()

	expected := models.TestRunStarted{StartedAt: time.Now()}
	s.MustInsertTestRunStarted(expected)

	actual := s.MustGetTestRunStarted()
	assert.Equal(t, expected, actual)
}

func Test_MustGetPickleResult(t *testing.T) {
	s := storage.NewStorage()

	const pickleID = "1"
	expected := models.PickleResult{PickleID: pickleID}
	s.MustInsertPickleResult(expected)

	actual := s.MustGetPickleResult(pickleID)
	assert.Equal(t, expected, actual)
}

func Test_MustGetPickleResults(t *testing.T) {
	s := storage.NewStorage()

	expected := []models.PickleResult{{PickleID: "1"}, {PickleID: "2"}}
	for _, pr := range expected {
		s.MustInsertPickleResult(pr)
	}

	actual := s.MustGetPickleResults()
	assert.Equal(t, expected, actual)
}

func Test_MustGetPickleStepResult(t *testing.T) {
	s := storage.NewStorage()

	const pickleID = "1"
	const stepID = "2"

	expected := models.PickleStepResult{
		Status:       models.Passed,
		PickleID:     pickleID,
		PickleStepID: stepID,
	}
	s.MustInsertPickleStepResult(expected)

	actual := s.MustGetPickleStepResult(stepID)
	assert.Equal(t, expected, actual)
}

func Test_MustGetPickleStepResultsByPickleID(t *testing.T) {
	s := storage.NewStorage()

	const pickleID = "p1"

	expected := []models.PickleStepResult{
		{
			Status:       models.Passed,
			PickleID:     pickleID,
			PickleStepID: "s1",
		},
		{
			Status:       models.Passed,
			PickleID:     pickleID,
			PickleStepID: "s2",
		},
	}

	for _, psr := range expected {
		s.MustInsertPickleStepResult(psr)
	}

	actual := s.MustGetPickleStepResultsByPickleID(pickleID)
	assert.Equal(t, expected, actual)
}

func Test_MustGetPickleStepResultsByStatus(t *testing.T) {
	s := storage.NewStorage()

	const pickleID = "p1"

	expected := []models.PickleStepResult{
		{
			Status:       models.Passed,
			PickleID:     pickleID,
			PickleStepID: "s1",
		},
	}

	testdata := []models.PickleStepResult{
		expected[0],
		{
			Status:       models.Failed,
			PickleID:     pickleID,
			PickleStepID: "s2",
		},
	}

	for _, psr := range testdata {
		s.MustInsertPickleStepResult(psr)
	}

	actual := s.MustGetPickleStepResultsByStatus(models.Passed)
	assert.Equal(t, expected, actual)
}

func Test_MustGetFeature(t *testing.T) {
	s := storage.NewStorage()

	const uri = "<uri>"

	expected := &models.Feature{GherkinDocument: &messages.GherkinDocument{Uri: uri}}
	s.MustInsertFeature(expected)

	actual := s.MustGetFeature(uri)
	assert.Equal(t, expected, actual)
}

func Test_MustGetFeatures(t *testing.T) {
	s := storage.NewStorage()

	expected := []*models.Feature{
		{GherkinDocument: &messages.GherkinDocument{Uri: "uri1"}},
		{GherkinDocument: &messages.GherkinDocument{Uri: "uri2"}},
	}

	for _, f := range expected {
		s.MustInsertFeature(f)
	}

	actual := s.MustGetFeatures()
	assert.Equal(t, expected, actual)
}

func Test_MustGetStepDefintionMatch(t *testing.T) {
	s := storage.NewStorage()

	const stepID = "<step_id>"

	expected := &models.StepDefinition{}
	s.MustInsertStepDefintionMatch(stepID, expected)

	actual := s.MustGetStepDefintionMatch(stepID)
	assert.Equal(t, expected, actual)
}
