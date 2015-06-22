package godog

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/DATA-DOG/godog/gherkin"
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
	features    []*gherkin.Feature

	failed    []*failed
	passed    []*passed
	skipped   []*skipped
	undefined []*undefined
}

func (f *progress) Node(node interface{}) {
	switch t := node.(type) {
	case *gherkin.Feature:
		f.features = append(f.features, t)
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
			fmt.Println(s(4) + cl(fail.step.Token.Keyword+" "+fail.step.Text, red) + cl(" # "+fail.line(), black))
			fmt.Println(s(6) + cl("Error: ", red) + bcl(fail.err, red) + "\n")
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
	}
	f.steps += 1
	if math.Mod(float64(f.steps), float64(f.stepsPerRow)) == 0 {
		fmt.Printf(" %d\n", f.steps)
	}
}

func (f *progress) Passed(step *gherkin.Step, match *StepDef) {
	s := &passed{step: step, def: match}
	f.passed = append(f.passed, s)
	f.step(s)
}

func (f *progress) Skipped(step *gherkin.Step) {
	s := &skipped{step: step}
	f.skipped = append(f.skipped, s)
	f.step(s)
}

func (f *progress) Undefined(step *gherkin.Step) {
	s := &undefined{step: step}
	f.undefined = append(f.undefined, s)
	f.step(s)
}

func (f *progress) Failed(step *gherkin.Step, match *StepDef, err error) {
	s := &failed{step: step, def: match, err: err}
	f.failed = append(f.failed, s)
	f.step(s)
}
