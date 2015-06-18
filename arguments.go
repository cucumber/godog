package godog

import (
	"fmt"
	"strconv"

	"github.com/DATA-DOG/godog/gherkin"
)

// Arg is an argument for StepHandler parsed from
// the regexp submatch to handle the step
type Arg struct {
	value interface{}
}

// Float64 converts an argument to float64
// or panics if unable to convert it
func (a *Arg) Float64() float64 {
	s, ok := a.value.(string)
	a.must(ok, "string")
	v, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return v
	}
	panic(fmt.Sprintf(`cannot convert "%s" to float64: %s`, s, err))
}

// Float32 converts an argument to float32
// or panics if unable to convert it
func (a *Arg) Float32() float32 {
	s, ok := a.value.(string)
	a.must(ok, "string")
	v, err := strconv.ParseFloat(s, 32)
	if err == nil {
		return float32(v)
	}
	panic(fmt.Sprintf(`cannot convert "%s" to float32: %s`, s, err))
}

// Int converts an argument to int
// or panics if unable to convert it
func (a *Arg) Int() int {
	s, ok := a.value.(string)
	a.must(ok, "string")
	v, err := strconv.ParseInt(s, 10, 0)
	if err == nil {
		return int(v)
	}
	panic(fmt.Sprintf(`cannot convert "%s" to int: %s`, s, err))
}

// Int64 converts an argument to int64
// or panics if unable to convert it
func (a *Arg) Int64() int64 {
	s, ok := a.value.(string)
	a.must(ok, "string")
	v, err := strconv.ParseInt(s, 10, 64)
	if err == nil {
		return v
	}
	panic(fmt.Sprintf(`cannot convert "%s" to int64: %s`, s, err))
}

// Int32 converts an argument to int32
// or panics if unable to convert it
func (a *Arg) Int32() int32 {
	s, ok := a.value.(string)
	a.must(ok, "string")
	v, err := strconv.ParseInt(s, 10, 32)
	if err == nil {
		return int32(v)
	}
	panic(fmt.Sprintf(`cannot convert "%s" to int32: %s`, s, err))
}

// Int16 converts an argument to int16
// or panics if unable to convert it
func (a *Arg) Int16() int16 {
	s, ok := a.value.(string)
	a.must(ok, "string")
	v, err := strconv.ParseInt(s, 10, 16)
	if err == nil {
		return int16(v)
	}
	panic(fmt.Sprintf(`cannot convert "%s" to int16: %s`, s, err))
}

// Int8 converts an argument to int8
// or panics if unable to convert it
func (a *Arg) Int8() int8 {
	s, ok := a.value.(string)
	a.must(ok, "string")
	v, err := strconv.ParseInt(s, 10, 8)
	if err == nil {
		return int8(v)
	}
	panic(fmt.Sprintf(`cannot convert "%s" to int8: %s`, s, err))
}

// String converts an argument to string
func (a *Arg) String() string {
	s, ok := a.value.(string)
	a.must(ok, "string")
	return s
}

// Bytes converts an argument string to bytes
func (a *Arg) Bytes() []byte {
	s, ok := a.value.(string)
	a.must(ok, "string")
	return []byte(s)
}

// PyString converts an argument gherkin PyString node
func (a *Arg) PyString() *gherkin.PyString {
	s, ok := a.value.(*gherkin.PyString)
	a.must(ok, "*gherkin.PyString")
	return s
}

func (a *Arg) must(ok bool, expected string) {
	if !ok {
		panic(fmt.Sprintf(`cannot convert "%v" of type "%T" to type "%s"`, a.value, a.value, expected))
	}
}
