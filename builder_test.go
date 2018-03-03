package godog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDeps(t *testing.T) {
	t.Log("hh")

	abs, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}

	// we allow package to be nil, if godog is run only when
	// there is a feature file in empty directory
	pkg := importPackage(abs)
	deps := make(map[string]string)
	err = dependencies(pkg, deps)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(deps)
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
