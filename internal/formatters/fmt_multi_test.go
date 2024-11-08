package formatters

import (
	"errors"
	"testing"

	"github.com/cucumber/godog/formatters"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/stretchr/testify/assert"
)

var (
	mock = DummyFormatter{}
	base = BaseFormatter{}

	document   = &messages.GherkinDocument{}
	str        = "theString"
	byt        = []byte("bytes")
	pickle     = &messages.Pickle{}
	step       = &messages.PickleStep{}
	definition = &formatters.StepDefinition{}
	err        = errors.New("expected")
)

// TestRepeater tests the delegation of the repeater functions.
func TestRepeater(t *testing.T) {
	mock.tt = t
	f := make(repeater, 0)
	f = append(f, &mock)
	f = append(f, &mock)
	f = append(f, &base)

	f.Feature(document, str, byt)
	f.TestRunStarted()
	f.Pickle(pickle)
	f.Defined(pickle, step, definition)
	f.Passed(pickle, step, definition)
	f.Skipped(pickle, step, definition)
	f.Undefined(pickle, step, definition)
	f.Failed(pickle, step, definition, err)
	f.Pending(pickle, step, definition)
	f.Ambiguous(pickle, step, definition, err)

	assert.Equal(t, 2, mock.CountFeature)
	assert.Equal(t, 2, mock.CountTestRunStarted)
	assert.Equal(t, 2, mock.CountPickle)
	assert.Equal(t, 2, mock.CountDefined)
	assert.Equal(t, 2, mock.CountPassed)
	assert.Equal(t, 2, mock.CountSkipped)
	assert.Equal(t, 2, mock.CountUndefined)
	assert.Equal(t, 2, mock.CountFailed)
	assert.Equal(t, 2, mock.CountPending)
	assert.Equal(t, 2, mock.CountAmbiguous)
}

type BaseFormatter struct {
	*Base
}

type DummyFormatter struct {
	*Base

	tt                  *testing.T
	CountFeature        int
	CountTestRunStarted int
	CountPickle         int
	CountDefined        int
	CountPassed         int
	CountSkipped        int
	CountUndefined      int
	CountFailed         int
	CountPending        int
	CountAmbiguous      int
	CountSummary        int
}

// SetStorage assigns gherkin data storage.
// func (f *DummyFormatter) SetStorage(st *storage.Storage) {
// }

// TestRunStarted is triggered on test start.
func (f *DummyFormatter) TestRunStarted() {
	f.CountTestRunStarted++
}

// Feature receives gherkin document.
func (f *DummyFormatter) Feature(d *messages.GherkinDocument, s string, b []byte) {
	assert.Equal(f.tt, document, d)
	assert.Equal(f.tt, str, s)
	assert.Equal(f.tt, byt, b)
	f.CountFeature++
}

// Pickle receives scenario.
func (f *DummyFormatter) Pickle(p *messages.Pickle) {
	assert.Equal(f.tt, pickle, p)
	f.CountPickle++
}

// Defined receives step definition.
func (f *DummyFormatter) Defined(p *messages.Pickle, s *messages.PickleStep, d *formatters.StepDefinition) {
	assert.Equal(f.tt, pickle, p)
	assert.Equal(f.tt, s, step)
	assert.Equal(f.tt, d, definition)
	f.CountDefined++
}

// Passed captures passed step.
func (f *DummyFormatter) Passed(p *messages.Pickle, s *messages.PickleStep, d *formatters.StepDefinition) {
	assert.Equal(f.tt, pickle, p)
	assert.Equal(f.tt, s, step)
	assert.Equal(f.tt, d, definition)
	f.CountPassed++
}

// Skipped captures skipped step.
func (f *DummyFormatter) Skipped(p *messages.Pickle, s *messages.PickleStep, d *formatters.StepDefinition) {
	assert.Equal(f.tt, pickle, p)
	assert.Equal(f.tt, s, step)
	assert.Equal(f.tt, d, definition)

	f.CountSkipped++
}

// Undefined captures undefined step.
func (f *DummyFormatter) Undefined(p *messages.Pickle, s *messages.PickleStep, d *formatters.StepDefinition) {
	assert.Equal(f.tt, pickle, p)
	assert.Equal(f.tt, s, step)
	assert.Equal(f.tt, d, definition)

	f.CountUndefined++
}

// Failed captures failed step.
func (f *DummyFormatter) Failed(p *messages.Pickle, s *messages.PickleStep, d *formatters.StepDefinition, e error) {
	assert.Equal(f.tt, pickle, p)
	assert.Equal(f.tt, s, step)
	assert.Equal(f.tt, d, definition)
	assert.Equal(f.tt, err, e)

	f.CountFailed++
}

// Pending captures pending step.
func (f *DummyFormatter) Pending(p *messages.Pickle, s *messages.PickleStep, d *formatters.StepDefinition) {
	assert.Equal(f.tt, pickle, p)
	assert.Equal(f.tt, s, step)
	assert.Equal(f.tt, d, definition)

	f.CountPending++
}

// Ambiguous captures ambiguous step.
func (f *DummyFormatter) Ambiguous(p *messages.Pickle, s *messages.PickleStep, d *formatters.StepDefinition, e error) {
	assert.Equal(f.tt, pickle, p)
	assert.Equal(f.tt, s, step)
	assert.Equal(f.tt, d, definition)
	f.CountAmbiguous++
}

// Pickle receives scenario.
func (f *DummyFormatter) Summary() {
	f.CountSummary++
}
