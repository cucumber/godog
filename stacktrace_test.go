package godog

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func trimLineSpaces(s string) string {
	var res []string
	for _, ln := range strings.Split(s, "\n") {
		res = append(res, strings.TrimSpace(ln))
	}
	return strings.Join(res, "\n")
}

func callstack1() *stack {
	return callstack2()
}

func callstack2() *stack {
	return callstack3()
}

func callstack3() *stack {
	const depth = 4
	var pcs [depth]uintptr
	n := runtime.Callers(1, pcs[:])
	var st stack = pcs[0:n]
	return &st
}

func TestStacktrace(t *testing.T) {
	err := &traceError{
		msg:   "err msg",
		stack: callstack1(),
	}

	expect := "err msg"
	actual := fmt.Sprintf("%s", err)
	if expect != actual {
		t.Fatalf("expected formatted trace error message to be: %s, but got %s", expect, actual)
	}

	expect = trimLineSpaces(`err msg
github.com/DATA-DOG/godog.callstack3
github.com/DATA-DOG/godog/stacktrace_test.go:29
github.com/DATA-DOG/godog.callstack2
github.com/DATA-DOG/godog/stacktrace_test.go:23
github.com/DATA-DOG/godog.callstack1
github.com/DATA-DOG/godog/stacktrace_test.go:19
github.com/DATA-DOG/godog.TestStacktrace
github.com/DATA-DOG/godog/stacktrace_test.go:37`)

	actual = trimLineSpaces(fmt.Sprintf("%+v", err))
	if expect != actual {
		t.Fatalf("detaily formatted actual: %s", actual)
	}
}
