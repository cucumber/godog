package godog

import (
	"testing"

	"github.com/cucumber/godog/internal/flags"
	"github.com/stretchr/testify/assert"
)

func Test_BindFlagsShouldRespectFlagDefaults(t *testing.T) {
	opts := flags.Options{}

	BindCommandLineFlags("flagDefaults.", &opts)

	assert.Equal(t, "pretty", opts.Format)
	assert.Equal(t, "", opts.Tags)
	assert.Equal(t, 1, opts.Concurrency)
	assert.False(t, opts.ShowStepDefinitions)
	assert.False(t, opts.StopOnFailure)
	assert.False(t, opts.Strict)
	assert.False(t, opts.NoColors)
	assert.Equal(t, int64(0), opts.Randomize)
}
