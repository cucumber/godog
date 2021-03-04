package formatters

import (
	"io"

	"github.com/cucumber/godog/formatters"
	"github.com/cucumber/godog/internal/storage"
)

// BaseFormatterFunc implements the FormatterFunc for the base formatter
func BaseFormatterFunc(suite string, out io.Writer) formatters.Formatter

// NewBaseFmt creates a new base formatter
func NewBaseFmt(suite string, out io.Writer) *Basefmt

// Basefmt ...
type Basefmt struct{}

// SetStorage ...
func (f *Basefmt) SetStorage(st *storage.Storage)

// TestRunStarted ...
func (f *Basefmt) TestRunStarted()

// Feature ...
func (f *Basefmt) Feature(featureURI string)

// Pickle ...
func (f *Basefmt) Pickle(pickleID string)

// Defined ...
func (f *Basefmt) Defined(pickleStepID string)

// Passed ...
func (f *Basefmt) Passed(pickleStepID string)

// Skipped ...
func (f *Basefmt) Skipped(pickleStepID string)

// Undefined ...
func (f *Basefmt) Undefined(pickleStepID string)

// Failed ...
func (f *Basefmt) Failed(pickleStepID string, err error)

// Pending ...
func (f *Basefmt) Pending(pickleStepID string)

// Summary ...
func (f *Basefmt) Summary()
