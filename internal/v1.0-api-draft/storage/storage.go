package storage

import (
	"github.com/cucumber/messages-go/v10"
)

// Storage is an interface for retrieving data regarding the test execution.
type Storage interface {

	// MustGetTestRunStartedEvent will retrieve the test run started event and panic on error.
	MustGetTestRunStartedEvent() TestRunStartedEvent

	// MustGetFeature will retrieve a feature by URI and panic on error.
	MustGetFeature(featureURI string) Feature

	// MustGetFeatures will retrieve all features by and panic on error.
	MustGetFeatures() []Feature

	// MustGetPickle will retrieve a pickle by id and panic on error.
	MustGetPickle(pickleID string) messages.Pickle

	// MustGetPickleByPickleStepID will retrieve a pickle by the id of a pickle step and panic on error.
	MustGetPickleByPickleStepID(pickleStepID string) messages.Pickle

	// MustGetPickles will retrieve pickles by Feature URI and panic on error.
	MustGetPickles(featureURI string) []messages.Pickle

	// MustGetPickleStep will retrieve a pickle step and panic on error.
	MustGetPickleStep(pickleStepID string) messages.Pickle_PickleStep

	// MustGetPickleStartedEvent will retrieve a pickle started event by id and panic on error.
	MustGetPickleStartedEvent(pickleID string) PickleStartedEvent

	// MustGetPickleStartedEvents will retrieve all pickle started event and panic on error.
	MustGetPickleStartedEvents() []PickleStartedEvent

	// MustGetPickleStepResultEvent will retrieve a pickle strep result by id and panic on error.
	MustGetPickleStepResultEvent(pickleStepID string) PickleStepResultEvent

	// MustGetPickleStepResultEventsByPickleID will retrieve pickle strep results by pickle id and panic on error.
	MustGetPickleStepResultEventsByPickleID(pickleID string) []PickleStepResultEvent

	// MustGetPickleStepResultEventsByStatus will retrieve pickle strep results by status and panic on error.
	MustGetPickleStepResultEventsByStatus(PickleStepResultStatus) []PickleStepResultEvent

	// MustGetStepDefintionMatch will retrieve the matched StepDefintion for the step ID and panic on error.
	MustGetStepDefintionMatch(pickleStepID string) StepDefinition
}

// InMemStorage ...
type InMemStorage struct{}

// NewInMemStorage will create an in-mem storage that is used across concurrent runners and formatters
func NewInMemStorage() *InMemStorage { return &InMemStorage{} }
