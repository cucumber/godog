package godog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

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
func Build() error {
	abs, err := filepath.Abs(".")
	if err != nil {
		return err
	}
	pkg, err := build.ImportDir(abs, 0)
	if err != nil {
		return err
	}

	src, err := buildTestMain(pkg)
	if err != nil {
		return err
	}

	// first of all compile test package dependencies
	out, err := exec.Command("go", "test", "-i").CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			return fmt.Errorf("failed to compile package %s deps - %v, output - %s", pkg.Name, err, string(out))
		}
		return fmt.Errorf("failed to compile package %s deps - %v", pkg.Name, err)
	}

	// next build and compile the tested package
	// test executable will be piped to /dev/null
	// since we do not need it for godog suite
	// we also print back the temp WORK directory
	// go has build to test this package. We will
	// reuse it for our suite
	out, err = exec.Command("go", "test", "-c", "-work", "-o", os.DevNull).CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			return fmt.Errorf("failed to compile tested package %s - %v, output - %s", pkg.Name, err, string(out))
		}
		return fmt.Errorf("failed to compile tested package %s - %v", pkg.Name, err)
	}
	workdir := strings.TrimSpace(string(out))
	if !strings.HasPrefix(workdir, "WORK=") {
		return fmt.Errorf("expected WORK dir path, but got: %s", workdir)
	}
	workdir = strings.Replace(workdir, "WORK=", "", 1)
	testdir := filepath.Join(workdir, pkg.ImportPath, "_test")

	// replace testmain.go file with our own
	testmain := filepath.Join(testdir, "_testmain.go")
	err = ioutil.WriteFile(testmain, src, 0644)
	if err != nil {
		return err
	}

	// now we need to ensure godod library can be linked
	// need to check for it and compile
	// first try vendor directory and then all go src dirs
	try := []string{filepath.Join(pkg.Dir, "vendor", godogImportPath)}
	for _, d := range build.Default.SrcDirs() {
		try = append(try, filepath.Join(d, godogImportPath))
	}
	godogPkg, err := locatePackage(try)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	pkgDirs := []string{testdir, workdir, filepath.Join(godogPkg.PkgRoot, runtime.GOOS+"_"+runtime.GOARCH)}
	// build godog testmain package archive
	testMainPkgOut := filepath.Join(testdir, "main.a")
	args := []string{
		"tool", "compile",
		"-o", testMainPkgOut,
		"-trimpath", workdir,
		"-p", "main",
		"-complete",
	}
	if i := strings.LastIndex(godogPkg.ImportPath, "vendor/"); i != -1 {
		args = append(args, "-importmap", godogImportPath+"="+godogPkg.ImportPath)
	}
	for _, inc := range pkgDirs {
		args = append(args, "-I", inc)
	}
	args = append(args, "-pack", testmain)
	cmd := exec.Command("go", args...)
	cmd.Env = os.Environ()
	cmd.Dir = testdir
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err = cmd.Run()
	if err != nil {
		fmt.Println("command:", cmd.Path, cmd.Args)
		return fmt.Errorf("failed to compile testmain package %v, output - %s", err, buf.String())
	}

	// build test suite executable
	bin := filepath.Join(pkg.Dir, "godog.test")
	cmd = exec.Command(filepath.Join(build.ToolDir, "link"), "-o", bin)
	for _, link := range pkgDirs {
		cmd.Args = append(cmd.Args, "-L", link)
	}
	cmd.Args = append(cmd.Args, "-buildmode=exe", "-extld", build.Default.Compiler, testMainPkgOut)
	cmd.Env = os.Environ()
	cmd.Dir = testdir
	out, err = cmd.CombinedOutput()
	if err != nil {
		if len(out) > 0 {
			return fmt.Errorf("failed to compile testmain package %v, output - %s", err, string(out))
		}
		return fmt.Errorf("failed to compile testmain package %v", err)
	}

	return nil
}

func locatePackage(try []string) (*build.Package, error) {
	for _, path := range try {
		abs, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		return build.ImportDir(abs, 0)
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
	return ctxs, nil
}

func debug(i interface{}) {
	dat, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Fprintln(os.Stdout, string(dat))
}
