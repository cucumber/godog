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

	actual = trimLineSpaces(fmt.Sprintf("%+v", err))
	if strings.Index(actual, "stacktrace_test.go") == -1 {
		t.Fatalf("does not have stacktrace in actual: %s", actual)
	}
}
