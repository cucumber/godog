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
	"strconv"
	"strings"
	"text/template"
)

type builder struct {
	files    map[string]*ast.File
	fset     *token.FileSet
	Contexts []string
	Internal bool
	tpl      *template.Template

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
{{ if not .Internal }}import (
	"github.com/DATA-DOG/godog"
){{ end }}

func main() {

	{{ if not .Internal }}godog.{{ end }}Run(func (suite *{{ if not .Internal }}godog.{{ end }}Suite) {
		{{range .Contexts}}
			{{ . }}(suite)
		{{end}}
	})
}`)),
	}
}

func newBuilder(buildPath string) (*builder, error) {
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
			b.register(f, path)
		}
		return err
	})

	return b, err
}

func (b *builder) register(f *ast.File, path string) {
	// mark godog package as internal
	if f.Name.Name == "godog" && !b.Internal {
		b.Internal = true
	}
	b.deleteMainFunc(f)
	b.registerContexts(f)
	b.deleteImports(f)
	b.files[path] = f
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

func (b *builder) merge() ([]byte, error) {
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

	ret, err := ast.MergePackageFiles(pkg, 0), nil
	if err != nil {
		return nil, err
	}

	// @TODO: we reread the file, probably something goes wrong with position
	buf.Reset()
	if err = format.Node(&buf, b.fset, ret); err != nil {
		return nil, err
	}

	ret, err = parser.ParseFile(b.fset, "", buf.Bytes(), 0)
	if err != nil {
		return nil, err
	}

	used := b.usedPackages(ret)
	isUsed := func(p string) bool {
		for _, ref := range used {
			if p == ref {
				return true
			}
		}
		return p == "_"
	}
	for _, spec := range b.imports {
		var name string
		ipath := strings.Trim(spec.Path.Value, `\"`)
		check := importPathToName(ipath)
		if spec.Name != nil {
			name = spec.Name.Name
			check = spec.Name.Name
		}
		if isUsed(check) {
			addImport(b.fset, ret, name, ipath)
		}
	}

	buf.Reset()
	if err := format.Node(&buf, b.fset, ret); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
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

	return b.merge()
}

// taken from https://github.com/golang/tools/blob/master/go/ast/astutil/imports.go#L17
func addImport(fset *token.FileSet, f *ast.File, name, ipath string) {
	newImport := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(ipath),
		},
	}
	if name != "" {
		newImport.Name = &ast.Ident{Name: name}
	}

	// Find an import decl to add to.
	// The goal is to find an existing import
	// whose import path has the longest shared
	// prefix with ipath.
	var (
		bestMatch  = -1         // length of longest shared prefix
		lastImport = -1         // index in f.Decls of the file's final import decl
		impDecl    *ast.GenDecl // import decl containing the best match
		impIndex   = -1         // spec index in impDecl containing the best match
	)
	for i, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if ok && gen.Tok == token.IMPORT {
			lastImport = i
			// Do not add to import "C", to avoid disrupting the
			// association with its doc comment, breaking cgo.
			if declImports(gen, "C") {
				continue
			}

			// Match an empty import decl if that's all that is available.
			if len(gen.Specs) == 0 && bestMatch == -1 {
				impDecl = gen
			}

			// Compute longest shared prefix with imports in this group.
			for j, spec := range gen.Specs {
				impspec := spec.(*ast.ImportSpec)
				n := matchLen(importPath(impspec), ipath)
				if n > bestMatch {
					bestMatch = n
					impDecl = gen
					impIndex = j
				}
			}
		}
	}

	// If no import decl found, add one after the last import.
	if impDecl == nil {
		impDecl = &ast.GenDecl{
			Tok: token.IMPORT,
		}
		if lastImport >= 0 {
			impDecl.TokPos = f.Decls[lastImport].End()
		} else {
			// There are no existing imports.
			// Our new import goes after the package declaration and after
			// the comment, if any, that starts on the same line as the
			// package declaration.
			impDecl.TokPos = f.Package

			file := fset.File(f.Package)
			pkgLine := file.Line(f.Package)
			for _, c := range f.Comments {
				if file.Line(c.Pos()) > pkgLine {
					break
				}
				impDecl.TokPos = c.End()
			}
		}
		f.Decls = append(f.Decls, nil)
		copy(f.Decls[lastImport+2:], f.Decls[lastImport+1:])
		f.Decls[lastImport+1] = impDecl
	}

	// Insert new import at insertAt.
	insertAt := 0
	if impIndex >= 0 {
		// insert after the found import
		insertAt = impIndex + 1
	}
	impDecl.Specs = append(impDecl.Specs, nil)
	copy(impDecl.Specs[insertAt+1:], impDecl.Specs[insertAt:])
	impDecl.Specs[insertAt] = newImport
	pos := impDecl.Pos()
	if insertAt > 0 {
		// Assign same position as the previous import,
		// so that the sorter sees it as being in the same block.
		pos = impDecl.Specs[insertAt-1].Pos()
	}
	if newImport.Name != nil {
		newImport.Name.NamePos = pos
	}
	newImport.Path.ValuePos = pos
	newImport.EndPos = pos

	// Clean up parens. impDecl contains at least one spec.
	if len(impDecl.Specs) == 1 {
		// Remove unneeded parens.
		impDecl.Lparen = token.NoPos
	} else if !impDecl.Lparen.IsValid() {
		// impDecl needs parens added.
		impDecl.Lparen = impDecl.Specs[0].Pos()
	}

	f.Imports = append(f.Imports, newImport)
}

func declImports(gen *ast.GenDecl, path string) bool {
	if gen.Tok != token.IMPORT {
		return false
	}
	for _, spec := range gen.Specs {
		impspec := spec.(*ast.ImportSpec)
		if importPath(impspec) == path {
			return true
		}
	}
	return false
}

func matchLen(x, y string) int {
	n := 0
	for i := 0; i < len(x) && i < len(y) && x[i] == y[i]; i++ {
		if x[i] == '/' {
			n++
		}
	}
	return n
}

func importPath(s *ast.ImportSpec) string {
	return strings.Trim(s.Path.Value, `\"`)
}

var importPathToName = importPathToNameGoPath

// importPathToNameBasic assumes the package name is the base of import path.
func importPathToNameBasic(importPath string) (packageName string) {
	return path.Base(importPath)
}

// importPathToNameGoPath finds out the actual package name, as declared in its .go files.
// If there's a problem, it falls back to using importPathToNameBasic.
func importPathToNameGoPath(importPath string) (packageName string) {
	if buildPkg, err := build.Import(importPath, "", 0); err == nil {
		return buildPkg.Name
	}
	return importPathToNameBasic(importPath)
}
