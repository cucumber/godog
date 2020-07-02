// +build go1.13

package builder_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func testWithVendoredGodogAndMod(t *testing.T) {
	builderTC := builderTestCase{}

	gopath := filepath.Join(os.TempDir(), t.Name(), "_gpc")
	defer os.RemoveAll(gopath)

	builderTC.dir = filepath.Join(gopath, "src", "godogs")
	builderTC.files = map[string]string{
		"godogs.feature": builderFeatureFile,
		"godogs.go":      builderMainCodeFile,
		"godogs_test.go": builderTestFile,
		"go.mod":         builderModFile,
	}

	builderTC.goModCmds = make([]*exec.Cmd, 1)
	builderTC.goModCmds[0] = exec.Command("go", "mod", "vendor")
	builderTC.testCmdEnv = append(envVarsWithoutGopath(), "GOPATH="+gopath)

	builderTC.run(t)
}
