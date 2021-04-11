package godog

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/cucumber/godog/colors"
	"github.com/cucumber/godog/formatters"
	internal_fmt "github.com/cucumber/godog/internal/formatters"
	"github.com/cucumber/godog/internal/models"
	"github.com/cucumber/godog/internal/storage"
)

var (
	Passed    = models.Passed
	Failed    = models.Failed
	Skipped   = models.Skipped
	Undefined = models.Undefined
	Pending   = models.Pending
)

// FindFmt searches available formatters registered
// and returns FormaterFunc matched by given
// format name or nil otherwise
func FindFmt(name string) FormatterFunc {
	return formatters.FindFmt(name)
}

// Format registers a feature suite output
// formatter by given name, description and
// FormatterFunc constructor function, to initialize
// formatter with the output recorder.
func Format(name, description string, f FormatterFunc) {
	formatters.Format(name, description, f)
}

// AvailableFormatters gives a map of all
// formatters registered with their name as key
// and description as value
func AvailableFormatters() map[string]string {
	return formatters.AvailableFormatters()
}

// Formatter is an interface for feature runner
// output summary presentation.
//
// New formatters may be created to represent
// suite results in different ways. These new
// formatters needs to be registered with a
// godog.Format function call
type Formatter = formatters.Formatter

type storageFormatter interface {
	SetStorage(*storage.Storage)
}

// FormatterFunc builds a formatter with given
// suite name and io.Writer to record output
type FormatterFunc = formatters.FormatterFunc

func printStepDefinitions(steps []*models.StepDefinition, w io.Writer) {
	var longest int
	for _, def := range steps {
		n := utf8.RuneCountInString(def.Expr.String())
		if longest < n {
			longest = n
		}
	}

	for _, def := range steps {
		n := utf8.RuneCountInString(def.Expr.String())
		location := internal_fmt.DefinitionID(def)
		spaces := strings.Repeat(" ", longest-n)
		fmt.Fprintln(w,
			colors.Yellow(def.Expr.String())+spaces,
			colors.Bold(colors.Black)("# "+location))
	}

	if len(steps) == 0 {
		fmt.Fprintln(w, "there were no contexts registered, could not find any step definition..")
	}
}

func NewBasefmt(suite string, out io.Writer) *Basefmt {
	return &Basefmt{
		Basefmt: internal_fmt.NewBaseFmt(suite, out),
	}
}

func NewProgressfmt(suite string, out io.Writer) *Progress {
	return &Progress{
		Progress: internal_fmt.NewProgressfmt(suite, out),
	}
}

type Basefmt struct {
	*internal_fmt.Basefmt
}

type Progress struct {
	*internal_fmt.Progress
}
