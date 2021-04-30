package storage

import (
	"github.com/cucumber/messages-go/v10"

	"github.com/cucumber/godog/internal/v1.0-api-draft/storage"
)

// InternalStorage is an interface for storing and retrieving data regarding the test execution.
type InternalStorage interface {
	storage.Storage

	// MustInsertTestRunStartedEvent will set the test run started event and panic on error.
	MustInsertTestRunStartedEvent(storage.TestRunStartedEvent)

	// MustInsertFeature will insert a feature and panic on error.
	MustInsertFeature(storage.Feature)

	// MustInsertPickle will insert a pickle and it's steps, will panic on error.
	MustInsertPickle(messages.Pickle)

	// MustInsertPickleStartedEvent will instert a pickle started event and panic on error.
	MustInsertPickleStartedEvent(storage.PickleStartedEvent)

	// MustInsertPickleStepResultEvent will insert a pickle step result and panic on error.
	MustInsertPickleStepResultEvent(storage.PickleStepResultEvent)

	// MustInsertStepDefintionMatch will insert the matched StepDefintion for the step ID and panic on error.
	MustInsertStepDefintionMatch(pickleStepID string, match storage.StepDefinition)
}
