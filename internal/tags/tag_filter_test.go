package tags_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cucumber/godog/internal/tags"
	messages "github.com/cucumber/messages/go/v31"
)

type tag = messages.PickleTag
type pickle = messages.Pickle

type testcase struct {
	filter   string
	expected []*pickle
}

var testdata = []*pickle{p1, p2, p3}
var p1 = &pickle{Id: "one", Tags: []*tag{{Name: "@one"}, {Name: "@wip"}}}
var p2 = &pickle{Id: "two", Tags: []*tag{{Name: "@two"}, {Name: "@wip"}}}
var p3 = &pickle{Id: "three", Tags: []*tag{{Name: "@hree"}, {Name: "@wip"}}}

var testcases = []testcase{
	{filter: "", expected: testdata},

	{filter: "@one", expected: []*pickle{p1}},
	{filter: "~@one", expected: []*pickle{p2, p3}},
	{filter: "one", expected: []*pickle{p1}},
	{filter: " one ", expected: []*pickle{p1}},

	{filter: "@one,@two", expected: []*pickle{p1, p2}},
	{filter: "@one,~@two", expected: []*pickle{p1, p3}},
	{filter: " @one , @two ", expected: []*pickle{p1, p2}},

	{filter: "@one&&@two", expected: []*pickle{}},
	{filter: "@one&&~@two", expected: []*pickle{p1}},
	{filter: "@one&&@wip", expected: []*pickle{p1}},

	{filter: "@one&&@two,@wip", expected: []*pickle{p1}},
}

func Test_ApplyTagFilter(t *testing.T) {
	for _, tc := range testcases {
		t.Run("", func(t *testing.T) {
			actual := tags.ApplyTagFilter(tc.filter, testdata)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
