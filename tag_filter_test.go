package godog

import (
	"testing"
)

func assertNotMatchesTagFilter(tags []string, filter string, t *testing.T) {
	if matchesTags(filter, tags) {
		t.Errorf(`expected tags: %v not to match tag filter "%s", but it did`, tags, filter)
	}
}

func assertMatchesTagFilter(tags []string, filter string, t *testing.T) {
	if !matchesTags(filter, tags) {
		t.Errorf(`expected tags: %v to match tag filter "%s", but it did not`, tags, filter)
	}
}

func TestTagFilter(t *testing.T) {
	assertMatchesTagFilter([]string{"wip"}, "@wip", t)
	assertMatchesTagFilter([]string{}, "~@wip", t)
	assertMatchesTagFilter([]string{"one", "two"}, "@two,@three", t)
	assertMatchesTagFilter([]string{"one", "two"}, "@one&&@two", t)
	assertMatchesTagFilter([]string{"one", "two"}, "one && two", t)

	assertNotMatchesTagFilter([]string{}, "@wip", t)
	assertNotMatchesTagFilter([]string{"one", "two"}, "@one&&~@two", t)
	assertNotMatchesTagFilter([]string{"one", "two"}, "@one && ~@two", t)
}
