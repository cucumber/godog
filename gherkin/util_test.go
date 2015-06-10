package gherkin

import "strings"

func indent(n int, s string) string {
	return strings.Repeat(" ", n) + s
}
