package godog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestSuite_WriteManifest(t *testing.T) {
	ts := TestSuite{
		ScenarioInitializer: func(ctx *ScenarioContext) {
			ctx.Given(`^passing step$`, okStep)
		},
	}

	dir := t.TempDir()
	require.NoError(t, ts.WriteManifest(dir))

	b, err := os.ReadFile(filepath.Join(dir, manifestFileName))
	require.NoError(t, err)

	var dump manifest
	require.NoError(t, json.Unmarshal(b, &dump))
	require.Len(t, dump.Steps, 1)

	assert.Equal(t, "^passing step$", dump.Steps[0].Expr)
	assert.Equal(t, "manifest_test.go", filepath.Base(dump.Steps[0].File))

	require.NotNil(t, dump.Test)
	assert.Equal(t, "TestTestSuite_WriteManifest", dump.Test.Name)
	assert.Equal(t, "manifest_test.go", filepath.Base(dump.Test.File))
}

func TestPortablePath(t *testing.T) {
	manifestDir := filepath.FromSlash("/repo/pkg")
	modCache := filepath.FromSlash("/home/me/go/pkg/mod")

	t.Run("in-module file becomes manifest-relative", func(t *testing.T) {
		got := portablePath(manifestDir, modCache, filepath.FromSlash("/repo/pkg/sub/file.go"))
		assert.Equal(t, "sub/file.go", got)
	})

	t.Run("vendored file becomes manifest-relative too", func(t *testing.T) {
		got := portablePath(manifestDir, modCache, filepath.FromSlash("/repo/pkg/vendor/github.com/x/y/z.go"))
		assert.Equal(t, "vendor/github.com/x/y/z.go", got)
	})

	t.Run("module cache file becomes a MOD_PATH placeholder", func(t *testing.T) {
		got := portablePath(manifestDir, modCache, filepath.FromSlash("/home/me/go/pkg/mod/github.com/some/dep@v1.2.3/file.go"))
		assert.Equal(t, "<MOD_PATH>/github.com/some/dep@v1.2.3/file.go", got)
	})

	t.Run("relative input is passed through unchanged", func(t *testing.T) {
		assert.Equal(t, "already/relative.go", portablePath(manifestDir, modCache, "already/relative.go"))
	})

	t.Run("empty GOMODCACHE just skips that check", func(t *testing.T) {
		got := portablePath(manifestDir, "", filepath.FromSlash("/repo/pkg/sub/file.go"))
		assert.Equal(t, "sub/file.go", got)
	})
}
