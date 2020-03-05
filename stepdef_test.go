package godog

import (
	"reflect"
	"strings"
	"testing"

	"github.com/cucumber/messages-go/v9"
)

func TestShouldSupportIntTypes(t *testing.T) {
	fn := func(a int64, b int32, c int16, d int8) error { return nil }

	def := &StepDefinition{
		Handler: fn,
		hv:      reflect.ValueOf(fn),
	}

	def.args = []interface{}{"1", "1", "1", "1"}
	if err := def.run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	def.args = []interface{}{"1", "1", "1", strings.Repeat("1", 9)}
	if err := def.run(); err == nil {
		t.Fatalf("expected convertion fail for int8, but got none")
	}
}

func TestShouldSupportFloatTypes(t *testing.T) {
	fn := func(a float64, b float32) error { return nil }

	def := &StepDefinition{
		Handler: fn,
		hv:      reflect.ValueOf(fn),
	}

	def.args = []interface{}{"1.1", "1.09"}
	if err := def.run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	def.args = []interface{}{"1.08", strings.Repeat("1", 65) + ".67"}
	if err := def.run(); err == nil {
		t.Fatalf("expected convertion fail for float32, but got none")
	}
}

func TestShouldNotSupportOtherPointerTypesThanGherkin(t *testing.T) {
	fn1 := func(a *int) error { return nil }
	fn2 := func(a *messages.PickleStepArgument_PickleDocString) error { return nil }
	fn3 := func(a *messages.PickleStepArgument_PickleTable) error { return nil }

	def1 := &StepDefinition{Handler: fn1, hv: reflect.ValueOf(fn1), args: []interface{}{(*int)(nil)}}
	def2 := &StepDefinition{Handler: fn2, hv: reflect.ValueOf(fn2), args: []interface{}{&messages.PickleStepArgument_PickleDocString{}}}
	def3 := &StepDefinition{Handler: fn3, hv: reflect.ValueOf(fn3), args: []interface{}{(*messages.PickleStepArgument_PickleTable)(nil)}}

	if err := def1.run(); err == nil {
		t.Fatalf("expected conversion error, but got none")
	}
	if err := def2.run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := def3.run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestShouldSupportOnlyByteSlice(t *testing.T) {
	fn1 := func(a []byte) error { return nil }
	fn2 := func(a []string) error { return nil }

	def1 := &StepDefinition{Handler: fn1, hv: reflect.ValueOf(fn1), args: []interface{}{"str"}}
	def2 := &StepDefinition{Handler: fn2, hv: reflect.ValueOf(fn2), args: []interface{}{[]string{}}}

	if err := def1.run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := def2.run(); err == nil {
		t.Fatalf("expected conversion error, but got none")
	}
}

func TestUnexpectedArguments(t *testing.T) {
	fn := func(a, b int) error { return nil }
	def := &StepDefinition{Handler: fn, hv: reflect.ValueOf(fn)}

	def.args = []interface{}{"1"}
	if err := def.run(); err == nil {
		t.Fatalf("expected an error due to wrong number of arguments, but got none")
	}

	def.args = []interface{}{"one", "two"}
	if err := def.run(); err == nil {
		t.Fatalf("expected conversion error, but got none")
	}

	// @TODO maybe we should support duration
	// fn2 := func(err time.Duration) error { return nil }
	// def = &StepDefinition{Handler: fn2, hv: reflect.ValueOf(fn2)}

	// def.args = []interface{}{"1"}
	// if err := def.run(); err == nil {
	// 	t.Fatalf("expected an error due to wrong argument type, but got none")
	// }
}
