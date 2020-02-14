package godog

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/cucumber/messages-go/v9"
)

func init() {
	Format("junit", "Prints junit compatible xml to stdout", junitFunc)
}

func junitFunc(suite string, out io.Writer) Formatter {
	return &junitFormatter{
		basefmt: basefmt{
			suiteName: suite,
			started:   timeNowFunc(),
			indent:    2,
			out:       out,
		},
		lock: new(sync.Mutex),
	}
}

type junitFormatter struct {
	basefmt
	lock *sync.Mutex
}

func (f *junitFormatter) Pickle(pickle *messages.Pickle) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.basefmt.Pickle(pickle)
}

func (f *junitFormatter) Feature(ft *messages.GherkinDocument, p string, c []byte) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.basefmt.Feature(ft, p, c)
}

func (f *junitFormatter) Summary() {
	suite := buildJUNITPackageSuite(f.suiteName, f.started, f.features)

	_, err := io.WriteString(f.out, xml.Header)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to write junit string:", err)
	}

	enc := xml.NewEncoder(f.out)
	enc.Indent("", s(2))
	if err = enc.Encode(suite); err != nil {
		fmt.Fprintln(os.Stderr, "failed to write junit xml:", err)
	}
}

func (f *junitFormatter) Passed(pickle *messages.Pickle, step *messages.Pickle_PickleStep, match *StepDefinition) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.basefmt.Passed(pickle, step, match)
}

func (f *junitFormatter) Skipped(pickle *messages.Pickle, step *messages.Pickle_PickleStep, match *StepDefinition) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.basefmt.Skipped(pickle, step, match)
}

func (f *junitFormatter) Undefined(pickle *messages.Pickle, step *messages.Pickle_PickleStep, match *StepDefinition) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.basefmt.Undefined(pickle, step, match)
}

func (f *junitFormatter) Failed(pickle *messages.Pickle, step *messages.Pickle_PickleStep, match *StepDefinition, err error) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.basefmt.Failed(pickle, step, match, err)
}

func (f *junitFormatter) Pending(pickle *messages.Pickle, step *messages.Pickle_PickleStep, match *StepDefinition) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.basefmt.Pending(pickle, step, match)
}

func (f *junitFormatter) Sync(cf ConcurrentFormatter) {
	if source, ok := cf.(*junitFormatter); ok {
		f.lock = source.lock
	}
}

func (f *junitFormatter) Copy(cf ConcurrentFormatter) {
	if source, ok := cf.(*junitFormatter); ok {
		for _, v := range source.features {
			f.features = append(f.features, v)
		}
	}
}

func buildJUNITPackageSuite(suiteName string, startedAt time.Time, features []*feature) junitPackageSuite {
	suite := junitPackageSuite{
		Name:       suiteName,
		TestSuites: make([]*junitTestSuite, len(features)),
		Time:       timeNowFunc().Sub(startedAt).String(),
	}

	sort.Sort(sortByName(features))

	for idx, feat := range features {
		ts := junitTestSuite{
			Name:      feat.GherkinDocument.Feature.Name,
			Time:      feat.finishedAt().Sub(feat.startedAt()).String(),
			TestCases: make([]*junitTestCase, len(feat.pickleResults)),
		}

		var testcaseNames = make(map[string]int)
		for _, pickleResult := range feat.pickleResults {
			testcaseNames[pickleResult.Name] = testcaseNames[pickleResult.Name] + 1
		}

		var outlineNo = make(map[string]int)
		for idx, pickleResult := range feat.pickleResults {
			tc := junitTestCase{}
			tc.Time = pickleResult.finishedAt().Sub(pickleResult.startedAt()).String()

			tc.Name = pickleResult.Name
			if testcaseNames[tc.Name] > 1 {
				outlineNo[tc.Name] = outlineNo[tc.Name] + 1
				tc.Name += fmt.Sprintf(" #%d", outlineNo[tc.Name])
			}

			ts.Tests++
			suite.Tests++

			for _, stepResult := range pickleResult.stepResults {
				switch stepResult.status {
				case passed:
					tc.Status = passed.String()
				case failed:
					tc.Status = failed.String()
					tc.Failure = &junitFailure{
						Message: fmt.Sprintf("Step %s: %s", stepResult.step.Text, stepResult.err),
					}
				case skipped:
					tc.Error = append(tc.Error, &junitError{
						Type:    "skipped",
						Message: fmt.Sprintf("Step %s", stepResult.step.Text),
					})
				case undefined:
					tc.Status = undefined.String()
					tc.Error = append(tc.Error, &junitError{
						Type:    "undefined",
						Message: fmt.Sprintf("Step %s", stepResult.step.Text),
					})
				case pending:
					tc.Status = pending.String()
					tc.Error = append(tc.Error, &junitError{
						Type:    "pending",
						Message: fmt.Sprintf("Step %s: TODO: write pending definition", stepResult.step.Text),
					})
				}
			}

			switch tc.Status {
			case failed.String():
				ts.Failures++
				suite.Failures++
			case undefined.String(), pending.String():
				ts.Errors++
				suite.Errors++
			}

			ts.TestCases[idx] = &tc
		}

		suite.TestSuites[idx] = &ts
	}

	return suite
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr,omitempty"`
}

type junitError struct {
	XMLName xml.Name `xml:"error,omitempty"`
	Message string   `xml:"message,attr"`
	Type    string   `xml:"type,attr"`
}

type junitTestCase struct {
	XMLName xml.Name      `xml:"testcase"`
	Name    string        `xml:"name,attr"`
	Status  string        `xml:"status,attr"`
	Time    string        `xml:"time,attr"`
	Failure *junitFailure `xml:"failure,omitempty"`
	Error   []*junitError
}

type junitTestSuite struct {
	XMLName   xml.Name `xml:"testsuite"`
	Name      string   `xml:"name,attr"`
	Tests     int      `xml:"tests,attr"`
	Skipped   int      `xml:"skipped,attr"`
	Failures  int      `xml:"failures,attr"`
	Errors    int      `xml:"errors,attr"`
	Time      string   `xml:"time,attr"`
	TestCases []*junitTestCase
}

type junitPackageSuite struct {
	XMLName    xml.Name `xml:"testsuites"`
	Name       string   `xml:"name,attr"`
	Tests      int      `xml:"tests,attr"`
	Skipped    int      `xml:"skipped,attr"`
	Failures   int      `xml:"failures,attr"`
	Errors     int      `xml:"errors,attr"`
	Time       string   `xml:"time,attr"`
	TestSuites []*junitTestSuite
}
