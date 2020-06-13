package godog

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cucumber/gherkin-go/v11"
	"github.com/cucumber/godog/colors"
	"github.com/cucumber/messages-go/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_TestContext(t *testing.T) {
	const path = "any.feature"

	var buf bytes.Buffer
	w := colors.Uncolored(&buf)

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(basicGherkinFeature), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)

	pickles := gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	r := runner{
		fmt:                  progressFunc("progress", w),
		features:             []*feature{{GherkinDocument: gd, pickles: pickles}},
		testSuiteInitializer: nil,
		scenarioInitializer: func(sc *ScenarioContext) {
			sc.Step(`^one$`, func() error { return nil })
			sc.Step(`^two$`, func() error { return nil })
		},
	}

	r.storage = newStorage()
	for _, pickle := range pickles {
		r.storage.mustInsertPickle(pickle)
	}

	failed := r.concurrent(1, func() Formatter { return progressFunc("progress", w) })
	require.False(t, failed)

	expected := `.. 2


1 scenarios (1 passed)
2 steps (2 passed)
0s
`

	actual := buf.String()
	assert.Equal(t, expected, actual)
}
