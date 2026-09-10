package godog

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// registeredStep describes a step definition discovered by running the
// test suite's initializers, without executing any scenarios.
type registeredStep struct {
	Expr string `json:"expr"`

	// File and Line point at the call site of Step/Given/When/Then that
	// registered this step - not the handler's own definition, which isn't
	// reliably resolvable via reflection (see test_context.go). An editor
	// can reach the handler from here with a plain "go to declaration" on
	// the reference.
	File string `json:"file"`
	Line int    `json:"line"`
}

// registeredTest identifies the test function that called WriteManifest -
// the same one that's expected to call Run - so editor tooling can build a
// run/debug configuration for a feature or scenario deterministically,
// instead of guessing which test exercises a given features directory.
type registeredTest struct {
	Name string `json:"name"`
	File string `json:"file"`
	Line int    `json:"line"`
}

// manifest is the on-disk shape WriteManifest writes.
type manifest struct {
	Steps []registeredStep `json:"steps"`
	Test  *registeredTest  `json:"test,omitempty"`
}

// manifestFileName is the fixed name WriteManifest writes under the
// directory it's given - fixed, rather than caller-chosen, so that editor
// tooling looking for it doesn't depend on every caller spelling it the
// same way.
const manifestFileName = ".godog-gherkin.json"

// modPathPlaceholder prefixes a file living in the (machine-specific) Go
// module cache, e.g. "<MOD_PATH>/github.com/some/dep@v1.2.3/file.go" - a
// reader resolves it by substituting its own GOMODCACHE, keeping the rest
// (an already Go-escaped "path@version" cache directory name, identical on
// every machine) as-is.
const modPathPlaceholder = "<MOD_PATH>"

// portablePath rewrites an absolute path as reported by the runtime into a
// form that survives the manifest being committed and checked out
// elsewhere: relative to the manifest's own directory for anything
// reachable that way (the module being tested, including its vendor/ tree),
// or a modPathPlaceholder path for a dependency living in the module cache,
// which is machine-specific but always shaped goModCache/path@version/....
func portablePath(manifestDir, goModCache, absPath string) string {
	if !filepath.IsAbs(absPath) {
		return absPath
	}

	if goModCache != "" {
		if rel, ok := cutDirPrefix(absPath, goModCache); ok {
			return modPathPlaceholder + "/" + filepath.ToSlash(rel)
		}
	}

	if rel, err := filepath.Rel(manifestDir, absPath); err == nil {
		return filepath.ToSlash(rel)
	}

	return absPath
}

// cutDirPrefix reports whether path is inside dir and, if so, returns the
// part of path after it.
func cutDirPrefix(path, dir string) (string, bool) {
	dir = strings.TrimSuffix(dir, string(filepath.Separator)) + string(filepath.Separator)
	if !strings.HasPrefix(path, dir) {
		return "", false
	}

	return path[len(dir):], true
}

// goModCache returns `go env GOMODCACHE`, or "" if that fails - callers
// should degrade to manifest-relative/absolute paths rather than error out,
// since WriteManifest is a convenience, not something a build should fail on.
func goModCache() string {
	out, err := exec.Command("go", "env", "GOMODCACHE").Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}

// retrieveSteps runs the test suite's initializers and returns every
// registered step definition, without parsing or running any features.
func (ts TestSuite) retrieveSteps() []registeredStep {
	s := &suite{}

	if ts.TestSuiteInitializer != nil {
		ts.TestSuiteInitializer(&TestSuiteContext{suite: s})
	}

	if ts.ScenarioInitializer != nil {
		ts.ScenarioInitializer(&ScenarioContext{suite: s})
	}

	steps := make([]registeredStep, 0, len(s.steps))
	for _, d := range s.steps {
		steps = append(steps, registeredStep{
			Expr: d.Expr.String(),
			File: d.File,
			Line: d.Line,
		})
	}

	return steps
}

// WriteManifest runs the test suite's initializers and writes every
// registered step definition to <dir>/manifestFileName as JSON, without
// parsing or running any features. The file is left untouched if its
// content would not change, so its mtime only changes when steps actually
// do.
//
// It is meant for tooling (e.g. editor integrations) that need to resolve
// Gherkin steps to their Go implementation ahead of time, and to know which
// test to run for a given feature/scenario. For the latter to be
// deterministic, WriteManifest must be called directly from the same test
// function that calls Run, not through a helper:
//
//	func TestFeatures(t *testing.T) {
//	    ts := godog.TestSuite{ScenarioInitializer: InitializeScenario}
//	    _ = ts.WriteManifest(".") // records this call site as "the test"
//	    status := ts.Run()
//	    ...
//	}
//
// The written file is a good candidate for .gitignore - called alongside
// Run so it stays fresh on every test run.
func (ts TestSuite) WriteManifest(dir string) error {
	path := filepath.Join(dir, manifestFileName)

	dump := manifest{Steps: ts.retrieveSteps()}

	if pc, file, line, ok := runtime.Caller(1); ok {
		name := runtime.FuncForPC(pc).Name()
		if i := strings.LastIndexByte(name, '.'); i >= 0 {
			name = name[i+1:]
		}
		dump.Test = &registeredTest{Name: name, File: file, Line: line}
	}

	if manifestDir, err := filepath.Abs(dir); err == nil {
		modCache := goModCache()

		for i, s := range dump.Steps {
			dump.Steps[i].File = portablePath(manifestDir, modCache, s.File)
		}

		if dump.Test != nil {
			dump.Test.File = portablePath(manifestDir, modCache, dump.Test.File)
		}
	}

	b, err := json.MarshalIndent(dump, "", "  ")
	if err != nil {
		return err
	}

	if existing, err := os.ReadFile(path); err == nil && bytes.Equal(existing, b) {
		return nil
	}

	return os.WriteFile(path, b, 0o600)
}
