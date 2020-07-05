package models_test

import (
	"testing"

	"github.com/cucumber/godog/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func Test_Find(t *testing.T) {
	ft := testutils.BuildTestFeature(t)

	t.Run("scenario", func(t *testing.T) {
		sc := ft.FindScenario(ft.Pickles[0].AstNodeIds[0])
		assert.NotNilf(t, sc, "expected scenario to not be nil")
	})

	t.Run("background", func(t *testing.T) {
		bg := ft.FindBackground(ft.Pickles[0].AstNodeIds[0])
		assert.NotNilf(t, bg, "expected background to not be nil")
	})

	t.Run("example", func(t *testing.T) {
		example, row := ft.FindExample(ft.Pickles[1].AstNodeIds[1])
		assert.NotNilf(t, example, "expected example to not be nil")
		assert.NotNilf(t, row, "expected table row to not be nil")
	})

	t.Run("step", func(t *testing.T) {
		for _, ps := range ft.Pickles[0].Steps {
			step := ft.FindStep(ps.AstNodeIds[0])
			assert.NotNilf(t, step, "expected step to not be nil")
		}
	})
}

func Test_NotFind(t *testing.T) {
	ft := testutils.BuildTestFeature(t)

	t.Run("scenario", func(t *testing.T) {
		sc := ft.FindScenario("-")
		assert.Nilf(t, sc, "expected scenario to be nil")
	})

	t.Run("background", func(t *testing.T) {
		bg := ft.FindBackground("-")
		assert.Nilf(t, bg, "expected background to be nil")
	})

	t.Run("example", func(t *testing.T) {
		example, row := ft.FindExample("-")
		assert.Nilf(t, example, "expected example to be nil")
		assert.Nilf(t, row, "expected table row to be nil")
	})

	t.Run("step", func(t *testing.T) {
		step := ft.FindStep("-")
		assert.Nilf(t, step, "expected step to be nil")
	})
}
