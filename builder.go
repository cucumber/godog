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

var runnerTemplate = template.Must(template.New("main").Parse(`package main
import (
{{ if not .Internal }}	"github.com/DATA-DOG/godog"{{ end }}
	"os"
	"testing"
)

const GodogSuiteName = "{{ .SuiteName }}"

func TestMain(m *testing.M) {
	status := {{ if not .Internal }}godog.{{ end }}Run(func (suite *{{ if not .Internal }}godog.{{ end }}Suite) {
		{{range .Contexts}}
			{{ . }}(suite)
		{{end}}
	})
	os.Exit(status)
}`))

type builder struct {
	files     map[string]*ast.File
	Contexts  []string
	Internal  bool
	SuiteName string
}

func (b *builder) register(f *ast.File, name string) {
	// mark godog package as internal
	if f.Name.Name == "godog" && !b.Internal {
		b.Internal = true
	}
	b.SuiteName = f.Name.Name
	deleteTestMainFunc(f)
	f.Name.Name = "main"
	b.registerContexts(f)
	b.files[name] = f
}

func (b *builder) registerContexts(f *ast.File) {
	for _, d := range f.Decls {
		switch fun := d.(type) {
		case *ast.FuncDecl:
			for _, param := range fun.Type.Params.List {
				switch expr := param.Type.(type) {
				case *ast.StarExpr:
					switch x := expr.X.(type) {
					case *ast.Ident:
						if x.Name == "Suite" {
							b.Contexts = append(b.Contexts, fun.Name.Name)
						}
					case *ast.SelectorExpr:
						switch t := x.X.(type) {
						case *ast.Ident:
							if t.Name == "godog" && x.Sel.Name == "Suite" {
								b.Contexts = append(b.Contexts, fun.Name.Name)
							}
						}
					}
				}
			}
		}
	}
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
