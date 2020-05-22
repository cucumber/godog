// +build go1.13

package godog_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGodogBuildWithVendoredGodogAndMod(t *testing.T) {
	gopath := filepath.Join(os.TempDir(), "_gpc")
	dir := filepath.Join(gopath, "src", "godogs")
	err := buildTestPackage(dir, map[string]string{
		"godogs.feature": builderFeatureFile,
		"godogs.go":      builderMainCodeFile,
		"godogs_test.go": builderTestFile,
		"go.mod":         builderModFile,
	})
	if err != nil {
		os.RemoveAll(gopath)
		t.Fatal(err)
	}
	defer os.RemoveAll(gopath)

	prevDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	if err = exec.Command("go", "mod", "vendor").Run(); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(prevDir)

	cmd := buildTestCommand(t, "godogs.feature")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = append(envVarsWithoutGopath(), "GOPATH="+gopath)

	if err := cmd.Run(); err != nil {
		t.Log(stdout.String())
		t.Log(stderr.String())
		t.Fatal(err)
	}
}
