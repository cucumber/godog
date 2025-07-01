package builder

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// buildStubTool builds a tiny cross-platform binary that simply honours “-o
// <file>” (creating an empty file) and otherwise exits 0.  We can use the same
// stub for both `compile` and `link`, allowing `Build` to be executed
// without invoking the real tool-chain.
func buildStubTool(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	stubSrc := filepath.Join(dir, "stub.go")
	stubBin := filepath.Join(dir, "stub")
	if runtime.GOOS == "windows" {
		stubBin += ".exe"
	}

	const program = `package main
import (
	"os"
)
func main() {
	for i, a := range os.Args {
		if a == "-o" && i+1 < len(os.Args) {
			f, _ := os.Create(os.Args[i+1])
			f.Close()
		}
	}
}`
	if err := os.WriteFile(stubSrc, []byte(program), 0644); err != nil {
		t.Fatalf("cannot write stub source: %v", err)
	}
	cmd := exec.Command("go", "build", "-o", stubBin, stubSrc)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build stub tool: %v\n%s", err, string(out))
	}
	return stubBin
}

// makePackage creates a throw-away Go package (and *_test.go) inside dir.
func makePackage(t *testing.T, dir, name string) {
	t.Helper()

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("cannot create package dir: %v", err)
	}
	src := `package ` + name + `
func Foo() {}
`
	testSrc := `package ` + name + `
import "testing"
func TestFoo(t *testing.T) { Foo() }
`
	if err := os.WriteFile(filepath.Join(dir, "foo.go"), []byte(src), 0644); err != nil {
		t.Fatalf("write foo.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "foo_test.go"), []byte(testSrc), 0644); err != nil {
		t.Fatalf("write foo_test.go: %v", err)
	}
}

// We create a vendor copy of godog so that `maybeVendoredGodog()` is non-nil,
// and use stub tools so the heavy compile/link work is skipped.
func TestRewritesImportCfgForVendoredGodog(t *testing.T) {
	origWD, _ := os.Getwd()
	defer os.Chdir(origWD)

	// GOPATH sandbox ----------------------------------------------------------
	gopath := t.TempDir()
	if err := os.Setenv("GOPATH", gopath); err != nil {
		t.Fatalf("set GOPATH: %v", err)
	}
	gopaths = filepath.SplitList(gopath)

	// Disable modules for a classic GOPATH build (avoids 'go mod tidy').
	if err := os.Setenv("GO111MODULE", "off"); err != nil {
		t.Fatalf("set GO111MODULE: %v", err)
	}
	defer os.Setenv("GO111MODULE", "on")

	pkgDir := filepath.Join(gopath, "src", "vendoredproj")
	defer os.Remove(pkgDir)
	makePackage(t, pkgDir, "vendoredproj")

	// Add a *vendored* copy of godog so maybeVendoredGodog() is triggered.
	vendorGodog := filepath.Join(pkgDir, "vendor", "github.com", "cucumber", "godog")
	if err := os.MkdirAll(vendorGodog, 0755); err != nil {
		t.Fatalf("mkdir vendor godog: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vendorGodog, "doc.go"),
		[]byte("package godog\nconst Version = \"0.0.1\""), 0644); err != nil {
		t.Fatalf("write vendor godog/doc.go: %v", err)
	}

	// -------------------------------------------------------------------------
	// Replace compile / link with stub tools
	// -------------------------------------------------------------------------
	stub := buildStubTool(t)

	oldCompile, oldLinker := compiler, linker
	compiler, linker = stub, stub
	defer func() { compiler, linker = oldCompile, oldLinker }()

	if err := os.Chdir(pkgDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	bin := filepath.Join(pkgDir, "godog_suite_bin")
	defer os.Remove(bin)
	if err := Build(bin); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if _, err := os.Stat(bin); err != nil {
		t.Fatalf("expected binary %s to exist: %v", bin, err)
	}
}
