package gherkin

import (
	"testing"
)

func (a *AST) assertMatchesTypes(expected []TokenType, t *testing.T) {
	key := -1
	for item := a.head; item != nil; item = item.next {
		key += 1
		if len(expected) <= key {
			t.Fatalf("there are more tokens in AST then expected, next is '%s'", item.value.Type)
		}
		if expected[key] != item.value.Type {
			t.Fatalf("expected ast token '%s', but got '%s' at position: %d", expected[key], item.value.Type, key)
		}
	}
	if len(expected)-1 != key {
		t.Fatalf("expected ast length %d, does not match actual: %d", len(expected), key+1)
	}
}
