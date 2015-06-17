package godog

import (
	"fmt"
	"strconv"
)

// Arg is an argument for StepHandler parsed from
// the regexp submatch to handle the step
type Arg string

// Float64 converts an argument to float64
// or panics if unable to convert it
func (a Arg) Float64() float64 {
	v, err := strconv.ParseFloat(string(a), 64)
	if err == nil {
		return v
	}
	panic(fmt.Sprintf(`cannot convert "%s" to float64: %s`, a, err))
}

// Float32 converts an argument to float32
// or panics if unable to convert it
func (a Arg) Float32() float32 {
	v, err := strconv.ParseFloat(string(a), 32)
	if err == nil {
		return float32(v)
	}
	panic(fmt.Sprintf(`cannot convert "%s" to float32: %s`, a, err))
}

// Int converts an argument to int
// or panics if unable to convert it
func (a Arg) Int() int {
	v, err := strconv.ParseInt(string(a), 10, 0)
	if err == nil {
		return int(v)
	}
	panic(fmt.Sprintf(`cannot convert "%s" to int: %s`, a, err))
}

// Int64 converts an argument to int64
// or panics if unable to convert it
func (a Arg) Int64() int64 {
	v, err := strconv.ParseInt(string(a), 10, 64)
	if err == nil {
		return v
	}
	panic(fmt.Sprintf(`cannot convert "%s" to int64: %s`, a, err))
}

// Int32 converts an argument to int32
// or panics if unable to convert it
func (a Arg) Int32() int32 {
	v, err := strconv.ParseInt(string(a), 10, 32)
	if err == nil {
		return int32(v)
	}
	panic(fmt.Sprintf(`cannot convert "%s" to int32: %s`, a, err))
}

// Int16 converts an argument to int16
// or panics if unable to convert it
func (a Arg) Int16() int16 {
	v, err := strconv.ParseInt(string(a), 10, 16)
	if err == nil {
		return int16(v)
	}
	panic(fmt.Sprintf(`cannot convert "%s" to int16: %s`, a, err))
}

// Int8 converts an argument to int8
// or panics if unable to convert it
func (a Arg) Int8() int8 {
	v, err := strconv.ParseInt(string(a), 10, 8)
	if err == nil {
		return int8(v)
	}
	panic(fmt.Sprintf(`cannot convert "%s" to int8: %s`, a, err))
}

// String converts an argument to string
func (a Arg) String() string {
	return string(a)
}

// Bytes converts an argument string to bytes
func (a Arg) Bytes() []byte {
	return []byte(a)
}
