package utils

import (
	"fmt"
	"strings"
	"time"
)

// S repeats a space n times
func S(n int) string {
	if n < 0 {
		n = 1
	}
	return strings.Repeat(" ", n)
}

// TimeNowFunc is a utility function to simply testing
// by allowing TimeNowFunc to be defined to zero time
// to remove the time domain from tests
var TimeNowFunc = func() time.Time {
	return time.Now()
}

// wrapString wraps a string into chunks of the given width.
func wrapString(s string, width int) []string {
	var result []string
	for len(s) > width {
		result = append(result, s[:width])
		s = s[width:]
	}
	result = append(result, s)
	return result
}

// compareLists compares two lists of strings and prints them with wrapped text.
func VDiffString(expected, actual string) {
	list1 := strings.Split(expected, "\n")
	list2 := strings.Split(actual, "\n")

	VDiffLists(list1, list2)
}

func VDiffLists(list1 []string, list2 []string) {
	// Get the length of the longer list
	maxLength := len(list1)
	if len(list2) > maxLength {
		maxLength = len(list2)
	}

	colWid := 60
	fmtTitle := fmt.Sprintf("%%4s: %%-%ds | %%-%ds\n", colWid+2, colWid+2)
	fmtData := fmt.Sprintf("%%4d: %%-%ds | %%-%ds   %%s\n", colWid+2, colWid+2)

	fmt.Printf(fmtTitle, "#", "expected", "actual")

	for i := 0; i < maxLength; i++ {
		var val1, val2 string

		// Get the value from list1 if it exists
		if i < len(list1) {
			val1 = list1[i]
		} else {
			val1 = "N/A"
		}

		// Get the value from list2 if it exists
		if i < len(list2) {
			val2 = list2[i]
		} else {
			val2 = "N/A"
		}

		// Wrap both strings into slices of strings with fixed width
		wrapped1 := wrapString(val1, colWid)
		wrapped2 := wrapString(val2, colWid)

		// Find the number of wrapped lines needed for the current pair
		maxWrappedLines := len(wrapped1)
		if len(wrapped2) > maxWrappedLines {
			maxWrappedLines = len(wrapped2)
		}

		// Print the wrapped lines with alignment
		for j := 0; j < maxWrappedLines; j++ {
			var line1, line2 string

			// Get the wrapped line or use an empty string if it doesn't exist
			if j < len(wrapped1) {
				line1 = wrapped1[j]
			} else {
				line1 = ""
			}

			if j < len(wrapped2) {
				line2 = wrapped2[j]
			} else {
				line2 = ""
			}

			status := "same"
			// if val1 != val2 {
			if line1 != line2 {
				status = "different"
			}

			delim := "¬"
			// Print the wrapped lines with fixed-width column
			fmt.Printf(fmtData, i+1, delim+line1+delim, delim+line2+delim, status)
		}
	}
}

func TrimAllLines(s string) string {
	var lines []string
	for _, ln := range strings.Split(strings.TrimSpace(s), "\n") {
		lines = append(lines, strings.TrimSpace(ln))
	}
	return strings.Join(lines, "\n")
}
