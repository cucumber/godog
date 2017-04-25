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

	// RandomSeed, if not `0`, will be used to run scenarios in a random order.
	//
	// Randomizing scenario order is especially helpful for detecting
	// situations where you have state leaking between scenarios, which can
	// cause flickering or fragile tests.
	//
	// The default value of `0` means "do not randomize".
	//
	// The magic value of `-1` means "pick a random seed for me", the resulting
	// seed will only be between `1-99999` for ease of specification.
	RandomSeed int64

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
