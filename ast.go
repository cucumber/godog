package godog

import (
	"go/ast"
	"go/build"
	"go/token"
	"path"
	"strings"
)

func removeUnusedImports(f *ast.File) {
	used := usedPackages(f)
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

func deleteTestMainFunc(f *ast.File) {
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
		removeUnusedImports(f)
	}
}

type visitFn func(node ast.Node) ast.Visitor

func (fn visitFn) Visit(node ast.Node) ast.Visitor {
	return fn(node)
}

func usedPackages(f *ast.File) []string {
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

// importPathToName finds out the actual package name, as declared in its .go files.
// If there's a problem, it falls back to using importPathToNameBasic.
func importPathToName(importPath string) (packageName string) {
	if buildPkg, err := build.Import(importPath, "", 0); err == nil {
		return buildPkg.Name
	}
	return path.Base(importPath)
}
