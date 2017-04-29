package godog

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/DATA-DOG/godog/colors"
)

type initializer func(*Suite)

type runner struct {
	randomSeed    int64
	stopOnFailure bool
	features      []*feature
	fmt           Formatter
	initializer   initializer
}

func (r *runner) concurrent(rate int) (failed bool) {
	queue := make(chan int, rate)
	for i, ft := range r.features {
		queue <- i // reserve space in queue
		go func(fail *bool, feat *feature) {
			defer func() {
				<-queue // free a space in queue
			}()
			if r.stopOnFailure && *fail {
				return
			}
			suite := &Suite{
				fmt:           r.fmt,
				randomSeed:    r.randomSeed,
				stopOnFailure: r.stopOnFailure,
				features:      []*feature{feat},
			}
			r.initializer(suite)
			suite.run()
			if suite.failed {
				*fail = true
			}
		}(&failed, ft)
	}
	// wait until last are processed
	for i := 0; i < rate; i++ {
		queue <- i
	}
	close(queue)

	// print summary
	r.fmt.Summary()
	return
}

func (r *runner) run() bool {
	suite := &Suite{
		fmt:           r.fmt,
		randomSeed:    r.randomSeed,
		stopOnFailure: r.stopOnFailure,
		features:      r.features,
	}
	r.initializer(suite)
	suite.run()

	r.fmt.Summary()
	return suite.failed
}

// RunWithOptions is same as Run function, except
// it uses Options provided in order to run the
// test suite without parsing flags
//
// This method is useful in case if you run
// godog in for example TestMain function together
// with go tests
func RunWithOptions(suite string, contextInitializer func(suite *Suite), opt Options) int {
	var output io.Writer = os.Stdout
	if nil != opt.Output {
		output = opt.Output
	}

	if opt.NoColors {
		output = colors.Uncolored(output)
	} else {
		output = colors.Colored(output)
	}

	if opt.ShowStepDefinitions {
		s := &Suite{}
		contextInitializer(s)
		s.printStepDefinitions(output)
		return 0
	}

	if len(opt.Paths) == 0 {
		inf, err := os.Stat("features")
		if err == nil && inf.IsDir() {
			opt.Paths = []string{"features"}
		}
	}

	if opt.Concurrency > 1 && !supportsConcurrency(opt.Format) {
		fatal(fmt.Errorf("format \"%s\" does not support concurrent execution", opt.Format))
	}
	formatter, err := findFmt(opt.Format)
	fatal(err)

	features, err := parseFeatures(opt.Tags, opt.Paths)
	fatal(err)

	r := runner{
		fmt:           formatter(suite, output),
		initializer:   contextInitializer,
		features:      features,
		randomSeed:    opt.Randomize,
		stopOnFailure: opt.StopOnFailure,
	}

	// store chosen seed in environment, so it could be seen in formatter summary report
	os.Setenv("GODOG_SEED", strconv.FormatInt(r.randomSeed, 10))

	var failed bool
	if opt.Concurrency > 1 {
		failed = r.concurrent(opt.Concurrency)
	} else {
		failed = r.run()
	}
	if failed && opt.Format != "events" {
		return 1
	}
	return 0
}

// Run creates and runs the feature suite.
// Reads all configuration options from flags.
// uses contextInitializer to register contexts
//
// the concurrency option allows runner to
// initialize a number of suites to be run
// separately. Only progress formatter
// is supported when concurrency level is
// higher than 1
//
// contextInitializer must be able to register
// the step definitions and event handlers.
func Run(suite string, contextInitializer func(suite *Suite)) int {
	var opt Options
	opt.Output = colors.Colored(os.Stdout)
	flagSet := FlagSet(&opt)
	err := flagSet.Parse(os.Args[1:])
	fatal(err)

	opt.Paths = flagSet.Args()

	return RunWithOptions(suite, contextInitializer, opt)
}

func supportsConcurrency(format string) bool {
	switch format {
	case "events":
	case "junit":
	case "pretty":
	case "cucumber":
	default:
		return true // supports concurrency
	}

	return true // all custom formatters are treated as supporting concurrency
}
