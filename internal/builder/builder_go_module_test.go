package builder_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func testOutsideGopathAndHavingOnlyFeature(t *testing.T) {
	builderTC := builderTestCase{}

	builderTC.dir = filepath.Join(os.TempDir(), t.Name(), "godogs")
	builderTC.files = map[string]string{
		"godogs.feature": builderFeatureFile,
	}

	builderTC.goModCmds = make([]*exec.Cmd, 2)
	builderTC.goModCmds[0] = exec.Command("go", "mod", "init", "godogs")

	builderTC.goModCmds[1] = exec.Command("go", "mod", "tidy")

	builderTC.run(t)
}

func testOutsideGopath(t *testing.T) {
	builderTC := builderTestCase{}

	builderTC.dir = filepath.Join(os.TempDir(), t.Name(), "godogs")
	builderTC.files = map[string]string{
		"godogs.feature": builderFeatureFile,
		"godogs.go":      builderMainCodeFile,
		"godogs_test.go": builderTestFile,
	}

	builderTC.goModCmds = make([]*exec.Cmd, 1)
	builderTC.goModCmds[0] = exec.Command("go", "mod", "init", "godogs")

	builderTC.run(t)
}

func testOutsideGopathWithXTest(t *testing.T) {
	builderTC := builderTestCase{}

	builderTC.dir = filepath.Join(os.TempDir(), t.Name(), "godogs")
	builderTC.files = map[string]string{
		"godogs.feature": builderFeatureFile,
		"godogs.go":      builderMainCodeFile,
		"godogs_test.go": builderXTestFile,
	}

	builderTC.goModCmds = make([]*exec.Cmd, 1)
	builderTC.goModCmds[0] = exec.Command("go", "mod", "init", "godogs")

	builderTC.run(t)
}

func testInsideGopath(t *testing.T) {
	builderTC := builderTestCase{}

	gopath := filepath.Join(os.TempDir(), t.Name(), "_gp")
	defer func() {
		if err := os.RemoveAll(gopath); err != nil {
			t.Fatal(err)
		}
	}()

	builderTC.dir = filepath.Join(gopath, "src", "godogs")
	builderTC.files = map[string]string{
		"godogs.feature": builderFeatureFile,
		"godogs.go":      builderMainCodeFile,
		"godogs_test.go": builderTestFile,
	}

	builderTC.goModCmds = make([]*exec.Cmd, 1)
	builderTC.goModCmds[0] = exec.Command("go", "mod", "init", "godogs")
	builderTC.goModCmds[0].Env = os.Environ()
	builderTC.goModCmds[0].Env = append(builderTC.goModCmds[0].Env, "GOPATH="+gopath)
	builderTC.goModCmds[0].Env = append(builderTC.goModCmds[0].Env, "GO111MODULE=on")

	builderTC.run(t)
}
