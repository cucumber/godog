package godog

import (
	"fmt"
	"go/build"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/cucumber/messages-go/v10"

	"github.com/cucumber/godog/colors"
	"github.com/cucumber/godog/formatters"
	"github.com/cucumber/godog/internal/models"
	"github.com/cucumber/godog/internal/parser"
	"github.com/cucumber/godog/internal/storage"
	"github.com/cucumber/godog/internal/utils"
)

const (
	exitSuccess int = iota
	exitFailure
	exitOptionError
)

type testSuiteInitializer func(*TestSuiteContext)
type scenarioInitializer func(*ScenarioContext)

type runner struct {
	randomSeed            int64
	stopOnFailure, strict bool

	features []*models.Feature

	testSuiteInitializer testSuiteInitializer
	scenarioInitializer  scenarioInitializer

	storage *storage.Storage
	fmt     Formatter
}

func (r *runner) concurrent(rate int) (failed bool) {
	var copyLock sync.Mutex

	if fmt, ok := r.fmt.(storageFormatter); ok {
		fmt.SetStorage(r.storage)
	}

	testSuiteContext := TestSuiteContext{}
	if r.testSuiteInitializer != nil {
		r.testSuiteInitializer(&testSuiteContext)
	}

	testRunStarted := models.TestRunStarted{StartedAt: utils.TimeNowFunc()}
	r.storage.MustInsertTestRunStarted(testRunStarted)
	r.fmt.TestRunStarted()

	// run before suite handlers
	for _, f := range testSuiteContext.beforeSuiteHandlers {
		f()
	}

	queue := make(chan int, rate)
	for _, ft := range r.features {
		pickles := make([]*messages.Pickle, len(ft.Pickles))
		if r.randomSeed != 0 {
			r := rand.New(rand.NewSource(r.randomSeed))
			perm := r.Perm(len(ft.Pickles))
			for i, v := range perm {
				pickles[v] = ft.Pickles[i]
			}
		} else {
			copy(pickles, ft.Pickles)
		}

		for i, p := range pickles {
			pickle := *p

			queue <- i // reserve space in queue

			if i == 0 {
				r.fmt.Feature(ft.GherkinDocument, ft.Uri, ft.Content)
			}

			go func(fail *bool, pickle *messages.Pickle) {
				defer func() {
					<-queue // free a space in queue
				}()

				if r.stopOnFailure && *fail {
					return
				}

				suite := &suite{
					fmt:        r.fmt,
					randomSeed: r.randomSeed,
					strict:     r.strict,
					storage:    r.storage,
				}

				if r.scenarioInitializer != nil {
					sc := ScenarioContext{suite: suite}
					r.scenarioInitializer(&sc)
				}

				err := suite.runPickle(pickle)
				if suite.shouldFail(err) {
					copyLock.Lock()
					*fail = true
					copyLock.Unlock()
				}
			}(&failed, &pickle)
		}
	}

	// wait until last are processed
	for i := 0; i < rate; i++ {
		queue <- i
	}

	close(queue)

	// run after suite handlers
	for _, f := range testSuiteContext.afterSuiteHandlers {
		f()
	}

	// print summary
	r.fmt.Summary()
	return
}

func runWithOptions(suiteName string, runner runner, opt Options) int {
	var output io.Writer = os.Stdout
	if nil != opt.Output {
		output = opt.Output
	}

	if formatterParts := strings.SplitN(opt.Format, ":", 2); len(formatterParts) > 1 {
		f, err := os.Create(formatterParts[1])
		if err != nil {
			err = fmt.Errorf(
				`couldn't create file with name: "%s", error: %s`,
				formatterParts[1], err.Error(),
			)
			fmt.Fprintln(os.Stderr, err)

			return exitOptionError
		}

		defer f.Close()

		output = f
		opt.Format = formatterParts[0]
	}

	if opt.NoColors {
		output = colors.Uncolored(output)
	} else {
		output = colors.Colored(output)
	}

	if opt.ShowStepDefinitions {
		s := suite{}
		sc := ScenarioContext{suite: &s}
		runner.scenarioInitializer(&sc)
		printStepDefinitions(s.steps, output)
		return exitOptionError
	}

	if len(opt.Paths) == 0 {
		inf, err := os.Stat("features")
		if err == nil && inf.IsDir() {
			opt.Paths = []string{"features"}
		}
	}

	if opt.Concurrency < 1 {
		opt.Concurrency = 1
	}

	formatter := formatters.FindFmt(opt.Format)
	if nil == formatter {
		var names []string
		for name := range formatters.AvailableFormatters() {
			names = append(names, name)
		}
		fmt.Fprintln(os.Stderr, fmt.Errorf(
			`unregistered formatter name: "%s", use one of: %s`,
			opt.Format,
			strings.Join(names, ", "),
		))
		return exitOptionError
	}
	runner.fmt = formatter(suiteName, output)

	var err error
	if runner.features, err = parser.ParseFeatures(opt.Tags, opt.Paths); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitOptionError
	}

	runner.storage = storage.NewStorage()
	for _, feat := range runner.features {
		runner.storage.MustInsertFeature(feat)

		for _, pickle := range feat.Pickles {
			runner.storage.MustInsertPickle(pickle)
		}
	}

	// user may have specified -1 option to create random seed
	runner.randomSeed = opt.Randomize
	if runner.randomSeed == -1 {
		runner.randomSeed = makeRandomSeed()
	}

	runner.stopOnFailure = opt.StopOnFailure
	runner.strict = opt.Strict

	// store chosen seed in environment, so it could be seen in formatter summary report
	os.Setenv("GODOG_SEED", strconv.FormatInt(runner.randomSeed, 10))
	// determine tested package
	_, filename, _, _ := runtime.Caller(1)
	os.Setenv("GODOG_TESTED_PACKAGE", runsFromPackage(filename))

	failed := runner.concurrent(opt.Concurrency)

	// @TODO: should prevent from having these
	os.Setenv("GODOG_SEED", "")
	os.Setenv("GODOG_TESTED_PACKAGE", "")
	if failed && opt.Format != "events" {
		return exitFailure
	}
	return exitSuccess
}

func runsFromPackage(fp string) string {
	dir := filepath.Dir(fp)

	gopaths := filepath.SplitList(build.Default.GOPATH)
	for _, gp := range gopaths {
		gp = filepath.Join(gp, "src")
		if strings.Index(dir, gp) == 0 {
			return strings.TrimLeft(strings.Replace(dir, gp, "", 1), string(filepath.Separator))
		}
	}
	return dir
}

// TestSuite allows for configuration
// of the Test Suite Execution
type TestSuite struct {
	Name                 string
	TestSuiteInitializer func(*TestSuiteContext)
	ScenarioInitializer  func(*ScenarioContext)
	Options              *Options
}

// Run will execute the test suite.
//
// If options are not set, it will reads
// all configuration options from flags.
//
// The exit codes may vary from:
//  0 - success
//  1 - failed
//  2 - command line usage error
//  128 - or higher, os signal related error exit codes
//
// If there are flag related errors they will be directed to os.Stderr
func (ts TestSuite) Run() int {
	if ts.Options == nil {
		ts.Options = &Options{}
		ts.Options.Output = colors.Colored(os.Stdout)

		flagSet := flagSet(ts.Options)
		if err := flagSet.Parse(os.Args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return exitOptionError
		}

		ts.Options.Paths = flagSet.Args()
	}

	r := runner{testSuiteInitializer: ts.TestSuiteInitializer, scenarioInitializer: ts.ScenarioInitializer}
	return runWithOptions(ts.Name, r, *ts.Options)
}
