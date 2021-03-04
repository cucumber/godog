package formatters

import (
	"io"

	"github.com/cucumber/godog/internal/v1.0/storage"
)

// FindFmt searches available formatters registered
// and returns FormaterFunc matched by given
// format name or nil otherwise
func FindFmt(name string) FormatterFunc

// Format registers a feature suite output
// formatter by given name, description and
// FormatterFunc constructor function, to initialize
// formatter with the output recorder.
func Format(name, description string, f FormatterFunc)

// AvailableFormatters gives a map of all
// formatters registered with their name as key
// and description as value
func AvailableFormatters() map[string]string

// Formatter is an interface for feature runner
// output summary presentation.
//
// New formatters may be created to represent
// suite results in different ways. These new
// formatters needs to be registered with a
// godog.Format function call
type Formatter interface {
	SetStorage(storage.Storage)
	TestRunStarted()
	Feature(featureURI string)
	Pickle(pickleID string)
	Defined(pickleStepID string)
	Failed(pickleStepID string, err error)
	Passed(pickleStepID string)
	Skipped(pickleStepID string)
	Undefined(pickleStepID string)
	Pending(pickleStepID string)
	Summary()
}

// FormatterFunc builds a formatter with given
// suite name and io.Writer to record output
type FormatterFunc func(string, io.Writer) Formatter
