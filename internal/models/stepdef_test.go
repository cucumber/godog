package models_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"

	cucumberexpressions "github.com/cucumber/cucumber-expressions/go/v20"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/formatters"
	"github.com/cucumber/godog/internal/models"
	messages "github.com/cucumber/messages/go/v34"
	"github.com/stretchr/testify/assert"
)

type ctxKey string

func parseArgs(t *testing.T, fn interface{}, args ...string) []interface{} {
	t.Helper()

	expression := cucumberexpressions.NewRegularExpression(regexp.MustCompile("(.*)"), cucumberexpressions.NewParameterTypeRegistry())
	paramCount := reflect.TypeOf(fn).NumIn()

	parsed := make([]interface{}, 0, len(args))
	for i, arg := range args {
		argType := reflect.TypeOf("")
		if i < paramCount {
			argType = reflect.TypeOf(fn).In(i)
		}
		match, err := expression.Match(arg, argType)
		if err != nil {
			t.Fatal("failed to match cucumber expression:", err)
		}
		parsed = append(parsed, match[0])
	}
	return parsed
}

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
			Expr:    cucumberexpressions.NewRegularExpression(regexp.MustCompile("some regex string"), cucumberexpressions.NewParameterTypeRegistry()),
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

	def.Args = parseArgs(t, fn, "1")
	_, err := def.Run(context.Background())
	assert.False(t, wasCalled)
	assert.Equal(t, `func expected more arguments than given: expected 2 arguments, matched 1 from step`, err.(error).Error())
	assert.True(t, errors.Is(err.(error), models.ErrUnmatchedStepArgumentNumber))

	// FIXME - extra args are ignored - but should be reported at runtime
	def.Args = parseArgs(t, fn, "1", "2", "IGNORED-EXTRA-ARG")
	_, err = def.Run(context.Background())
	assert.True(t, wasCalled)
	assert.Nil(t, err)
}

func TestArgumentCountChecksWithContext(t *testing.T) {
	wasCalled := false
	fn := func(ctx context.Context, a int, b int) {
		wasCalled = true
	}

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = parseArgs(t, fn, "1")
	_, err := def.Run(context.Background())
	assert.False(t, wasCalled)
	assert.Equal(t, `func expected more arguments than given: expected 2 arguments, matched 1 from step`, err.(error).Error())
	assert.True(t, errors.Is(err.(error), models.ErrUnmatchedStepArgumentNumber))
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

	def.Args = parseArgs(t, fn, "1", "2", "3", "4")
	_, err := def.Run(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, int64(1), aActual)
	assert.Equal(t, int32(2), bActual)
	assert.Equal(t, int16(3), cActual)
	assert.Equal(t, int8(4), dActual)

	// 128 doesn't fit in signed 8bit int
	def.Args = parseArgs(t, fn, "1", "2", "3", "128")
	_, err = def.Run(context.Background())
	assert.Equal(t, `cannot convert argument 3: "128" to int8: strconv.ParseInt: parsing "128": value out of range`, err.(error).Error())

	def.Args = parseArgs(t, fn, "1", "2", "99999", "4")
	_, err = def.Run(context.Background())
	assert.Equal(t, `cannot convert argument 2: "99999" to int16: strconv.ParseInt: parsing "99999": value out of range`, err.(error).Error())

	def.Args = parseArgs(t, fn, "1", strings.Repeat("2", 32), "3", "4")
	_, err = def.Run(context.Background())
	assert.Equal(t, `cannot convert argument 1: "22222222222222222222222222222222" to int32: strconv.ParseInt: parsing "22222222222222222222222222222222": value out of range`, err.(error).Error())

	def.Args = parseArgs(t, fn, strings.Repeat("1", 32), "2", "3", "4")
	_, err = def.Run(context.Background())
	assert.Equal(t, `cannot convert argument 0: "11111111111111111111111111111111" to int64: strconv.ParseInt: parsing "11111111111111111111111111111111": value out of range`, err.(error).Error())
}

func TestShouldSupportUintTypes(t *testing.T) {
	var aActual uint64
	var bActual uint32
	var cActual uint16
	var dActual uint8

	fn := func(a uint64, b uint32, c uint16, d uint8) {
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

	def.Args = parseArgs(t, fn, "1", "2", "3", "4")
	_, err := def.Run(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, uint64(1), aActual)
	assert.Equal(t, uint32(2), bActual)
	assert.Equal(t, uint16(3), cActual)
	assert.Equal(t, uint8(4), dActual)

	// 256 doesn't fit in uint8
	def.Args = parseArgs(t, fn, "1", "2", "3", "256")
	_, err = def.Run(context.Background())
	assert.Equal(t, `cannot convert argument 3: "256" to uint8: strconv.ParseUint: parsing "256": value out of range`, err.(error).Error())

	// 65536 doesn't fit in uint16
	def.Args = parseArgs(t, fn, "1", "2", "65536", "4")
	_, err = def.Run(context.Background())
	assert.Equal(t, `cannot convert argument 2: "65536" to uint16: strconv.ParseUint: parsing "65536": value out of range`, err.(error).Error())

	// 4294967296 too large for uint32
	def.Args = parseArgs(t, fn, "1", "4294967296", "3", "4")
	_, err = def.Run(context.Background())
	assert.Equal(t, `cannot convert argument 1: "4294967296" to uint32: strconv.ParseUint: parsing "4294967296": value out of range`, err.(error).Error())

	// 18446744073709551616 too large for uint64
	def.Args = parseArgs(t, fn, "18446744073709551616", "2", "3", "4")
	_, err = def.Run(context.Background())
	assert.Equal(t, `cannot convert argument 0: "18446744073709551616" to uint64: strconv.ParseUint: parsing "18446744073709551616": value out of range`, err.(error).Error())
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

	def.Args = parseArgs(t, fn, "1.1", "2.2")
	_, err := def.Run(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, float64(1.1), aActual)
	assert.Equal(t, float32(2.2), bActual)

	def.Args = parseArgs(t, fn, "1.1", strings.Repeat("2", 65)+".22")
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

func TestShouldSupportGherkinTableAndDocString(t *testing.T) {
	var actualTable *messages.PickleTable
	var actualDocString *messages.PickleDocString
	fnTable := func(a *messages.PickleTable, b *messages.PickleDocString) {
		actualTable = a
		actualDocString = b
	}

	expectedTable := &messages.PickleTable{}
	expectedDocString := &messages.PickleDocString{Content: "hello"}
	defTable := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fnTable,
		},
		HandlerValue: reflect.ValueOf(fnTable),
		Args:         []interface{}{expectedTable, expectedDocString},
	}

	_, err := defTable.Run(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, expectedTable, actualTable)
	assert.Equal(t, expectedDocString, actualDocString)
}

func TestShouldSupportByteSlice(t *testing.T) {
	var aActual []byte
	fn := func(a []byte) {
		aActual = a
	}

	def1 := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
		Args:         parseArgs(t, fn, "str"),
	}

	_, err := def1.Run(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, []byte{'s', 't', 'r'}, aActual)
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
			t.Fatalf("expected a conversion error, but got none")
		}

		err, ok := res.(error)
		if !ok {
			t.Fatalf("expected a conversion error, but got %T instead", res)
		}

		if !errors.Is(err, models.ErrCannotConvert) {
			t.Fatalf("expected a conversion error, but got '%v' instead", err)
		}

		assert.Equal(t, expectedError, err.Error())
	}

	// Ensure step type error if step argument is not a string
	// for all supported types.
	const toStringError = `cannot convert argument 0: expected a *cucumberexpressions.Argument, but got "int"`
	shouldNotBeCalled := func() { assert.Fail(t, "should not be called") }
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
	test(t, func(a []byte) { shouldNotBeCalled() }, `cannot convert argument 0: "12" of type "int" to []byte`)

}

func TestStepDefinition_Run_CustomParameterConversion(t *testing.T) {
	test := func(t *testing.T, input string, expected interface{}) {
		t.Helper()

		paramType := reflect.TypeOf(expected)
		transformerCalled := false
		parameterType, err := cucumberexpressions.NewParameterType(
			"custom", []*regexp.Regexp{regexp.MustCompile(".*")}, paramType.String(),
			func(args ...*string) interface{} {
				transformerCalled = true
				value := reflect.New(paramType)
				if err := json.Unmarshal([]byte(*args[0]), value.Interface()); err != nil {
					panic(err)
				}
				return value.Elem().Interface()
			}, false, false, false,
		)
		if err != nil {
			t.Fatal("failed to create custom parameter type:", err)
		}

		registry := cucumberexpressions.NewParameterTypeRegistry()
		err = registry.DefineParameterType(parameterType)
		if err != nil {
			t.Fatal("failed to define custom parameter type:", err)
		}

		expression, err := cucumberexpressions.NewCucumberExpression("{custom}", registry)
		if err != nil {
			t.Fatal("failed to create cucumber expression:", err)
		}

		args, err := expression.Match(input)
		if err != nil {
			t.Fatal("failed to match cucumber expression:", err)
		}

		assert.Len(t, args, 1)

		handlerCalled := false
		// Give the handler the exact parameter type to exercise reflective invocation.
		handler := reflect.MakeFunc(reflect.FuncOf([]reflect.Type{paramType}, nil, false), func(args []reflect.Value) []reflect.Value {
			handlerCalled = true
			assert.Equal(t, expected, args[0].Interface())
			return nil
		})

		def := &models.StepDefinition{
			StepDefinition: formatters.StepDefinition{
				Handler: handler.Interface(),
			},
			HandlerValue: handler,
			Args:         []interface{}{args[0]},
		}

		_, res := def.Run(context.Background())
		if res != nil {
			t.Fatal("expected no error, but got:", res)
		}

		assert.True(t, transformerCalled, "custom transformer should be called")
		assert.True(t, handlerCalled, "handler should be called")
	}

	type customStruct struct {
		Value int
	}
	intValue := 12
	stringValue := "hello"

	test(t, "12", &intValue)
	test(t, `"hello"`, &stringValue)
	test(t, `{"Value":12}`, &customStruct{Value: 12})
	test(t, `{"Value":12}`, customStruct{Value: 12})
	test(t, `{"value":12}`, map[string]int{"value": 12})
	test(t, "[1,2]", []int{1, 2})
	test(t, `["hello","world"]`, []string{"hello", "world"})
	test(t, "[true,false]", []bool{true, false})
	test(t, "true", true)
}

func TestStepDefinition_Run_CustomParameterConversionError(t *testing.T) {
	parameterType, err := cucumberexpressions.NewParameterType(
		"custom", []*regexp.Regexp{regexp.MustCompile(".*")}, "bool",
		func(args ...*string) interface{} {
			panic("custom conversion failed")
		}, false, false, false,
	)
	if err != nil {
		t.Fatal("failed to create custom parameter type:", err)
	}

	registry := cucumberexpressions.NewParameterTypeRegistry()
	err = registry.DefineParameterType(parameterType)
	if err != nil {
		t.Fatal("failed to define custom parameter type:", err)
	}

	expression, err := cucumberexpressions.NewCucumberExpression("{custom}", registry)
	if err != nil {
		t.Fatal("failed to create cucumber expression:", err)
	}

	args, err := expression.Match("invalid")
	if err != nil {
		t.Fatal("failed to match expression:", err)
	}

	assert.Len(t, args, 1)

	shouldNotBeCalled := func(value bool) { assert.Fail(t, "should not be called") }

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: shouldNotBeCalled,
		},
		HandlerValue: reflect.ValueOf(shouldNotBeCalled),
		Args:         []interface{}{args[0]},
	}

	_, res := def.Run(context.Background())
	if res == nil {
		t.Fatalf("expected a conversion error, but got none")
	}

	err, ok := res.(error)
	if !ok {
		t.Fatalf("expected a conversion error, but got %T instead", res)
	}

	if !errors.Is(err, models.ErrCannotConvert) {
		t.Fatalf("expected a conversion error, but got '%v' instead", err)
	}

	assert.Equal(t, `cannot convert argument 0: "invalid" to bool: custom conversion failed`, err.Error())
}

func TestStepDefinition_Run_StringConversionToFunctionType(t *testing.T) {
	test := func(t *testing.T, fn interface{}, arg string, expectedError string) {
		def := &models.StepDefinition{
			StepDefinition: formatters.StepDefinition{
				Handler: fn,
			},
			HandlerValue: reflect.ValueOf(fn),
			Args:         parseArgs(t, fn, arg),
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
	test(t, func(a int) { shouldNotBeCalled() }, "a", `cannot convert argument 0: "a" to int: strconv.Atoi: parsing "a": invalid syntax`)
	test(t, func(a int64) { shouldNotBeCalled() }, "a", `cannot convert argument 0: "a" to int64: strconv.ParseInt: parsing "a": invalid syntax`)
	test(t, func(a int32) { shouldNotBeCalled() }, "a", `cannot convert argument 0: "a" to int32: strconv.ParseInt: parsing "a": invalid syntax`)
	test(t, func(a int16) { shouldNotBeCalled() }, "a", `cannot convert argument 0: "a" to int16: strconv.ParseInt: parsing "a": invalid syntax`)
	test(t, func(a int8) { shouldNotBeCalled() }, "a", `cannot convert argument 0: "a" to int8: strconv.ParseInt: parsing "a": invalid syntax`)

	// Cannot convert invalid float
	test(t, func(a float32) { shouldNotBeCalled() }, "a", `cannot convert argument 0: "a" to float32: strconv.ParseFloat: parsing "a": invalid syntax`)
	test(t, func(a float64) { shouldNotBeCalled() }, "a", `cannot convert argument 0: "a" to float64: strconv.ParseFloat: parsing "a": invalid syntax`)

	// Cannot convert to DataArg
	test(t, func(a *godog.Table) { shouldNotBeCalled() }, "194", `cannot convert argument 0: "194" of type "*cucumberexpressions.Argument" to *messages.PickleTable`)

	// Cannot convert to DocString ?
	test(t, func(a *godog.DocString) { shouldNotBeCalled() }, "194", `cannot convert argument 0: "194" of type "*cucumberexpressions.Argument" to *messages.PickleDocString`)

}

// @TODO maybe we should support duration
// fn2 := func(err time.Duration) error { return nil }
// def = &models.StepDefinition{Handler: fn2, HandlerValue: reflect.ValueOf(fn2)}

// def.Args = []interface{}{"1"}
// if _, err := def.Run(context.Background()); err == nil {
// 	t.Fatalf("expected an error due to wrong argument type, but got none")
// }

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
