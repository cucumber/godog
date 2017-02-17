package godog

import (
	"fmt"
	"io"
	"time"
	"strings"
	"github.com/DATA-DOG/godog/gherkin"
	"encoding/json"
)

const cukeurl = "https://www.relishapp.com/cucumber/cucumber/docs/formatters/json-output-formatter"

func init() {
	Format("cucumber", fmt.Sprintf("Produces cucumber JSON stream, based on spec @: %s.", cukeurl), cucumberFunc)
}

func cucumberFunc(suite string, out io.Writer) Formatter {
	formatter := &cukefmt{
		basefmt: basefmt{
			started: time.Now(),
			indent:  2,
			out:     out,
		},
	}

	return formatter
}

// Replace spaces with -
func makeId(name string) string {
	return strings.Replace(strings.ToLower(name)," ","-",-1)
}

type cukeTag struct {
	Name string `json:"name"`
	Line int `json:"line"`
}

type cukeResult struct {
	Status string `json:"status"`
	Duration int `json:"duration"`
	Error string `json:"error_message,omitempty"`
}

type cukeMatch struct {
	Location string `json:"location"`
}

type cukeStep struct {
	Keyword string `json:"keyword"`
	Name string `json:"name"`
	Line int `json:"line"`
	Match cukeMatch `json:"match"`
	Result cukeResult `json:"result"`
}


type cukeElement struct {
	Keyword string `json:"keyword"`
	Id string `json:"id"`
	Name string `json:"name"`
	Line int `json:"line"`
	Description string `json:"description"`
	Tags []cukeTag `json:"tags"`
	Type string `json:type`
	Steps []cukeStep `json:steps`

}

type cukeFeatureJson struct {
	Uri string `json:"uri"`
	Id string `json:"id"`
	Keyword string `json:"keyword"`
	Name string `json:"name"`
	Line int `json:"line"`
	Description string `json:"description"`
	Tags []cukeTag `json:"tags"`
	Elements []cukeElement `json:"elements"`

}

type cukefmt struct {
	basefmt

	// currently running feature path, to be part of id.
	// this is sadly not passed by gherkin nodes.
	// it restricts this formatter to run only in synchronous single
	// threaded execution. Unless running a copy of formatter for each feature
	path         string
	stat         stepType // last step status, before skipped
	outlineSteps int      // number of current outline scenario steps
	id           string      // current test id.
	results      []cukeFeatureJson // structure that represent cuke results
	curStep     *cukeStep // track the current step
	curElement  *cukeElement  // track the current element
	curFeature  *cukeFeatureJson // track the current feature

}


func (f *cukefmt) Node(n interface{}) {
	f.basefmt.Node(n)

	switch t := n.(type) {

	case *gherkin.ScenarioOutline:
	case *gherkin.Scenario:
		f.curFeature.Elements = append(f.curFeature.Elements,cukeElement{})
		f.curElement = &f.curFeature.Elements[len(f.curFeature.Elements)-1]

		f.curElement.Name = t.Name
		f.curElement.Line = t.Location.Line
		f.curElement.Description = t.Description
		f.curElement.Keyword = t.Keyword
		f.curElement.Id = f.curFeature.Id+";"+makeId(t.Name)
		f.curElement.Type = t.Type
		f.curElement.Tags = make([]cukeTag,len(t.Tags))
		for idx,element := range t.Tags {
			f.curElement.Tags[idx].Line = element.Location.Line
			f.curElement.Tags[idx].Name = element.Name
		}

	case *gherkin.TableRow:
		fmt.Fprintf(f.out,"Entering Node TableRow: %s:%d\n",f.path, t.Location.Line)
	}

}

func (f *cukefmt) Feature(ft *gherkin.Feature, p string, c []byte) {


	f.basefmt.Feature(ft, p, c)
	f.path = p
	f.id = makeId(ft.Name)
	f.results = append(f.results,cukeFeatureJson{})

	f.curFeature = &f.results[len(f.results)-1]
	f.curFeature.Uri = p
	f.curFeature.Name = ft.Name
	f.curFeature.Keyword = ft.Keyword
	f.curFeature.Line = ft.Location.Line
	f.curFeature.Description = ft.Description
	f.curFeature.Id = f.id
	f.curFeature.Tags = make([]cukeTag, len(ft.Tags))

	for idx,element := range ft.Tags {
		f.curFeature.Tags[idx].Line = element.Location.Line
		f.curFeature.Tags[idx].Name = element.Name
	}

}

func (f *cukefmt) Summary() {
	dat, err := json.MarshalIndent(f.results, "", "\t")
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(f.out,"%s\n",string(dat))
}

func (f *cukefmt) step(res *stepResult) {


	// determine if test case has finished
	var finished bool
	var line int
	switch t := f.owner.(type) {
	case *gherkin.TableRow:
		line = t.Location.Line
		finished = f.isLastStep(res.step)
		fmt.Fprintf(f.out,"step: TableRow: line:%v finished:%v\n",line, finished)
	case *gherkin.Scenario:
		f.curStep.Result.Status = res.typ.String()
		if res.err != nil {
			f.curStep.Result.Error = res.err.Error()
		}
	}
}

func (f *cukefmt) Defined(step *gherkin.Step, def *StepDef) {
	fmt.Fprintf(f.out,"Defined: step:%v stepDef:%v\n",step,def)

	f.curElement.Steps = append(f.curElement.Steps,cukeStep{})
	f.curStep = &f.curElement.Steps[len(f.curElement.Steps)-1]

	f.curStep.Name = step.Text
	f.curStep.Line = step.Location.Line
	f.curStep.Keyword = step.Keyword

	if def != nil {
		f.curStep.Match.Location = strings.Split(def.definitionID()," ")[0]
	}
}

func (f *cukefmt) Passed(step *gherkin.Step, match *StepDef) {
	f.basefmt.Passed(step, match)
	f.stat = passed
	f.step(f.passed[len(f.passed)-1])
}

func (f *cukefmt) Skipped(step *gherkin.Step) {
	f.basefmt.Skipped(step)
	f.step(f.skipped[len(f.skipped)-1])
}

func (f *cukefmt) Undefined(step *gherkin.Step) {
	f.basefmt.Undefined(step)
	f.stat = undefined
	f.step(f.undefined[len(f.undefined)-1])
}

func (f *cukefmt) Failed(step *gherkin.Step, match *StepDef, err error) {
	f.basefmt.Failed(step, match, err)
	f.stat = failed
	f.step(f.failed[len(f.failed)-1])
}

func (f *cukefmt) Pending(step *gherkin.Step, match *StepDef) {
	f.stat = pending
	f.basefmt.Pending(step, match)
	f.step(f.pending[len(f.pending)-1])
}
