package godog

import (
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

var runnerTemplate = template.Must(template.New("main").Parse(`package {{ .Name }}

import (
	{{ if ne .Name "godog" }}"github.com/DATA-DOG/godog"{{ end }}
	"os"
	"testing"
)

const GodogSuiteName = "{{ .Name }}"

func TestMain(m *testing.M) {
	status := {{ if ne .Name "godog" }}godog.{{ end }}Run(func (suite *{{ if ne .Name "godog" }}godog.{{ end }}Suite) {
		{{range .Contexts}}
			{{ . }}(suite)
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
func Build(dir string) error {
	pkg, err := build.ImportDir(".", 0)
	if err != nil {
		return err
	}

	return buildTestPackage(pkg, dir)
}

// buildTestPackage clones a package and adds a godog
// entry point with TestMain func in order to
// run the test suite. If TestMain func is found in tested
// source, it will be removed so it can be replaced
func buildTestPackage(pkg *build.Package, dir string) error {
	// these file packs may be adjusted in the future, if there are complaints
	// in general, most important aspect is to replicate go test behavior
	err := copyNonTestPackageFiles(
		dir,
		pkg.CFiles,
		pkg.CgoFiles,
		pkg.CXXFiles,
		pkg.HFiles,
		pkg.GoFiles,
		pkg.IgnoredGoFiles,
	)

	if err != nil {
		return err
	}

	contexts, err := processPackageTestFiles(
		dir,
		pkg.TestGoFiles,
		pkg.XTestGoFiles,
	)
	if err != nil {
		return err
	}

	// build godog runner test file
	out, err := os.Create(filepath.Join(dir, "godog_runner_test.go"))
	if err != nil {
		return err
	}
	defer out.Close()

	data := struct {
		Name     string
		Contexts []string
	}{pkg.Name, contexts}

	return runnerTemplate.Execute(out, data)
}

// copyNonTestPackageFiles simply copies all given file packs
// to the destDir.
func copyNonTestPackageFiles(destDir string, packs ...[]string) error {
	for _, pack := range packs {
		for _, file := range pack {
			if err := copyPackageFile(file, filepath.Join(destDir, file)); err != nil {
				return err
			}
		}
	}
	return nil
}

// processPackageTestFiles runs through ast of each test
// file pack and removes TestMain func if found. it also
// looks for godog suite contexts to register on run
func processPackageTestFiles(destDir string, packs ...[]string) ([]string, error) {
	var ctxs []string
	fset := token.NewFileSet()
	for _, pack := range packs {
		for _, testFile := range pack {
			node, err := parser.ParseFile(fset, testFile, nil, 0)
			if err != nil {
				return ctxs, err
			}

			astDeleteTestMainFunc(node)
			ctxs = append(ctxs, astContexts(node)...)

			destFile := filepath.Join(destDir, testFile)
			if err = os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
				return ctxs, err
			}

			out, err := os.Create(destFile)
			if err != nil {
				return ctxs, err
			}
			defer out.Close()

			if err := format.Node(out, fset, node); err != nil {
				return ctxs, err
			}
		}
	}
	return ctxs, nil
}

// copyPackageFile simply copies the file, if dest file dir does
// not exist it creates it
func copyPackageFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	if err = os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return
	}
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
