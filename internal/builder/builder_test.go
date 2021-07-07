package builder_test

import (
	"bytes"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/internal/builder"
)

func InitializeScenario(ctx *godog.ScenarioContext) {}

func Test_GodogBuild(t *testing.T) {
	t.Run("WithSourceNotInGoPath", testWithSourceNotInGoPath)
	t.Run("WithoutSourceNotInGoPath", testWithoutSourceNotInGoPath)
	t.Run("WithoutTestSourceNotInGoPath", testWithoutTestSourceNotInGoPath)
	t.Run("WithinGopath", testWithinGopath)
	t.Run("WithVendoredGodogWithoutModule", testWithVendoredGodogWithoutModule)
	t.Run("WithVendoredGodogAndMod", testWithVendoredGodogAndMod)

	t.Run("WithModule", func(t *testing.T) {
		t.Run("OutsideGopathAndHavingOnlyFeature", testOutsideGopathAndHavingOnlyFeature)
		t.Run("OutsideGopath", testOutsideGopath)
		t.Run("OutsideGopathWithXTest", testOutsideGopathWithXTest)
		t.Run("InsideGopath", testInsideGopath)
	})
}

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

	"github.com/cucumber/godog"
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


func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step("^there are (\\d+) godogs$", thereAreGodogs)
	ctx.Step("^I eat (\\d+)$", iEat)
	ctx.Step("^there should be (\\d+) remaining$", thereShouldBeRemaining)

	ctx.BeforeScenario(func(*godog.Scenario) {
		Godogs = 0 // clean the state before every scenario
	})
}
`

var builderXTestFile = `package godogs_test

import (
	"fmt"

	"github.com/cucumber/godog"

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

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Step("^there are (\\d+) godogs$", thereAreGodogs)
	ctx.Step("^I eat (\\d+)$", iEat)
	ctx.Step("^there should be (\\d+) remaining$", thereShouldBeRemaining)

	ctx.BeforeScenario(func(*godog.Scenario) {
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

var builderModFile = `module godogs`

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

func buildTestCommand(t *testing.T, wd, featureFile string) *exec.Cmd {
	testBin := filepath.Join(wd, "godog.test")
	testBin, err := filepath.Abs(testBin)
	require.Nil(t, err)

	if build.Default.GOOS == "windows" {
		testBin += ".exe"
	}

	err = builder.Build(testBin)
	require.Nil(t, err)

	featureFilePath := filepath.Join(wd, featureFile)
	return exec.Command(testBin, featureFilePath)
}

func envVarsWithoutGopath() []string {
	var env []string

	for _, def := range os.Environ() {
		if strings.Index(def, "GOPATH=") == 0 {
			continue
		}

		env = append(env, def)
	}

	return env
}

func testWithSourceNotInGoPath(t *testing.T) {
	dir := filepath.Join(os.TempDir(), t.Name(), "godogs")
	files := map[string]string{
		"godogs.feature": builderFeatureFile,
		"godogs.go":      builderMainCodeFile,
		"godogs_test.go": builderTestFile,
		"go.mod":         builderModFile,
	}

	err := buildTestPackage(dir, files)
	defer os.RemoveAll(dir)
	require.Nil(t, err)

	prevDir, err := os.Getwd()
	require.Nil(t, err)

	err = os.Chdir(dir)
	require.Nil(t, err)
	defer os.Chdir(prevDir)

	testCmd := buildTestCommand(t, "", "godogs.feature")
	testCmd.Env = os.Environ()

	var stdout, stderr bytes.Buffer
	testCmd.Stdout = &stdout
	testCmd.Stderr = &stderr

	err = testCmd.Run()
	require.Nil(t, err, "stdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
}

func testWithoutSourceNotInGoPath(t *testing.T) {
	builderTC := builderTestCase{}

	builderTC.dir = filepath.Join(os.TempDir(), t.Name(), "godogs")
	builderTC.files = map[string]string{
		"godogs.feature": builderFeatureFile,
		"go.mod":         builderModFile,
	}

	builderTC.run(t)
}

func testWithoutTestSourceNotInGoPath(t *testing.T) {
	builderTC := builderTestCase{}

	builderTC.dir = filepath.Join(os.TempDir(), t.Name(), "godogs")
	builderTC.files = map[string]string{
		"godogs.feature": builderFeatureFile,
		"godogs.go":      builderMainCodeFile,
		"go.mod":         builderModFile,
	}

	builderTC.run(t)
}

func testWithinGopath(t *testing.T) {
	builderTC := builderTestCase{}

	gopath := filepath.Join(os.TempDir(), t.Name(), "_gp")
	defer os.RemoveAll(gopath)

	builderTC.dir = filepath.Join(gopath, "src", "godogs")
	builderTC.files = map[string]string{
		"godogs.feature": builderFeatureFile,
		"godogs.go":      builderMainCodeFile,
		"godogs_test.go": builderTestFile,
		"go.mod":         builderModFile,
	}

	pkg := filepath.Join(gopath, "src", "github.com", "cucumber")
	err := os.MkdirAll(pkg, 0755)
	require.Nil(t, err)

	wd, err := os.Getwd()
	require.Nil(t, err)

	// symlink godog package
	err = os.Symlink(wd, filepath.Join(pkg, "godog"))
	require.Nil(t, err)

	builderTC.testCmdEnv = []string{"GOPATH=" + gopath}
	builderTC.run(t)
}

func testWithVendoredGodogWithoutModule(t *testing.T) {
	builderTC := builderTestCase{}

	gopath := filepath.Join(os.TempDir(), t.Name(), "_gp")
	defer os.RemoveAll(gopath)

	builderTC.dir = filepath.Join(gopath, "src", "godogs")
	builderTC.files = map[string]string{
		"godogs.feature": builderFeatureFile,
	}

	pkg := filepath.Join(builderTC.dir, "vendor", "github.com", "cucumber")
	err := os.MkdirAll(pkg, 0755)
	require.Nil(t, err)

	wd, err := os.Getwd()
	require.Nil(t, err)

	// symlink godog package
	err = os.Symlink(wd, filepath.Join(pkg, "godog"))
	require.Nil(t, err)

	builderTC.testCmdEnv = append(envVarsWithoutGopath(), "GOPATH="+gopath)
	builderTC.run(t)
}

type builderTestCase struct {
	dir        string
	files      map[string]string
	goModCmds  []*exec.Cmd
	testCmdEnv []string
}

func (bt builderTestCase) run(t *testing.T) {
	t.Parallel()

	err := buildTestPackage(bt.dir, bt.files)
	defer os.RemoveAll(bt.dir)
	require.Nil(t, err)

	for _, c := range bt.goModCmds {
		c.Dir = bt.dir
		out, err := c.CombinedOutput()
		require.Nil(t, err, "%s", string(out))
	}

	testCmd := buildTestCommand(t, bt.dir, "godogs.feature")
	testCmd.Env = append(os.Environ(), bt.testCmdEnv...)

	var stdout, stderr bytes.Buffer
	testCmd.Stdout = &stdout
	testCmd.Stderr = &stderr

	err = testCmd.Run()
	require.Nil(t, err, "stdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
}
