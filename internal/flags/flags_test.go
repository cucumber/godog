package flags_test

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"

	"github.com/cucumber/godog/internal/flags"
)

func Test_BindFlagsShouldRespectFlagDefaults(t *testing.T) {
	opts := flags.Options{}
	flagSet := pflag.FlagSet{}

	flags.BindRunCmdFlags("optDefaults.", &flagSet, &opts)

	flagSet.Parse([]string{})

	assert.Equal(t, "pretty", opts.Format)
	assert.Equal(t, "", opts.Tags)
	assert.Equal(t, 1, opts.Concurrency)
	assert.False(t, opts.ShowStepDefinitions)
	assert.False(t, opts.StopOnFailure)
	assert.False(t, opts.Strict)
	assert.False(t, opts.NoColors)
	assert.Equal(t, int64(0), opts.Randomize)
}

func Test_BindFlagsShouldRespectFlagOverrides(t *testing.T) {
	opts := flags.Options{
		Format:              "progress",
		Tags:                "test",
		Concurrency:         2,
		ShowStepDefinitions: true,
		StopOnFailure:       true,
		Strict:              true,
		NoColors:            true,
		Randomize:           11,
	}
	flagSet := pflag.FlagSet{}

	flags.BindRunCmdFlags("optOverrides.", &flagSet, &opts)

	flagSet.Parse([]string{
		"--optOverrides.format=junit",
		"--optOverrides.tags=test2",
		"--optOverrides.concurrency=3",
		"--optOverrides.definitions=false",
		"--optOverrides.stop-on-failure=false",
		"--optOverrides.strict=false",
		"--optOverrides.no-colors=false",
		"--optOverrides.random=2",
	})

	assert.Equal(t, "junit", opts.Format)
	assert.Equal(t, "test2", opts.Tags)
	assert.Equal(t, 3, opts.Concurrency)
	assert.False(t, opts.ShowStepDefinitions)
	assert.False(t, opts.StopOnFailure)
	assert.False(t, opts.Strict)
	assert.False(t, opts.NoColors)
	assert.Equal(t, int64(2), opts.Randomize)
}
