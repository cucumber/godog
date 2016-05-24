package godog

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

var astMainFile = `package main

import "fmt"

func main() {
	fmt.Println("hello")
}`

var astNormalFile = `package main

import "fmt"

func hello() {
	fmt.Println("hello")
}`

var astTestMainFile = `package main

import (
	"fmt"
	"testing"
	"os"
)

func TestMain(m *testing.M) {
	fmt.Println("hello")
	os.Exit(0)
}`

var astPackAliases = `package main

import (
	"testing"
	a "fmt"
	b "fmt"
)

func TestMain(m *testing.M) {
	a.Println("a")
	b.Println("b")
}`

var astAnonymousImport = `package main

import (
	"testing"
	_ "github.com/go-sql-driver/mysql"
)

func TestMain(m *testing.M) {
}`

var astLibrarySrc = `package lib

import "fmt"

func test() {
	fmt.Println("hello")
}`

var astInternalPackageSrc = `package godog

import "fmt"

func test() {
	fmt.Println("hello")
}`

func astProcess(src string, t *testing.T) string {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", []byte(src), 0)
	if err != nil {
		t.Fatalf("unexpected error while parsing ast: %v", err)
	}

	deleteTestMainFunc(f)

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		t.Fatalf("failed to build source file: %v", err)
	}

	return buf.String()
}

func TestShouldCleanTestMainFromSimpleTestFile(t *testing.T) {
	actual := strings.TrimSpace(astProcess(astTestMainFile, t))
	expect := `package main`

	if actual != expect {
		t.Fatalf("expected output does not match: %s", actual)
	}
}

func TestShouldCleanTestMainFromFileWithPackageAliases(t *testing.T) {
	actual := strings.TrimSpace(astProcess(astPackAliases, t))
	expect := `package main`

	if actual != expect {
		t.Fatalf("expected output does not match: %s", actual)
	}
}

func TestShouldNotModifyNormalFile(t *testing.T) {
	actual := strings.TrimSpace(astProcess(astNormalFile, t))
	expect := astNormalFile

	if actual != expect {
		t.Fatalf("expected output does not match: %s", actual)
	}
}

func TestShouldNotModifyMainFile(t *testing.T) {
	actual := strings.TrimSpace(astProcess(astMainFile, t))
	expect := astMainFile

	if actual != expect {
		t.Fatalf("expected output does not match: %s", actual)
	}
}

func TestShouldMaintainAnonymousImport(t *testing.T) {
	actual := strings.TrimSpace(astProcess(astAnonymousImport, t))
	expect := `package main

import (
	_ "github.com/go-sql-driver/mysql"
)`

	if actual != expect {
		t.Fatalf("expected output does not match: %s", actual)
	}
}

func TestShouldNotModifyLibraryPackageSource(t *testing.T) {
	actual := strings.TrimSpace(astProcess(astLibrarySrc, t))
	expect := astLibrarySrc

	if actual != expect {
		t.Fatalf("expected output does not match: %s", actual)
	}
}

func TestShouldNotModifyGodogPackageSource(t *testing.T) {
	actual := strings.TrimSpace(astProcess(astInternalPackageSrc, t))
	expect := astInternalPackageSrc

	if actual != expect {
		t.Fatalf("expected output does not match: %s", actual)
	}
}
