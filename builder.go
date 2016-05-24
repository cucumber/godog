package godog

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var runnerTemplate = template.Must(template.New("main").Parse(`package {{ .PackageName }}
import (
{{ if ne .PackageName "godog" }}	"github.com/DATA-DOG/godog"{{ end }}
	"os"
	"testing"
)

const GodogSuiteName = "{{ .PackageName }}"

func TestMain(m *testing.M) {
	status := {{ if ne .PackageName "godog" }}godog.{{ end }}Run(func (suite *{{ if ne .PackageName "godog" }}godog.{{ end }}Suite) {
		{{range .Contexts}}
			{{ . }}(suite)
		{{end}}
	})
	os.Exit(status)
}`))

type builder struct {
	files       map[string]*ast.File
	Contexts    []string
	PackageName string
}

func (b *builder) register(f *ast.File, name string) {
	b.PackageName = f.Name.Name
	deleteTestMainFunc(f)
	// f.Name.Name = "main"
	b.Contexts = append(b.Contexts, contexts(f)...)
	b.files[name] = f
}

// Build scans all go files in current directory,
// copies them to temporary build directory.
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
	fset := token.NewFileSet()
	b := &builder{files: make(map[string]*ast.File)}

	err := filepath.Walk(".", func(path string, file os.FileInfo, err error) error {
		if file.IsDir() && file.Name() != "." {
			return filepath.SkipDir
		}
		// @TODO: maybe should copy all files in root dir (may contain CGO)
		// or use build.Import go tool, to manage package details
		if err == nil && strings.HasSuffix(path, ".go") {
			f, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return err
			}
			b.register(f, file.Name())
		}
		return err
	})

	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := runnerTemplate.Execute(&buf, b); err != nil {
		return err
	}

	f, err := parser.ParseFile(fset, "", &buf, 0)
	if err != nil {
		return err
	}

	b.files["godog_test.go"] = f

	os.Mkdir(dir, 0755)
	for name, node := range b.files {
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		if err := format.Node(f, fset, node); err != nil {
			return err
		}
	}

	return nil
}
