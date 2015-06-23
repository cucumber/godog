package godog

import (
	"fmt"
	"math"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/DATA-DOG/godog/gherkin"
)

func init() {
	RegisterFormatter("pretty", "Prints every feature with runtime statuses.", &pretty{
		started: time.Now(),
	})
}

var outlinePlaceholderRegexp = regexp.MustCompile("<[^>]+>")

// a built in default pretty formatter
type pretty struct {
	feature         *gherkin.Feature
	commentPos      int
	backgroundSteps int

	// outline
	outlineExamples int
	outlineNumSteps int
	outlineSteps    []interface{}

	// summary
	started   time.Time
	features  []*gherkin.Feature
	failed    []*failed
	passed    []*passed
	skipped   []*skipped
	undefined []*undefined
}

// a line number representation in feature file
func (f *pretty) line(tok *gherkin.Token) string {
	return cl(fmt.Sprintf("# %s:%d", f.feature.Path, tok.Line), black)
}

// Node takes a gherkin node for formatting
func (f *pretty) Node(node interface{}) {
	switch t := node.(type) {
	case *gherkin.Feature:
		if f.feature != nil {
			// not a first feature, add a newline
			fmt.Println("")
		}
		f.feature = t
		f.features = append(f.features, t)
		// print feature header
		fmt.Println(bcl(t.Token.Keyword+": ", white) + t.Title)
		fmt.Println(t.Description)
		// print background header
		if t.Background != nil {
			f.commentPos = longestStep(t.Background.Steps, t.Background.Token.Length())
			f.backgroundSteps = len(t.Background.Steps)
			fmt.Println("\n" + s(t.Background.Token.Indent) + bcl(t.Background.Token.Keyword+":", white))
		}
	case *gherkin.Scenario:
		f.commentPos = longestStep(t.Steps, t.Token.Length())
		if t.Outline != nil {
			f.outlineSteps = []interface{}{} // reset steps list
			f.commentPos = longestStep(t.Outline.Steps, t.Token.Length())
			if f.outlineExamples == 0 {
				f.outlineNumSteps = len(t.Outline.Steps)
				f.outlineExamples = len(t.Outline.Examples.Rows) - 1
			} else {
				return // already printed an outline
			}
		}
		text := s(t.Token.Indent) + bcl(t.Token.Keyword+": ", white) + t.Title
		text += s(f.commentPos-t.Token.Length()+1) + f.line(t.Token)
		fmt.Println("\n" + text)
	}
}

// Summary sumarize the feature formatter output
func (f *pretty) Summary() {
	// failed steps on background are not scenarios
	var failedScenarios []*failed
	for _, fail := range f.failed {
		if fail.step.Scenario != nil {
			failedScenarios = append(failedScenarios, fail)
		}
	}
	if len(failedScenarios) > 0 {
		fmt.Println("\n--- " + cl("Failed scenarios:", red) + "\n")
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
			fmt.Println("    " + cl(fail, red))
		}
	}
	var total, passed int
	for _, ft := range f.features {
		total += len(ft.Scenarios)
	}
	passed = total

	var steps, parts, scenarios []string
	nsteps := len(f.passed) + len(f.failed) + len(f.skipped) + len(f.undefined)
	if len(f.passed) > 0 {
		steps = append(steps, cl(fmt.Sprintf("%d passed", len(f.passed)), green))
	}
	if len(f.failed) > 0 {
		passed -= len(f.failed)
		parts = append(parts, cl(fmt.Sprintf("%d failed", len(f.failed)), red))
		steps = append(steps, parts[len(parts)-1])
	}
	if len(f.skipped) > 0 {
		steps = append(steps, cl(fmt.Sprintf("%d skipped", len(f.skipped)), cyan))
	}
	if len(f.undefined) > 0 {
		passed -= len(f.undefined)
		parts = append(parts, cl(fmt.Sprintf("%d undefined", len(f.undefined)), yellow))
		steps = append(steps, parts[len(parts)-1])
	}
	if passed > 0 {
		scenarios = append(scenarios, cl(fmt.Sprintf("%d passed", passed), green))
	}
	scenarios = append(scenarios, parts...)
	elapsed := time.Since(f.started)

	fmt.Println("")
	if total == 0 {
		fmt.Println("No scenarios")
	} else {
		fmt.Println(fmt.Sprintf("%d scenarios (%s)", total, strings.Join(scenarios, ", ")))
	}

	if nsteps == 0 {
		fmt.Println("No steps")
	} else {
		fmt.Println(fmt.Sprintf("%d steps (%s)", nsteps, strings.Join(steps, ", ")))
	}
	fmt.Println(elapsed)
}

func (f *pretty) printOutlineExample(scenario *gherkin.Scenario) {
	var failed error
	clr := green
	tbl := scenario.Outline.Examples
	firstExample := f.outlineExamples == len(tbl.Rows)-1

	for i, act := range f.outlineSteps {
		var c color
		var def *StepDef
		var err error

		_, def, c, err = f.stepDetails(act)
		// determine example row status
		switch {
		case err != nil:
			failed = err
			clr = red
		case c == yellow:
			clr = yellow
		case c == cyan && clr == green:
			clr = cyan
		}
		if firstExample {
			// in first example, we need to print steps
			var text string
			ostep := scenario.Outline.Steps[i]
			if def != nil {
				if m := outlinePlaceholderRegexp.FindAllStringIndex(ostep.Text, -1); len(m) > 0 {
					var pos int
					for i := 0; i < len(m); i++ {
						pair := m[i]
						text += cl(ostep.Text[pos:pair[0]], cyan)
						text += bcl(ostep.Text[pair[0]:pair[1]], cyan)
						pos = pair[1]
					}
					text += cl(ostep.Text[pos:len(ostep.Text)], cyan)
				} else {
					text = cl(ostep.Text, cyan)
				}
				// use reflect to get step handler function name
				name := runtime.FuncForPC(reflect.ValueOf(def.Handler).Pointer()).Name()
				text += s(f.commentPos-ostep.Token.Length()+1) + cl(fmt.Sprintf("# %s", name), black)
			} else {
				text = cl(ostep.Text, cyan)
			}
			// print the step outline
			fmt.Println(s(ostep.Token.Indent) + cl(ostep.Token.Keyword, cyan) + " " + text)
		}
	}

	cols := make([]string, len(tbl.Rows[0]))
	max := longest(tbl)
	// an example table header
	if firstExample {
		out := scenario.Outline
		fmt.Println("")
		fmt.Println(s(out.Token.Indent) + bcl(out.Token.Keyword+":", white))
		row := tbl.Rows[0]

		for i, col := range row {
			cols[i] = cl(col, cyan) + s(max[i]-len(col))
		}
		fmt.Println(s(tbl.Token.Indent) + "| " + strings.Join(cols, " | ") + " |")
	}

	// an example table row
	row := tbl.Rows[len(tbl.Rows)-f.outlineExamples]
	for i, col := range row {
		cols[i] = cl(col, clr) + s(max[i]-len(col))
	}
	fmt.Println(s(tbl.Token.Indent) + "| " + strings.Join(cols, " | ") + " |")

	// if there is an error
	if failed != nil {
		fmt.Println(s(tbl.Token.Indent) + bcl(failed, red))
	}
}

func (f *pretty) printStep(step *gherkin.Step, def *StepDef, c color) {
	text := s(step.Token.Indent) + cl(step.Token.Keyword, c) + " "
	switch {
	case def != nil:
		if m := (def.Expr.FindStringSubmatchIndex(step.Text))[2:]; len(m) > 0 {
			var pos, i int
			for pos, i = 0, 0; i < len(m); i++ {
				if math.Mod(float64(i), 2) == 0 {
					text += cl(step.Text[pos:m[i]], c)
				} else {
					text += bcl(step.Text[pos:m[i]], c)
				}
				pos = m[i]
			}
			text += cl(step.Text[pos:len(step.Text)], c)
		} else {
			text += cl(step.Text, c)
		}
		// use reflect to get step handler function name
		name := runtime.FuncForPC(reflect.ValueOf(def.Handler).Pointer()).Name()
		text += s(f.commentPos-step.Token.Length()+1) + cl(fmt.Sprintf("# %s", name), black)
	default:
		text += cl(step.Text, c)
	}

	fmt.Println(text)
	if step.PyString != nil {
		fmt.Println(s(step.Token.Indent+2) + cl(`"""`, c))
		fmt.Println(cl(step.PyString.Raw, c))
		fmt.Println(s(step.Token.Indent+2) + cl(`"""`, c))
	}
	if step.Table != nil {
		f.printTable(step.Table, c)
	}
}

func (f *pretty) stepDetails(stepAction interface{}) (step *gherkin.Step, def *StepDef, c color, err error) {
	switch typ := stepAction.(type) {
	case *passed:
		step = typ.step
		def = typ.def
		c = green
	case *failed:
		step = typ.step
		def = typ.def
		err = typ.err
		c = red
	case *skipped:
		step = typ.step
		c = cyan
	case *undefined:
		step = typ.step
		c = yellow
	default:
		fatal(fmt.Errorf("unexpected step type received: %T", typ))
	}
	return
}

func (f *pretty) printStepKind(stepAction interface{}) {
	var c color
	var step *gherkin.Step
	var def *StepDef
	var err error

	step, def, c, err = f.stepDetails(stepAction)

	// do not print background more than once
	switch {
	case step.Background != nil && f.backgroundSteps == 0:
		return
	case step.Background != nil && f.backgroundSteps > 0:
		f.backgroundSteps--
	}

	if f.outlineExamples != 0 {
		f.outlineSteps = append(f.outlineSteps, stepAction)
		if len(f.outlineSteps) == f.outlineNumSteps {
			// an outline example steps has went through
			f.printOutlineExample(step.Scenario)
			f.outlineExamples--
		}
		return // wait till example steps
	}

	f.printStep(step, def, c)
	if err != nil {
		fmt.Println(s(step.Token.Indent) + bcl(err, red))
	}
}

// print table with aligned table cells
func (f *pretty) printTable(t *gherkin.Table, c color) {
	var l = longest(t)
	var cols = make([]string, len(t.Rows[0]))
	for _, row := range t.Rows {
		for i, col := range row {
			cols[i] = col + s(l[i]-len(col))
		}
		fmt.Println(s(t.Token.Indent) + cl("| "+strings.Join(cols, " | ")+" |", c))
	}
}

// Passed is called to represent a passed step
func (f *pretty) Passed(step *gherkin.Step, match *StepDef) {
	s := &passed{step: step, def: match}
	f.printStepKind(s)
	f.passed = append(f.passed, s)
}

// Skipped is called to represent a passed step
func (f *pretty) Skipped(step *gherkin.Step) {
	s := &skipped{step: step}
	f.printStepKind(s)
	f.skipped = append(f.skipped, s)
}

// Undefined is called to represent a pending step
func (f *pretty) Undefined(step *gherkin.Step) {
	s := &undefined{step: step}
	f.printStepKind(s)
	f.undefined = append(f.undefined, s)
}

// Failed is called to represent a failed step
func (f *pretty) Failed(step *gherkin.Step, match *StepDef, err error) {
	s := &failed{step: step, def: match, err: err}
	f.printStepKind(s)
	f.failed = append(f.failed, s)
}

// longest gives a list of longest columns of all rows in Table
func longest(t *gherkin.Table) []int {
	var longest = make([]int, len(t.Rows[0]))
	for _, row := range t.Rows {
		for i, col := range row {
			if longest[i] < len(col) {
				longest[i] = len(col)
			}
		}
	}
	return longest
}

func longestStep(steps []*gherkin.Step, base int) int {
	ret := base
	for _, step := range steps {
		length := step.Token.Length()
		if length > ret {
			ret = length
		}
	}
	return ret
}
