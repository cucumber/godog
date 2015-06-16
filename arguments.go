package godog

import (
	"fmt"
	"strconv"
)

// Arg is an argument for StepHandler parsed from
// the regexp submatch to handle the step
type Arg string

// Float converts an argument to float64
// or panics if unable to convert it
func (a Arg) Float() float64 {
	v, err := strconv.ParseFloat(string(a), 64)
	if err == nil {
		return v
	}
	panic(fmt.Sprintf(`cannot convert "%s" to float64: %s`, a, err))
}

// Int converts an argument to int64
// or panics if unable to convert it
func (a Arg) Int() int64 {
	v, err := strconv.ParseInt(string(a), 10, 0)
	if err == nil {
		return v
	}
	panic(fmt.Sprintf(`cannot convert "%s" to int64: %s`, a, err))
}

// String converts an argument to string
func (a Arg) String() string {
	return string(a)
}
