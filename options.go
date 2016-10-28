package godog

// Options are suite run options
// flags are mapped to these options.
//
// It can also be used together with godog.RunWithOptions
// to run test suite from go source directly
//
// See the flags for more details
type Options struct {
	ShowStepDefinitions bool
	StopOnFailure       bool
	NoColors            bool
	Tags                string
	Format              string
	Concurrency         int
	Paths               []string
}
