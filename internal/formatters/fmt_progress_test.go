package formatters

import (
	"bytes"
	"github.com/cucumber/godog/internal/storage"
	"strings"
	"testing"
)

// TestSummaryWhenLessThanOneRow verifies that Summary() prints the simple
// single-row form (“ <steps>\n”) when the total number of executed steps is
// smaller than StepsPerRow.
func TestSummaryWhenLessThanOneRow(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgress("sample-suite", &buf)

	// Pretend we executed 5 steps (default StepsPerRow is 70).
	*p.Steps = 5

	p.SetStorage(storage.NewStorage())

	p.Summary()

	got := buf.String()
	want := " 5\n" // a single space, the number of steps, newline

	if !strings.HasPrefix(got, want) {
		t.Fatalf("unexpected summary output\nwant prefix: %q\ngot: %q", want, got)
	}
}

// TestSummaryWithMultipleRows verifies the branch that pads with spaces in
// order to align the last line when at least one full row is already printed.
func TestSummaryWithMultipleRows(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgress("sample-suite", &buf)

	// 83 is >70, so one full line (70 chars) would already have been printed
	// during execution; Summary() must therefore pad with 70-(83-70)=57 spaces
	// before the final “ 83\n”.  We only assert on the suffix to avoid
	// hard-wiring the exact amount of padding.
	*p.Steps = p.StepsPerRow + 13 // 83

	p.SetStorage(storage.NewStorage())

	p.Summary()

	got := buf.String()
	stepsIndication := " 83\n"

	if !strings.Contains(got, stepsIndication) {
		t.Fatalf("summary does not end with expected step count\nwant stepsIndication: %q\ngot: %q", stepsIndication, got)
	}
	if !strings.Contains(got, s(57)) {
		t.Fatalf("summary does contains 57 blank spaces before the number of steps")
	}
}

// TestSummaryWhenExactlyOnRowBoundary ensures that nothing is printed by the
// special-case block when the total number of steps is an exact multiple of
// StepsPerRow.  In that situation Summary() should NOT start with a leading
// space-line; instead it should go straight to Base.Summary() (whose exact
// output we purposely ignore here).
func TestSummaryWhenExactlyOnRowBoundary(t *testing.T) {
	var buf bytes.Buffer
	p := NewProgress("sample-suite", &buf)

	p.SetStorage(storage.NewStorage())

	*p.Steps = p.StepsPerRow // 70

	p.Summary()

	got := buf.String()

	// The first character printed by the special block would be a space.
	// Absence of that space tells us the branch was correctly skipped.
	if strings.HasPrefix(got, " ") {
		t.Fatalf("summary should not start with a space when steps %% StepsPerRow == 0; got %q", got)
	}
}
