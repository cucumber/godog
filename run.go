package godog

import (
	"fmt"
	"os"
	"os/exec"
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
func Run(suite string, contextInitializer func(suite *Suite)) int {
	var defs, sof, noclr bool
	var tags, format string
	var concurrency int
	flagSet := FlagSet(&format, &tags, &defs, &sof, &noclr, &concurrency)
	err := flagSet.Parse(os.Args[1:])
	fatal(err)

	if defs {
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

	var formatter FormatterFunc
	// cunicorn formatter is installed
	_, err = exec.LookPath("cunicorn")
	var output *os.File
	var cunicorn *exec.Cmd
	runWithCunicorn := err == nil && format == "progress"
	if runWithCunicorn {
		formatter, err = findFmt("events")
		fatal(err)
		reader, writer, err := os.Pipe()
		fatal(err)

		output = writer
		cunicorn = exec.Command("cunicorn", "-f", format)
		cunicorn.Stdin = reader
		cunicorn.Stdout = os.Stdout
		cunicorn.Stderr = os.Stderr
		err = cunicorn.Start()
		fatal(err)
	} else {
		formatter, err = findFmt(format)
		output = os.Stdout
		fatal(err)
	}

	features, err := parseFeatures(tags, paths)
	fatal(err)

	r := runner{
		fmt:           formatter(suite, output),
		initializer:   contextInitializer,
		features:      features,
		stopOnFailure: sof,
	}

	var failed bool
	if concurrency > 1 {
		failed = r.concurrent(concurrency)
	} else {
		failed = r.run()
	}

	if runWithCunicorn {
		output.Close() // need to close output in case if it was a pipe
		err = cunicorn.Wait()
		fatal(err)
	}

	if failed && format != "events" {
		return 1
	}
	return 0
}
