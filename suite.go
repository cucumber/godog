package godog

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cucumber/messages-go/v10"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()
var typeOfBytes = reflect.TypeOf([]byte(nil))

// ErrUndefined is returned in case if step definition was not found
var ErrUndefined = fmt.Errorf("step is undefined")

// ErrPending should be returned by step definition if
// step implementation is pending
var ErrPending = fmt.Errorf("step implementation is pending")

type suite struct {
	steps []*StepDefinition

	fmt     Formatter
	storage *storage

	failed        bool
	randomSeed    int64
	stopOnFailure bool
	strict        bool

	// suite event handlers
	beforeScenarioHandlers []func(*Scenario)
	beforeStepHandlers     []func(*Step)
	afterStepHandlers      []func(*Step, error)
	afterScenarioHandlers  []func(*Scenario, error)
}

func (s *suite) matchStep(step *messages.Pickle_PickleStep) *StepDefinition {
	def := s.matchStepText(step.Text)
	if def != nil && step.Argument != nil {
		def.args = append(def.args, step.Argument)
	}
	return def
}

func (s *suite) runStep(pickle *messages.Pickle, step *messages.Pickle_PickleStep, prevStepErr error) (err error) {
	// run before step handlers
	for _, f := range s.beforeStepHandlers {
		f(step)
	}

	match := s.matchStep(step)
	s.fmt.Defined(pickle, step, match)

	// user multistep definitions may panic
	defer func() {
		if e := recover(); e != nil {
			err = &traceError{
				msg:   fmt.Sprintf("%v", e),
				stack: callStack(),
			}
		}

		if prevStepErr != nil {
			return
		}

		if err == ErrUndefined {
			return
		}

		sr := newStepResult(pickle.Id, step.Id, match)

		switch err {
		case nil:
			sr.Status = passed
			s.storage.mustInsertPickleStepResult(sr)

			s.fmt.Passed(pickle, step, match)
		case ErrPending:
			sr.Status = pending
			s.storage.mustInsertPickleStepResult(sr)

			s.fmt.Pending(pickle, step, match)
		default:
			sr.Status = failed
			sr.err = err
			s.storage.mustInsertPickleStepResult(sr)

			s.fmt.Failed(pickle, step, match, err)
		}

		// run after step handlers
		for _, f := range s.afterStepHandlers {
			f(step, err)
		}
	}()

	if undef, err := s.maybeUndefined(step.Text, step.Argument); err != nil {
		return err
	} else if len(undef) > 0 {
		if match != nil {
			match = &StepDefinition{
				args:      match.args,
				hv:        match.hv,
				Expr:      match.Expr,
				Handler:   match.Handler,
				nested:    match.nested,
				undefined: undef,
			}
		}

		sr := newStepResult(pickle.Id, step.Id, match)
		sr.Status = undefined
		s.storage.mustInsertPickleStepResult(sr)

		s.fmt.Undefined(pickle, step, match)
		return ErrUndefined
	}

	if prevStepErr != nil {
		sr := newStepResult(pickle.Id, step.Id, match)
		sr.Status = skipped
		s.storage.mustInsertPickleStepResult(sr)

		s.fmt.Skipped(pickle, step, match)
		return nil
	}

	err = s.maybeSubSteps(match.run())
	return
}

func (s *suite) maybeUndefined(text string, arg interface{}) ([]string, error) {
	step := s.matchStepText(text)
	if nil == step {
		return []string{text}, nil
	}

	var undefined []string
	if !step.nested {
		return undefined, nil
	}

	if arg != nil {
		step.args = append(step.args, arg)
	}

	for _, next := range step.run().(Steps) {
		lines := strings.Split(next, "\n")
		// @TODO: we cannot currently parse table or content body from nested steps
		if len(lines) > 1 {
			return undefined, fmt.Errorf("nested steps cannot be multiline and have table or content body argument")
		}
		if len(lines[0]) > 0 && lines[0][len(lines[0])-1] == ':' {
			return undefined, fmt.Errorf("nested steps cannot be multiline and have table or content body argument")
		}
		undef, err := s.maybeUndefined(next, nil)
		if err != nil {
			return undefined, err
		}
		undefined = append(undefined, undef...)
	}
	return undefined, nil
}

func (s *suite) maybeSubSteps(result interface{}) error {
	if nil == result {
		return nil
	}

	if err, ok := result.(error); ok {
		return err
	}

	steps, ok := result.(Steps)
	if !ok {
		return fmt.Errorf("unexpected error, should have been []string: %T - %+v", result, result)
	}

	for _, text := range steps {
		if def := s.matchStepText(text); def == nil {
			return ErrUndefined
		} else if err := s.maybeSubSteps(def.run()); err != nil {
			return fmt.Errorf("%s: %+v", text, err)
		}
	}
	return nil
}

func (s *suite) matchStepText(text string) *StepDefinition {
	for _, h := range s.steps {
		if m := h.Expr.FindStringSubmatch(text); len(m) > 0 {
			var args []interface{}
			for _, m := range m[1:] {
				args = append(args, m)
			}

			// since we need to assign arguments
			// better to copy the step definition
			return &StepDefinition{
				args:    args,
				hv:      h.hv,
				Expr:    h.Expr,
				Handler: h.Handler,
				nested:  h.nested,
			}
		}
	}
	return nil
}

func (s *suite) runSteps(pickle *messages.Pickle, steps []*messages.Pickle_PickleStep) (err error) {
	for _, step := range steps {
		stepErr := s.runStep(pickle, step, err)
		switch stepErr {
		case ErrUndefined:
			// do not overwrite failed error
			if err == ErrUndefined || err == nil {
				err = stepErr
			}
		case ErrPending:
			err = stepErr
		case nil:
		default:
			err = stepErr
		}
	}
	return
}

func (s *suite) shouldFail(err error) bool {
	if err == nil {
		return false
	}

	if err == ErrUndefined || err == ErrPending {
		return s.strict
	}

	return true
}

func isEmptyFeature(pickles []*messages.Pickle) bool {
	for _, pickle := range pickles {
		if len(pickle.Steps) > 0 {
			return false
		}
	}

	return true
}

func (s *suite) runPickle(pickle *messages.Pickle) (err error) {
	if len(pickle.Steps) == 0 {
		pr := pickleResult{PickleID: pickle.Id, StartedAt: timeNowFunc()}
		s.storage.mustInsertPickleResult(pr)

		s.fmt.Pickle(pickle)
		return ErrUndefined
	}

	// run before scenario handlers
	for _, f := range s.beforeScenarioHandlers {
		f(pickle)
	}

	pr := pickleResult{PickleID: pickle.Id, StartedAt: timeNowFunc()}
	s.storage.mustInsertPickleResult(pr)

	s.fmt.Pickle(pickle)

	// scenario
	err = s.runSteps(pickle, pickle.Steps)

	// run after scenario handlers
	for _, f := range s.afterScenarioHandlers {
		f(pickle, err)
	}

	return
}
