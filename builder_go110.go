// +build go1.10

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
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"
	"unicode"
)

var (
	tooldir         = findToolDir()
	compiler        = filepath.Join(tooldir, "compile")
	linker          = filepath.Join(tooldir, "link")
	gopaths         = filepath.SplitList(build.Default.GOPATH)
	godogImportPath = "github.com/DATA-DOG/godog"

	// godep
	runnerTemplate = template.Must(template.New("testmain").Parse(`package main

import (
	"github.com/DATA-DOG/godog"
	{{if .Contexts}}_test "{{.ImportPath}}"{{end}}
	{{if .XContexts}}_xtest "{{.ImportPath}}_test"{{end}}
	{{if .XContexts}}"testing/internal/testdeps"{{end}}
	"os"
)

{{if .XContexts}}
func init() {
	testdeps.ImportPath = "{{.ImportPath}}"
}
{{end}}

func main() {
	status := godog.Run("{{ .Name }}", func (suite *godog.Suite) {
		os.Setenv("GODOG_TESTED_PACKAGE", "{{.ImportPath}}")
		{{range .Contexts}}
			_test.{{ . }}(suite)
		{{end}}
		{{range .XContexts}}
			_xtest.{{ . }}(suite)
		{{end}}
	})
	os.Exit(status)
}`))
)

type module struct {
	Path, Dir string
}

func (mod *module) match(name string) *build.Package {
	if strings.Index(name, mod.Path) == -1 {
		return nil
	}

	suffix := strings.Replace(name, mod.Path, "", 1)
	add := strings.Replace(suffix, "/", string(filepath.Separator), -1)
	pkg, err := build.ImportDir(mod.Dir+add, 0)
	if err != nil {
		return nil
	}

	return pkg
}

// Build creates a test package like go test command at given target path.
// If there are no go files in tested directory, then
// it simply builds a godog executable to scan features.
//
// If there are go test files, it first builds a test
// package with standard go test command.
//
// Finally it generates godog suite executable which
// registers exported godog contexts from the test files
// of tested package.
//
// Returns the path to generated executable
func Build(bin string) error {
	abs, err := filepath.Abs(".")
	if err != nil {
		return err
	}

	// we allow package to be nil, if godog is run only when
	// there is a feature file in empty directory
	pkg := importPackage(abs)
	src, anyContexts, err := buildTestMain(pkg)
	if err != nil {
		return err
	}

	workdir := fmt.Sprintf(filepath.Join("%s", "godog-%d"), os.TempDir(), time.Now().UnixNano())
	testdir := workdir

	// if none of test files exist, or there are no contexts found
	// we will skip test package compilation, since it is useless
	if anyContexts {
		// build and compile the tested package.
		// generated test executable will be removed
		// since we do not need it for godog suite.
		// we also print back the temp WORK directory
		// go has built. We will reuse it for our suite workdir.
		temp := fmt.Sprintf(filepath.Join("%s", "temp-%d.test"), os.TempDir(), time.Now().UnixNano())
		out, err := exec.Command("go", "test", "-c", "-work", "-o", temp).CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to compile tested package: %s, reason: %v, output: %s", pkg.Name, err, string(out))
		}
		defer os.Remove(temp)

		// extract go-build temporary directory as our workdir
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		// it may have some compilation warnings, in the output, but these are not
		// considered to be errors, since command exit status is 0
		for _, ln := range lines {
			if !strings.HasPrefix(ln, "WORK=") {
				continue
			}
			workdir = strings.Replace(ln, "WORK=", "", 1)
			break
		}

		// may not locate it in output
		if workdir == testdir {
			return fmt.Errorf("expected WORK dir path to be present in output: %s", string(out))
		}

		// check whether workdir exists
		stats, err := os.Stat(workdir)
		if os.IsNotExist(err) {
			return fmt.Errorf("expected WORK dir: %s to be available", workdir)
		}

		if !stats.IsDir() {
			return fmt.Errorf("expected WORK dir: %s to be directory", workdir)
		}
		testdir = filepath.Join(workdir, "b001")
	} else {
		// still need to create temporary workdir
		if err = os.MkdirAll(testdir, 0755); err != nil {
			return err
		}
	}
	defer os.RemoveAll(workdir)

	// replace _testmain.go file with our own
	testmain := filepath.Join(testdir, "_testmain.go")
	err = ioutil.WriteFile(testmain, src, 0644)
	if err != nil {
		return err
	}

	mods := readModules()
	// if it was not located as module
	// we look it up in available source paths
	// including vendor directory, supported since 1.5.
	godogPkg, err := locatePackage(godogImportPath, mods)
	if err != nil {
		return err
	}

	if !isModule(godogImportPath, mods) {
		// must make sure that package is installed
		// modules are installed on download
		cmd := exec.Command("go", "install", "-i", godogPkg.ImportPath)
		cmd.Env = os.Environ()
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to install godog package: %s, reason: %v", string(out), err)
		}
	}

	// compile godog testmain package archive
	// we do not depend on CGO so a lot of checks are not necessary
	testMainPkgOut := filepath.Join(testdir, "main.a")
	args := []string{
		"-o", testMainPkgOut,
		"-p", "main",
		"-complete",
	}

	cfg := filepath.Join(testdir, "importcfg.link")
	args = append(args, "-importcfg", cfg)
	if _, err := os.Stat(cfg); err != nil {
		// there were no go sources in the directory
		// so we need to build all dependency tree ourselves
		in, err := os.Create(cfg)
		if err != nil {
			return err
		}
		fmt.Fprintln(in, "# import config")

		deps := make(map[string]string)
		if err := dependencies(godogPkg, mods, deps, false); err != nil {
			in.Close()
			return err
		}

		for pkgName, pkgObj := range deps {
			if i := strings.LastIndex(pkgName, "vendor/"); i != -1 {
				name := pkgName[i+7:]
				fmt.Fprintf(in, "importmap %s=%s\n", name, pkgName)
			}
			fmt.Fprintf(in, "packagefile %s=%s\n", pkgName, pkgObj)
		}
		in.Close()
	} else {
		// need to make sure that vendor dependencies are mapped
		in, err := os.OpenFile(cfg, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		deps := make(map[string]string)
		if err := dependencies(pkg, mods, deps, true); err != nil {
			in.Close()
			return err
		}
		if err := dependencies(godogPkg, mods, deps, false); err != nil {
			in.Close()
			return err
		}
		for pkgName := range deps {
			if i := strings.LastIndex(pkgName, "vendor/"); i != -1 {
				name := pkgName[i+7:]
				fmt.Fprintf(in, "importmap %s=%s\n", name, pkgName)
			}
		}
		in.Close()
	}

	args = append(args, "-pack", testmain)
	cmd := exec.Command(compiler, args...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to compile testmain package: %v - output: %s", err, string(out))
	}

	// link test suite executable
	args = []string{
		"-o", bin,
		"-importcfg", cfg,
		"-buildmode=exe",
	}
	args = append(args, testMainPkgOut)
	cmd = exec.Command(linker, args...)
	cmd.Env = os.Environ()

	// in case if build is without contexts, need to remove import maps
	data, err := ioutil.ReadFile(cfg)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var fixed []string
	for _, line := range lines {
		if strings.Index(line, "importmap") == 0 {
			continue
		}
		fixed = append(fixed, line)
	}
	if err := ioutil.WriteFile(cfg, []byte(strings.Join(fixed, "\n")), 0600); err != nil {
		return err
	}

	out, err = cmd.CombinedOutput()
	if err != nil {
		msg := `failed to link test executable:
	reason: %s
	command: %s`
		return fmt.Errorf(msg, string(out), linker+" '"+strings.Join(args, "' '")+"'")
	}

	return nil
}

func locatePackage(name string, mods []*module) (*build.Package, error) {
	// search vendor paths first since that takes priority
	dir, err := filepath.Abs(".")
	if err != nil {
		return nil, err
	}

	// first of all check modules
	if mods != nil {
		for _, mod := range mods {
			if pkg := mod.match(name); pkg != nil {
				return pkg, nil
			}
		}
	}

	for _, gopath := range gopaths {
		gopath = filepath.Join(gopath, "src")
		for strings.HasPrefix(dir, gopath) && dir != gopath {
			pkg, err := build.ImportDir(filepath.Join(dir, "vendor", name), 0)
			if err != nil {
				dir = filepath.Dir(dir)
				continue
			}
			return pkg, nil
		}
	}

	// search source paths otherwise
	for _, p := range build.Default.SrcDirs() {
		abs, err := filepath.Abs(filepath.Join(p, name))
		if err != nil {
			continue
		}
		pkg, err := build.ImportDir(abs, 0)
		if err != nil {
			continue
		}
		return pkg, nil
	}

	return nil, fmt.Errorf("failed to find %s package in any of:\n%s", name, strings.Join(build.Default.SrcDirs(), "\n"))
}

func importPackage(dir string) *build.Package {
	pkg, _ := build.ImportDir(dir, 0)

	// normalize import path for local import packages
	// taken from go source code
	// see: https://github.com/golang/go/blob/go1.7rc5/src/cmd/go/pkg.go#L279
	if pkg != nil && pkg.ImportPath == "." {
		pkg.ImportPath = path.Join("_", strings.Map(makeImportValid, filepath.ToSlash(dir)))
	}

	return pkg
}

// from go src
func makeImportValid(r rune) rune {
	// Should match Go spec, compilers, and ../../go/parser/parser.go:/isValidImport.
	const illegalChars = `!"#$%&'()*,:;<=>?[\]^{|}` + "`\uFFFD"
	if !unicode.IsGraphic(r) || unicode.IsSpace(r) || strings.ContainsRune(illegalChars, r) {
		return '_'
	}
	return r
}

// buildTestMain if given package is valid
// it scans test files for contexts
// and produces a testmain source code.
func buildTestMain(pkg *build.Package) ([]byte, bool, error) {
	var (
		contexts         []string
		xcontexts        []string
		err              error
		name, importPath string
	)
	if nil != pkg {
		contexts, err = processPackageTestFiles(pkg.TestGoFiles)
		if err != nil {
			return nil, false, err
		}
		xcontexts, err = processPackageTestFiles(pkg.XTestGoFiles)
		if err != nil {
			return nil, false, err
		}
		importPath = parseImport(pkg.ImportPath, pkg.Root)
		name = pkg.Name
	} else {
		name = "main"
	}
	data := struct {
		Name       string
		Contexts   []string
		XContexts  []string
		ImportPath string
	}{
		Name:       name,
		Contexts:   contexts,
		XContexts:  xcontexts,
		ImportPath: importPath,
	}

	hasContext := len(contexts) > 0 || len(xcontexts) > 0
	var buf bytes.Buffer
	if err = runnerTemplate.Execute(&buf, data); err != nil {
		return nil, hasContext, err
	}
	return buf.Bytes(), hasContext, nil
}

// parseImport parses the import path to deal with go module.
func parseImport(rawPath, rootPath string) string {
	// with go > 1.11 and go module enabled out of the GOPATH,
	// the import path begins with an underscore and the GOPATH is unknown on build.
	if rootPath != "" {
		// go < 1.11 or it's a module inside the GOPATH
		return rawPath
	}
	// for module support, query the module import path
	cmd := exec.Command("go", "list", "-m", "-json")
	out, err := cmd.StdoutPipe()
	if err != nil {
		// Unable to read stdout
		return rawPath
	}
	if cmd.Start() != nil {
		// Does not using modules
		return rawPath
	}
	var mod struct {
		Dir  string `json:"Dir"`
		Path string `json:"Path"`
	}
	if json.NewDecoder(out).Decode(&mod) != nil {
		// Unexpected result
		return rawPath
	}
	if cmd.Wait() != nil {
		return rawPath
	}
	// Concatenates the module path with the current sub-folders if needed
	return mod.Path + filepath.ToSlash(strings.TrimPrefix(strings.TrimPrefix(rawPath, "_"), mod.Dir))
}

// processPackageTestFiles runs through ast of each test
// file pack and looks for godog suite contexts to register
// on run
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

func findToolDir() string {
	if out, err := exec.Command("go", "env", "GOTOOLDIR").Output(); err != nil {
		return filepath.Clean(strings.TrimSpace(string(out)))
	}
	return filepath.Clean(build.ToolDir)
}

func dependencies(pkg *build.Package, mods []*module, visited map[string]string, vendor bool) error {
	visited[pkg.ImportPath] = pkg.PkgObj
	imports := pkg.Imports
	if vendor {
		imports = append(imports, pkg.TestImports...)
	}
	for _, name := range imports {
		if i := strings.LastIndex(name, "vendor/"); vendor && i == -1 {
			continue // only interested in vendor packages
		}
		if _, ok := visited[name]; ok {
			continue
		}

		next, err := locatePackage(name, mods)
		if err != nil {
			return err
		}

		visited[name] = pkg.PkgObj
		if err := dependencies(next, mods, visited, vendor); err != nil {
			return err
		}
	}
	return nil
}

func readModules() []*module {
	// for module support, query the module import path
	out, err := exec.Command("go", "mod", "download", "-json").Output()
	if err != nil {
		// Unable to read stdout
		return nil
	}

	var mods []*module
	reader := json.NewDecoder(bytes.NewReader(out))
	for {
		var mod *module
		if err := reader.Decode(&mod); err != nil {
			break // might be also EOF
		}
		mods = append(mods, mod)
	}
	return mods
}

func isModule(name string, mods []*module) bool {
	if mods == nil {
		return false
	}

	for _, mod := range mods {
		if pkg := mod.match(name); pkg != nil {
			return true
		}
	}

	return false
}
