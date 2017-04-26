package godog

import (
	"io"
	"math/rand"
	"strconv"
	"time"
)

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

	// Randomize, if not `0`, will be used to run scenarios in a random order.
	//
	// Randomizing scenario order is especially helpful for detecting
	// situations where you have state leaking between scenarios, which can
	// cause flickering or fragile tests.
	//
	// The default value of `0` means "do not randomize".
	//
	// The magic value of `-1` means "pick a random seed for me", and godog will
	// assign a seed on it's own during the `RunWithOptions` phase, similar to if
	// you specified `--random` on the command line.
	//
	// Any other value will be used as the random seed for shuffling. Re-using the
	// same seed will allow you to reproduce the shuffle order of a previous run
	// to isolate an error condition.
	Randomize randomSeed

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

// randomSeed implements `flag.Value`, see https://golang.org/pkg/flag/#Value
type randomSeed int64

// Choose randomly assigns a convenient pseudo-random seed value.
// The resulting seed will be between `1-99999` for later ease of specification.
func (rs *randomSeed) Choose() {
	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	*rs = randomSeed(r.Int63n(99998) + 1)
}

func (rs *randomSeed) Set(s string) error {
	if s == "true" {
		rs.Choose()
		return nil
	}

	if s == "false" {
		*rs = 0
		return nil
	}

	i, err := strconv.ParseInt(s, 10, 64)
	*rs = randomSeed(i)
	return err
}

func (rs randomSeed) String() string {
	return strconv.FormatInt(int64(rs), 10)
}

// If a Value has an IsBoolFlag() bool method returning true, the command-line
// parser makes -name equivalent to -name=true rather than using the next
// command-line argument.
func (rs *randomSeed) IsBoolFlag() bool {
	return *rs == 0
}
