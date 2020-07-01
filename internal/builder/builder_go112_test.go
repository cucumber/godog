// +build go1.12
// +build !go1.13

package builder_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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
