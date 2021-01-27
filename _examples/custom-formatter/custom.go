package main

import (
	"io"

	"github.com/cucumber/godog"
)

func init() {
	godog.Format("emoji", "Progress formatter with emojis", emojiFormatterFunc)
}

func emojiFormatterFunc(suite string, out io.Writer) godog.Formatter {
	return newEmojiFmt(suite, out)
}

func newEmojiFmt(suite string, out io.Writer) *emojiFmt {
	return &emojiFmt{
		Progress: godog.NewProgressfmt(suite, out),
	}
}

type emojiFmt struct {
	*godog.Progress
}

// func (f *customFmt) TestRunStarted()                                                   {}
// func (f *customFmt) Feature(*messages.GherkinDocument, string, []byte)                 {}
// func (f *customFmt) Pickle(*godog.Scenario)                                            {}
// func (f *customFmt) Defined(*godog.Scenario, *godog.Step, *godog.StepDefinition)       {}
// func (f *customFmt) Passed(*godog.Scenario, *godog.Step, *godog.StepDefinition)        {}
// func (f *customFmt) Skipped(*godog.Scenario, *godog.Step, *godog.StepDefinition)       {}
// func (f *customFmt) Undefined(*godog.Scenario, *godog.Step, *godog.StepDefinition)     {}
// func (f *customFmt) Failed(*godog.Scenario, *godog.Step, *godog.StepDefinition, error) {}
// func (f *customFmt) Pending(*godog.Scenario, *godog.Step, *godog.StepDefinition)       {}
// func (f *customFmt) Summary()                                                          {}
