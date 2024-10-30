package models_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/formatters"
	"github.com/cucumber/godog/internal/models"
	messages "github.com/cucumber/messages/go/v21"
)

type ctxKey string

func TestShouldSupportVoidHandlerReturn(t *testing.T) {
	wasCalled := false
	initialCtx := context.WithValue(context.Background(), ctxKey("original"), 123)

	fn := func(ctx context.Context) {
		wasCalled = true
		assert.Equal(t, 123, ctx.Value(ctxKey("original")))
	}

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{}

	ctx, err := def.Run(initialCtx)
	assert.True(t, wasCalled)
	// ctx is passed thru
	assert.Equal(t, initialCtx, ctx)
	assert.Nil(t, err)

}

func TestShouldSupportNilContextReturn(t *testing.T) {
	initialCtx := context.WithValue(context.Background(), ctxKey("original"), 123)

	wasCalled := false
	fn := func(ctx context.Context) context.Context {
		wasCalled = true
		assert.Equal(t, 123, ctx.Value(ctxKey("original")))

		// nil context is permitted if is single return value
		return nil
	}

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{}
	ctx, err := def.Run(initialCtx)
	assert.True(t, wasCalled)
	// original context is substituted for a nil return value
	//  << JL : IS THIS A BUG? TWO ARG API DOESN'T ALLOW THIS
	assert.Equal(t, initialCtx, ctx)
	assert.Nil(t, err)
}

func TestShouldSupportNilErrorReturn(t *testing.T) {
	initialCtx := context.WithValue(context.Background(), ctxKey("original"), 123)

	wasCalled := false
	fn := func(ctx context.Context) error {
		wasCalled = true
		assert.Equal(t, 123, ctx.Value(ctxKey("original")))

		// nil error is permitted
		return nil
	}

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{}
	ctx, err := def.Run(initialCtx)
	assert.True(t, wasCalled)
	// original context is passed thru if method doesn't return context.
	assert.Equal(t, initialCtx, ctx)
	assert.Nil(t, err)
}

func TestShouldSupportContextReturn(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxKey("original"), 123)

	fn := func(ctx context.Context) context.Context {
		assert.Equal(t, 123, ctx.Value(ctxKey("original")))

		return context.WithValue(ctx, ctxKey("updated"), 321)
	}

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{}
	ctx, err := def.Run(ctx)
	assert.Nil(t, err)
	// converys the context
	assert.Equal(t, 123, ctx.Value(ctxKey("original")))
	assert.Equal(t, 321, ctx.Value(ctxKey("updated")))
}

func TestShouldSupportErrorReturn(t *testing.T) {
	ctx := context.WithValue(context.Background(), ctxKey("original"), 123)
	expectedErr := fmt.Errorf("expected error")

	fn := func(ctx context.Context) error {
		assert.Equal(t, 123, ctx.Value(ctxKey("original")))

		return expectedErr
	}

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{}
	ctx, err := def.Run(ctx)
	// conveys the returned error
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 123, ctx.Value(ctxKey("original")))
}

func TestShouldSupportContextAndErrorReturn(t *testing.T) {

	ctx := context.WithValue(context.Background(), ctxKey("original"), 123)
	expectedErr := fmt.Errorf("expected error")

	fn := func(ctx context.Context) (context.Context, error) {
		assert.Equal(t, 123, ctx.Value(ctxKey("original")))

		return context.WithValue(ctx, ctxKey("updated"), 321), expectedErr
	}

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{}
	ctx, err := def.Run(ctx)
	// conveys error and context
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 123, ctx.Value(ctxKey("original")))
	assert.Equal(t, 321, ctx.Value(ctxKey("updated")))
}

func TestShouldSupportContextAndNilErrorReturn(t *testing.T) {

	ctx := context.WithValue(context.Background(), ctxKey("original"), 123)

	fn := func(ctx context.Context) (context.Context, error) {
		assert.Equal(t, 123, ctx.Value(ctxKey("original")))

		return context.WithValue(ctx, ctxKey("updated"), 321), nil
	}

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{}
	ctx, err := def.Run(ctx)
	// conveys nil error and context
	assert.Nil(t, err)
	assert.Equal(t, 123, ctx.Value(ctxKey("original")))
	assert.Equal(t, 321, ctx.Value(ctxKey("updated")))
}

func TestShouldRejectNilContextWhenMultiValueReturn(t *testing.T) {

	ctx := context.WithValue(context.Background(), ctxKey("original"), 123)

	fn := func(ctx context.Context) (context.Context, error) {
		assert.Equal(t, 123, ctx.Value(ctxKey("original")))

		// nil context is illegal.
		return nil, fmt.Errorf("expected error")
	}

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
			Expr:    regexp.MustCompile("some regex string"),
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{}

	defer func() {
		if e := recover(); e != nil {
			pe := e.(string)
			assert.Equal(t, "step definition 'some regex string' with return type (context.Context, error) must not return <nil> for the context.Context value, step def also returned an error: expected error", pe)
		}
	}()

	def.Run(ctx)

	assert.Fail(t, "should not get here")
}

func TestArgumentCountChecks(t *testing.T) {

	wasCalled := false
	fn := func(a int, b int) {
		wasCalled = true
	}

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{"1"}
	_, err := def.Run(context.Background())
	assert.False(t, wasCalled)
	assert.Equal(t, `func expected more arguments than given: expected 2 arguments, matched 1 from step`, err.(error).Error())
	assert.True(t, errors.Is(err.(error), models.ErrUnmatchedStepArgumentNumber))

	// FIXME - extra args are ignored - but should be reported at runtime
	def.Args = []interface{}{"1", "2", "IGNORED-EXTRA-ARG"}
	_, err = def.Run(context.Background())
	assert.True(t, wasCalled)
	assert.Nil(t, err)
}

func TestShouldSupportIntTypes(t *testing.T) {
	var aActual int64
	var bActual int32
	var cActual int16
	var dActual int8

	fn := func(a int64, b int32, c int16, d int8) {
		aActual = a
		bActual = b
		cActual = c
		dActual = d
	}

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{"1", "2", "3", "4"}
	_, err := def.Run(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, int64(1), aActual)
	assert.Equal(t, int32(2), bActual)
	assert.Equal(t, int16(3), cActual)
	assert.Equal(t, int8(4), dActual)

	// 128 doesn't fit in signed 8bit int
	def.Args = []interface{}{"1", "2", "3", "128"}
	_, err = def.Run(context.Background())
	assert.Equal(t, `cannot convert argument 3: "128" to int8: strconv.ParseInt: parsing "128": value out of range`, err.(error).Error())

	def.Args = []interface{}{"1", "2", "99999", "4"}
	_, err = def.Run(context.Background())
	assert.Equal(t, `cannot convert argument 2: "99999" to int16: strconv.ParseInt: parsing "99999": value out of range`, err.(error).Error())

	def.Args = []interface{}{"1", strings.Repeat("2", 32), "3", "4"}
	_, err = def.Run(context.Background())
	assert.Equal(t, `cannot convert argument 1: "22222222222222222222222222222222" to int32: strconv.ParseInt: parsing "22222222222222222222222222222222": value out of range`, err.(error).Error())

	def.Args = []interface{}{strings.Repeat("1", 32), "2", "3", "4"}
	_, err = def.Run(context.Background())
	assert.Equal(t, `cannot convert argument 0: "11111111111111111111111111111111" to int64: strconv.ParseInt: parsing "11111111111111111111111111111111": value out of range`, err.(error).Error())
}

func TestShouldSupportFloatTypes(t *testing.T) {
	var aActual float64
	var bActual float32
	fn := func(a float64, b float32) {
		aActual = a
		bActual = b
	}

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{"1.1", "2.2"}
	_, err := def.Run(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, float64(1.1), aActual)
	assert.Equal(t, float32(2.2), bActual)

	def.Args = []interface{}{"1.1", strings.Repeat("2", 65) + ".22"}
	_, err = def.Run(context.Background())
	assert.Equal(t, `cannot convert argument 1: "22222222222222222222222222222222222222222222222222222222222222222.22" to float32: strconv.ParseFloat: parsing "22222222222222222222222222222222222222222222222222222222222222222.22": value out of range`, err.(error).Error())
}

func TestShouldSupportGherkinDocstring(t *testing.T) {
	var actualDocString *messages.PickleDocString
	fnDocstring := func(a *messages.PickleDocString) {
		actualDocString = a
	}

	expectedDocString := &messages.PickleDocString{Content: "hello"}
	defDocstring := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fnDocstring,
		},
		HandlerValue: reflect.ValueOf(fnDocstring),
		Args:         []interface{}{expectedDocString},
	}

	_, err := defDocstring.Run(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, expectedDocString, actualDocString)
}

func TestShouldSupportGherkinTable(t *testing.T) {

	var actualTable *messages.PickleTable
	fnTable := func(a *messages.PickleTable) {
		actualTable = a
	}

	expectedTable := &messages.PickleTable{}
	defTable := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fnTable,
		},
		HandlerValue: reflect.ValueOf(fnTable),
		Args:         []interface{}{expectedTable},
	}

	_, err := defTable.Run(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, expectedTable, actualTable)
}

func TestShouldSupportOnlyByteSlice(t *testing.T) {
	var aActual []byte
	fn1 := func(a []byte) {
		aActual = a
	}
	fn2 := func(a []string) {
		assert.Fail(t, "fn2 should not be called")
	}

	def1 := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn1,
		},
		HandlerValue: reflect.ValueOf(fn1),
		Args:         []interface{}{"str"},
	}

	def2 := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn2,
		},
		HandlerValue: reflect.ValueOf(fn2),
		Args:         []interface{}{[]string{}},
	}

	_, err := def1.Run(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []byte{'s', 't', 'r'}, aActual)

	_, err = def2.Run(context.Background())
	assert.Equal(t, `func has unsupported parameter type: the slice parameter 0 type []string is not supported`, err.(error).Error())
	assert.True(t, errors.Is(err.(error), models.ErrUnsupportedParameterType))
}

// this test is superficial compared to the ones above where the actual error messages the user woudl see are verified
func TestStepDefinition_Run_StepArgsShouldBeString(t *testing.T) {
	test := func(t *testing.T, fn interface{}, expectedError string) {
		def := &models.StepDefinition{
			StepDefinition: formatters.StepDefinition{
				Handler: fn,
			},
			HandlerValue: reflect.ValueOf(fn),
		}

		// some value that is not a string
		def.Args = []interface{}{12}

		_, res := def.Run(context.Background())
		if res == nil {
			t.Fatalf("expected a string convertion error, but got none")
		}

		err, ok := res.(error)
		if !ok {
			t.Fatalf("expected a string convertion error, but got %T instead", res)
		}

		if !errors.Is(err, models.ErrCannotConvert) {
			t.Fatalf("expected a string convertion error, but got '%v' instead", err)
		}

		assert.Equal(t, expectedError, err.Error())
	}

	// Ensure step type error if step argument is not a string
	// for all supported types.
	const toStringError = `cannot convert argument 0: "12" of type "int" to string`
	shouldNotBeCalled := func() { assert.Fail(t, "shound not be called") }
	test(t, func(a int) { shouldNotBeCalled() }, toStringError)
	test(t, func(a int64) { shouldNotBeCalled() }, toStringError)
	test(t, func(a int32) { shouldNotBeCalled() }, toStringError)
	test(t, func(a int16) { shouldNotBeCalled() }, toStringError)
	test(t, func(a int8) { shouldNotBeCalled() }, toStringError)
	test(t, func(a string) { shouldNotBeCalled() }, toStringError)
	test(t, func(a float64) { shouldNotBeCalled() }, toStringError)
	test(t, func(a float32) { shouldNotBeCalled() }, toStringError)
	test(t, func(a *godog.Table) { shouldNotBeCalled() }, `cannot convert argument 0: "12" of type "int" to *messages.PickleTable`)
	test(t, func(a *godog.DocString) { shouldNotBeCalled() }, `cannot convert argument 0: "12" of type "int" to *messages.PickleDocString`)
	test(t, func(a []byte) { shouldNotBeCalled() }, toStringError)

}

func TestStepDefinition_Run_InvalidHandlerParamConversion(t *testing.T) {
	test := func(t *testing.T, fn interface{}, expectedError string) {
		def := &models.StepDefinition{
			StepDefinition: formatters.StepDefinition{
				Handler: fn,
			},
			HandlerValue: reflect.ValueOf(fn),
		}

		def.Args = []interface{}{12}

		_, res := def.Run(context.Background())
		if res == nil {
			t.Fatalf("expected an unsupported argument type error, but got none")
		}

		err, ok := res.(error)
		if !ok {
			t.Fatalf("expected an unsupported argument type error, but got %T instead", res)
		}

		if !errors.Is(err, models.ErrUnsupportedParameterType) {
			t.Fatalf("expected an unsupported argument type error, but got '%v' instead", err)
		}

		assert.Equal(t, expectedError, err.Error())
	}

	shouldNotBeCalled := func() { assert.Fail(t, "shound not be called") }

	// Lists some unsupported argument types for step handler.

	// Pointers should work only for godog.Table/godog.DocString
	test(t, func(a *int) { shouldNotBeCalled() }, "func has unsupported parameter type: the data type of parameter 0 type *int is not supported")
	test(t, func(a *int64) { shouldNotBeCalled() }, "func has unsupported parameter type: the data type of parameter 0 type *int64 is not supported")
	test(t, func(a *int32) { shouldNotBeCalled() }, "func has unsupported parameter type: the data type of parameter 0 type *int32 is not supported")
	test(t, func(a *int16) { shouldNotBeCalled() }, "func has unsupported parameter type: the data type of parameter 0 type *int16 is not supported")
	test(t, func(a *int8) { shouldNotBeCalled() }, "func has unsupported parameter type: the data type of parameter 0 type *int8 is not supported")
	test(t, func(a *string) { shouldNotBeCalled() }, "func has unsupported parameter type: the data type of parameter 0 type *string is not supported")
	test(t, func(a *float64) { shouldNotBeCalled() }, "func has unsupported parameter type: the data type of parameter 0 type *float64 is not supported")
	test(t, func(a *float32) { shouldNotBeCalled() }, "func has unsupported parameter type: the data type of parameter 0 type *float32 is not supported")

	// I cannot pass structures
	test(t, func(a godog.Table) { shouldNotBeCalled() }, "func has unsupported parameter type: the struct parameter 0 type messages.PickleTable is not supported")
	test(t, func(a godog.DocString) { shouldNotBeCalled() }, "func has unsupported parameter type: the struct parameter 0 type messages.PickleDocString is not supported")
	test(t, func(a testStruct) { shouldNotBeCalled() }, "func has unsupported parameter type: the struct parameter 0 type models_test.testStruct is not supported")

	// // I cannot use maps
	test(t, func(a map[string]interface{ body() }) { shouldNotBeCalled() }, "func has unsupported parameter type: the parameter 0 type map is not supported")
	test(t, func(a map[string]int) { shouldNotBeCalled() }, "func has unsupported parameter type: the parameter 0 type map is not supported")

	// // Slice works only for byte
	test(t, func(a []int) { shouldNotBeCalled() }, "func has unsupported parameter type: the slice parameter 0 type []int is not supported")
	test(t, func(a []string) { shouldNotBeCalled() }, "func has unsupported parameter type: the slice parameter 0 type []string is not supported")
	test(t, func(a []bool) { shouldNotBeCalled() }, "func has unsupported parameter type: the slice parameter 0 type []bool is not supported")

	// // I cannot use bool
	test(t, func(a bool) { shouldNotBeCalled() }, "func has unsupported parameter type: the parameter 0 type bool is not supported")

}

func TestStepDefinition_Run_StringConversionToFunctionType(t *testing.T) {
	test := func(t *testing.T, fn interface{}, args []interface{}, expectedError string) {
		def := &models.StepDefinition{
			StepDefinition: formatters.StepDefinition{
				Handler: fn,
			},
			HandlerValue: reflect.ValueOf(fn),
			Args:         args,
		}

		_, res := def.Run(context.Background())
		if res == nil {
			t.Fatalf("expected a cannot convert argument type error, but got none")
		}

		err, ok := res.(error)
		if !ok {
			t.Fatalf("expected a cannot convert argument type error, but got %T instead", res)
		}

		if !errors.Is(err, models.ErrCannotConvert) {
			t.Fatalf("expected a cannot convert argument type error, but got '%v' instead", err)
		}

		assert.Equal(t, expectedError, err.Error())
	}

	shouldNotBeCalled := func() { assert.Fail(t, "shound not be called") }

	// Lists some unsupported argument types for step handler.

	// Cannot convert invalid int
	test(t, func(a int) { shouldNotBeCalled() }, []interface{}{"a"}, `cannot convert argument 0: "a" to int: strconv.ParseInt: parsing "a": invalid syntax`)
	test(t, func(a int64) { shouldNotBeCalled() }, []interface{}{"a"}, `cannot convert argument 0: "a" to int64: strconv.ParseInt: parsing "a": invalid syntax`)
	test(t, func(a int32) { shouldNotBeCalled() }, []interface{}{"a"}, `cannot convert argument 0: "a" to int32: strconv.ParseInt: parsing "a": invalid syntax`)
	test(t, func(a int16) { shouldNotBeCalled() }, []interface{}{"a"}, `cannot convert argument 0: "a" to int16: strconv.ParseInt: parsing "a": invalid syntax`)
	test(t, func(a int8) { shouldNotBeCalled() }, []interface{}{"a"}, `cannot convert argument 0: "a" to int8: strconv.ParseInt: parsing "a": invalid syntax`)

	// Cannot convert invalid float
	test(t, func(a float32) { shouldNotBeCalled() }, []interface{}{"a"}, `cannot convert argument 0: "a" to float32: strconv.ParseFloat: parsing "a": invalid syntax`)
	test(t, func(a float64) { shouldNotBeCalled() }, []interface{}{"a"}, `cannot convert argument 0: "a" to float64: strconv.ParseFloat: parsing "a": invalid syntax`)

	// Cannot convert to DataArg
	test(t, func(a *godog.Table) { shouldNotBeCalled() }, []interface{}{"194"}, `cannot convert argument 0: "194" of type "string" to *messages.PickleTable`)

	// Cannot convert to DocString ?
	test(t, func(a *godog.DocString) { shouldNotBeCalled() }, []interface{}{"194"}, `cannot convert argument 0: "194" of type "string" to *messages.PickleDocString`)

}

// @TODO maybe we should support duration
// fn2 := func(err time.Duration) error { return nil }
// def = &models.StepDefinition{Handler: fn2, HandlerValue: reflect.ValueOf(fn2)}

// def.Args = []interface{}{"1"}
// if _, err := def.Run(context.Background()); err == nil {
// 	t.Fatalf("expected an error due to wrong argument type, but got none")
// }

type testStruct struct {
	_ string
}

func TestShouldSupportDocStringToStringConversion(t *testing.T) {
	var aActual string
	fn := func(a string) {
		aActual = a
	}

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
		Args: []interface{}{&messages.PickleDocString{
			Content: "hello",
		}},
	}

	_, err := def.Run(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, "hello", aActual)
}
