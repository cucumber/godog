package models_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/cucumber/godog/internal/models"
	"github.com/cucumber/messages-go/v10"
)

func TestShouldSupportIntTypes(t *testing.T) {
	fn := func(a int64, b int32, c int16, d int8) error { return nil }

	def := &models.StepDefinition{
		Handler:      fn,
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{"1", "1", "1", "1"}
	if err := def.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	def.Args = []interface{}{"1", "1", "1", strings.Repeat("1", 9)}
	if err := def.Run(); err == nil {
		t.Fatalf("expected convertion fail for int8, but got none")
	}
}

func TestShouldSupportFloatTypes(t *testing.T) {
	fn := func(a float64, b float32) error { return nil }

	def := &models.StepDefinition{
		Handler:      fn,
		HandlerValue: reflect.ValueOf(fn),
	}

	def.Args = []interface{}{"1.1", "1.09"}
	if err := def.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	def.Args = []interface{}{"1.08", strings.Repeat("1", 65) + ".67"}
	if err := def.Run(); err == nil {
		t.Fatalf("expected convertion fail for float32, but got none")
	}
}

func TestShouldNotSupportOtherPointerTypesThanGherkin(t *testing.T) {
	fn1 := func(a *int) error { return nil }
	fn2 := func(a *messages.PickleStepArgument_PickleDocString) error { return nil }
	fn3 := func(a *messages.PickleStepArgument_PickleTable) error { return nil }

	def1 := &models.StepDefinition{Handler: fn1, HandlerValue: reflect.ValueOf(fn1), Args: []interface{}{(*int)(nil)}}
	def2 := &models.StepDefinition{Handler: fn2, HandlerValue: reflect.ValueOf(fn2), Args: []interface{}{&messages.PickleStepArgument_PickleDocString{}}}
	def3 := &models.StepDefinition{Handler: fn3, HandlerValue: reflect.ValueOf(fn3), Args: []interface{}{(*messages.PickleStepArgument_PickleTable)(nil)}}

	if err := def1.Run(); err == nil {
		t.Fatalf("expected conversion error, but got none")
	}
	if err := def2.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := def3.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShouldSupportOnlyByteSlice(t *testing.T) {
	fn1 := func(a []byte) error { return nil }
	fn2 := func(a []string) error { return nil }

	def1 := &models.StepDefinition{Handler: fn1, HandlerValue: reflect.ValueOf(fn1), Args: []interface{}{"str"}}
	def2 := &models.StepDefinition{Handler: fn2, HandlerValue: reflect.ValueOf(fn2), Args: []interface{}{[]string{}}}

	if err := def1.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := def2.Run(); err == nil {
		t.Fatalf("expected conversion error, but got none")
	}
}

func TestUnexpectedArguments(t *testing.T) {
	fn := func(a, b int) error { return nil }
	def := &models.StepDefinition{Handler: fn, HandlerValue: reflect.ValueOf(fn)}

	def.Args = []interface{}{"1"}
	if err := def.Run(); err == nil {
		t.Fatalf("expected an error due to wrong number of arguments, but got none")
	}

	def.Args = []interface{}{"one", "two"}
	if err := def.Run(); err == nil {
		t.Fatalf("expected conversion error, but got none")
	}

	// @TODO maybe we should support duration
	// fn2 := func(err time.Duration) error { return nil }
	// def = &models.StepDefinition{Handler: fn2, HandlerValue: reflect.ValueOf(fn2)}

	// def.Args = []interface{}{"1"}
	// if err := def.Run(); err == nil {
	// 	t.Fatalf("expected an error due to wrong argument type, but got none")
	// }
}
