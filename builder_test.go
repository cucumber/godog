package godog

import (
	"go/build"
	"path/filepath"
	"reflect"
	"testing"
)

func TestVendorPaths(t *testing.T) {
	gopaths = []string{"/go"}

	type Case struct {
		dir    string
		expect []string
	}

	cases := []Case{
		{"/go", []string{}},
		{"/go/src", []string{}},
		{"/go/src/project", []string{"/go/src/project/vendor"}},
		{"/go/src/party/project", []string{"/go/src/party/project/vendor", "/go/src/party/vendor"}},
	}

	for i, c := range cases {
		actual := maybeVendorPaths(c.dir)
		var expect []string
		for _, s := range c.expect {
			expect = append(expect, filepath.Join(s, godogImportPath))
		}
		if !reflect.DeepEqual(expect, actual) {
			t.Fatalf("case %d expected %+v, got %+v", i, expect, actual)
		}
	}

	gopaths = filepath.SplitList(build.Default.GOPATH)
}
