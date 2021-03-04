package colors

import (
	"io"
)

// Colored creates and initializes a new ansiColorWriter
// using io.Writer w as its initial contents.
// In the console of Windows, which change the foreground and background
// colors of the text by the escape sequence.
// In the console of other systems, which writes to w all text.
func Colored(w io.Writer) io.Writer

// Uncolored will accept and io.Writer and return a
// new io.Writer that won't include colors.
func Uncolored(w io.Writer) io.Writer
