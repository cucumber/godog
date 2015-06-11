package godog

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/imports"
)

var mainTpl = `package main

import (
	"github.com/DATA-DOG/godog"
	"os"
)

func main() {
	godog.Run()
	os.Exit(0)
}
`

type builder struct {
	files map[string]*ast.File
	fset  *token.FileSet
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
	b.deleteImports(f)
	b.files[f.Name.Name] = f
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

func (b *builder) merge() (*ast.File, error) {
	f, err := parser.ParseFile(b.fset, "", mainTpl, 0)
	if err != nil {
		log.Println("fail here")
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
