package godog

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/cucumber/gherkin-go"
)

func init() {
	RegisterFormatter("progress", "Prints a character per step.", &progress{
		started:     time.Now(),
		stepsPerRow: 70,
	})
}

type progress struct {
	stepsPerRow int
	started     time.Time
	steps       int

	features []*feature
	owner    interface{}

	failed    []*failed
	passed    []*passed
	skipped   []*skipped
	undefined []*undefined
	pending   []*pending
}

func (f *progress) Feature(ft *gherkin.Feature, p string) {
	f.features = append(f.features, &feature{Path: p, Feature: ft})
}

func (f *progress) Node(n interface{}) {
	switch t := n.(type) {
	case *gherkin.ScenarioOutline:
		f.owner = t
	case *gherkin.Scenario:
		f.owner = t
	case *gherkin.Background:
		f.owner = t
	}
}

func (f *progress) Summary() {
	left := math.Mod(float64(f.steps), float64(f.stepsPerRow))
	if left != 0 {
		if int(f.steps) > f.stepsPerRow {
			fmt.Printf(s(f.stepsPerRow-int(left)) + fmt.Sprintf(" %d\n", f.steps))
		} else {
			fmt.Printf(" %d\n", f.steps)
		}
	}
	fmt.Println("")

	if len(f.failed) > 0 {
		fmt.Println("\n--- " + cl("Failed steps:", red) + "\n")
		for _, fail := range f.failed {
			fmt.Println(s(4) + cl(fail.step.Keyword+" "+fail.step.Text, red) + cl(" # "+fail.line(), black))
			fmt.Println(s(6) + cl("Error: ", red) + bcl(fail.err, red) + "\n")
		}
	}
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
	if len(f.pending) > 0 {
		steps = append(steps, cl(fmt.Sprintf("%d pending", len(f.pending)), yellow))
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
		passed -= undefined
		parts = append(parts, cl(fmt.Sprintf("%d undefined", undefined), yellow))
		steps = append(steps, cl(fmt.Sprintf("%d undefined", len(f.undefined)), yellow))
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

func (f *progress) step(step interface{}) {
	switch step.(type) {
	case *passed:
		fmt.Print(cl(".", green))
	case *skipped:
		fmt.Print(cl("-", cyan))
	case *failed:
		fmt.Print(cl("F", red))
	case *undefined:
		fmt.Print(cl("U", yellow))
	case *pending:
		fmt.Print(cl("P", yellow))
	}
	f.steps++
	if math.Mod(float64(f.steps), float64(f.stepsPerRow)) == 0 {
		fmt.Printf(" %d\n", f.steps)
	}
}

func (f *progress) Passed(step *gherkin.Step, match *StepDef) {
	s := &passed{owner: f.owner, feature: f.features[len(f.features)-1], step: step, def: match}
	f.passed = append(f.passed, s)
	f.step(s)
}

func (f *progress) Skipped(step *gherkin.Step) {
	s := &skipped{owner: f.owner, feature: f.features[len(f.features)-1], step: step}
	f.skipped = append(f.skipped, s)
	f.step(s)
}

func (f *progress) Undefined(step *gherkin.Step) {
	s := &undefined{owner: f.owner, feature: f.features[len(f.features)-1], step: step}
	f.undefined = append(f.undefined, s)
	f.step(s)
}

func (f *progress) Failed(step *gherkin.Step, match *StepDef, err error) {
	s := &failed{owner: f.owner, feature: f.features[len(f.features)-1], step: step, def: match, err: err}
	f.failed = append(f.failed, s)
	f.step(s)
}

func (f *progress) Pending(step *gherkin.Step, match *StepDef) {
	s := &pending{owner: f.owner, feature: f.features[len(f.features)-1], step: step, def: match}
	f.pending = append(f.pending, s)
	f.step(s)
}
