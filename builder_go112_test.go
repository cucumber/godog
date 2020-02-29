// +build go1.12
// +build !go1.13

package godog

import (
	"bytes"
	"os"
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

	pkg := filepath.Join(dir, "vendor", "github.com", "cucumber")
	if err := os.MkdirAll(pkg, 0755); err != nil {
		t.Fatal(err)
	}

	prevDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	// symlink godog package
	if err := os.Symlink(prevDir, filepath.Join(pkg, "godog")); err != nil {
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
