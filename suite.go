package godog

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/cucumber/messages-go/v16"

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

	defaultContext context.Context

	// suite event handlers
	beforeScenarioHandlers []BeforeScenarioHook
	beforeStepHandlers     []BeforeStepHook
	afterStepHandlers      []AfterStepHook
	afterScenarioHandlers  []AfterScenarioHook
}

func (s *suite) matchStep(step *messages.PickleStep) *models.StepDefinition {
	def := s.matchStepText(step.Text)
	if def != nil && step.Argument != nil {
		def.Args = append(def.Args, step.Argument)
	}
	return def
}

func (s *suite) runStep(ctx context.Context, pickle *Scenario, step *Step, prevStepErr error) (rctx context.Context, err error) {
	// run before step handlers
	for _, f := range s.beforeStepHandlers {
		ctx, err = f(ctx, step)
		if err != nil {
			return ctx, err
		}
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
			hctx, herr := f(rctx, step, err)

			// Adding hook error to resulting error without breaking hooks loop.
			if herr != nil {
				if err == nil {
					err = herr
				} else {
					err = fmt.Errorf("%v: %w", herr, err)
				}
			}

			rctx = hctx
		}
	}()

	if ctx, undef, err := s.maybeUndefined(ctx, step.Text, step.Argument); err != nil {
		return ctx, err
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
		return ctx, ErrUndefined
	}

	if prevStepErr != nil {
		sr := models.NewStepResult(pickle.Id, step.Id, match)
		sr.Status = models.Skipped
		s.storage.MustInsertPickleStepResult(sr)

		s.fmt.Skipped(pickle, step, match.GetInternalStepDefinition())
		return ctx, nil
	}

	ctx, err = s.maybeSubSteps(match.Run(ctx))

	return ctx, err
}

func (s *suite) maybeUndefined(ctx context.Context, text string, arg interface{}) (context.Context, []string, error) {
	step := s.matchStepText(text)
	if nil == step {
		return ctx, []string{text}, nil
	}

	var undefined []string
	if !step.Nested {
		return ctx, undefined, nil
	}

	if arg != nil {
		step.Args = append(step.Args, arg)
	}

	ctx, steps := step.Run(ctx)

	for _, next := range steps.(Steps) {
		lines := strings.Split(next, "\n")
		// @TODO: we cannot currently parse table or content body from nested steps
		if len(lines) > 1 {
			return ctx, undefined, fmt.Errorf("nested steps cannot be multiline and have table or content body argument")
		}
		if len(lines[0]) > 0 && lines[0][len(lines[0])-1] == ':' {
			return ctx, undefined, fmt.Errorf("nested steps cannot be multiline and have table or content body argument")
		}
		ctx, undef, err := s.maybeUndefined(ctx, next, nil)
		if err != nil {
			return ctx, undefined, err
		}
		undefined = append(undefined, undef...)
	}
	return ctx, undefined, nil
}

func (s *suite) maybeSubSteps(ctx context.Context, result interface{}) (context.Context, error) {
	if nil == result {
		return ctx, nil
	}

	if err, ok := result.(error); ok {
		return ctx, err
	}

	steps, ok := result.(Steps)
	if !ok {
		return ctx, fmt.Errorf("unexpected error, should have been []string: %T - %+v", result, result)
	}

	for _, text := range steps {
		if def := s.matchStepText(text); def == nil {
			return ctx, ErrUndefined
		} else if ctx, err := s.maybeSubSteps(def.Run(ctx)); err != nil {
			return ctx, fmt.Errorf("%s: %+v", text, err)
		}
	}
	return ctx, nil
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

func (s *suite) runSteps(ctx context.Context, pickle *Scenario, steps []*Step) (context.Context, error) {
	var (
		stepErr, err error
	)

	for _, step := range steps {
		ctx, stepErr = s.runStep(ctx, pickle, step, err)
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

	return ctx, err
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
	ctx := s.defaultContext
	if ctx == nil {
		ctx = context.Background()
	}

	if len(pickle.Steps) == 0 {
		pr := models.PickleResult{PickleID: pickle.Id, StartedAt: utils.TimeNowFunc()}
		s.storage.MustInsertPickleResult(pr)

		s.fmt.Pickle(pickle)
		return ErrUndefined
	}

	// run before scenario handlers
	for _, f := range s.beforeScenarioHandlers {
		ctx, err = f(ctx, pickle)
		if err != nil {
			return err
		}
	}

	pr := models.PickleResult{PickleID: pickle.Id, StartedAt: utils.TimeNowFunc()}
	s.storage.MustInsertPickleResult(pr)

	s.fmt.Pickle(pickle)

	// scenario
	ctx, err = s.runSteps(ctx, pickle, pickle.Steps)

	// run after scenario handlers
	for _, f := range s.afterScenarioHandlers {
		hctx, herr := f(ctx, pickle, err)

		// Adding hook error to resulting error without breaking hooks loop.
		if herr != nil {
			if err == nil {
				err = herr
			} else {
				err = fmt.Errorf("%v: %w", herr, err)
			}
		}

		ctx = hctx
	}

	return err
}
