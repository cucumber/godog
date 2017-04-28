package godog

import (
	"fmt"
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

func TestStacktrace(t *testing.T) {
	err := &traceError{
		msg:   "err msg",
		stack: callStack(),
	}

	expect := "err msg"
	actual := fmt.Sprintf("%s", err)
	if expect != actual {
		t.Fatalf("expected formatted trace error message to be: %s, but got %s", expect, actual)
	}

	expect = trimLineSpaces(`err msg
testing.tRunner
	/usr/lib/go/src/testing/testing.go:657
runtime.goexit
	/usr/lib/go/src/runtime/asm_amd64.s:2197`)

	actual = trimLineSpaces(fmt.Sprintf("%+v", err))
	if expect != actual {
		t.Fatalf("detaily formatted actual: %s", actual)
	}
}
