package formatters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlushWrapOnFormatter(t *testing.T) {
	mock.tt = t

	fmt := WrapOnFlush(&mock)

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

	assert.Equal(t, 0, mock.CountFeature)
	assert.Equal(t, 0, mock.CountTestRunStarted)
	assert.Equal(t, 0, mock.CountPickle)
	assert.Equal(t, 0, mock.CountDefined)
	assert.Equal(t, 0, mock.CountPassed)
	assert.Equal(t, 0, mock.CountSkipped)
	assert.Equal(t, 0, mock.CountUndefined)
	assert.Equal(t, 0, mock.CountFailed)
	assert.Equal(t, 0, mock.CountPending)
	assert.Equal(t, 0, mock.CountAmbiguous)

	fmt.Flush()

	assert.Equal(t, 1, mock.CountFeature)
	assert.Equal(t, 1, mock.CountTestRunStarted)
	assert.Equal(t, 1, mock.CountPickle)
	assert.Equal(t, 1, mock.CountDefined)
	assert.Equal(t, 1, mock.CountPassed)
	assert.Equal(t, 1, mock.CountSkipped)
	assert.Equal(t, 1, mock.CountUndefined)
	assert.Equal(t, 1, mock.CountFailed)
	assert.Equal(t, 1, mock.CountPending)
	assert.Equal(t, 1, mock.CountAmbiguous)
}
