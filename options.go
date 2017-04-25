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

	// Randomize causes scenarios to be run in random order.
	//
	// Randomizing scenario order is especially helpful for detecting
	// situations where you have state leaking between scenarios, which can
	// cause flickering or fragile tests.
	Randomize bool

	// RandomSeed allows specifying the seed to reproduce the random scenario
	// shuffling from a previous run.
	//
	// When `RandomSeed` is left at the nil value (`0`), but `Randomize`
	// has been set to `true`, then godog will automatically pick a random
	// seed between `1-99999` for ease of specification.
	//
	// If RandomSeed is set to anything other than the default nil value (`0`),
	// then `Randomize = true` will be implied.
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
