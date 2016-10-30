package godog

import "io"

// Options are suite run options
// flags are mapped to these options.
//
// It can also be used together with godog.RunWithOptions
// to run test suite from go source directly
//
// See the flags for more details
type Options struct {
	// Print step definitions found and exit
	ShowStepDefinitions bool

	// Stops on the first failure
	StopOnFailure bool

	// Forces ansi color stripping
	NoColors bool

	// Various filters for scenarios parsed
	// from feature files
	Tags string

	// The formatter name
	Format string

	// Concurrency rate, not all formatters accepts this
	Concurrency int

	// All feature file paths
	Paths []string

	// Where it should print formatter output
	Output io.Writer
}
