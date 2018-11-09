// +build go1.11

package godog

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGodogBuildWithModuleOutsideGopath(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "godogs")
	err := buildTestPackage(dir, builderFeatureFile, builderMainCodeFile, builderTestFile)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	prevDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(prevDir)

	if out, err := exec.Command("go", "mod", "init", "godogs").CombinedOutput(); err != nil {
		t.Log(string(out))
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("godog", "godogs.feature")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		t.Log(stdout.String())
		t.Log(stderr.String())
		t.Fatal(err)
	}
}

func TestGodogBuildWithModuleInsideGopath(t *testing.T) {
	gopath := filepath.Join(os.TempDir(), "_gp")
	dir := filepath.Join(gopath, "src", "godogs")
	err := buildTestPackage(dir, builderFeatureFile, builderMainCodeFile, builderTestFile)
	if err != nil {
		os.RemoveAll(gopath)
		t.Fatal(err)
	}
	defer os.RemoveAll(gopath)

	prevDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(prevDir)

	c := exec.Command("go", "mod", "init", "godogs")
	c.Env = os.Environ()
	c.Env = append(c.Env, "GOPATH="+gopath)
	c.Env = append(c.Env, "GO111MODULE=on")
	if out, err := c.CombinedOutput(); err != nil {
		t.Log(string(out))
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("godog", "godogs.feature")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOPATH="+gopath)
	cmd.Env = append(cmd.Env, "GO111MODULE=on")

	if err := cmd.Run(); err != nil {
		t.Log(stdout.String())
		t.Log(stderr.String())
		t.Fatal(err)
	}
}
