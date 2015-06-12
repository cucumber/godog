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

var mainTpl = `package main

import (
	"github.com/DATA-DOG/godog"
)

func main() {
	suite := &GodogSuite{
		steps: make(map[*regexp.Regexp]StepHandler),
	}
	{{range $c := .Contexts}}
		{{$c}}(suite)
	{{end}}
	suite.Run()
}
`

type builder struct {
	files    map[string]*ast.File
	fset     *token.FileSet
	Contexts []string
}

func newBuilder() *builder {
	return &builder{
		files: make(map[string]*ast.File),
		fset:  token.NewFileSet(),
	}
}

func (b *builder) parseFile(path string) error {
	f, err := parser.ParseFile(b.fset, path, nil, 0)
	if err != nil {
		return err
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
		fun, ok := d.(*ast.FuncDecl)
		if !ok {
			continue
		}
		for _, param := range fun.Type.Params.List {
			ident, ok := param.Type.(*ast.Ident)
			if !ok {
				continue
			}
			if ident.Name == "godog.Suite" || f.Name.Name == "godog" && ident.Name == "Suite" {
				b.Contexts = append(b.Contexts, fun.Name.Name)
			}
		}
	}
}

func (b *builder) merge() (*ast.File, error) {
	var buf bytes.Buffer
	t := template.Must(template.New("main").Parse(mainTpl))
	if err := t.Execute(&buf, b); err != nil {
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

// Build creates a runnable godog executable
// from current package go files
func Build() ([]byte, error) {
	b := newBuilder()
	err := filepath.Walk(".", func(path string, file os.FileInfo, err error) error {
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
