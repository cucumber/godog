package godog

import (
	"fmt"
	"go/parser"
	"go/token"
	"runtime"
	"strings"
	"testing"
)

var builderMainFile = `
package main
import "fmt"
func main() {
	fmt.Println("hello")
}`

var builderPackAliases = `
package main
import (
	a "fmt"
	b "fmt"
)
func Tester() {
	a.Println("a")
	b.Println("b")
}`

var builderAnonymousImport = `
package main
import (
	_ "github.com/go-sql-driver/mysql"
)
`

var builderContextSrc = `
package main
import (
	"github.com/DATA-DOG/godog"
)

func myContext(s *godog.Suite) {

}
`

var builderLibrarySrc = `
package lib
import "fmt"
func test() {
	fmt.Println("hello")
}
`

var builderInternalPackageSrc = `
package godog
import "fmt"
func test() {
	fmt.Println("hello")
}
`

func (b *builder) registerMulti(contents []string) error {
	for i, c := range contents {
		f, err := parser.ParseFile(token.NewFileSet(), "", []byte(c), 0)
		if err != nil {
			return err
		}
		b.register(f, fmt.Sprintf("path%d", i))
	}
	return nil
}

func (b *builder) cleanSpacing(src string) string {
	var lines []string
	for _, ln := range strings.Split(src, "\n") {
		if ln == "" {
			continue
		}
		lines = append(lines, strings.TrimSpace(ln))
	}
	return strings.Join(lines, "\n")
}

func TestUsualSourceFileMerge(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.1") {
		t.Skip("skipping this test for go1.1")
	}
	b := newBuilderSkel()
	err := b.registerMulti([]string{
		builderMainFile, builderPackAliases, builderAnonymousImport,
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	data, err := b.merge()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected := `package main

import (
	"fmt"
	a "fmt"
	b "fmt"
	"github.com/DATA-DOG/godog"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	godog.Run(func(suite *godog.Suite) {

	})
}
func Tester() {
	a.Println("a")
	b.Println("b")
}`

	actual := string(data)
	if b.cleanSpacing(expected) != b.cleanSpacing(actual) {
		t.Fatalf("expected output does not match: %s", actual)
	}
}

func TestShouldCallContextOnMerged(t *testing.T) {
	b := newBuilderSkel()
	err := b.registerMulti([]string{
		builderMainFile, builderContextSrc,
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	data, err := b.merge()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected := `package main
import (
	"fmt"
	"github.com/DATA-DOG/godog"
)

func main() {
	godog.Run(func(suite *godog.Suite) {
		myContext(suite)
	})
}

func myContext(s *godog.Suite) {
}`

	actual := string(data)
	if b.cleanSpacing(expected) != b.cleanSpacing(actual) {
		t.Fatalf("expected output does not match: %s", actual)
	}
}

func TestBuildLibraryPackage(t *testing.T) {
	b := newBuilderSkel()
	err := b.registerMulti([]string{
		builderLibrarySrc,
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	data, err := b.merge()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected := `package main
import (
	"fmt"
	"github.com/DATA-DOG/godog"
)

func main() {
	godog.Run(func(suite *godog.Suite) {

	})
}

func test() {
	fmt.Println(
		"hello",
	)
}`

	actual := string(data)
	if b.cleanSpacing(expected) != b.cleanSpacing(actual) {
		t.Fatalf("expected output does not match: %s", actual)
	}
}

func TestBuildInternalPackage(t *testing.T) {
	b := newBuilderSkel()
	err := b.registerMulti([]string{
		builderInternalPackageSrc,
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	data, err := b.merge()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expected := `package main
import "fmt"

func main() {
	Run(func(suite *Suite) {

	})
}

func test() {
	fmt.Println("hello")
}`

	actual := string(data)
	if b.cleanSpacing(expected) != b.cleanSpacing(actual) {
		t.Fatalf("expected output does not match: %s", actual)
	}
}
