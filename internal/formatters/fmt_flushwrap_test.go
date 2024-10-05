package formatters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var flushMock = DummyFormatter{}

func TestFlushWrapOnFormatter(t *testing.T) {
	flushMock.tt = t

	fmt := WrapOnFlush(&flushMock)

	fmt.Feature(document, str, byt)
	fmt.TestRunStarted()
	fmt.Pickle(pickle)
	fmt.Defined(pickle, step, definition)
	fmt.Passed(pickle, step, definition)
	fmt.Skipped(pickle, step, definition)
	fmt.Undefined(pickle, step, definition)
	fmt.Failed(pickle, step, definition, err)
	fmt.Pending(pickle, step, definition)
	fmt.Ambiguous(pickle, step, definition, err)

	assert.Equal(t, 0, flushMock.CountFeature)
	assert.Equal(t, 0, flushMock.CountTestRunStarted)
	assert.Equal(t, 0, flushMock.CountPickle)
	assert.Equal(t, 0, flushMock.CountDefined)
	assert.Equal(t, 0, flushMock.CountPassed)
	assert.Equal(t, 0, flushMock.CountSkipped)
	assert.Equal(t, 0, flushMock.CountUndefined)
	assert.Equal(t, 0, flushMock.CountFailed)
	assert.Equal(t, 0, flushMock.CountPending)
	assert.Equal(t, 0, flushMock.CountAmbiguous)

	fmt.Flush()

	assert.Equal(t, 1, flushMock.CountFeature)
	assert.Equal(t, 1, flushMock.CountTestRunStarted)
	assert.Equal(t, 1, flushMock.CountPickle)
	assert.Equal(t, 1, flushMock.CountDefined)
	assert.Equal(t, 1, flushMock.CountPassed)
	assert.Equal(t, 1, flushMock.CountSkipped)
	assert.Equal(t, 1, flushMock.CountUndefined)
	assert.Equal(t, 1, flushMock.CountFailed)
	assert.Equal(t, 1, flushMock.CountPending)
	assert.Equal(t, 1, flushMock.CountAmbiguous)
}
