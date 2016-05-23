package godog

var builderMainFile = `
package main
import "fmt"
func main() {
	fmt.Println("hello")
}`

var builderTestMainFile = `
package main
import (
	"fmt"
	"testing"
	"os"
)
func TestMain(m *testing.M) {
	fmt.Println("hello")
	os.Exit(0)
}`

var builderPackAliases = `
package main
import (
	"testing"
	a "fmt"
	b "fmt"
)
func TestMain(m *testing.M) {
	a.Println("a")
	b.Println("b")
}`

var builderAnonymousImport = `
package main
import (
	"testing"
	_ "github.com/go-sql-driver/mysql"
)
func TestMain(m *testing.M) {
}`

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

// func builderProcess(src string, t *testing.T) string {
// 	fset := token.NewFileSet()
// 	f, err := parser.ParseFile(fset, "", []byte(builderTestMainFile), 0)
// 	if err != nil {
// 		t.Fatalf("unexpected error while parsing ast: %v", err)
// 	}

// 	deleteTestMainFunc(f)

// 	var buf strings.Buffer
// 	if err := format.Node(&buf, fset, node); err != nil {
// 		return err
// 	}
// }

// func TestShouldCleanTestMainFromSimpleTestFile(t *testing.T) {

// 	b := newBuilderSkel()
// 	err := b.registerMulti([]string{
// 		builderMainFile, builderPackAliases, builderAnonymousImport,
// 	})
// 	if err != nil {
// 		t.Fatalf("unexpected error: %s", err)
// 	}

// 	data, err := b.merge()
// 	if err != nil {
// 		t.Fatalf("unexpected error: %s", err)
// 	}
// 	expected := `package main
// import (
// 	a "fmt"
// 	b "fmt"
// 	"github.com/DATA-DOG/godog"
// 	_ "github.com/go-sql-driver/mysql"
// )
// func main() {
// 	godog.Run(func(suite *godog.Suite) {
// 	})
// }
// func Tester() {
// 	a.Println("a")
// 	b.Println("b")
// }`

// 	actual := string(data)
// 	if b.cleanSpacing(expected) != b.cleanSpacing(actual) {
// 		t.Fatalf("expected output does not match: %s", actual)
// 	}
// }
