package godog

import (
	"bytes"
	"context"
	"strings"
	"testing"

	gherkin "github.com/cucumber/gherkin/go/v26"
	messages "github.com/cucumber/messages/go/v21"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cucumber/godog/colors"
	"github.com/cucumber/godog/internal/formatters"
	"github.com/cucumber/godog/internal/models"
	"github.com/cucumber/godog/internal/storage"
)

const oneStepFeature = `
Feature: basic

  Scenario: passing scenario
	When one
`

// Test_AfterStepHookPanic covers cucumber/godog#662: a panic in an after-step
// hook used to escape runStep's deferred recover (which had already finished),
// so the panic killed the worker goroutine and the run died with a stack
// trace. The expected behaviour matches the before-step side: convert the
// panic into a hook error attached to the failing step.
func Test_AfterStepHookPanic(t *testing.T) {
	const path = "any.feature"

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(oneStepFeature), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)
	gd.Uri = path
	ft := models.Feature{GherkinDocument: gd}
	ft.Pickles = gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	r := runner{
		fmt:      formatters.ProgressFormatterFunc("progress", w),
		features: []*models.Feature{&ft},
		scenarioInitializer: func(ctx *ScenarioContext) {
			ctx.Step(`^one$`, func() error { return nil })
			ctx.StepContext().After(func(ctx context.Context, st *Step, status StepResultStatus, err error) (context.Context, error) {
				panic("boom from after-step hook")
			})
		},
	}

	r.storage = storage.NewStorage()
	r.storage.MustInsertFeature(&ft)
	for _, pickle := range ft.Pickles {
		r.storage.MustInsertPickle(pickle)
	}

	failed := r.concurrent(1)
	require.True(t, failed, "panicking after-step hook should fail the run")
	assert.Contains(t, buf.String(), "after step hook panicked")
}

// Test_AfterScenarioHookPanic covers the same shape on the scenario hook.
func Test_AfterScenarioHookPanic(t *testing.T) {
	const path = "any.feature"

	gd, err := gherkin.ParseGherkinDocument(strings.NewReader(oneStepFeature), (&messages.Incrementing{}).NewId)
	require.NoError(t, err)
	gd.Uri = path
	ft := models.Feature{GherkinDocument: gd}
	ft.Pickles = gherkin.Pickles(*gd, path, (&messages.Incrementing{}).NewId)

	var buf bytes.Buffer
	w := colors.Uncolored(&buf)
	r := runner{
		fmt:      formatters.ProgressFormatterFunc("progress", w),
		features: []*models.Feature{&ft},
		scenarioInitializer: func(ctx *ScenarioContext) {
			ctx.Step(`^one$`, func() error { return nil })
			ctx.After(func(ctx context.Context, sc *Scenario, err error) (context.Context, error) {
				panic("boom from after-scenario hook")
			})
		},
	}

	r.storage = storage.NewStorage()
	r.storage.MustInsertFeature(&ft)
	for _, pickle := range ft.Pickles {
		r.storage.MustInsertPickle(pickle)
	}

	failed := r.concurrent(1)
	require.True(t, failed, "panicking after-scenario hook should fail the run")
	assert.Contains(t, buf.String(), "after scenario hook panicked")
}
