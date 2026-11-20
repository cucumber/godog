package models

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"

	cucumberexpressions "github.com/cucumber/cucumber-expressions/go/v20"
	messages "github.com/cucumber/messages/go/v34"

	"github.com/cucumber/godog/formatters"
)

var typeOfBytes = reflect.TypeOf([]byte(nil))

// matchable errors
var (
	ErrUnmatchedStepArgumentNumber = errors.New("func expected more arguments than given")
	ErrCannotConvert               = errors.New("cannot convert argument")
)

// StepDefinition ...
type StepDefinition struct {
	formatters.StepDefinition

	Args         []interface{}
	HandlerValue reflect.Value
	File         string
	Line         int

	// multistep related
	Nested    bool
	Undefined []string
}

var typeOfContext = reflect.TypeOf((*context.Context)(nil)).Elem()

// Run a step with the matched arguments using reflect
// Returns one of ...
// (context, error)
// (context, godog.Steps)
func (sd *StepDefinition) Run(ctx context.Context) (context.Context, interface{}) {
	var values []reflect.Value

	typ := sd.HandlerValue.Type()
	numIn := typ.NumIn()
	hasCtxIn := numIn > 0 && typ.In(0).Implements(typeOfContext)
	ctxOffset := 0

	if hasCtxIn {
		values = append(values, reflect.ValueOf(ctx))
		ctxOffset = 1
		numIn--
	}

	if len(sd.Args) < numIn {
		return ctx, fmt.Errorf("%w: expected %d arguments, matched %d from step", ErrUnmatchedStepArgumentNumber, numIn, len(sd.Args))
	}

	for i := 0; i < numIn; i++ {
		param := typ.In(i + ctxOffset)
		switch param.Kind() {
		case reflect.String:
			arg := sd.Args[i]
			switch v := arg.(type) {
			case *messages.PickleStepArgument:
				if v.DocString == nil {
					return ctx, fmt.Errorf(`%w %d: "%v" of type "%T": DocString is not set`, ErrCannotConvert, i, arg, arg)
				}
				values = append(values, reflect.ValueOf(v.DocString.Content).Convert(param))
			case *messages.PickleDocString:
				values = append(values, reflect.ValueOf(v.Content).Convert(param))
			default:
				value, err := sd.shouldBeCucumberArgument(i)
				if err != nil {
					return ctx, err
				}
				values = append(values, reflect.ValueOf(value))
			}
		case reflect.Ptr:
			arg := sd.Args[i]
			switch param.Elem().String() {
			case "messages.PickleDocString":
				switch v := arg.(type) {
				case *messages.PickleStepArgument:
					values = append(values, reflect.ValueOf(v.DocString))
				case *messages.PickleDocString:
					values = append(values, reflect.ValueOf(v))
				case *cucumberexpressions.Argument:
					return ctx, fmt.Errorf(`%w %d: "%v" of type "%T" to *messages.PickleDocString`, ErrCannotConvert, i, *v.Group().Value(), arg)
				default:
					return ctx, fmt.Errorf(`%w %d: "%v" of type "%T" to *messages.PickleDocString`, ErrCannotConvert, i, arg, arg)
				}
			case "messages.PickleTable":
				switch v := arg.(type) {
				case *messages.PickleStepArgument:
					values = append(values, reflect.ValueOf(v.DataTable))
				case *messages.PickleTable:
					values = append(values, reflect.ValueOf(v))
				case *cucumberexpressions.Argument:
					return ctx, fmt.Errorf(`%w %d: "%v" of type "%T" to *messages.PickleTable`, ErrCannotConvert, i, *v.Group().Value(), arg)
				default:
					return ctx, fmt.Errorf(`%w %d: "%v" of type "%T" to *messages.PickleTable`, ErrCannotConvert, i, arg, arg)
				}
			default:
				value, err := sd.shouldBeCucumberArgument(i)
				if err != nil {
					return ctx, err
				}
				values = append(values, reflect.ValueOf(value))
			}
		case reflect.Slice:
			switch param {
			case typeOfBytes:
				arg := sd.Args[i]
				switch v := arg.(type) {
				case *messages.PickleStepArgument:
					if v.DocString == nil {
						return ctx, fmt.Errorf(`%w %d: "%v" of type "%T": DocString is not set`, ErrCannotConvert, i, arg, arg)
					}
					values = append(values, reflect.ValueOf([]byte(v.DocString.Content)).Convert(param))
				case *messages.PickleDocString:
					values = append(values, reflect.ValueOf([]byte(v.Content)).Convert(param))
				case *cucumberexpressions.Argument:
					if v.Group() == nil || v.Group().Value() == nil {
						values = append(values, reflect.ValueOf([]byte{}).Convert(param))
					} else {
						values = append(values, reflect.ValueOf([]byte(*v.Group().Value())).Convert(param))
					}
				default:
					return ctx, fmt.Errorf(`%w %d: "%v" of type "%T" to []byte`, ErrCannotConvert, i, arg, arg)
				}
			default:
				value, err := sd.shouldBeCucumberArgument(i)
				if err != nil {
					return ctx, err
				}
				values = append(values, reflect.ValueOf(value))
			}
		default:
			value, err := sd.shouldBeCucumberArgument(i)
			if err != nil {
				return ctx, err
			}
			values = append(values, reflect.ValueOf(value))
		}
	}

	res := sd.HandlerValue.Call(values)
	if len(res) == 0 {
		return ctx, nil
	}

	// Note that the step fn return types were validated at Initialise in test_context.go stepWithKeyword()

	// single return value may be one of ...
	// error
	// context.Context
	// godog.Steps
	result0 := res[0].Interface()
	if len(res) == 1 {

		// if the single return value is a context then just return it
		if ctx, ok := result0.(context.Context); ok {
			return ctx, nil
		}

		// return type is presumably one of nil, "error" or "Steps" so place it into second return position
		return ctx, result0
	}

	// multi-value value return must be
	//  (context, error) and the context value must not be nil
	if ctx, ok := result0.(context.Context); ok {
		return ctx, res[1].Interface()
	}

	result1 := res[1].Interface()
	errMsg := ""
	if result1 != nil {
		errMsg = fmt.Sprintf(", step def also returned an error: %v", result1)
	}

	text := sd.StepDefinition.Expr.Source()

	if result0 == nil {
		panic(fmt.Sprintf("step definition '%v' with return type (context.Context, error) must not return <nil> for the context.Context value%s", text, errMsg))
	}

	panic(fmt.Errorf("step definition '%v' has return type (context.Context, error), but found %v rather than a context.Context value%s", text, result0, errMsg))
}

func (sd *StepDefinition) shouldBeCucumberArgument(idx int) (value interface{}, err error) {
	arg, ok := sd.Args[idx].(*cucumberexpressions.Argument)
	if !ok {
		return nil, fmt.Errorf(`%w %d: expected a *cucumberexpressions.Argument, but got "%T"`, ErrCannotConvert, idx, sd.Args[idx])
	}

	defer func() {
		r := recover()
		if r != nil {
			value = nil
			err = fmt.Errorf(`%w %d: "%s" to %s: %s`, ErrCannotConvert, idx, *arg.Group().Value(), arg.ParameterType().Type(), r)
		}
	}()

	value = arg.GetValue()

	return
}

// GetInternalStepDefinition ...
func (sd *StepDefinition) GetInternalStepDefinition() *formatters.StepDefinition {
	if sd == nil {
		return nil
	}

	return &sd.StepDefinition
}

func (sd *StepDefinition) TypeHints() []reflect.Type {
	types := slices.Collect(sd.HandlerValue.Type().Ins())
	if len(types) > 0 && types[0].AssignableTo(typeOfContext) {
		types = types[1:]
	}

	return types
}
