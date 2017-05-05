package godog

import (
	"fmt"
	"strings"
)

// ErrUndefined is returned in case if step definition was not found
var ErrUndefined = fmt.Errorf("step is undefined")

// ErrPending should be returned by step definition if
// step implementation is pending
var ErrPending = fmt.Errorf("step implementation is pending")

// errFailure is used to wrap all godog errors
// apart from ErrUndefined or ErrPending
type errFailure struct {
	main        error
	attachments []string // @TODO: may support media types
}

// Attach some additional context to an error message
func (e *errFailure) Attach(s string) {
	e.attachments = append(e.attachments, strings.TrimSpace(s))
}

func (e *errFailure) ErrorIndent(spaces int) string {
	// @TODO: may filter only plain/text media type based attachments
	all := strings.Join(append([]string{fmt.Sprintf("%+v", e.main)}, e.attachments...), "\n\n")
	lines := strings.Split(all, "\n")

	var indented []string
	for i, ln := range lines {
		// strip trailing spaces
		ln = strings.TrimRight(ln, " ")

		// do not indent first or empty line
		if len(ln) != 0 && i != 0 {
			ln = s(spaces) + ln
		}
		indented = append(indented, ln)
	}
	return strings.Join(indented, "\n")
}

func (e *errFailure) Error() string {
	return e.ErrorIndent(0)
}
