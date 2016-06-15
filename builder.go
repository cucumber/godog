package godog

import (
	"bytes"
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"unicode"
)

var compiler = filepath.Join(build.ToolDir, "compile")
var linker = filepath.Join(build.ToolDir, "link")
var gopaths = filepath.SplitList(build.Default.GOPATH)
var goarch = build.Default.GOARCH
var goos = build.Default.GOOS
var supportVendor = os.Getenv("GO15VENDOREXPERIMENT") != "0"

var godogImportPath = "github.com/DATA-DOG/godog"
var runnerTemplate = template.Must(template.New("testmain").Parse(`package main

import (
	"github.com/DATA-DOG/godog"
	_test "{{ .ImportPath }}"
	"os"
)

func main() {
	status := godog.Run(func (suite *godog.Suite) {
		{{range .Contexts}}
			_test.{{ . }}(suite)
		{{end}}
	})
	os.Exit(status)
}`))

// Build scans clones current package into a temporary
// godog suite test package.
//
// If there is a TestMain func in any of test.go files
// it removes it and all necessary unused imports related
// to this function.
//
// It also looks for any godog suite contexts and registers
// them in order to call them on execution.
//
// The test entry point which uses go1.4 TestMain func
// is generated from the template above.
func Build() (string, error) {
	abs, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	pkg, err := build.ImportDir(abs, 0)
	if err != nil {
		return "", err
	}

	bin := filepath.Join(pkg.Dir, "godog.test")
	// suffix with .exe for windows
	if goos == "windows" {
		bin += ".exe"
	}
	src, err := buildTestMain(pkg)
	if err != nil {
		return bin, err
	}

	// first of all compile test package dependencies
	// that will save was many compilations for dependencies
	// go does it better
	out, err := exec.Command("go", "test", "-i").CombinedOutput()
	if err != nil {
		return bin, fmt.Errorf("failed to compile package %s:\n%s", pkg.Name, string(out))
	}

	// let go do the dirty work and compile test
	// package with it's dependencies. Older go
	// versions does not accept existing file output
	// so we create a temporary executable which will
	// removed.
	temp := fmt.Sprintf(filepath.Join("%s", "temp-%d.test"), os.TempDir(), time.Now().UnixNano())

	// builds and compile the tested package.
	// generated test executable will be removed
	// since we do not need it for godog suite.
	// we also print back the temp WORK directory
	// go has built. We will reuse it for our suite workdir.
	out, err = exec.Command("go", "test", "-c", "-work", "-o", temp).CombinedOutput()
	if err != nil {
		return bin, fmt.Errorf("failed to compile tested package %s:\n%s", pkg.Name, string(out))
	}
	defer os.Remove(temp)

	// extract go-build temporary directory as our workdir
	workdir := strings.TrimSpace(string(out))
	if !strings.HasPrefix(workdir, "WORK=") {
		return bin, fmt.Errorf("expected WORK dir path, but got: %s", workdir)
	}
	workdir = strings.Replace(workdir, "WORK=", "", 1)
	testdir := filepath.Join(workdir, pkg.ImportPath, "_test")
	defer os.RemoveAll(workdir)

	// replace _testmain.go file with our own
	testmain := filepath.Join(testdir, "_testmain.go")
	err = ioutil.WriteFile(testmain, src, 0644)
	if err != nil {
		return bin, err
	}

	// godog library may not be imported in tested package
	// but we need it for our testmain package.
	// So we look it up in available source paths
	// including vendor directory, supported since 1.5.
	try := []string{filepath.Join(pkg.Dir, "vendor", godogImportPath)}
	for _, d := range build.Default.SrcDirs() {
		try = append(try, filepath.Join(d, godogImportPath))
	}
	godogPkg, err := locatePackage(try)
	if err != nil {
		return bin, err
	}

	// make sure godog package archive is installed, gherkin
	// will be installed as dependency of godog
	cmd := exec.Command("go", "install", godogPkg.ImportPath)
	cmd.Env = os.Environ()
	out, err = cmd.CombinedOutput()
	if err != nil {
		return bin, fmt.Errorf("failed to install godog package:\n%s", string(out))
	}

	// collect all possible package dirs, will be
	// used for includes and linker
	pkgDirs := []string{testdir, workdir}
	for _, gopath := range gopaths {
		pkgDirs = append(pkgDirs, filepath.Join(gopath, "pkg", goos+"_"+goarch))
	}

	// compile godog testmain package archive
	// we do not depend on CGO so a lot of checks are not necessary
	testMainPkgOut := filepath.Join(testdir, "main.a")
	args := []string{
		"-o", testMainPkgOut,
		// "-trimpath", workdir,
		"-p", "main",
		"-complete",
	}
	// if godog library is in vendor directory
	// link it with import map
	if i := strings.LastIndex(godogPkg.ImportPath, "vendor/"); i != -1 && supportVendor {
		args = append(args, "-importmap", godogImportPath+"="+godogPkg.ImportPath)
	}
	for _, inc := range pkgDirs {
		args = append(args, "-I", inc)
	}
	args = append(args, "-pack", testmain)
	cmd = exec.Command(compiler, args...)
	cmd.Env = os.Environ()
	out, err = cmd.CombinedOutput()
	if err != nil {
		return bin, fmt.Errorf("failed to compile testmain package - %v:\n%s", err, string(out))
	}

	// link test suite executable
	args = []string{
		"-o", bin,
		"-extld", build.Default.Compiler,
		"-buildmode=exe",
	}
	for _, link := range pkgDirs {
		args = append(args, "-L", link)
	}
	args = append(args, testMainPkgOut)
	cmd = exec.Command(linker, args...)
	cmd.Env = os.Environ()
	out, err = cmd.CombinedOutput()
	if err != nil {
		return bin, fmt.Errorf("failed to link test executable:\n%s", string(out))
	}

	return bin, nil
}

func locatePackage(try []string) (*build.Package, error) {
	for _, path := range try {
		abs, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		pkg, err := build.ImportDir(abs, 0)
		if err != nil {
			continue
		}
		return pkg, nil
	}
	return nil, fmt.Errorf("failed to find godog package in any of:\n%s", strings.Join(try, "\n"))
}

// buildTestPackage clones a package and adds a godog
// entry point with TestMain func in order to
// run the test suite. If TestMain func is found in tested
// source, it will be removed so it can be replaced
func buildTestMain(pkg *build.Package) ([]byte, error) {
	contexts, err := processPackageTestFiles(
		pkg.TestGoFiles,
		pkg.XTestGoFiles,
	)
	if err != nil {
		return nil, err
	}

	data := struct {
		Name       string
		Contexts   []string
		ImportPath string
	}{pkg.Name, contexts, pkg.ImportPath}

	var buf bytes.Buffer
	if err = runnerTemplate.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// processPackageTestFiles runs through ast of each test
// file pack and removes TestMain func if found. it also
// looks for godog suite contexts to register on run
func processPackageTestFiles(packs ...[]string) ([]string, error) {
	var ctxs []string
	fset := token.NewFileSet()
	for _, pack := range packs {
		for _, testFile := range pack {
			node, err := parser.ParseFile(fset, testFile, nil, 0)
			if err != nil {
				return ctxs, err
			}

			ctxs = append(ctxs, astContexts(node)...)
		}
	}
	var failed []string
	for _, ctx := range ctxs {
		runes := []rune(ctx)
		if unicode.IsLower(runes[0]) {
			expected := append([]rune{unicode.ToUpper(runes[0])}, runes[1:]...)
			failed = append(failed, fmt.Sprintf("%s - should be: %s", ctx, string(expected)))
		}
	}
	if len(failed) > 0 {
		return ctxs, fmt.Errorf("godog contexts must be exported:\n\t%s", strings.Join(failed, "\n\t"))
	}
	return ctxs, nil
}
