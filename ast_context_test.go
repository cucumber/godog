package godog

import (
	"go/parser"
	"go/token"
	"testing"
)

var astContextSrc = `package main

import (
	"github.com/DATA-DOG/godog"
)

func myContext(s *godog.Suite) {
}`

var astTwoContextSrc = `package lib

import (
	"github.com/DATA-DOG/godog"
)

func apiContext(s *godog.Suite) {
}

func dbContext(s *godog.Suite) {
}`

func astContexts(src string, t *testing.T) []string {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", []byte(src), 0)
	if err != nil {
		t.Fatalf("unexpected error while parsing ast: %v", err)
	}

	return contexts(f)
}

func TestShouldGetSingleContextFromSource(t *testing.T) {
	actual := astContexts(astContextSrc, t)
	expect := []string{"myContext"}

	if len(actual) != len(expect) {
		t.Fatalf("number of found contexts do not match, expected %d, but got %d", len(expect), len(actual))
	}

	for i, c := range expect {
		if c != actual[i] {
			t.Fatalf("expected context '%s' at pos %d, but got: '%s'", c, i, actual[i])
		}
	}
}

func TestShouldGetTwoContextsFromSource(t *testing.T) {
	actual := astContexts(astTwoContextSrc, t)
	expect := []string{"apiContext", "dbContext"}

	if len(actual) != len(expect) {
		t.Fatalf("number of found contexts do not match, expected %d, but got %d", len(expect), len(actual))
	}

	for i, c := range expect {
		if c != actual[i] {
			t.Fatalf("expected context '%s' at pos %d, but got: '%s'", c, i, actual[i])
		}
	}
}

func TestShouldNotFindAnyContextsInEmptyFile(t *testing.T) {
	actual := astContexts(`package main`, t)

	if len(actual) != 0 {
		t.Fatalf("expected no contexts to be found, but there was some: %v", actual)
	}
}
