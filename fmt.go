package godog

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/cucumber/gherkin-go"
)

// some snippet formatting regexps
var snippetExprCleanup = regexp.MustCompile("([\\/\\[\\]\\(\\)\\\\^\\$\\.\\|\\?\\*\\+\\'])")
var snippetExprQuoted = regexp.MustCompile("(\\s*|^)\"(?:[^\"]*)\"(\\s+|$)")
var snippetMethodName = regexp.MustCompile("[^a-zA-Z\\_\\ ]")
var snippetNumbers = regexp.MustCompile("(\\d+)")

var snippetHelperFuncs = template.FuncMap{
	"backticked": func(s string) string {
		return "`" + s + "`"
	},
}

var undefinedSnippetsTpl = template.Must(template.New("snippets").Funcs(snippetHelperFuncs).Parse(`
{{ range . }}func {{ .Method }}({{ .Args }}) error {
	return godog.ErrPending
}

{{end}}func featureContext(s godog.Suite) { {{ range . }}
	s.Step({{ backticked .Expr }}, {{ .Method }}){{end}}
}
`))

type undefinedSnippet struct {
	Method   string
	Expr     string
	argument interface{} // gherkin step argument
}

type registeredFormatter struct {
	name        string
	fmt         Formatter
	description string
}

var formatters []*registeredFormatter

// RegisterFormatter registers a feature suite output
// Formatter as the name and descriptiongiven.
// Formatter is used to represent suite output
func RegisterFormatter(name, description string, f Formatter) {
	formatters = append(formatters, &registeredFormatter{
		name:        name,
		fmt:         f,
		description: description,
	})
}

// Formatter is an interface for feature runner
// output summary presentation.
//
// New formatters may be created to represent
// suite results in different ways. These new
// formatters needs to be registered with a
// RegisterFormatter function call
type Formatter interface {
	Feature(*gherkin.Feature, string)
	Node(interface{})
	Failed(*gherkin.Step, *StepDef, error)
	Passed(*gherkin.Step, *StepDef)
	Skipped(*gherkin.Step)
	Undefined(*gherkin.Step)
	Pending(*gherkin.Step, *StepDef)
	Summary()
}

// failed represents a failed step data structure
// with all necessary references
type failed struct {
	feature *feature
	owner   interface{}
	step    *gherkin.Step
	def     *StepDef
	err     error
}

func (f failed) line() string {
	return fmt.Sprintf("%s:%d", f.feature.Path, f.step.Location.Line)
}

// passed represents a successful step data structure
// with all necessary references
type passed struct {
	feature *feature
	owner   interface{}
	step    *gherkin.Step
	def     *StepDef
}

// skipped represents a skipped step data structure
// with all necessary references
type skipped struct {
	feature *feature
	owner   interface{}
	step    *gherkin.Step
}

// undefined represents an undefined step data structure
// with all necessary references
type undefined struct {
	feature *feature
	owner   interface{}
	step    *gherkin.Step
}

// pending represents a pending step data structure
// with all necessary references
type pending struct {
	feature *feature
	owner   interface{}
	step    *gherkin.Step
	def     *StepDef
}

type basefmt struct {
	owner  interface{}
	indent int

	started   time.Time
	features  []*feature
	failed    []*failed
	passed    []*passed
	skipped   []*skipped
	undefined []*undefined
	pending   []*pending
}

func (f *basefmt) Node(n interface{}) {
	switch t := n.(type) {
	case *gherkin.ScenarioOutline:
		f.owner = t
	case *gherkin.Scenario:
		f.owner = t
	case *gherkin.Background:
		f.owner = t
	}
}

func (f *basefmt) Feature(ft *gherkin.Feature, p string) {
	f.features = append(f.features, &feature{Path: p, Feature: ft})
}

func (f *basefmt) Passed(step *gherkin.Step, match *StepDef) {
	s := &passed{owner: f.owner, feature: f.features[len(f.features)-1], step: step, def: match}
	f.passed = append(f.passed, s)
}

func (f *basefmt) Skipped(step *gherkin.Step) {
	s := &skipped{owner: f.owner, feature: f.features[len(f.features)-1], step: step}
	f.skipped = append(f.skipped, s)
}

func (f *basefmt) Undefined(step *gherkin.Step) {
	s := &undefined{owner: f.owner, feature: f.features[len(f.features)-1], step: step}
	f.undefined = append(f.undefined, s)
}

func (f *basefmt) Failed(step *gherkin.Step, match *StepDef, err error) {
	s := &failed{owner: f.owner, feature: f.features[len(f.features)-1], step: step, def: match, err: err}
	f.failed = append(f.failed, s)
}

func (f *basefmt) Pending(step *gherkin.Step, match *StepDef) {
	s := &pending{owner: f.owner, feature: f.features[len(f.features)-1], step: step, def: match}
	f.pending = append(f.pending, s)
}

func (f *basefmt) Summary() {
	var total, passed, undefined int
	for _, ft := range f.features {
		for _, def := range ft.ScenarioDefinitions {
			switch t := def.(type) {
			case *gherkin.Scenario:
				total++
			case *gherkin.ScenarioOutline:
				for _, ex := range t.Examples {
					total += len(ex.TableBody)
				}
			}
		}
	}
	passed = total
	var owner interface{}
	for _, undef := range f.undefined {
		if owner != undef.owner {
			undefined++
			owner = undef.owner
		}
	}

	var steps, parts, scenarios []string
	nsteps := len(f.passed) + len(f.failed) + len(f.skipped) + len(f.undefined) + len(f.pending)
	if len(f.passed) > 0 {
		steps = append(steps, cl(fmt.Sprintf("%d passed", len(f.passed)), green))
	}
	if len(f.failed) > 0 {
		passed -= len(f.failed)
		parts = append(parts, cl(fmt.Sprintf("%d failed", len(f.failed)), red))
		steps = append(steps, parts[len(parts)-1])
	}
	if len(f.pending) > 0 {
		passed -= len(f.pending)
		parts = append(parts, cl(fmt.Sprintf("%d pending", len(f.pending)), yellow))
		steps = append(steps, cl(fmt.Sprintf("%d pending", len(f.pending)), yellow))
	}
	if len(f.undefined) > 0 {
		passed -= undefined
		parts = append(parts, cl(fmt.Sprintf("%d undefined", undefined), yellow))
		steps = append(steps, cl(fmt.Sprintf("%d undefined", len(f.undefined)), yellow))
	}
	if len(f.skipped) > 0 {
		steps = append(steps, cl(fmt.Sprintf("%d skipped", len(f.skipped)), cyan))
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

	f.snippets()
}

func (s *undefinedSnippet) Args() string {
	var args []string
	var pos, idx int
	var breakLoop bool
	for !breakLoop {
		part := s.Expr[pos:]
		ipos := strings.Index(part, "(\\d+)")
		spos := strings.Index(part, "\"([^\"]*)\"")
		switch {
		case spos == -1 && ipos == -1:
			breakLoop = true
		case spos == -1:
			idx++
			pos += ipos + len("(\\d+)")
			args = append(args, fmt.Sprintf("arg%d int", idx))
		case ipos == -1:
			idx++
			pos += spos + len("\"([^\"]*)\"")
			args = append(args, fmt.Sprintf("arg%d string", idx))
		case ipos < spos:
			idx++
			pos += ipos + len("(\\d+)")
			args = append(args, fmt.Sprintf("arg%d int", idx))
		case spos < ipos:
			idx++
			pos += spos + len("\"([^\"]*)\"")
			args = append(args, fmt.Sprintf("arg%d string", idx))
		}
	}
	if s.argument != nil {
		idx++
		switch s.argument.(type) {
		case *gherkin.DocString:
			args = append(args, fmt.Sprintf("arg%d *gherkin.DocString", idx))
		case *gherkin.DataTable:
			args = append(args, fmt.Sprintf("arg%d *gherkin.DataTable", idx))
		}
	}
	return strings.Join(args, ", ")
}

func (f *basefmt) snippets() {
	if len(f.undefined) == 0 {
		return
	}

	fmt.Println(cl("\nYou can implement step definitions for undefined steps with these snippets:", yellow))

	var index int
	var snips []*undefinedSnippet
	// build snippets
	for _, u := range f.undefined {
		expr := snippetExprCleanup.ReplaceAllString(u.step.Text, "\\\\$1")
		expr = snippetNumbers.ReplaceAllString(expr, "(\\d+)")
		expr = snippetExprQuoted.ReplaceAllString(expr, " \"([^\"]*)\" ")
		expr = "^" + strings.TrimSpace(expr) + "$"

		name := snippetNumbers.ReplaceAllString(u.step.Text, "")
		name = snippetExprQuoted.ReplaceAllString(name, "")
		name = snippetMethodName.ReplaceAllString(name, "")
		var words []string
		for i, w := range strings.Split(name, " ") {
			if i != 0 {
				w = strings.Title(w)
			} else {
				w = string(unicode.ToLower(rune(w[0]))) + w[1:]
			}
			words = append(words, w)
		}
		name = strings.Join(words, "")
		if len(name) == 0 {
			index++
			name = fmt.Sprintf("stepDefinition%d", index)
		}

		var found bool
		for _, snip := range snips {
			if snip.Expr == expr {
				found = true
				break
			}
		}
		if !found {
			snips = append(snips, &undefinedSnippet{Method: name, Expr: expr, argument: u.step.Argument})
		}
	}

	var buf bytes.Buffer
	if err := undefinedSnippetsTpl.Execute(&buf, snips); err != nil {
		panic(err)
	}
	fmt.Println(cl(buf.String(), yellow))
}
