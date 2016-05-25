package godog

import (
	"fmt"
	"os"
	"sync"
)

type initializer func(*Suite)

type runner struct {
	sync.WaitGroup

	semaphore   chan int
	features    []*feature
	fmt         Formatter // needs to support concurrency
	initializer initializer
}

func (r *runner) run() (failed bool) {
	r.Add(len(r.features))
	for _, ft := range r.features {
		go func(fail *bool, feat *feature) {
			r.semaphore <- 1
			suite := &Suite{
				fmt:      r.fmt,
				features: []*feature{feat},
			}
			r.initializer(suite)
			suite.run()
			if suite.failed {
				*fail = true
			}
			<-r.semaphore
			r.Done()
		}(&failed, ft)
	}
	r.Wait()

	r.fmt.Summary()
	return
}

// Run creates and runs the feature suite.
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
func Run(contextInitializer func(suite *Suite)) int {
	var vers, defs, sof bool
	var tags, format string
	var concurrency int
	flagSet := flags(&format, &tags, &defs, &sof, &vers, &concurrency)
	err := flagSet.Parse(os.Args[1:])
	fatal(err)

	switch {
	case vers:
		fmt.Println(cl("Godog", green) + " version is " + cl(Version, yellow))
		return 0
	case defs:
		s := &Suite{}
		contextInitializer(s)
		s.printStepDefinitions()
		return 0
	}

	paths := flagSet.Args()
	if len(paths) == 0 {
		inf, err := os.Stat("features")
		if err == nil && inf.IsDir() {
			paths = []string{"features"}
		}
	}

	if concurrency > 1 && format != "progress" {
		fatal(fmt.Errorf("when concurrency level is higher than 1, only progress format is supported"))
	}
	if concurrency > 1 && sof {
		fatal(fmt.Errorf("when concurrency level is higher than 1, cannot stop on first failure for now"))
	}
	formatter, err := findFmt(format)
	fatal(err)

	features, err := parseFeatures(tags, paths)
	fatal(err)

	r := runner{
		fmt:         formatter,
		initializer: contextInitializer,
		semaphore:   make(chan int, concurrency),
		features:    features,
	}

	if failed := r.run(); failed {
		return 1
	}
	return 0
}
