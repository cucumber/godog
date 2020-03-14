package godog

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	expected := "err msg"
	actual := fmt.Sprintf("%s", err)

	assert.Equal(t, expected, actual)
	assert.NotContains(t, actual, "stacktrace_test.go")
}
