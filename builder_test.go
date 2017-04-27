package godog

import (
	"go/build"
	"os"
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

func TestBuildTestRunner(t *testing.T) {
	bin := filepath.Join(os.TempDir(), "godog.test")
	if err := Build(bin); err != nil {
		t.Fatalf("failed to build godog test binary: %v", err)
	}
	os.Remove(bin)
}

func TestBuildTestRunnerWithoutGoFiles(t *testing.T) {
	bin := filepath.Join(os.TempDir(), "godog.test")
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	wd := filepath.Join(pwd, "features")
	if err := os.Chdir(wd); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	defer func() {
		os.Chdir(pwd) // get back to current dir
	}()

	if err := Build(bin); err != nil {
		t.Fatalf("failed to build godog test binary: %v", err)
	}
	os.Remove(bin)
}
