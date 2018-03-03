package godog

import (
	"strconv"
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

var timeNowFunc = func() time.Time {
	return time.Now()
}

// Code stolen here: https://stackoverflow.com/questions/18409373/how-to-compare-two-version-number-strings-in-golang
// This avoid to import an external lib
// Return true if v1 is equal or greater than v2
func isVersionGreaterOrEqual(v1, v2 string) (bool, error) {
	var ret int
	as := strings.Split(v1, ".")
	bs := strings.Split(v2, ".")
	loopMax := len(bs)
	if len(as) > len(bs) {
		loopMax = len(as)
	}
	for i := 0; i < loopMax; i++ {
		var x, y string
		if len(as) > i {
			x = as[i]
		}
		if len(bs) > i {
			y = bs[i]
		}
		xi, err := strconv.Atoi(x)

		if err != nil {
			return false, err
		}

		yi, err := strconv.Atoi(y)

		if err != nil {
			return false, err
		}
		if xi > yi {
			ret = -1
		} else if xi < yi {
			ret = 1
		}
		if ret != 0 {
			break
		}
	}
	return (ret == -1 || ret == 0), nil
}
