package godog

import (
	"testing"

	"github.com/cucumber/messages-go/v10"
)

func assertNotMatchesTagFilter(tags []*tag, filter string, t *testing.T) {
	if matchesTags(filter, tags) {
		t.Errorf(`expected tags: %v not to match tag filter "%s", but it did`, tags, filter)
	}
}

func assertMatchesTagFilter(tags []*tag, filter string, t *testing.T) {
	if !matchesTags(filter, tags) {
		t.Errorf(`expected tags: %v to match tag filter "%s", but it did not`, tags, filter)
	}
}

func TestTagFilter(t *testing.T) {
	assertMatchesTagFilter([]*tag{{Name: "wip"}}, "@wip", t)
	assertMatchesTagFilter([]*tag{}, "~@wip", t)
	assertMatchesTagFilter([]*tag{{Name: "one"}, {Name: "two"}}, "@two,@three", t)
	assertMatchesTagFilter([]*tag{{Name: "one"}, {Name: "two"}}, "@one&&@two", t)
	assertMatchesTagFilter([]*tag{{Name: "one"}, {Name: "two"}}, "one && two", t)

	assertNotMatchesTagFilter([]*tag{}, "@wip", t)
	assertNotMatchesTagFilter([]*tag{{Name: "one"}, {Name: "two"}}, "@one&&~@two", t)
	assertNotMatchesTagFilter([]*tag{{Name: "one"}, {Name: "two"}}, "@one && ~@two", t)
}

type tag = messages.Pickle_PickleTag
