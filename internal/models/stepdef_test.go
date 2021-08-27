package models_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cucumber/godog/internal/messages"
	"github.com/stretchr/testify/assert"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/formatters"
	"github.com/cucumber/godog/internal/models"
)

func TestShouldSupportContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), "original", 123)

	fn := func(ctx context.Context, a int64, b int32, c int16, d int8) context.Context {
		assert.Equal(t, 123, ctx.Value("original"))

		return context.WithValue(ctx, "updated", 321)
	}

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{"1", "1", "1", "1"}
	ctx, err := def.Run(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 123, ctx.Value("original"))
	assert.Equal(t, 321, ctx.Value("updated"))
}

func TestShouldSupportContextAndError(t *testing.T) {
	ctx := context.WithValue(context.Background(), "original", 123)

	fn := func(ctx context.Context, a int64, b int32, c int16, d int8) (context.Context, error) {
		assert.Equal(t, 123, ctx.Value("original"))

		return context.WithValue(ctx, "updated", 321), nil
	}

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{"1", "1", "1", "1"}
	ctx, err := def.Run(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 123, ctx.Value("original"))
	assert.Equal(t, 321, ctx.Value("updated"))
}

func TestShouldSupportEmptyHandlerReturn(t *testing.T) {
	fn := func(a int64, b int32, c int16, d int8) {}

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{"1", "1", "1", "1"}
	if _, err := def.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	def.Args = []interface{}{"1", "1", "1", strings.Repeat("1", 9)}
	if _, err := def.Run(context.Background()); err == nil {
		t.Fatalf("expected convertion fail for int8, but got none")
	}
}

func TestShouldSupportIntTypes(t *testing.T) {
	fn := func(a int64, b int32, c int16, d int8) error { return nil }

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{"1", "1", "1", "1"}
	if _, err := def.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	def.Args = []interface{}{"1", "1", "1", strings.Repeat("1", 9)}
	if _, err := def.Run(context.Background()); err == nil {
		t.Fatalf("expected convertion fail for int8, but got none")
	}
}

func TestShouldSupportFloatTypes(t *testing.T) {
	fn := func(a float64, b float32) error { return nil }

	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{"1.1", "1.09"}
	if _, err := def.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	def.Args = []interface{}{"1.08", strings.Repeat("1", 65) + ".67"}
	if _, err := def.Run(context.Background()); err == nil {
		t.Fatalf("expected convertion fail for float32, but got none")
	}
}

func TestShouldNotSupportOtherPointerTypesThanGherkin(t *testing.T) {
	fn1 := func(a *int) error { return nil }
	fn2 := func(a *messages.PickleDocString) error { return nil }
	fn3 := func(a *messages.PickleTable) error { return nil }

	def1 := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn1,
		},
		HandlerValue: reflect.ValueOf(fn1),
		Args:         []interface{}{(*int)(nil)},
	}
	def2 := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn2,
		},
		HandlerValue: reflect.ValueOf(fn2),
		Args:         []interface{}{&messages.PickleDocString{}},
	}
	def3 := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn3,
		},
		HandlerValue: reflect.ValueOf(fn3),
		Args:         []interface{}{(*messages.PickleTable)(nil)},
	}

	if _, err := def1.Run(context.Background()); err == nil {
		t.Fatalf("expected conversion error, but got none")
	}

	if _, err := def2.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := def3.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShouldSupportOnlyByteSlice(t *testing.T) {
	fn1 := func(a []byte) error { return nil }
	fn2 := func(a []string) error { return nil }

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

	if _, err := def1.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := def2.Run(context.Background()); err == nil {
		t.Fatalf("expected conversion error, but got none")
	}
}

func TestUnexpectedArguments(t *testing.T) {
	fn := func(a, b int) error { return nil }
	def := &models.StepDefinition{
		StepDefinition: formatters.StepDefinition{
			Handler: fn,
		},
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{"1"}

	_, res := def.Run(context.Background())
	if res == nil {
		t.Fatalf("expected an error due to wrong number of arguments, but got none")
	}

	err, ok := res.(error)
	if !ok {
		t.Fatalf("expected an error due to wrong number of arguments, but got %T instead", res)
	}

	if !errors.Is(err, models.ErrUnmatchedStepArgumentNumber) {
		t.Fatalf("expected an error due to wrong number of arguments, but got %v instead", err)
	}
}

func TestStepDefinition_Run_StepShouldBeString(t *testing.T) {
	test := func(t *testing.T, fn interface{}) {
		def := &models.StepDefinition{
			StepDefinition: formatters.StepDefinition{
				Handler: fn,
			},
			HandlerValue: reflect.ValueOf(fn),
		}

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
	}

	// Ensure step type error if step argument is not a string
	// for all supported types.
	test(t, func(a int) error { return nil })
	test(t, func(a int64) error { return nil })
	test(t, func(a int32) error { return nil })
	test(t, func(a int16) error { return nil })
	test(t, func(a int8) error { return nil })
	test(t, func(a string) error { return nil })
	test(t, func(a float64) error { return nil })
	test(t, func(a float32) error { return nil })
	test(t, func(a *godog.Table) error { return nil })
	test(t, func(a *godog.DocString) error { return nil })
	test(t, func(a []byte) error { return nil })

}

func TestStepDefinition_Run_InvalidHandlerParamConversion(t *testing.T) {
	test := func(t *testing.T, fn interface{}) {
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

		if !errors.Is(err, models.ErrUnsupportedArgumentType) {
			t.Fatalf("expected an unsupported argument type error, but got '%v' instead", err)
		}
	}

	// Lists some unsupported argument types for step handler.

	// Pointers should work only for godog.Table/godog.DocString
	test(t, func(a *int) error { return nil })
	test(t, func(a *int64) error { return nil })
	test(t, func(a *int32) error { return nil })
	test(t, func(a *int16) error { return nil })
	test(t, func(a *int8) error { return nil })
	test(t, func(a *string) error { return nil })
	test(t, func(a *float64) error { return nil })
	test(t, func(a *float32) error { return nil })

	// I cannot pass structures
	test(t, func(a godog.Table) error { return nil })
	test(t, func(a godog.DocString) error { return nil })
	test(t, func(a testStruct) error { return nil })

	// I cannot use maps
	test(t, func(a map[string]interface{}) error { return nil })
	test(t, func(a map[string]int) error { return nil })

	// Slice works only for byte
	test(t, func(a []int) error { return nil })
	test(t, func(a []string) error { return nil })
	test(t, func(a []bool) error { return nil })

	// I cannot use bool
	test(t, func(a bool) error { return nil })

}

func TestStepDefinition_Run_StringConversionToFunctionType(t *testing.T) {
	test := func(t *testing.T, fn interface{}, args []interface{}) {
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
	}

	// Lists some unsupported argument types for step handler.

	// Cannot convert invalid int
	test(t, func(a int) error { return nil }, []interface{}{"a"})
	test(t, func(a int64) error { return nil }, []interface{}{"a"})
	test(t, func(a int32) error { return nil }, []interface{}{"a"})
	test(t, func(a int16) error { return nil }, []interface{}{"a"})
	test(t, func(a int8) error { return nil }, []interface{}{"a"})

	// Cannot convert invalid float
	test(t, func(a float32) error { return nil }, []interface{}{"a"})
	test(t, func(a float64) error { return nil }, []interface{}{"a"})

	// Cannot convert to DataArg
	test(t, func(a *godog.Table) error { return nil }, []interface{}{"194"})

	// Cannot convert to DocString ?
	test(t, func(a *godog.DocString) error { return nil }, []interface{}{"194"})

}

// @TODO maybe we should support duration
// fn2 := func(err time.Duration) error { return nil }
// def = &models.StepDefinition{Handler: fn2, HandlerValue: reflect.ValueOf(fn2)}

// def.Args = []interface{}{"1"}
// if _, err := def.Run(context.Background()); err == nil {
// 	t.Fatalf("expected an error due to wrong argument type, but got none")
// }

type testStruct struct {
	a string
}

func TestShouldSupportDocStringToStringConversion(t *testing.T) {
	fn := func(a string) error {
		if a != "hello" {
			return fmt.Errorf("did not get hello")
		}
		return nil
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

	if _, err := def.Run(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
