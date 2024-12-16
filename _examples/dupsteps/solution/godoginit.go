package solution

import (
	"flag"

	"github.com/cucumber/godog"
)

func init() {
	// allow user overrides of preferred godog defaults via command-line flags
	godog.BindFlags("godog.", flag.CommandLine, &defaultOpts)
}

// holds preferred godog defaults to be used by the test case(s)
var defaultOpts = godog.Options{
	Strict: true,
	Format: "cucumber",
}
