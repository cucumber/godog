package godog

import (
	"bytes"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type builder struct {
	files     map[string]*ast.File
	fset      *token.FileSet
	Contexts  []string
	Internal  bool
	SuiteName string
	tpl       *template.Template

	imports []*ast.ImportSpec
}

func (b *builder) hasImport(a *ast.ImportSpec) bool {
	for _, b := range b.imports {
		var aname, bname string
		if a.Name != nil {
			aname = a.Name.Name
		}
		if b.Name != nil {
			bname = b.Name.Name
		}
		if bname == aname && a.Path.Value == b.Path.Value {
			return true
		}
	}
	return false
}

func newBuilderSkel() *builder {
	return &builder{
		files: make(map[string]*ast.File),
		fset:  token.NewFileSet(),
		tpl: template.Must(template.New("main").Parse(`package main
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
}`)),
	}
}

func doBuild(buildPath, dir string) error {
	b := newBuilderSkel()
	err := filepath.Walk(buildPath, func(path string, file os.FileInfo, err error) error {
		if file.IsDir() && file.Name() != "." {
			return filepath.SkipDir
		}
		if err == nil && strings.HasSuffix(path, ".go") {
			f, err := parser.ParseFile(b.fset, path, nil, 0)
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
	if err := b.tpl.Execute(&buf, b); err != nil {
		return err
	}

	f, err := parser.ParseFile(b.fset, "", &buf, 0)
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
		if err := format.Node(f, b.fset, node); err != nil {
			return err
		}
	}

	return nil
}

func (b *builder) register(f *ast.File, name string) {
	// mark godog package as internal
	if f.Name.Name == "godog" && !b.Internal {
		b.Internal = true
	}
	b.SuiteName = f.Name.Name
	b.deleteMainFunc(f)
	f.Name.Name = "main"
	b.registerContexts(f)
	b.files[name] = f
}

func (b *builder) removeUnusedImports(f *ast.File) {
	used := b.usedPackages(f)
	isUsed := func(p string) bool {
		for _, ref := range used {
			if p == ref {
				return true
			}
		}
		return p == "_"
	}
	var decls []ast.Decl
	for _, d := range f.Decls {
		gen, ok := d.(*ast.GenDecl)
		if ok && gen.Tok == token.IMPORT {
			var specs []ast.Spec
			for _, spec := range gen.Specs {
				impspec := spec.(*ast.ImportSpec)
				ipath := strings.Trim(impspec.Path.Value, `\"`)
				check := importPathToName(ipath)
				if impspec.Name != nil {
					check = impspec.Name.Name
				}

				if isUsed(check) {
					specs = append(specs, spec)
				}
			}

			if len(specs) == 0 {
				continue
			}
			gen.Specs = specs
		}
		decls = append(decls, d)
	}
	f.Decls = decls
}

func (b *builder) deleteImports(f *ast.File) {
	var decls []ast.Decl
	for _, d := range f.Decls {
		gen, ok := d.(*ast.GenDecl)
		if ok && gen.Tok == token.IMPORT {
			for _, spec := range gen.Specs {
				impspec := spec.(*ast.ImportSpec)
				if !b.hasImport(impspec) {
					b.imports = append(b.imports, impspec)
				}
			}
			continue
		}
		decls = append(decls, d)
	}
	f.Decls = decls
}

func (b *builder) deleteMainFunc(f *ast.File) {
	var decls []ast.Decl
	var hadMain bool
	for _, d := range f.Decls {
		fun, ok := d.(*ast.FuncDecl)
		if !ok {
			decls = append(decls, d)
			continue
		}
		if fun.Name.Name != "TestMain" {
			decls = append(decls, fun)
		} else {
			hadMain = true
		}
	}
	f.Decls = decls

	if hadMain {
		b.removeUnusedImports(f)
	}
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

type visitFn func(node ast.Node) ast.Visitor

func (fn visitFn) Visit(node ast.Node) ast.Visitor {
	return fn(node)
}

func (b *builder) usedPackages(f *ast.File) []string {
	var refs []string
	var visitor visitFn
	visitor = visitFn(func(node ast.Node) ast.Visitor {
		if node == nil {
			return visitor
		}
		switch v := node.(type) {
		case *ast.SelectorExpr:
			xident, ok := v.X.(*ast.Ident)
			if !ok {
				break
			}
			if xident.Obj != nil {
				// if the parser can resolve it, it's not a package ref
				break
			}
			refs = append(refs, xident.Name)
		}
		return visitor
	})
	ast.Walk(visitor, f)
	return refs
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
func Build(dir string) error {
	return doBuild(".", dir)
}

// importPathToName finds out the actual package name, as declared in its .go files.
// If there's a problem, it falls back to using importPathToNameBasic.
func importPathToName(importPath string) (packageName string) {
	if buildPkg, err := build.Import(importPath, "", 0); err == nil {
		return buildPkg.Name
	}
	return path.Base(importPath)
}
