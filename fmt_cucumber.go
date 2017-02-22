package godog

import (
	"fmt"
	"io"
	"time"
	"strings"
	"github.com/DATA-DOG/godog/gherkin"
	"encoding/json"
	"strconv"
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
	return strings.Replace(strings.ToLower(name), " ", "-", -1)
}

type cukeComment struct {
	Value string `json:"value"`
	Line  int `json:"line"`
}
type cukeTag struct {
	Name string `json:"name"`
	Line int `json:"line"`
}

type cukeResult struct {
	Status   string `json:"status"`
	Error    string `json:"error_message,omitempty"`
	Duration int `json:"duration"`
}

type cukeMatch struct {
	Location string `json:"location"`
}

type cukeStep struct {
	Keyword string `json:"keyword"`
	Name    string `json:"name"`
	Line    int `json:"line"`
	Match   cukeMatch `json:"match"`
	Result  cukeResult `json:"result"`
}

type cukeElement struct {
	Id          string `json:"id"`
	Keyword     string `json:"keyword"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Line        int `json:"line"`
	Type        string `json:"type"`
	Tags        []cukeTag `json:"tags,omitempty"`
	Steps       []cukeStep `json:"steps"`
}

type cukeFeatureJson struct {
	Uri         string `json:"uri"`
	Id          string `json:"id"`
	Keyword     string `json:"keyword"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Line        int `json:"line"`
	Comments []cukeComment `json:"comments,omitempty"`
	Tags        []cukeTag `json:"tags,omitempty"`
	Elements    []cukeElement `json:"elements"`
}

type cukefmt struct {
	basefmt

					 // currently running feature path, to be part of id.
					 // this is sadly not passed by gherkin nodes.
					 // it restricts this formatter to run only in synchronous single
					 // threaded execution. Unless running a copy of formatter for each feature
	path           string
	stat           stepType          // last step status, before skipped
	outlineSteps   int               // number of current outline scenario steps
	id             string            // current test id.
	results        []cukeFeatureJson // structure that represent cuke results
	curStep        *cukeStep         // track the current step
	curElement     *cukeElement      // track the current element
	curFeature     *cukeFeatureJson  // track the current feature
	curOutline     cukeElement       // Each example show up as an outline element but the outline is parsed only once
					 // so I need to keep track of the current outline
	curRow         int               // current row of the example table as it is being processed.
	curExampleTags []cukeTag         // temporary storage for tags associate with the current example table.
	startTime      time.Time
	curExampleName string
}

func (f *cukefmt) Node(n interface{}) {
	f.basefmt.Node(n)

	switch t := n.(type) {

	// When the example definition is seen we just need track the id and
	// append the name associated with the example as part of the id.
	case *gherkin.Examples:
		f.curExampleName = makeId(t.Name)
		f.curRow = 2 // there can be more than one example set per outline so reset row count.
		// cucumber counts the header row as an example when creating the id.

		// store any example level tags in a  temp location.
		f.curExampleTags = make([]cukeTag, len(t.Tags))
		for idx, element := range t.Tags {
			f.curExampleTags[idx].Line = element.Location.Line
			f.curExampleTags[idx].Name = element.Name
		}

	// The outline node creates a placeholder and the actual element is added as each TableRow is processed.
	case *gherkin.ScenarioOutline:

		f.curOutline = cukeElement{}
		f.curOutline.Name = t.Name
		f.curOutline.Line = t.Location.Line
		f.curOutline.Description = t.Description
		f.curOutline.Keyword = t.Keyword
		f.curOutline.Id = f.curFeature.Id + ";" + makeId(t.Name)
		f.curOutline.Type = "scenario"
		f.curOutline.Tags = make([]cukeTag, len(t.Tags) + len(f.curFeature.Tags))

		// apply feature level tags
		if (len(f.curOutline.Tags) > 0) {
			copy(f.curOutline.Tags, f.curFeature.Tags)

			// apply outline level tags.
			for idx, element := range t.Tags {
				f.curOutline.Tags[idx + len(f.curFeature.Tags)].Line = element.Location.Line
				f.curOutline.Tags[idx + len(f.curFeature.Tags)].Name = element.Name
			}
		}

	// This scenario adds the element to the output immediately.
	case *gherkin.Scenario:
		f.curFeature.Elements = append(f.curFeature.Elements, cukeElement{})
		f.curElement = &f.curFeature.Elements[len(f.curFeature.Elements) - 1]

		f.curElement.Name = t.Name
		f.curElement.Line = t.Location.Line
		f.curElement.Description = t.Description
		f.curElement.Keyword = t.Keyword
		f.curElement.Id = f.curFeature.Id + ";" + makeId(t.Name)
		f.curElement.Type = "scenario"
		f.curElement.Tags = make([]cukeTag, len(t.Tags) + len(f.curFeature.Tags))

		if (len(f.curElement.Tags) > 0) {
			// apply feature level tags
			copy(f.curElement.Tags, f.curFeature.Tags)

			// apply scenario level tags.
			for idx, element := range t.Tags {
				f.curElement.Tags[idx + len(f.curFeature.Tags)].Line = element.Location.Line
				f.curElement.Tags[idx + len(f.curFeature.Tags)].Name = element.Name
			}
		}


	// This is an outline scenario and the element is added to the output as
	// the TableRows are encountered.
	case *gherkin.TableRow:
		tmpElem := f.curOutline
		tmpElem.Line = t.Location.Line
		tmpElem.Id = tmpElem.Id + ";" + f.curExampleName + ";" + strconv.Itoa(f.curRow)
		f.curRow++
		f.curFeature.Elements = append(f.curFeature.Elements, tmpElem)
		f.curElement = &f.curFeature.Elements[len(f.curFeature.Elements) - 1]

		// copy in example level tags.
		f.curElement.Tags = append(f.curElement.Tags, f.curExampleTags...)

	}

}

func (f *cukefmt) Feature(ft *gherkin.Feature, p string, c []byte) {

	f.basefmt.Feature(ft, p, c)
	f.path = p
	f.id = makeId(ft.Name)
	f.results = append(f.results, cukeFeatureJson{})

	f.curFeature = &f.results[len(f.results) - 1]
	f.curFeature.Uri = p
	f.curFeature.Name = ft.Name
	f.curFeature.Keyword = ft.Keyword
	f.curFeature.Line = ft.Location.Line
	f.curFeature.Description = ft.Description
	f.curFeature.Id = f.id
	f.curFeature.Tags = make([]cukeTag, len(ft.Tags))

	for idx, element := range ft.Tags {
		f.curFeature.Tags[idx].Line = element.Location.Line
		f.curFeature.Tags[idx].Name = element.Name
	}

	f.curFeature.Comments = make([]cukeComment, len(ft.Comments))
	for idx, comment := range ft.Comments {
		f.curFeature.Comments[idx].Value = comment.Text
		f.curFeature.Comments[idx].Line = comment.Location.Line
	}

}

func (f *cukefmt) Summary() {
	dat, err := json.MarshalIndent(f.results, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(f.out, "%s\n", string(dat))
}

func (f *cukefmt) step(res *stepResult) {

	// determine if test case has finished
	switch t := f.owner.(type) {
	case *gherkin.TableRow:
		f.curStep.Result.Duration = int(time.Since(f.startTime).Nanoseconds())
		f.curStep.Line = t.Location.Line
		f.curStep.Result.Status = res.typ.String()
		if res.err != nil {
			f.curStep.Result.Error = res.err.Error()
		}
	case *gherkin.Scenario:
		f.curStep.Result.Duration = int(time.Since(f.startTime).Nanoseconds())
		f.curStep.Result.Status = res.typ.String()
		if res.err != nil {
			f.curStep.Result.Error = res.err.Error()
		}
	}
}

func (f *cukefmt) Defined(step *gherkin.Step, def *StepDef) {

	f.startTime = time.Now() // start timing the step
	f.curElement.Steps = append(f.curElement.Steps, cukeStep{})
	f.curStep = &f.curElement.Steps[len(f.curElement.Steps) - 1]

	f.curStep.Name = step.Text
	f.curStep.Line = step.Location.Line
	f.curStep.Keyword = step.Keyword

	if def != nil {
		f.curStep.Match.Location = strings.Split(def.definitionID(), " ")[0]
	}
}

func (f *cukefmt) Passed(step *gherkin.Step, match *StepDef) {
	f.basefmt.Passed(step, match)
	f.stat = passed
	f.step(f.passed[len(f.passed) - 1])
}

func (f *cukefmt) Skipped(step *gherkin.Step) {
	f.basefmt.Skipped(step)
	f.step(f.skipped[len(f.skipped) - 1])
}

func (f *cukefmt) Undefined(step *gherkin.Step) {
	f.basefmt.Undefined(step)
	f.stat = undefined
	f.step(f.undefined[len(f.undefined) - 1])
}

func (f *cukefmt) Failed(step *gherkin.Step, match *StepDef, err error) {
	f.basefmt.Failed(step, match, err)
	f.stat = failed
	f.step(f.failed[len(f.failed) - 1])
}

func (f *cukefmt) Pending(step *gherkin.Step, match *StepDef) {
	f.stat = pending
	f.basefmt.Pending(step, match)
	f.step(f.pending[len(f.pending) - 1])
}
