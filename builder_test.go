package godog

import (
	"bytes"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var builderFeatureFile = `Feature: eat godogs
  In order to be happy
  As a hungry gopher
  I need to be able to eat godogs

  Scenario: Eat 5 out of 12
    Given there are 12 godogs
    When I eat 5
    Then there should be 7 remaining
`

var builderTestFile = `package godogs

import (
	"fmt"

	"github.com/DATA-DOG/godog"
)

func thereAreGodogs(available int) error {
	Godogs = available
	return nil
}

func iEat(num int) error {
	if Godogs < num {
		return fmt.Errorf("you cannot eat %d godogs, there are %d available", num, Godogs)
	}
	Godogs -= num
	return nil
}

func thereShouldBeRemaining(remaining int) error {
	if Godogs != remaining {
		return fmt.Errorf("expected %d godogs to be remaining, but there is %d", remaining, Godogs)
	}
	return nil
}

func FeatureContext(s *godog.Suite) {
	s.Step("^there are (\\d+) godogs$", thereAreGodogs)
	s.Step("^I eat (\\d+)$", iEat)
	s.Step("^there should be (\\d+) remaining$", thereShouldBeRemaining)

	s.BeforeScenario(func(interface{}) {
		Godogs = 0 // clean the state before every scenario
	})
}
`

var builderXTestFile = `package godogs_test

import (
	"fmt"

	"github.com/DATA-DOG/godog"
	"godogs"
)

func thereAreGodogs(available int) error {
	godogs.Godogs = available
	return nil
}

func iEat(num int) error {
	if godogs.Godogs < num {
		return fmt.Errorf("you cannot eat %d godogs, there are %d available", num, godogs.Godogs)
	}
	godogs.Godogs -= num
	return nil
}

func thereShouldBeRemaining(remaining int) error {
	if godogs.Godogs != remaining {
		return fmt.Errorf("expected %d godogs to be remaining, but there is %d", remaining, godogs.Godogs)
	}
	return nil
}

func FeatureContext(s *godog.Suite) {
	s.Step("^there are (\\d+) godogs$", thereAreGodogs)
	s.Step("^I eat (\\d+)$", iEat)
	s.Step("^there should be (\\d+) remaining$", thereShouldBeRemaining)

	s.BeforeScenario(func(interface{}) {
		godogs.Godogs = 0 // clean the state before every scenario
	})
}
`

var builderMainCodeFile = `package godogs

// Godogs available to eat
var Godogs int

func main() {
}
`

func buildTestPackage(dir string, files map[string]string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	for name, content := range files {
		if err := ioutil.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			return err
		}
	}

	return nil
}

func buildTestCommand(t *testing.T, args ...string) *exec.Cmd {
	bin, err := filepath.Abs("godog.test")
	if err != nil {
		t.Fatal(err)
	}
	if build.Default.GOOS == "windows" {
		bin += ".exe"
	}
	if err = Build(bin); err != nil {
		t.Fatal(err)
	}

	return exec.Command(bin, args...)
}

func TestGodogBuildWithSourceNotInGoPath(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "godogs")
	err := buildTestPackage(dir, map[string]string{
		"godogs.feature": builderFeatureFile,
		"godogs.go":      builderMainCodeFile,
		"godogs_test.go": builderTestFile,
	})
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

	cmd := buildTestCommand(t, "godogs.feature")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Log(stdout.String())
		t.Log(stderr.String())
		t.Fatal(err)
	}
}

func TestGodogBuildWithoutSourceNotInGoPath(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "godogs")
	err := buildTestPackage(dir, map[string]string{
		"godogs.feature": builderFeatureFile,
	})
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

	cmd := buildTestCommand(t, "godogs.feature")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Log(stdout.String())
		t.Log(stderr.String())
		t.Fatal(err)
	}
}

func TestGodogBuildWithoutTestSourceNotInGoPath(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "godogs")
	err := buildTestPackage(dir, map[string]string{
		"godogs.feature": builderFeatureFile,
		"godogs.go":      builderMainCodeFile,
	})
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

	cmd := buildTestCommand(t, "godogs.feature")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Log(stdout.String())
		t.Log(stderr.String())
		t.Fatal(err)
	}
}

func TestGodogBuildWithinGopath(t *testing.T) {
	gopath := filepath.Join(os.TempDir(), "_gp")
	dir := filepath.Join(gopath, "src", "godogs")
	err := buildTestPackage(dir, map[string]string{
		"godogs.feature": builderFeatureFile,
		"godogs.go":      builderMainCodeFile,
		"godogs_test.go": builderTestFile,
	})
	if err != nil {
		os.RemoveAll(gopath)
		t.Fatal(err)
	}
	defer os.RemoveAll(gopath)

	pkg := filepath.Join(gopath, "src", "github.com", "DATA-DOG")
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
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOPATH="+gopath)

	if err := cmd.Run(); err != nil {
		t.Log(stdout.String())
		t.Log(stderr.String())
		t.Fatal(err)
	}
}

func TestGodogBuildWithVendoredGodog(t *testing.T) {
	gopath := filepath.Join(os.TempDir(), "_gp")
	dir := filepath.Join(gopath, "src", "godogs")
	err := buildTestPackage(dir, map[string]string{
		"godogs.feature": builderFeatureFile,
		"godogs.go":      builderMainCodeFile,
		"godogs_test.go": builderTestFile,
	})
	if err != nil {
		os.RemoveAll(gopath)
		t.Fatal(err)
	}
	defer os.RemoveAll(gopath)

	pkg := filepath.Join(dir, "vendor", "github.com", "DATA-DOG")
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
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GOPATH="+gopath)

	if err := cmd.Run(); err != nil {
		t.Log(stdout.String())
		t.Log(stderr.String())
		t.Fatal(err)
	}
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
		_ = os.Chdir(pwd) // get back to current dir
	}()

	if err := Build(bin); err != nil {
		t.Fatalf("failed to build godog test binary: %v", err)
	}
	os.Remove(bin)
}
