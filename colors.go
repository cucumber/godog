package godog

import "fmt"

type color int

const ansiEscape = "\x1b"

const (
	black color = iota + 30
	red
	green
	yellow
	blue
	magenta
	cyan
	white
)

func cl(s interface{}, c color) string {
	return fmt.Sprintf("%s[%dm%v%s[0m", ansiEscape, c, s, ansiEscape)
}

func bcl(s interface{}, c color) string {
	return fmt.Sprintf("%s[1;%dm%v%s[0m", ansiEscape, c, s, ansiEscape)
}
