package godog

import (
	"os"
	"path/filepath"
	"testing"
)

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
