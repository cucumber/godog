package godog

import (
	"context"
	"flag"
	"fmt"
	"go/build"
	"io"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"

	messages "github.com/cucumber/messages/go/v21"

	"github.com/cucumber/godog/colors"
	"github.com/cucumber/godog/formatters"
	ifmt "github.com/cucumber/godog/internal/formatters"
	"github.com/cucumber/godog/internal/models"
	"github.com/cucumber/godog/internal/parser"
	"github.com/cucumber/godog/internal/storage"
	"github.com/cucumber/godog/internal/utils"
)

const (
	ExitSuccess = iota
	ExitFailure
	ExitOptionError
)

type testSuiteInitializer func(*TestSuiteContext)
type scenarioInitializer func(*ScenarioContext)

type runner struct {
	randomSeed            int64
	stopOnFailure, strict bool

	defaultContext context.Context
	testingT       *testing.T

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

	testSuiteContext := TestSuiteContext{
		suite: &suite{
			fmt:            r.fmt,
			randomSeed:     r.randomSeed,
			strict:         r.strict,
			storage:        r.storage,
			defaultContext: r.defaultContext,
			testingT:       r.testingT,
		},
	}
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

			runPickle := func(fail *bool, pickle *messages.Pickle) {
				defer func() {
					<-queue // free a space in queue
				}()

				if r.stopOnFailure && *fail {
					return
				}

				// Copy base suite.
				suite := *testSuiteContext.suite

				if r.scenarioInitializer != nil {
					sc := ScenarioContext{suite: &suite}
					r.scenarioInitializer(&sc)
				}

				err := suite.runPickle(pickle)

				if suite.shouldFail(err) {
					copyLock.Lock()
					*fail = true
					copyLock.Unlock()
				}
			}

			if rate == 1 {
				// Running within the same goroutine for concurrency 1
				// to preserve original stacks and simplify debugging.
				runPickle(&failed, &pickle)
			} else {
				go runPickle(&failed, &pickle)
			}
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

func configureFormatter(opt Options, suiteName string, output io.WriteCloser) (Formatter, error) {
	multiFmt, err := configureMultiFormatter(opt, output)
	if err != nil {
		return nil, err
	}

	return multiFmt.FormatterFunc(suiteName, output), nil
}

func configureMultiFormatter(opt Options, output io.WriteCloser) (multiFmt ifmt.MultiFormatter, err error) {

	for _, formatter := range strings.Split(opt.Format, ",") {
		out := output
		formatterParts := strings.SplitN(formatter, ":", 2)

		if len(formatterParts) > 1 {
			f, err := os.Create(formatterParts[1])
			if err != nil {
				err = fmt.Errorf(
					`couldn't create file with name: "%s", error: %s`,
					formatterParts[1], err.Error(),
				)
				return ifmt.MultiFormatter{}, err
			}

			out = f
		}

		if opt.NoColors {
			out = colors.Uncolored(out)
		} else {
			out = colors.Colored(out)
		}

		if nil == formatters.FindFmt(formatterParts[0]) {
			var names []string
			for name := range formatters.AvailableFormatters() {
				names = append(names, name)
			}
			err := fmt.Errorf(
				`unregistered formatter name: "%s", use one of: %s`,
				opt.Format,
				strings.Join(names, ", "),
			)
			return ifmt.MultiFormatter{}, err
		}

		multiFmt.Add(formatterParts[0], out)
	}
	return multiFmt, nil
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
	Options              *Options // TODO mutable value - is this necessary?
}

// Run will execute the test suite.
//
// If options are not set, it will reads
// all configuration options from flags.
//
// The exit codes may vary from:
//
//	0 - success
//	1 - failed
//	2 - command line usage error
//	128 - or higher, os signal related error exit codes
//
// If there are flag related errors they will be directed to os.Stderr
func (ts TestSuite) Run() int {
	result := ts.RunWithResult()
	return result.exitCode
}

func (ts TestSuite) RunWithResult() RunResult {
	if ts.Options == nil {
		var err error
		ts.Options, err = getDefaultOptions()
		if err != nil {
			return RunResult{ExitOptionError, nil}
		}
	}
	if ts.Options.FS == nil {
		ts.Options.FS = storage.FS{}
	}
	if ts.Options.ShowHelp {
		flag.CommandLine.Usage()

		return RunResult{0, nil}
	}

	runner := runner{
		testSuiteInitializer: ts.TestSuiteInitializer,
		scenarioInitializer:  ts.ScenarioInitializer,
	}

	var output io.WriteCloser = NopCloser(os.Stdout)
	if nil != ts.Options.Output {
		output = ts.Options.Output
	}

	if ts.Options.ShowStepDefinitions {
		s := suite{}
		sc := ScenarioContext{suite: &s}
		runner.scenarioInitializer(&sc)
		printStepDefinitions(s.steps, output)
		return RunResult{ExitOptionError, nil}
	}

	if len(ts.Options.Paths) == 0 && len(ts.Options.FeatureContents) == 0 {
		inf, err := func() (fs.FileInfo, error) {
			file, err := ts.Options.FS.Open("features")
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return nil, err
			}
			defer file.Close()

			return file.Stat()
		}()
		if err == nil && inf.IsDir() {
			ts.Options.Paths = []string{"features"}
		}
	}

	if ts.Options.Concurrency < 1 {
		ts.Options.Concurrency = 1
	}

	var err error
	runner.fmt, err = configureFormatter(*ts.Options, ts.Name, output)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return RunResult{ExitOptionError, nil}
	}
	defer func() {
		runner.fmt.Close()
	}()

	ts.Options.FS = storage.FS{FS: ts.Options.FS}

	if len(ts.Options.FeatureContents) > 0 {
		features, err := parser.ParseFromBytes(ts.Options.Tags, ts.Options.FeatureContents)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Options.FeatureContents contains an error: %s\n", err.Error())
			return RunResult{ExitOptionError, nil}
		}
		runner.features = append(runner.features, features...)
	}

	if len(ts.Options.Paths) > 0 {
		features, err := parser.ParseFeatures(ts.Options.FS, ts.Options.Tags, ts.Options.Paths)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return RunResult{ExitOptionError, nil}
		}
		runner.features = append(runner.features, features...)
	}

	runner.storage = storage.NewStorage()
	for _, feat := range runner.features {
		runner.storage.MustInsertFeature(feat)

		for _, pickle := range feat.Pickles {
			runner.storage.MustInsertPickle(pickle)
		}
	}

	// user may have specified -1 option to create random seed
	runner.randomSeed = ts.Options.Randomize
	if runner.randomSeed == -1 {
		runner.randomSeed = makeRandomSeed()
	}

	// TOD - move all these up to the initializer at top of func
	runner.stopOnFailure = ts.Options.StopOnFailure
	runner.strict = ts.Options.Strict
	runner.defaultContext = ts.Options.DefaultContext
	runner.testingT = ts.Options.TestingT

	// TODO using env vars to pass args to formatter instead of traditional arg passing seems less that ideal
	// store chosen seed in environment, so it could be seen in formatter summary report
	os.Setenv("GODOG_SEED", strconv.FormatInt(runner.randomSeed, 10))
	// determine tested package
	_, filename, _, _ := runtime.Caller(1)
	os.Setenv("GODOG_TESTED_PACKAGE", runsFromPackage(filename))

	failed := runner.concurrent(ts.Options.Concurrency)

	// @TODO: should prevent from having these
	os.Setenv("GODOG_SEED", "")
	os.Setenv("GODOG_TESTED_PACKAGE", "")
	if failed && ts.Options.Format != "events" {
		return RunResult{ExitFailure, runner.storage}

	}
	return RunResult{ExitSuccess, runner.storage}
}

// RetrieveFeatures will parse and return the features based on test suite option
// Any modification on the parsed features will not have any impact on the next Run of the Test Suite
func (ts TestSuite) RetrieveFeatures() ([]*models.Feature, error) {
	opt := ts.Options

	if opt == nil {
		var err error
		opt, err = getDefaultOptions()
		if err != nil {
			return nil, err
		}
	}

	if ts.Options.FS == nil {
		ts.Options.FS = storage.FS{}
	}

	if len(opt.Paths) == 0 {
		inf, err := func() (fs.FileInfo, error) {
			file, err := opt.FS.Open("features")
			if err != nil {
				return nil, err
			}
			defer file.Close()

			return file.Stat()
		}()
		if err == nil && inf.IsDir() {
			opt.Paths = []string{"features"}
		}
	}

	return parser.ParseFeatures(opt.FS, opt.Tags, opt.Paths)
}

func getDefaultOptions() (*Options, error) {
	opt := &Options{}
	opt.Output = colors.Colored(os.Stdout)

	flagSet := flagSet(opt)
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}

	opt.Paths = flagSet.Args()
	opt.FS = storage.FS{}

	return opt, nil
}

type noopCloser struct {
	out io.Writer
}

func (*noopCloser) Close() error { return nil }

func (n *noopCloser) Write(p []byte) (int, error) {
	return n.out.Write(p)
}

// NopCloser will return an io.WriteCloser that ignores Close() calls
func NopCloser(file io.Writer) io.WriteCloser {
	return &noopCloser{out: file}
}

type RunResult struct {
	exitCode int
	storage  *storage.Storage
}

func (r RunResult) ExitCode() int {
	return r.exitCode
}

func (r RunResult) Storage() *storage.Storage {
	return r.storage
}
