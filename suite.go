package godog

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/cucumber/messages-go/v16"

	"github.com/cucumber/godog/formatters"
	"github.com/cucumber/godog/internal/models"
	"github.com/cucumber/godog/internal/storage"
	"github.com/cucumber/godog/internal/utils"
)

var (
	errorInterface   = reflect.TypeOf((*error)(nil)).Elem()
	contextInterface = reflect.TypeOf((*context.Context)(nil)).Elem()
)

// ErrUndefined is returned in case if step definition was not found
var ErrUndefined = fmt.Errorf("step is undefined")

// ErrPending should be returned by step definition if
// step implementation is pending
var ErrPending = fmt.Errorf("step implementation is pending")

// StepResultStatus describes step result.
type StepResultStatus = models.StepResultStatus

const (
	// StepPassed indicates step that passed.
	StepPassed StepResultStatus = models.Passed
	// StepFailed indicates step that failed.
	StepFailed = models.Failed
	// StepSkipped indicates step that was skipped.
	StepSkipped = models.Skipped
	// StepUndefined indicates undefined step.
	StepUndefined = models.Undefined
	// StepPending indicates step with pending implementation.
	StepPending = models.Pending
)

type testSuite struct {
	stepDefs []*models.StepDefinition

	fmt     Formatter
	storage *storage.Storage

	failed        bool
	randomSeed    int64
	stopOnFailure bool
	strict        bool

	defaultContext context.Context
	testingT       *testing.T

	// suite event handlers
	beforeScenarioHandlers []BeforeScenarioHook
	beforeStepHandlers     []BeforeStepHook
	afterStepHandlers      []AfterStepHook
	afterScenarioHandlers  []AfterScenarioHook
}

func (suite *testSuite) matchStep(step *messages.PickleStep) *models.StepDefinition {
	stepDef := suite.matchStepText(step.Text)
	if stepDef != nil && step.Argument != nil {
		stepDef.Args = append(stepDef.Args, step.Argument)
	}
	return stepDef
}

func (suite *testSuite) runStep(ctx context.Context, pickle *Scenario, step *Step, prevStepErr error, isFirst, isLast bool) (rctx context.Context, err error) {
	var (
		stepDef    *models.StepDefinition
		stepResult = models.PickleStepResult{Status: models.Undefined}
	)

	rctx = ctx

	// user multistep definitions may panic
	defer func() {
		if e := recover(); e != nil {
			err = &traceError{
				msg:   fmt.Sprintf("%v", e),
				stack: callStack(),
			}
		}

		earlyReturn := prevStepErr != nil || err == ErrUndefined

		if !earlyReturn {
			stepResult = models.NewStepResult(pickle.Id, step.Id, stepDef)
		}

		// Run after step handlers.
		rctx, err = suite.runAfterStepHooks(ctx, step, stepResult.Status, err)

		// Trigger after scenario on failing or last step to attach possible hook error to step.
		if stepResult.Status != StepSkipped && ((err == nil && isLast) || err != nil) {
			rctx, err = suite.runAfterScenarioHooks(rctx, pickle, err)
		}

		if earlyReturn {
			return
		}

		switch err {
		case nil:
			stepResult.Status = models.Passed
			suite.storage.MustInsertPickleStepResult(stepResult)

			suite.fmt.Passed(pickle, step, stepDef.GetInternalStepDefinition())
		case ErrPending:
			stepResult.Status = models.Pending
			suite.storage.MustInsertPickleStepResult(stepResult)

			suite.fmt.Pending(pickle, step, stepDef.GetInternalStepDefinition())
		default:
			stepResult.Status = models.Failed
			stepResult.Err = err
			suite.storage.MustInsertPickleStepResult(stepResult)

			suite.fmt.Failed(pickle, step, stepDef.GetInternalStepDefinition(), err)
		}
	}()

	// run before scenario handlers
	if isFirst {
		ctx, err = suite.runBeforeScenarioHooks(ctx, pickle)
	}

	// run before step handlers
	ctx, err = suite.runBeforeStepHooks(ctx, step, err)

	stepDef = suite.matchStep(step)
	suite.storage.MustInsertStepDefintionMatch(step.AstNodeIds[0], stepDef)
	suite.fmt.Defined(pickle, step, stepDef.GetInternalStepDefinition())

	if err != nil {
		stepResult = models.NewStepResult(pickle.Id, step.Id, stepDef)
		stepResult.Status = models.Failed
		suite.storage.MustInsertPickleStepResult(stepResult)

		return ctx, err
	}

	if ctx, undef, err := suite.maybeUndefined(ctx, step.Text, step.Argument); err != nil {
		return ctx, err
	} else if len(undef) > 0 {
		if stepDef != nil {
			stepDef = &models.StepDefinition{
				StepDefinition: formatters.StepDefinition{
					Expr:    stepDef.Expr,
					Handler: stepDef.Handler,
				},
				Args:         stepDef.Args,
				HandlerValue: stepDef.HandlerValue,
				Nested:       stepDef.Nested,
				Undefined:    undef,
			}
		}

		stepResult = models.NewStepResult(pickle.Id, step.Id, stepDef)
		stepResult.Status = models.Undefined
		suite.storage.MustInsertPickleStepResult(stepResult)

		suite.fmt.Undefined(pickle, step, stepDef.GetInternalStepDefinition())
		return ctx, ErrUndefined
	}

	if prevStepErr != nil {
		stepResult = models.NewStepResult(pickle.Id, step.Id, stepDef)
		stepResult.Status = models.Skipped
		suite.storage.MustInsertPickleStepResult(stepResult)

		suite.fmt.Skipped(pickle, step, stepDef.GetInternalStepDefinition())
		return ctx, nil
	}

	ctx, err = suite.maybeSubSteps(stepDef.Run(ctx))

	return ctx, err
}

func (suite *testSuite) runBeforeStepHooks(ctx context.Context, step *Step, err error) (context.Context, error) {
	hooksFailed := false

	for _, handler := range suite.beforeStepHandlers {
		hctx, herr := handler(ctx, step)
		if herr != nil {
			hooksFailed = true

			if err == nil {
				err = herr
			} else {
				err = fmt.Errorf("%v, %w", herr, err)
			}
		}

		if hctx != nil {
			ctx = hctx
		}
	}

	if hooksFailed {
		err = fmt.Errorf("before step hook failed: %w", err)
	}

	return ctx, err
}

func (suite *testSuite) runAfterStepHooks(ctx context.Context, step *Step, status StepResultStatus, err error) (context.Context, error) {
	for _, handler := range suite.afterStepHandlers {
		hctx, herr := handler(ctx, step, status, err)

		// Adding hook error to resulting error without breaking hooks loop.
		if herr != nil {
			if err == nil {
				err = herr
			} else {
				err = fmt.Errorf("%v, %w", herr, err)
			}
		}

		if hctx != nil {
			ctx = hctx
		}
	}

	return ctx, err
}

func (suite *testSuite) runBeforeScenarioHooks(ctx context.Context, pickle *messages.Pickle) (context.Context, error) {
	var err error

	// run before scenario handlers
	for _, handler := range suite.beforeScenarioHandlers {
		hctx, herr := handler(ctx, pickle)
		if herr != nil {
			if err == nil {
				err = herr
			} else {
				err = fmt.Errorf("%v, %w", herr, err)
			}
		}

		if hctx != nil {
			ctx = hctx
		}
	}

	if err != nil {
		err = fmt.Errorf("before scenario hook failed: %w", err)
	}

	return ctx, err
}

func (suite *testSuite) runAfterScenarioHooks(ctx context.Context, pickle *messages.Pickle, lastStepErr error) (context.Context, error) {
	err := lastStepErr

	hooksFailed := false
	isStepErr := true

	// run after scenario handlers
	for _, handler := range suite.afterScenarioHandlers {
		hctx, herr := handler(ctx, pickle, err)

		// Adding hook error to resulting error without breaking hooks loop.
		if herr != nil {
			hooksFailed = true

			if err == nil {
				isStepErr = false
				err = herr
			} else {
				if isStepErr {
					err = fmt.Errorf("step error: %w", err)
					isStepErr = false
				}
				err = fmt.Errorf("%v, %w", herr, err)
			}
		}

		if hctx != nil {
			ctx = hctx
		}
	}

	if hooksFailed {
		err = fmt.Errorf("after scenario hook failed: %w", err)
	}

	return ctx, err
}

func (suite *testSuite) maybeUndefined(ctx context.Context, text string, arg interface{}) (context.Context, []string, error) {
	stepDef := suite.matchStepText(text)
	if nil == stepDef {
		return ctx, []string{text}, nil
	}

	var undefined []string
	if !stepDef.Nested {
		return ctx, undefined, nil
	}

	if arg != nil {
		stepDef.Args = append(stepDef.Args, arg)
	}

	ctx, steps := stepDef.Run(ctx)

	for _, next := range steps.(Steps) {
		lines := strings.Split(next, "\n")
		// @TODO: we cannot currently parse table or content body from nested steps
		if len(lines) > 1 {
			return ctx, undefined, fmt.Errorf("nested steps cannot be multiline and have table or content body argument")
		}
		if len(lines[0]) > 0 && lines[0][len(lines[0])-1] == ':' {
			return ctx, undefined, fmt.Errorf("nested steps cannot be multiline and have table or content body argument")
		}
		ctx, undef, err := suite.maybeUndefined(ctx, next, nil)
		if err != nil {
			return ctx, undefined, err
		}
		undefined = append(undefined, undef...)
	}
	return ctx, undefined, nil
}

func (suite *testSuite) maybeSubSteps(ctx context.Context, result interface{}) (context.Context, error) {
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
		if def := suite.matchStepText(text); def == nil {
			return ctx, ErrUndefined
		} else if ctx, err := suite.maybeSubSteps(def.Run(ctx)); err != nil {
			return ctx, fmt.Errorf("%s: %+v", text, err)
		}
	}
	return ctx, nil
}

func (suite *testSuite) matchStepText(text string) *models.StepDefinition {
	for _, stepDef := range suite.stepDefs {
		if m := stepDef.Expr.FindStringSubmatch(text); len(m) > 0 {
			var args []interface{}
			for _, m := range m[1:] {
				args = append(args, m)
			}

			// since we need to assign arguments
			// better to copy the step definition
			return &models.StepDefinition{
				StepDefinition: formatters.StepDefinition{
					Expr:    stepDef.Expr,
					Handler: stepDef.Handler,
				},
				Args:         args,
				HandlerValue: stepDef.HandlerValue,
				Nested:       stepDef.Nested,
			}
		}
	}
	return nil
}

func (suite *testSuite) runSteps(ctx context.Context, pickle *Scenario, steps []*Step) (context.Context, error) {
	var (
		stepErr, err error
	)

	for i, step := range steps {
		isLast := i == len(steps)-1
		isFirst := i == 0
		ctx, stepErr = suite.runStep(ctx, pickle, step, err, isFirst, isLast)
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

func (suite *testSuite) shouldFail(err error) bool {
	if err == nil {
		return false
	}

	if err == ErrUndefined || err == ErrPending {
		return suite.strict
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

func (suite *testSuite) runPickle(pickle *messages.Pickle) (err error) {
	ctx := suite.defaultContext
	if ctx == nil {
		ctx = context.Background()
	}

	if len(pickle.Steps) == 0 {
		pickleResult := models.PickleResult{PickleID: pickle.Id, StartedAt: utils.TimeNowFunc()}
		suite.storage.MustInsertPickleResult(pickleResult)

		suite.fmt.Pickle(pickle)
		return ErrUndefined
	}

	// Before scenario hooks are called in context of first evaluated step
	// so that error from handler can be added to step.

	pickleResult := models.PickleResult{PickleID: pickle.Id, StartedAt: utils.TimeNowFunc()}
	suite.storage.MustInsertPickleResult(pickleResult)

	suite.fmt.Pickle(pickle)

	// scenario
	if suite.testingT != nil {
		// Running scenario as a subtest.
		suite.testingT.Run(pickle.Name, func(t *testing.T) {
			ctx, err = suite.runSteps(ctx, pickle, pickle.Steps)
			if err != nil {
				t.Error(err)
			}
		})
	} else {
		ctx, err = suite.runSteps(ctx, pickle, pickle.Steps)
	}

	// After scenario handlers are called in context of last evaluated step
	// so that error from handler can be added to step.

	return err
}
