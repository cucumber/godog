package godog

import (
	"fmt"
	"os"
)

type initializer func(*Suite)

type runner struct {
	stopOnFailure bool
	features      []*feature
	fmt           Formatter // needs to support concurrency
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

func (r *runner) run() (failed bool) {
	suite := &Suite{
		fmt:           r.fmt,
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
	if opt.ShowStepDefinitions {
		s := &Suite{}
		contextInitializer(s)
		s.printStepDefinitions()
		return 0
	}

	if len(opt.Paths) == 0 {
		inf, err := os.Stat("features")
		if err == nil && inf.IsDir() {
			opt.Paths = []string{"features"}
		}
	}

	if opt.Concurrency > 1 && opt.Format != "progress" {
		fatal(fmt.Errorf("when concurrency level is higher than 1, only progress format is supported"))
	}
	formatter, err := findFmt(opt.Format)
	fatal(err)

	features, err := parseFeatures(opt.Tags, opt.Paths)
	fatal(err)

	r := runner{
		fmt:           formatter(suite, os.Stdout),
		initializer:   contextInitializer,
		features:      features,
		stopOnFailure: opt.StopOnFailure,
	}

	var failed bool
	if opt.Concurrency > 1 {
		failed = r.concurrent(opt.Concurrency)
	} else {
		failed = r.run()
	}
	if failed && format != "events" {
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
	flagSet := FlagSet(
		&opt.Format,
		&opt.Tags,
		&opt.ShowStepDefinitions,
		&opt.StopOnFailure,
		&opt.NoColors,
		&opt.Concurrency,
	)
	err := flagSet.Parse(os.Args[1:])
	fatal(err)
	opt.Paths = flagSet.Args()

	return RunWithOptions(suite, contextInitializer, opt)
}
