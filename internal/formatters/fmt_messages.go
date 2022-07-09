package formatters

import (
	"encoding/json"
	"io"

	"github.com/cucumber/godog/formatters"
	"github.com/cucumber/messages-go/v16"
)

func init() {
	formatters.Format("message", "Outputs protobuf messages", MessageFormatterFunc)
}

// MessageFormatterFunc implements the FormatterFunc for the message formatter
func MessageFormatterFunc(suite string, out io.Writer) formatters.Formatter {
	return &Message{Base: NewBase(suite, out)}
}

// TODO: meta, source, and pickle messages (https://github.com/cucumber/compatibility-kit/blob/main/devkit/samples/rules/rules.feature.ndjson)
type Message struct {
	*Base
}

func (f *Message) MetaData(m *formatters.MetaData) {
	f.Base.MetaData(m)
	msg := &messages.Envelope{
		Meta: &messages.Meta{},
	}
	f.send(msg)
}

func (f *Message) Source(m *formatters.Source) {
	f.Base.Source(m)
	msg := &messages.Envelope{
		Source: &messages.Source{
			Uri:       m.Uri,
			Data:      m.Data,
			MediaType: messages.SourceMediaType(m.MediaType),
		},
	}
	f.send(msg)
}

func (f *Message) TestRunStarted(t *messages.TestRunStarted) {
	f.Base.TestRunStarted(t)
	msg := &messages.Envelope{
		TestRunStarted: &messages.TestRunStarted{},
	}
	f.send(msg)
}

func (f *Message) GherkinDocument(doc *messages.GherkinDocument, p string, c []byte) {
	f.Base.Feature(doc, p, c)
	msg := &messages.Envelope{
		GherkinDocument: doc,
	}
	f.send(msg)
}

func (f *Message) Pickle(p *messages.Pickle) {
	f.Base.Pickle(p)
	msg := &messages.Envelope{
		Pickle: p,
	}
	f.send(msg)
}

func (f *Message) Defined(def *formatters.StepDefinition) {
	f.Base.Defined(def)

	f.Lock.Lock()
	defer f.Lock.Unlock()

	// feature := f.Storage.MustGetFeature(pickle.Uri)
	// step := feature.FindStep(pickleStep.AstNodeIds[0])

	if def != nil {
		// matchedDef := f.Storage.MustGetStepDefintionMatch(pickleStep.AstNodeIds[0])

		// m := def.Expr.FindStringSubmatchIndex(pickleStep.Text)[2:]
		// var args [][2]int
		// for i := 0; i < len(m)/2; i++ {
		// 	pair := m[i : i*2+2]
		// 	var idxs [2]int
		// 	idxs[0] = pair[0]
		// 	idxs[1] = pair[1]
		// 	args = append(args, idxs)
		// }

		// if len(args) == 0 {
		// 	args = make([][2]int, 0)
		// }

		msg := &messages.Envelope{
			StepDefinition: &messages.StepDefinition{},
		}
		f.send(msg)
	}
}

func (f *Message) TestCase(pickle *messages.Pickle) {
	f.Base.TestCase(pickle)

	msg := &messages.Envelope{
		TestCase: &messages.TestCase{
			Id:       pickle.Id,
			PickleId: pickle.Id,
		},
	}
	f.send(msg)

	msg = &messages.Envelope{ // TODO: NOT HERE!
		TestCaseStarted: &messages.TestCaseStarted{
			Attempt:    1,
			Id:         pickle.Id,
			TestCaseId: "",  // TODO: ???
			Timestamp:  nil, // TODO: ???
		},
	}
	f.send(msg)

	if len(pickle.Steps) == 0 {
		// TestCaseFinished
		msg := &messages.Envelope{
			TestCaseFinished: &messages.TestCaseFinished{},
		}
		f.send(msg)
	}
}

func (f *Message) TestStepStarted(pickle *messages.Pickle, pickleStep *messages.PickleStep, def *formatters.StepDefinition) {
	f.Base.TestStepStarted(pickle, pickleStep, def)

	f.Lock.Lock()
	defer f.Lock.Unlock()

	msg := &messages.Envelope{
		TestStepStarted: &messages.TestStepStarted{},
	}
	f.send(msg)
}

func (f *Message) Failed(pickle *messages.Pickle, step *messages.PickleStep, match *formatters.StepDefinition, err error) {
	f.Base.Failed(pickle, step, match, err)
	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.step(pickle, step)
}

func (f *Message) Passed(pickle *messages.Pickle, step *messages.PickleStep, match *formatters.StepDefinition) {
	f.Base.Passed(pickle, step, match)
	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.step(pickle, step)
}

func (f *Message) Skipped(pickle *messages.Pickle, step *messages.PickleStep, match *formatters.StepDefinition) {
	f.Base.Skipped(pickle, step, match)
	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.step(pickle, step)
}

func (f *Message) Undefined(pickle *messages.Pickle, step *messages.PickleStep, match *formatters.StepDefinition) {
	f.Base.Undefined(pickle, step, match)
	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.step(pickle, step)
}

func (f *Message) Pending(pickle *messages.Pickle, step *messages.PickleStep, match *formatters.StepDefinition) {
	f.Base.Pending(pickle, step, match)
	f.Lock.Lock()
	defer f.Lock.Unlock()

	f.step(pickle, step)
}

func (f *Message) Summary() {
	// f.Base.Summary() - should we skip this, or encapsulate writes in Base - Base.Summary writes stuff like the following:
	// scenarios (1 pending)
	// 1 steps (1 pending)
	// 0s
	msg := &messages.Envelope{
		TestRunFinished: &messages.TestRunFinished{},
	}
	f.send(msg)
}

func (f *Message) step(pickle *messages.Pickle, pickleStep *messages.PickleStep) {
	// feature := f.Storage.MustGetFeature(pickle.Uri)
	// pickleStepResult := f.Storage.MustGetPickleStepResult(pickleStep.Id)
	// step := feature.FindStep(pickleStep.AstNodeIds[0])

	// var errMsg string
	// if pickleStepResult.Err != nil {
	// 	errMsg = pickleStepResult.Err.Error()
	// }
	msg := &messages.Envelope{
		TestStepFinished: &messages.TestStepFinished{},
	}
	f.send(msg)

	if isLastStep(pickle, pickleStep) {
		// var status string

		// pickleStepResults := f.Storage.MustGetPickleStepResultsByPickleID(pickle.Id)
		// for _, stepResult := range pickleStepResults {
		// 	switch stepResult.Status {
		// 	case passed, failed, undefined, pending:
		// 		status = stepResult.Status.String()
		// 	}
		// }

		msg := &messages.Envelope{
			TestCaseFinished: &messages.TestCaseFinished{},
		}
		f.send(msg)
	}
}

func (f *Message) send(msg *messages.Envelope) {
	bytes, err := json.Marshal(msg)
	if err != nil {
		// TODO: ??
	}
	// fmt.Printf("send %v\n", string(bytes))
	_, err = f.out.Write(bytes)
	if err != nil {
		// TODO: ??
	}
}
