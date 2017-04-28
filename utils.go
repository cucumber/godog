package godog

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/DATA-DOG/godog/colors"
)

// empty struct value takes no space allocation
type void struct{}

var red = colors.Red
var redb = colors.Bold(colors.Red)
var green = colors.Green
var black = colors.Black
var blackb = colors.Bold(colors.Black)
var yellow = colors.Yellow
var cyan = colors.Cyan
var cyanb = colors.Bold(colors.Cyan)
var whiteb = colors.Bold(colors.White)

// repeats a space n times
func s(n int) string {
	return strings.Repeat(" ", n)
}

// checks the error and exits with error status code
func fatal(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var timeNowFunc = func() time.Time {
	return time.Now()
}
