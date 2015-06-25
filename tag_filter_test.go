package godog

import (
	"testing"

	"github.com/DATA-DOG/godog/gherkin"
)

func assertNotMatchesTagFilter(tags []string, filter string, t *testing.T) {
	gtags := gherkin.Tags{}
	for _, tag := range tags {
		gtags = append(gtags, gherkin.Tag(tag))
	}
	s := &suite{tags: filter}
	if s.matchesTags(gtags) {
		t.Errorf(`expected tags: %v not to match tag filter "%s", but it did`, gtags, filter)
	}
}

func assertMatchesTagFilter(tags []string, filter string, t *testing.T) {
	gtags := gherkin.Tags{}
	for _, tag := range tags {
		gtags = append(gtags, gherkin.Tag(tag))
	}
	s := &suite{tags: filter}
	if !s.matchesTags(gtags) {
		t.Errorf(`expected tags: %v to match tag filter "%s", but it did not`, gtags, filter)
	}
}

func Test_tag_filter(t *testing.T) {
	assertMatchesTagFilter([]string{"wip"}, "@wip", t)
	assertMatchesTagFilter([]string{}, "~@wip", t)
	assertMatchesTagFilter([]string{"one", "two"}, "@two,@three", t)
	assertMatchesTagFilter([]string{"one", "two"}, "@one&&@two", t)
	assertMatchesTagFilter([]string{"one", "two"}, "one && two", t)

	assertNotMatchesTagFilter([]string{}, "@wip", t)
	assertNotMatchesTagFilter([]string{"one", "two"}, "@one&&~@two", t)
	assertNotMatchesTagFilter([]string{"one", "two"}, "@one && ~@two", t)
}
