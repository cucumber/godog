package godog

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/DATA-DOG/godog/gherkin"
)

func init() {
	Format("text", "Prints every feature with runtime statuses without ANSI colors.", &text{
		basefmt: basefmt{
			started: time.Now(),
			indent:  2,
		},
	})
}

//var outlinePlaceholderRegexp = regexp.MustCompile("<[^>]+>")

// a built in default text formatter
type text struct {
	basefmt

	// currently processed
	feature            *gherkin.Feature
	scenario           *gherkin.Scenario
	outline            *gherkin.ScenarioOutline

	// state
	bgSteps            int
	steps              int
	commentPos         int

	// outline
	outlineSteps       []*stepResult
	outlineNumExample  int
	outlineNumExamples int
}

func (f *text) Feature(ft *gherkin.Feature, p string) {
	if len(f.features) != 0 {
		// not a first feature, add a newline
		fmt.Println("")
	}
	f.features = append(f.features, &feature{Path: p, Feature: ft})
	fmt.Println(ft.Keyword + ": " + ft.Name)
	if strings.TrimSpace(ft.Description) != "" {
		for _, line := range strings.Split(ft.Description, "\n") {
			fmt.Println(s(f.indent) + strings.TrimSpace(line))
		}
	}

	f.feature = ft
	f.scenario = nil
	f.outline = nil
	f.bgSteps = 0
	if ft.Background != nil {
		f.bgSteps = len(ft.Background.Steps)
	}
}

// Node takes a gherkin node for formatting
func (f *text) Node(node interface{}) {
	f.basefmt.Node(node)

	switch t := node.(type) {
	case *gherkin.Examples:
		f.outlineNumExamples = len(t.TableBody)
		f.outlineNumExample++
	case *gherkin.Scenario:
		f.scenario = t
		f.outline = nil
		f.steps = len(t.Steps)
	case *gherkin.ScenarioOutline:
		f.outline = t
		f.scenario = nil
		f.steps = len(t.Steps)
		f.outlineNumExample = -1
	}
}

// Summary sumarize the feature formatter output
func (f *text) Summary() {
	// failed steps on background are not scenarios
	var failedScenarios []*stepResult
	for _, fail := range f.failed {
		switch fail.owner.(type) {
		case *gherkin.Scenario:
			failedScenarios = append(failedScenarios, fail)
		case *gherkin.ScenarioOutline:
			failedScenarios = append(failedScenarios, fail)
		}
	}
	if len(failedScenarios) > 0 {
		fmt.Println("\n--- " + "Failed scenarios:" + "\n")
		var unique []string
		for _, fail := range failedScenarios {
			var found bool
			for _, in := range unique {
				if in == fail.line() {
					found = true
					break
				}
			}
			if !found {
				unique = append(unique, fail.line())
			}
		}

		for _, fail := range unique {
			fmt.Println("    " + fail, red)
		}
	}
	f.basefmt.Summary()
}

func (f *text) printOutlineExample(outline *gherkin.ScenarioOutline) {
	var msg string

	ex := outline.Examples[f.outlineNumExample]
	example, hasExamples := examples(ex)
	if !hasExamples {
		// do not print empty examples
		return
	}

	firstExample := f.outlineNumExamples == len(example.TableBody)
	printSteps := firstExample && f.outlineNumExample == 0

	for i, res := range f.outlineSteps {
		// determine example row status
		switch {
		case res.typ == failed:
			msg = res.err.Error()
		}
		if printSteps {
			// in first example, we need to print steps
			var text string
			ostep := outline.Steps[i]
			if res.def != nil {
				if m := outlinePlaceholderRegexp.FindAllStringIndex(ostep.Text, -1); len(m) > 0 {
					var pos int
					for i := 0; i < len(m); i++ {
						pair := m[i]
						text += ostep.Text[pos:pair[0]]
						text += ostep.Text[pair[0]:pair[1]]
						pos = pair[1]
					}
					text += ostep.Text[pos:len(ostep.Text)]
				} else {
					text = ostep.Text
				}
				text += s(f.commentPos - f.length(ostep) + 1) + fmt.Sprintf("# %s", res.def.funcName())
			} else {
				text = ostep.Text
			}
			// print the step outline
			fmt.Println(s(f.indent * 2) + strings.TrimSpace(ostep.Keyword) + " " + text)
		}
	}

	cells := make([]string, len(example.TableHeader.Cells))
	max := longest(example)
	// an example table header
	if firstExample {
		fmt.Println("")
		fmt.Println(s(f.indent * 2) + example.Keyword + ": " + example.Name)

		for i, cell := range example.TableHeader.Cells {
			cells[i] = cell.Value + s(max[i] - len(cell.Value))
		}
		fmt.Println(s(f.indent * 3) + "| " + strings.Join(cells, " | ") + " |")
	}

	// an example table row
	row := example.TableBody[len(example.TableBody) - f.outlineNumExamples]
	for i, cell := range row.Cells {
		cells[i] = cell.Value + s(max[i] - len(cell.Value))
	}
	fmt.Println(s(f.indent * 3) + "| " + strings.Join(cells, " | ") + " |")

	// if there is an error
	if msg != "" {
		fmt.Println(s(f.indent * 4) + msg)
	}
}

func (f *text) printStep(step *gherkin.Step, def *StepDef) {
	text := s(f.indent * 2) + strings.TrimSpace(step.Keyword) + " "
	switch {
	case def != nil:
		if m := (def.Expr.FindStringSubmatchIndex(step.Text))[2:]; len(m) > 0 {
			var pos, i int
			for pos, i = 0, 0; i < len(m); i++ {
				if math.Mod(float64(i), 2) == 0 {
					text += step.Text[pos:m[i]]
				} else {
					text += step.Text[pos:m[i]]
				}
				pos = m[i]
			}
			text += step.Text[pos:len(step.Text)]
		} else {
			text += step.Text
		}
		text += s(f.commentPos - f.length(step) + 1) + fmt.Sprintf("# %s", def.funcName())
	default:
		text += step.Text
	}

	fmt.Println(text)
	switch t := step.Argument.(type) {
	case *gherkin.DataTable:
		f.printTable(t)
	case *gherkin.DocString:
		var ct string
		if len(t.ContentType) > 0 {
			ct = " " + t.ContentType
		}
		fmt.Println(s(f.indent * 3) + t.Delimitter + ct)
		for _, ln := range strings.Split(t.Content, "\n") {
			fmt.Println(s(f.indent * 3) + ln)
		}
		fmt.Println(s(f.indent * 3) + t.Delimitter)
	}
}

func (f *text) printStepKind(res *stepResult) {
	_, isBgStep := res.owner.(*gherkin.Background)

	// if has not printed background yet
	switch {
	// first background step
	case f.bgSteps > 0 && f.bgSteps == len(f.feature.Background.Steps):
		f.commentPos = f.longestStep(f.feature.Background.Steps, f.length(f.feature.Background))
		fmt.Println("\n" + s(f.indent) + f.feature.Background.Keyword + ": " + f.feature.Background.Name)
		f.bgSteps--
	// subsequent background steps
	case f.bgSteps > 0:
		f.bgSteps--
	// a background step for another scenario, but all bg steps are
	// already printed. so just skip it
	case isBgStep:
		return
	// first step of scenario, print header and calculate comment position
	case f.scenario != nil && f.steps == len(f.scenario.Steps):
		f.commentPos = f.longestStep(f.scenario.Steps, f.length(f.scenario))
		text := s(f.indent) + f.scenario.Keyword + ": " + f.scenario.Name
		text += s(f.commentPos - f.length(f.scenario) + 1) + f.line(f.scenario.Location)
		fmt.Println("\n" + text)
		f.steps--
	// all subsequent scenario steps
	case f.scenario != nil:
		f.steps--
	// first step of outline scenario, print header and calculate comment position
	case f.outline != nil && f.steps == len(f.outline.Steps):
		f.commentPos = f.longestStep(f.outline.Steps, f.length(f.outline))
		text := s(f.indent) + f.outline.Keyword + ": " + f.outline.Name
		text += s(f.commentPos - f.length(f.outline) + 1) + f.line(f.outline.Location)
		fmt.Println("\n" + text)
		f.outlineSteps = append(f.outlineSteps, res)
		f.steps--
		if len(f.outlineSteps) == len(f.outline.Steps) {
			// an outline example steps has went through
			f.printOutlineExample(f.outline)
			f.outlineSteps = []*stepResult{}
			f.outlineNumExamples--
		}
		return
	// all subsequent outline steps
	case f.outline != nil:
		f.outlineSteps = append(f.outlineSteps, res)
		f.steps--
		if len(f.outlineSteps) == len(f.outline.Steps) {
			// an outline example steps has went through
			f.printOutlineExample(f.outline)
			f.outlineSteps = []*stepResult{}
			f.outlineNumExamples--
		}
		return
	}

	f.printStep(res.step, res.def)
	if res.err != nil {
		fmt.Println(s(f.indent * 2) + res.err.Error())
	}
	if res.typ == pending {
		fmt.Println(s(f.indent * 3) + "TODO: write pending definition")
	}
}

// print table with aligned table cells
func (f *text) printTable(t *gherkin.DataTable) {
	var l = longest(t)
	var cols = make([]string, len(t.Rows[0].Cells))
	for _, row := range t.Rows {
		for i, cell := range row.Cells {
			cols[i] = cell.Value + s(l[i] - len(cell.Value))
		}
		fmt.Println(s(f.indent * 3) + "| " + strings.Join(cols, " | ") + " |")
	}
}

func (f *text) Passed(step *gherkin.Step, match *StepDef) {
	f.basefmt.Passed(step, match)
	f.printStepKind(f.passed[len(f.passed) - 1])
}

func (f *text) Skipped(step *gherkin.Step) {
	f.basefmt.Skipped(step)
	f.printStepKind(f.skipped[len(f.skipped) - 1])
}

func (f *text) Undefined(step *gherkin.Step) {
	f.basefmt.Undefined(step)
	f.printStepKind(f.undefined[len(f.undefined) - 1])
}

func (f *text) Failed(step *gherkin.Step, match *StepDef, err error) {
	f.basefmt.Failed(step, match, err)
	f.printStepKind(f.failed[len(f.failed) - 1])
}

func (f *text) Pending(step *gherkin.Step, match *StepDef) {
	f.basefmt.Pending(step, match)
	f.printStepKind(f.pending[len(f.pending) - 1])
}


func (f *text) longestStep(steps []*gherkin.Step, base int) int {
	ret := base
	for _, step := range steps {
		length := f.length(step)
		if length > ret {
			ret = length
		}
	}
	return ret
}

// a line number representation in feature file
func (f *text) line(loc *gherkin.Location) string {
	return fmt.Sprintf("# %s:%d", f.features[len(f.features) - 1].Path, loc.Line)
}

func (f *text) length(node interface{}) int {
	switch t := node.(type) {
	case *gherkin.Background:
		return f.indent + utf8.RuneCountInString(strings.TrimSpace(t.Keyword) + ": " + t.Name)
	case *gherkin.Step:
		return f.indent * 2 + utf8.RuneCountInString(strings.TrimSpace(t.Keyword) + " " + t.Text)
	case *gherkin.Scenario:
		return f.indent + utf8.RuneCountInString(strings.TrimSpace(t.Keyword) + ": " + t.Name)
	case *gherkin.ScenarioOutline:
		return f.indent + utf8.RuneCountInString(strings.TrimSpace(t.Keyword) + ": " + t.Name)
	}
	panic(fmt.Sprintf("unexpected node %T to determine length", node))
}
