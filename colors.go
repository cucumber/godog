package godog

import "fmt"

type color int

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

func cl(s string, c color) string {
	return fmt.Sprintf("\033[%dm%s\033[0m", c, s)
}
