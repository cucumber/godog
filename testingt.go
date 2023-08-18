package godog

import (
	"context"
	"fmt"
	"testing"
)

// errFailNow should be returned inside a panic within the test to immediately halt execution of that
// test
var errFailNow = fmt.Errorf("FailNow called")

type dogTestingT struct {
	t      *testing.T
	failed bool
}

// check interface:
var _ TestingT = &dogTestingT{}

func (dt *dogTestingT) Log(args ...interface{}) {
	if dt.t != nil {
		dt.t.Log(args...)
		return
	}
	fmt.Println(args...)
}

func (dt *dogTestingT) Logf(format string, args ...interface{}) {
	if dt.t != nil {
		dt.t.Logf(format, args...)
		return
	}
	fmt.Printf(format+"\n", args...)
}

func (dt *dogTestingT) Errorf(format string, args ...interface{}) {
	dt.Logf(format, args...)
	dt.failed = true
}

func (dt *dogTestingT) Fail() {
	dt.failed = true
}

func (dt *dogTestingT) FailNow() {
	panic(errFailNow)
}

func (dt *dogTestingT) check() error {
	if dt.failed {
		return fmt.Errorf("one or more checks failed")
	}

	return nil
}

type testingTCtxVal struct{}

func setContextDogTester(ctx context.Context, dt *dogTestingT) context.Context {
	return context.WithValue(ctx, testingTCtxVal{}, dt)
}
func getDogTestingT(ctx context.Context) *dogTestingT {
	dt, ok := ctx.Value(testingTCtxVal{}).(*dogTestingT)
	if !ok {
		return nil
	}
	return dt
}

// GetTestingT returns a TestingT compatible interface from the current test context. It will return
// nil if called outside the context of a test. This can be used with (for example) testify's assert
// and require packages.
func GetTestingT(ctx context.Context) TestingT {
	return getDogTestingT(ctx)
}
