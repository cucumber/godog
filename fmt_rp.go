package godog

/*
	This formatter is based on the cucumber GoDog Formatter
*/

import (
	"fmt"
	"github.com/avarabyeu/goRP/gorp"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/DATA-DOG/godog/gherkin"
)

func init() {
	Format("rp", "Sends report to Report Portal.", rpFunc)
}

func rpFunc(suite string, out io.Writer) Formatter {

	// Start RP Connection
	goRP := gorp.NewClient(
		"http://localhost:8088",
		"frontoffice",
		"d9d5c243-6821-4e07-95eb-590dc49cf4e9",
	)

	formatter := &rp{
		basefmt: basefmt{
			started: timeNowFunc(),
			indent:  2,
			out:     out,
		},
		goRP: goRP,
	}

	return formatter
}

// The sequence of type structs are used to marshall the json object.
type rpComment struct {
	Value string `json:"value"`
	Line  int    `json:"line"`
}

type rpDocstring struct {
	Value       string `json:"value"`
	ContentType string `json:"content_type"`
	Line        int    `json:"line"`
}

type rpTag struct {
	Name string `json:"name"`
	Line int    `json:"line"`
}

type rpResult struct {
	Status   string `json:"status"`
	Error    string `json:"error_message,omitempty"`
	Duration *int   `json:"duration,omitempty"`
	Comment  string `json:"comment,omitempty"`
}

type rpMatch struct {
	Location string `json:"location"`
}

type rpStep struct {
	Keyword   string            `json:"keyword"`
	Name      string            `json:"name"`
	Line      int               `json:"line"`
	Docstring *rpDocstring      `json:"doc_string,omitempty"`
	Match     rpMatch           `json:"match"`
	Result    rpResult          `json:"result"`
	DataTable []*rpDataTableRow `json:"rows,omitempty"`
}

type rpDataTableRow struct {
	Cells []string `json:"cells"`
}

type rpElement struct {
	ID          string   `json:"id"`
	Keyword     string   `json:"keyword"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Line        int      `json:"line"`
	Type        string   `json:"type"`
	Tags        []rpTag  `json:"tags,omitempty"`
	Steps       []rpStep `json:"steps,omitempty"`
}

type rpFeatureJSON struct {
	URI         string      `json:"uri"`
	ID          string      `json:"id"`
	Keyword     string      `json:"keyword"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Line        int         `json:"line"`
	Comments    []rpComment `json:"comments,omitempty"`
	Tags        []rpTag     `json:"tags,omitempty"`
	Elements    []rpElement `json:"elements,omitempty"`
}

type rp struct {
	basefmt
	path                 string
	stat                 stepType        // last step status, before skipped
	ID                   string          // current test id.
	results              []rpFeatureJSON // structure that represent rp results
	curStep              *rpStep         // track the current step
	curElement           *rpElement      // track the current element
	curFeature           *rpFeatureJSON  // track the current feature
	curOutline           rpElement       // Each example show up as an outline element but the outline is parsed only once
	curRow               int             // current row of the example table as it is being processed.
	curExampleTags       []rpTag         // temporary storage for tags associate with the current example table.
	startTime            time.Time       // used to time duration of the step execution
	curExampleName       string          // Due to the fact that examples are parsed once and then iterated over for each result then we need to keep track
	goRP                 *gorp.Client
	env                  string
	tag                  string
	launchStatus         string
	currentLaunchID      string
	currentSuiteID       string
	currentTestID        string
	currentStepID        string
	currentTestStepLen   int
	currentTestStepCount int
	suiteStatus          string
	testStatus           string
}

func (f *rp) Node(n interface{}) {
	f.basefmt.Node(n)
	switch t := n.(type) {
	// When the example definition is seen we just need track the id and
	// append the name associated with the example as part of the id.
	case *gherkin.Examples:

		f.curExampleName = makeID(t.Name)
		f.curRow = 2 // there can be more than one example set per outline so reset row count.
		// cucumber counts the header row as an example when creating the id.

		// store any example level tags in a  temp location.
		f.curExampleTags = make([]rpTag, len(t.Tags))
		for idx, element := range t.Tags {
			f.curExampleTags[idx].Line = element.Location.Line
			f.curExampleTags[idx].Name = element.Name
		}

	// The outline node creates a placeholder and the actual element is added as each TableRow is processed.
	case *gherkin.ScenarioOutline:
		f.currentTestStepLen = len(t.Steps)
		f.curOutline = rpElement{}
		f.curOutline.Name = t.Name
		f.curOutline.Line = t.Location.Line
		f.curOutline.Description = t.Description
		f.curOutline.Keyword = t.Keyword
		f.curOutline.ID = f.curFeature.ID + ";" + makeID(t.Name)
		f.curOutline.Type = "scenario"
		f.curOutline.Tags = make([]rpTag, len(t.Tags)+len(f.curFeature.Tags))

		// apply feature level tags
		if len(f.curOutline.Tags) > 0 {
			copy(f.curOutline.Tags, f.curFeature.Tags)

			// apply outline level tags.
			for idx, element := range t.Tags {
				f.curOutline.Tags[idx+len(f.curFeature.Tags)].Line = element.Location.Line
				f.curOutline.Tags[idx+len(f.curFeature.Tags)].Name = element.Name
			}
		}

	// This scenario adds the element to the output immediately.
	case *gherkin.Scenario:
		f.currentTestStepLen = len(t.Steps)
		f.curFeature.Elements = append(f.curFeature.Elements, rpElement{})
		f.curElement = &f.curFeature.Elements[len(f.curFeature.Elements)-1]

		f.curElement.Name = t.Name
		f.curElement.Line = t.Location.Line
		f.curElement.Description = t.Description
		f.curElement.Keyword = t.Keyword
		f.curElement.ID = f.curFeature.ID + ";" + makeID(t.Name)
		f.curElement.Type = "scenario"
		f.curElement.Tags = make([]rpTag, len(t.Tags)+len(f.curFeature.Tags))

		if len(f.curElement.Tags) > 0 {
			// apply feature level tags
			copy(f.curElement.Tags, f.curFeature.Tags)

			// apply scenario level tags.
			for idx, element := range t.Tags {
				f.curElement.Tags[idx+len(f.curFeature.Tags)].Line = element.Location.Line
				f.curElement.Tags[idx+len(f.curFeature.Tags)].Name = element.Name
			}
		}

	// This is an outline scenario and the element is added to the output as
	// the TableRows are encountered.
	case *gherkin.TableRow:
		tmpElem := f.curOutline
		tmpElem.Line = t.Location.Line
		tmpElem.ID = tmpElem.ID + ";" + f.curExampleName + ";" + strconv.Itoa(f.curRow)
		f.curRow++
		f.curFeature.Elements = append(f.curFeature.Elements, tmpElem)
		f.curElement = &f.curFeature.Elements[len(f.curFeature.Elements)-1]

		// copy in example level tags.
		f.curElement.Tags = append(f.curElement.Tags, f.curExampleTags...)

	}
	f.currentTestStepCount = 0

	fmt.Println("-- Start Scenario --")
	f.startChildContainer()
}

func (f *rp) Feature(ft *gherkin.Feature, p string, c []byte) {
	f.basefmt.Feature(ft, p, c)
	f.path = p
	f.ID = makeID(ft.Name)
	f.results = append(f.results, rpFeatureJSON{})

	f.curFeature = &f.results[len(f.results)-1]
	f.curFeature.URI = p
	f.curFeature.Name = ft.Name
	f.curFeature.Keyword = ft.Keyword
	f.curFeature.Line = ft.Location.Line
	f.curFeature.Description = ft.Description
	f.curFeature.ID = f.ID
	f.curFeature.Tags = make([]rpTag, len(ft.Tags))

	for idx, element := range ft.Tags {
		f.curFeature.Tags[idx].Line = element.Location.Line
		f.curFeature.Tags[idx].Name = element.Name
	}

	f.curFeature.Comments = make([]rpComment, len(ft.Comments))
	for idx, comment := range ft.Comments {
		f.curFeature.Comments[idx].Value = strings.TrimSpace(comment.Text)
		f.curFeature.Comments[idx].Line = comment.Location.Line
	}

	fmt.Println("-- Start Launch --")
	f.startLaunch()

	fmt.Println("-- Start Suite aka Feature --")
	f.startSuite()
}

func (f *rp) Summary() {
	fmt.Println("-- Finish Suite aka Feature --")
	f.finishSuite()
	fmt.Println("-- Finish Launch --")
	f.finishLaunch()
}
func (f *rp) step(res *stepResult) {
	// determine if test case has finished
	switch t := f.owner.(type) {
	case *gherkin.TableRow:
		d := int(timeNowFunc().Sub(f.startTime).Nanoseconds())
		f.curStep.Result.Duration = &d
		f.curStep.Line = t.Location.Line
		f.curStep.Result.Status = res.typ.String()
		if res.err != nil {
			f.curStep.Result.Error = res.err.Error()
		}
	case *gherkin.Scenario:
		d := int(timeNowFunc().Sub(f.startTime).Nanoseconds())
		f.curStep.Result.Duration = &d
		f.curStep.Result.Status = res.typ.String()
		if res.err != nil {
			f.curStep.Result.Error = res.err.Error()
		}
	}
	fmt.Println("-- Finish Child Step --")
	f.finishChildStep()
	//check if its last step in scenario
	if f.currentTestStepCount == f.currentTestStepLen {
		f.finishChildContainer()
	}
}
func (f *rp) Defined(step *gherkin.Step, def *StepDef) {
	f.startTime = timeNowFunc() // start timing the step
	f.curElement.Steps = append(f.curElement.Steps, rpStep{})
	f.curStep = &f.curElement.Steps[len(f.curElement.Steps)-1]

	f.curStep.Name = step.Text
	f.curStep.Line = step.Location.Line
	f.curStep.Keyword = step.Keyword

	if _, ok := step.Argument.(*gherkin.DocString); ok {
		f.curStep.Docstring = &rpDocstring{}
		f.curStep.Docstring.ContentType = strings.TrimSpace(step.Argument.(*gherkin.DocString).ContentType)
		f.curStep.Docstring.Line = step.Argument.(*gherkin.DocString).Location.Line
		f.curStep.Docstring.Value = step.Argument.(*gherkin.DocString).Content
	}

	if _, ok := step.Argument.(*gherkin.DataTable); ok {
		dataTable := step.Argument.(*gherkin.DataTable)

		f.curStep.DataTable = make([]*rpDataTableRow, len(dataTable.Rows))
		for i, row := range dataTable.Rows {
			cells := make([]string, len(row.Cells))
			for j, cell := range row.Cells {
				cells[j] = cell.Value
			}
			f.curStep.DataTable[i] = &rpDataTableRow{Cells: cells}
		}
	}

	if def != nil {
		f.curStep.Match.Location = strings.Split(def.definitionID(), " ")[0]
	}

	fmt.Println("-- Start Child Step --")
	f.currentTestStepCount = f.currentTestStepCount + 1
	f.startChildStep()
}
func (f *rp) Passed(step *gherkin.Step, match *StepDef) {
	f.basefmt.Passed(step, match)
	f.stat = passed
	f.step(f.passed[len(f.passed)-1])
}
func (f *rp) Skipped(step *gherkin.Step, match *StepDef) {
	f.basefmt.Skipped(step, match)
	f.step(f.skipped[len(f.skipped)-1])

	// no duration reported for skipped.
	f.curStep.Result.Duration = nil
}
func (f *rp) Undefined(step *gherkin.Step, match *StepDef) {
	f.basefmt.Undefined(step, match)
	f.stat = undefined
	f.step(f.undefined[len(f.undefined)-1])

	// the location for undefined is the feature file location not the step file.
	f.curStep.Match.Location = fmt.Sprintf("%s:%d", f.path, step.Location.Line)
	f.curStep.Result.Duration = nil
}
func (f *rp) Failed(step *gherkin.Step, match *StepDef, err error) {
	f.basefmt.Failed(step, match, err)
	f.stat = failed
	f.step(f.failed[len(f.failed)-1])
}
func (f *rp) Pending(step *gherkin.Step, match *StepDef) {
	f.stat = pending
	f.basefmt.Pending(step, match)
	f.step(f.pending[len(f.pending)-1])

	// the location for pending is the feature file location not the step file.
	f.curStep.Match.Location = fmt.Sprintf("%s:%d", f.path, step.Location.Line)
	f.curStep.Result.Duration = nil
}

///////////////////////////
// Report Portal Functions
///////////////////////////

// startLaunch - Start Launch in RP
func (f *rp) startLaunch() {
	fmt.Println("Start launch")
	launchRqData := &gorp.StartLaunchRQ{
		StartRQ: gorp.StartRQ{
			Name:       f.ID,
			Attributes: nil,
			StartTime:  gorp.Timestamp{},
		},
		Mode: "Default",
	}
	launchRsData, err := f.goRP.StartLaunch(launchRqData)
	if err != nil {
		fmt.Println("Error no adding launch to RP")
		fmt.Println(err.Error())
	}
	f.currentLaunchID = launchRsData.ID
}

// finishLaunch - Finish Launch in RP
func (f *rp) finishLaunch() {
	launchRqData := &gorp.FinishExecutionRQ{
		EndTime: gorp.Timestamp{},
		Status:  f.launchStatus,
	}
	_, err := f.goRP.FinishLaunch(f.currentLaunchID, launchRqData)
	if err != nil {
		fmt.Println("Error no finishing launch to RP")
		fmt.Println(err.Error())
	}
}

// startSuite - Start BDD Feature
func (f *rp) startSuite() {
	testRqData := &gorp.StartTestRQ{
		StartRQ: gorp.StartRQ{
			Name:        f.curFeature.Name,
			Description: f.curFeature.Description,
			StartTime:   gorp.Timestamp{},
		},

		LaunchID: f.currentLaunchID,
		Type:     "SUITE",
	}
	testRsData, err := f.goRP.StartTest(testRqData)
	if err != nil {
		fmt.Println("Error on adding suite (feature) to RP")
		fmt.Println(err.Error())
	}
	f.currentSuiteID = testRsData.ID
}

// finishSuite - Finish BDD Feature
func (f *rp) finishSuite() {
	testRqData := &gorp.FinishTestRQ{
		FinishExecutionRQ: gorp.FinishExecutionRQ{
			EndTime: gorp.Timestamp{},
			Status:  f.suiteStatus,
		},
	}
	_, err := f.goRP.FinishTest(f.currentSuiteID, testRqData)
	if err != nil {
		fmt.Println("Error on finishing suite (feature) to RP")
		fmt.Println(err.Error())
	}
}

// startChildContainer - Start BDD Scenario
func (f *rp) startChildContainer() {
	testRqData := &gorp.StartTestRQ{
		StartRQ: gorp.StartRQ{
			Name:        f.curElement.Name,
			Description: f.curElement.Description,
			StartTime:   gorp.Timestamp{},
		},
		LaunchID: f.currentLaunchID,
		Type:     "TEST",
	}
	testRsData, err := f.goRP.StartChildTest(f.currentSuiteID, testRqData)
	if err != nil {
		fmt.Println("Error on adding child container (scenario) to RP")
		fmt.Println(err.Error())
	}
	f.currentTestID = testRsData.ID
}

// finishChildContainer - Finish BDD Scenario
func (f *rp) finishChildContainer() {
	fmt.Println("Finish Child Container")
	testRqData := &gorp.FinishTestRQ{
		FinishExecutionRQ: gorp.FinishExecutionRQ{
			EndTime: gorp.Timestamp{},
			Status:  f.suiteStatus,
		},
	}
	_, err := f.goRP.FinishTest(f.currentTestID, testRqData)
	if err != nil {
		fmt.Println("Error on finishing child container (scenario) to RP")
		fmt.Println(err.Error())
	}
}

// startChildStep - Start BDD Step
func (f *rp) startChildStep() {
	testRqData := &gorp.StartTestRQ{
		StartRQ: gorp.StartRQ{
			Name: f.curStep.Keyword + " " + f.curStep.Name,
			//Description: f.curStep.Keyword,
			StartTime: gorp.Timestamp{},
		},
		LaunchID: f.currentLaunchID,
		Type:     "STEP",
	}
	testRsData, err := f.goRP.StartChildTest(f.currentTestID, testRqData)
	if err != nil {
		fmt.Println("Error on adding step to RP")
		fmt.Println(err.Error())
	}
	f.currentStepID = testRsData.ID
}

// finishChildStep - Finish BDD Step
func (f *rp) finishChildStep() {
	// Send log Error if there is an error in step
	if f.curStep.Result.Error != "" {
		logTest := &gorp.SaveLogRQ{
			ItemID:  f.currentStepID,
			LogTime: gorp.Timestamp{},
			Message: f.curStep.Result.Error,
			Level:   "Error",
		}
		_, err := f.goRP.SaveLog(logTest)
		if err != nil {
			fmt.Println("Error on saving log to RP")
			fmt.Println(err.Error())
		}
	}

	// Finish Step
	testRqData := &gorp.FinishTestRQ{
		FinishExecutionRQ: gorp.FinishExecutionRQ{
			EndTime: gorp.Timestamp{},
			Status:  f.curStep.Result.Status,
		},
	}
	_, err := f.goRP.FinishTest(f.currentStepID, testRqData)
	if err != nil {
		fmt.Println("Error on finishing step to RP")
		fmt.Println(err.Error())
	}
}
