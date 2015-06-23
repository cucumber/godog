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

	"golang.org/x/tools/imports"
)

type builder struct {
	files    map[string]*ast.File
	fset     *token.FileSet
	Contexts []string
	Internal bool
	tpl      *template.Template
}

func newBuilder(buildPath string) (*builder, error) {
	b := &builder{
		files: make(map[string]*ast.File),
		fset:  token.NewFileSet(),
		tpl: template.Must(template.New("main").Parse(`package main
{{ if not .Internal }}import (
	"github.com/DATA-DOG/godog"
){{ end }}

func main() {
	suite := {{ if not .Internal }}godog.{{ end }}New()
	{{range .Contexts}}
		{{ . }}(suite)
	{{end}}
	suite.Run()
}`)),
	}

	return b, filepath.Walk(buildPath, func(path string, file os.FileInfo, err error) error {
		if file.IsDir() && file.Name() != "." {
			return filepath.SkipDir
		}
		if err == nil && strings.HasSuffix(path, ".go") {
			if err := b.parseFile(path); err != nil {
				return err
			}
		}
		return err
	})
}

func (b *builder) parseFile(path string) error {
	f, err := parser.ParseFile(b.fset, path, nil, 0)
	if err != nil {
		return err
	}
	// mark godog package as internal
	if f.Name.Name == "godog" && !b.Internal {
		b.Internal = true
	}
	b.deleteMainFunc(f)
	b.registerSteps(f)
	b.deleteImports(f)
	b.files[path] = f
	return nil
}

func (b *builder) deleteImports(f *ast.File) {
	var decls []ast.Decl
	for _, d := range f.Decls {
		fun, ok := d.(*ast.GenDecl)
		if !ok {
			decls = append(decls, d)
			continue
		}
		if fun.Tok != token.IMPORT {
			decls = append(decls, fun)
		}
	}
	f.Decls = decls
}

func (b *builder) deleteMainFunc(f *ast.File) {
	var decls []ast.Decl
	for _, d := range f.Decls {
		fun, ok := d.(*ast.FuncDecl)
		if !ok {
			decls = append(decls, d)
			continue
		}
		if fun.Name.Name != "main" {
			decls = append(decls, fun)
		}
	}
	f.Decls = decls
}

func (b *builder) registerSteps(f *ast.File) {
	for _, d := range f.Decls {
		switch fun := d.(type) {
		case *ast.FuncDecl:
			for _, param := range fun.Type.Params.List {
				switch expr := param.Type.(type) {
				case *ast.SelectorExpr:
					switch x := expr.X.(type) {
					case *ast.Ident:
						if x.Name == "godog" && expr.Sel.Name == "Suite" {
							b.Contexts = append(b.Contexts, fun.Name.Name)
						}
					}
				case *ast.Ident:
					if expr.Name == "Suite" {
						b.Contexts = append(b.Contexts, fun.Name.Name)
					}
				}
			}
		}
	}
}

func (b *builder) merge() (*ast.File, error) {
	var buf bytes.Buffer
	if err := b.tpl.Execute(&buf, b); err != nil {
		return nil, err
	}

	f, err := parser.ParseFile(b.fset, "", &buf, 0)
	if err != nil {
		return nil, err
	}
	b.deleteImports(f)
	b.files["main.go"] = f

	pkg, _ := ast.NewPackage(b.fset, b.files, nil, nil)
	pkg.Name = "main"

	return ast.MergePackageFiles(pkg, ast.FilterImportDuplicates), nil
}

// Build creates a runnable Godog executable file
// from current package source and test source files.
//
// The package files are merged with the help of go/ast into
// a single main package file which has a custom
// main function to run test suite features.
//
// Currently, to manage imports we use "golang.org/x/tools/imports"
// package, but that may be replaced in order to have
// no external dependencies
func Build() ([]byte, error) {
	b, err := newBuilder(".")
	if err != nil {
		return nil, err
	}

	merged, err := b.merge()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	if err := format.Node(&buf, b.fset, merged); err != nil {
		return nil, err
	}

	return imports.Process("", buf.Bytes(), nil)
}
