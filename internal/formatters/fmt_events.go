package formatters

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cucumber/messages-go/v16"

	"github.com/cucumber/godog/formatters"
	"github.com/cucumber/godog/internal/utils"
)

const nanoSec = 1000000
const spec = "0.1.0"

func init() {
	formatters.Format("events", fmt.Sprintf("Produces JSON event stream, based on spec: %s.", spec), EventsFormatterFunc)
}

// EventsFormatterFunc implements the FormatterFunc for the events formatter
func EventsFormatterFunc(suite string, out io.Writer) formatters.Formatter {
	return &EventsFormatter{Basefmt: NewBaseFmt(suite, out)}
}

type EventsFormatter struct {
	*Basefmt
}

func (f *EventsFormatter) event(ev interface{}) {
	data, err := json.Marshal(ev)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal stream event: %+v - %v", ev, err))
	}
	fmt.Fprintln(f.out, string(data))
}

func (f *EventsFormatter) Pickle(pickle *messages.Pickle) {
	f.Basefmt.Pickle(pickle)

	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.event(&struct {
		Event     string `json:"event"`
		Location  string `json:"location"`
		Timestamp int64  `json:"timestamp"`
	}{
		"TestCaseStarted",
		f.scenarioLocation(pickle),
		utils.TimeNowFunc().UnixNano() / nanoSec,
	})

	if len(pickle.Steps) == 0 {
		// @TODO: is status undefined or passed? when there are no steps
		// for this scenario
		f.event(&struct {
			Event     string `json:"event"`
			Location  string `json:"location"`
			Timestamp int64  `json:"timestamp"`
			Status    string `json:"status"`
		}{
			"TestCaseFinished",
			f.scenarioLocation(pickle),
			utils.TimeNowFunc().UnixNano() / nanoSec,
			"undefined",
		})
	}
}

func (f *EventsFormatter) TestRunStarted() {
	f.Basefmt.TestRunStarted()

	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.event(&struct {
		Event     string `json:"event"`
		Version   string `json:"version"`
		Timestamp int64  `json:"timestamp"`
		Suite     string `json:"suite"`
	}{
		"TestRunStarted",
		spec,
		utils.TimeNowFunc().UnixNano() / nanoSec,
		f.suiteName,
	})
}

func (f *EventsFormatter) Feature(ft *messages.GherkinDocument, p string, c []byte) {
	f.Basefmt.Feature(ft, p, c)

	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.event(&struct {
		Event    string `json:"event"`
		Location string `json:"location"`
		Source   string `json:"source"`
	}{
		"TestSource",
		fmt.Sprintf("%s:%d", p, ft.Feature.Location.Line),
		string(c),
	})
}

func (f *EventsFormatter) Summary() {
	// @TODO: determine status
	status := passed

	f.Storage.MustGetPickleStepResultsByStatus(failed)

	if len(f.Storage.MustGetPickleStepResultsByStatus(failed)) > 0 {
		status = failed
	} else if len(f.Storage.MustGetPickleStepResultsByStatus(passed)) == 0 {
		if len(f.Storage.MustGetPickleStepResultsByStatus(undefined)) > len(f.Storage.MustGetPickleStepResultsByStatus(pending)) {
			status = undefined
		} else {
			status = pending
		}
	}

	snips := f.Snippets()
	if len(snips) > 0 {
		snips = "You can implement step definitions for undefined steps with these snippets:\n" + snips
	}

	f.event(&struct {
		Event     string `json:"event"`
		Status    string `json:"status"`
		Timestamp int64  `json:"timestamp"`
		Snippets  string `json:"snippets"`
		Memory    string `json:"memory"`
	}{
		"TestRunFinished",
		status.String(),
		utils.TimeNowFunc().UnixNano() / nanoSec,
		snips,
		"", // @TODO not sure that could be correctly implemented
	})
}

func (f *EventsFormatter) step(pickle *messages.Pickle, pickleStep *messages.PickleStep) {
	feature := f.Storage.MustGetFeature(pickle.Uri)
	pickleStepResult := f.Storage.MustGetPickleStepResult(pickleStep.Id)
	step := feature.FindStep(pickleStep.AstNodeIds[0])

	var errMsg string
	if pickleStepResult.Err != nil {
		errMsg = pickleStepResult.Err.Error()
	}
	f.event(&struct {
		Event     string `json:"event"`
		Location  string `json:"location"`
		Timestamp int64  `json:"timestamp"`
		Status    string `json:"status"`
		Summary   string `json:"summary,omitempty"`
	}{
		"TestStepFinished",
		fmt.Sprintf("%s:%d", pickle.Uri, step.Location.Line),
		utils.TimeNowFunc().UnixNano() / nanoSec,
		pickleStepResult.Status.String(),
		errMsg,
	})

	if isLastStep(pickle, pickleStep) {
		var status string

		pickleStepResults := f.Storage.MustGetPickleStepResultsByPickleID(pickle.Id)
		for _, stepResult := range pickleStepResults {
			switch stepResult.Status {
			case passed, failed, undefined, pending:
				status = stepResult.Status.String()
			}
		}

		f.event(&struct {
			Event     string `json:"event"`
			Location  string `json:"location"`
			Timestamp int64  `json:"timestamp"`
			Status    string `json:"status"`
		}{
			"TestCaseFinished",
			f.scenarioLocation(pickle),
			utils.TimeNowFunc().UnixNano() / nanoSec,
			status,
		})
	}
}

func (f *EventsFormatter) Defined(pickle *messages.Pickle, pickleStep *messages.PickleStep, def *formatters.StepDefinition) {
	f.Basefmt.Defined(pickle, pickleStep, def)

	f.Lock.Lock()
	defer f.Lock.Unlock()

	feature := f.Storage.MustGetFeature(pickle.Uri)
	step := feature.FindStep(pickleStep.AstNodeIds[0])

	if def != nil {
		matchedDef := f.Storage.MustGetStepDefintionMatch(pickleStep.AstNodeIds[0])

		m := def.Expr.FindStringSubmatchIndex(pickleStep.Text)[2:]
		var args [][2]int
		for i := 0; i < len(m)/2; i++ {
			pair := m[i : i*2+2]
			var idxs [2]int
			idxs[0] = pair[0]
			idxs[1] = pair[1]
			args = append(args, idxs)
		}

		if len(args) == 0 {
			args = make([][2]int, 0)
		}

		f.event(&struct {
			Event    string   `json:"event"`
			Location string   `json:"location"`
			DefID    string   `json:"definition_id"`
			Args     [][2]int `json:"arguments"`
		}{
			"StepDefinitionFound",
			fmt.Sprintf("%s:%d", pickle.Uri, step.Location.Line),
			DefinitionID(matchedDef),
			args,
		})
	}

	f.event(&struct {
		Event     string `json:"event"`
		Location  string `json:"location"`
		Timestamp int64  `json:"timestamp"`
	}{
		"TestStepStarted",
		fmt.Sprintf("%s:%d", pickle.Uri, step.Location.Line),
		utils.TimeNowFunc().UnixNano() / nanoSec,
	})
}

func (f *EventsFormatter) Passed(pickle *messages.Pickle, step *messages.PickleStep, match *formatters.StepDefinition) {
	f.Basefmt.Passed(pickle, step, match)

	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.step(pickle, step)
}

func (f *EventsFormatter) Skipped(pickle *messages.Pickle, step *messages.PickleStep, match *formatters.StepDefinition) {
	f.Basefmt.Skipped(pickle, step, match)

	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.step(pickle, step)
}

func (f *EventsFormatter) Undefined(pickle *messages.Pickle, step *messages.PickleStep, match *formatters.StepDefinition) {
	f.Basefmt.Undefined(pickle, step, match)

	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.step(pickle, step)
}

func (f *EventsFormatter) Failed(pickle *messages.Pickle, step *messages.PickleStep, match *formatters.StepDefinition, err error) {
	f.Basefmt.Failed(pickle, step, match, err)

	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.step(pickle, step)
}

func (f *EventsFormatter) Pending(pickle *messages.Pickle, step *messages.PickleStep, match *formatters.StepDefinition) {
	f.Basefmt.Pending(pickle, step, match)

	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.step(pickle, step)
}

func (f *EventsFormatter) scenarioLocation(pickle *messages.Pickle) string {
	feature := f.Storage.MustGetFeature(pickle.Uri)
	scenario := feature.FindScenario(pickle.AstNodeIds[0])

	line := scenario.Location.Line
	if len(pickle.AstNodeIds) == 2 {
		_, row := feature.FindExample(pickle.AstNodeIds[1])
		line = row.Location.Line
	}

	return fmt.Sprintf("%s:%d", pickle.Uri, line)
}

func isLastStep(pickle *messages.Pickle, step *messages.PickleStep) bool {
	return pickle.Steps[len(pickle.Steps)-1].Id == step.Id
}
