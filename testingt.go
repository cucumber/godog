package godog

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

type TestingT interface {
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fail()
	FailNow()
}

// GetTestingT returns a TestingT compatible interface from the current test context. It will return
// nil if called outside the context of a test. This can be used with (for example) testify's assert
// and require packages.
func GetTestingT(ctx context.Context) TestingT {
	return getDogTestingT(ctx)
}

// errFailNow should be returned inside a panic within the test to immediately halt execution of that
// test
var errFailNow = fmt.Errorf("FailNow called")

type dogTestingT struct {
	t            *testing.T
	failed       bool
	failMessages []string
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
	dt.failMessages = append(dt.failMessages, fmt.Sprintf(format, args...))
	dt.Fail()
}

func (dt *dogTestingT) Fail() {
	dt.failed = true
}

func (dt *dogTestingT) FailNow() {
	dt.Fail()
	panic(errFailNow)
}

// isFailed will return an error representing the calls to Fail made during this test
func (dt *dogTestingT) isFailed() error {
	if !dt.failed {
		return nil
	}
	switch len(dt.failMessages) {
	case 0:
		return fmt.Errorf("fail called on TestingT")
	case 1:
		return fmt.Errorf(dt.failMessages[0])
	default:
		return fmt.Errorf("checks failed:\n* %s", strings.Join(dt.failMessages, "\n* "))
	}
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
