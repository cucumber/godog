package godog

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cucumber/messages-go/v10"

	"github.com/cucumber/godog/formatters"
	"github.com/cucumber/godog/internal/models"
	"github.com/cucumber/godog/internal/storage"
	"github.com/cucumber/godog/internal/utils"
)

var errorInterface = reflect.TypeOf((*error)(nil)).Elem()

// ErrUndefined is returned in case if step definition was not found
var ErrUndefined = fmt.Errorf("step is undefined")

// ErrPending should be returned by step definition if
// step implementation is pending
var ErrPending = fmt.Errorf("step implementation is pending")

type suite struct {
	steps []*models.StepDefinition

	fmt     Formatter
	storage *storage.Storage

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

func (s *suite) matchStep(step *messages.Pickle_PickleStep) *models.StepDefinition {
	def := s.matchStepText(step.Text)
	if def != nil && step.Argument != nil {
		def.Args = append(def.Args, step.Argument)
	}
	return def
}

func (s *suite) runStep(pickle *messages.Pickle, step *messages.Pickle_PickleStep, prevStepErr error) (err error) {
	// run before step handlers
	for _, f := range s.beforeStepHandlers {
		f(step)
	}

	match := s.matchStep(step)
	s.storage.MustInsertStepDefintionMatch(step.AstNodeIds[0], match)
	s.fmt.Defined(pickle, step, match.GetInternalStepDefinition())

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

		sr := models.NewStepResult(pickle.Id, step.Id, match)

		switch err {
		case nil:
			sr.Status = models.Passed
			s.storage.MustInsertPickleStepResult(sr)

			s.fmt.Passed(pickle, step, match.GetInternalStepDefinition())
		case ErrPending:
			sr.Status = models.Pending
			s.storage.MustInsertPickleStepResult(sr)

			s.fmt.Pending(pickle, step, match.GetInternalStepDefinition())
		default:
			sr.Status = models.Failed
			sr.Err = err
			s.storage.MustInsertPickleStepResult(sr)

			s.fmt.Failed(pickle, step, match.GetInternalStepDefinition(), err)
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
			match = &models.StepDefinition{
				StepDefinition: formatters.StepDefinition{
					Expr:    match.Expr,
					Handler: match.Handler,
				},
				Args:         match.Args,
				HandlerValue: match.HandlerValue,
				Nested:       match.Nested,
				Undefined:    undef,
			}
		}

		sr := models.NewStepResult(pickle.Id, step.Id, match)
		sr.Status = models.Undefined
		s.storage.MustInsertPickleStepResult(sr)

		s.fmt.Undefined(pickle, step, match.GetInternalStepDefinition())
		return ErrUndefined
	}

	if prevStepErr != nil {
		sr := models.NewStepResult(pickle.Id, step.Id, match)
		sr.Status = models.Skipped
		s.storage.MustInsertPickleStepResult(sr)

		s.fmt.Skipped(pickle, step, match.GetInternalStepDefinition())
		return nil
	}

	err = s.maybeSubSteps(match.Run())
	return
}

func (s *suite) maybeUndefined(text string, arg interface{}) ([]string, error) {
	step := s.matchStepText(text)
	if nil == step {
		return []string{text}, nil
	}

	var undefined []string
	if !step.Nested {
		return undefined, nil
	}

	if arg != nil {
		step.Args = append(step.Args, arg)
	}

	for _, next := range step.Run().(Steps) {
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
		} else if err := s.maybeSubSteps(def.Run()); err != nil {
			return fmt.Errorf("%s: %+v", text, err)
		}
	}
	return nil
}

func (s *suite) matchStepText(text string) *models.StepDefinition {
	for _, h := range s.steps {
		if m := h.Expr.FindStringSubmatch(text); len(m) > 0 {
			var args []interface{}
			for _, m := range m[1:] {
				args = append(args, m)
			}

			// since we need to assign arguments
			// better to copy the step definition
			return &models.StepDefinition{
				StepDefinition: formatters.StepDefinition{
					Expr:    h.Expr,
					Handler: h.Handler,
				},
				Args:         args,
				HandlerValue: h.HandlerValue,
				Nested:       h.Nested,
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
		pr := models.PickleResult{PickleID: pickle.Id, StartedAt: utils.TimeNowFunc()}
		s.storage.MustInsertPickleResult(pr)

		s.fmt.Pickle(pickle)
		return ErrUndefined
	}

	// run before scenario handlers
	for _, f := range s.beforeScenarioHandlers {
		f(pickle)
	}

	pr := models.PickleResult{PickleID: pickle.Id, StartedAt: utils.TimeNowFunc()}
	s.storage.MustInsertPickleResult(pr)

	s.fmt.Pickle(pickle)

	// scenario
	err = s.runSteps(pickle, pickle.Steps)

	// run after scenario handlers
	for _, f := range s.afterScenarioHandlers {
		f(pickle, err)
	}

	return
}
