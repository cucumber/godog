package godog

import (
	"flag"
	"fmt"
	"regexp"

	"github.com/DATA-DOG/godog/gherkin"
)

// Suite is an interface which allows various contexts
// to register step definitions and event handlers
type Suite interface {
	Step(exp *regexp.Regexp, h StepHandler)
}

type suite struct {
	steps    map[*regexp.Regexp]StepHandler
	features []*gherkin.Feature
}

// New initializes a suite which supports the Suite
// interface. The instance is passed around to all
// context initialization functions from *_test.go files
func New() *suite {
	if !flag.Parsed() {
		flag.Parse()
	}
	return &suite{
		steps: make(map[*regexp.Regexp]StepHandler),
	}
}

func (s *suite) Step(exp *regexp.Regexp, h StepHandler) {
	s.steps[exp] = h
}

// Run - runs a godog feature suite
func (s *suite) Run() {
	var err error
	s.features, err = cfg.features()
	fatal(err)

	fmt.Println("running", cl("godog", cyan)+", num registered steps:", cl(len(s.steps), yellow))
	fmt.Println("have loaded", cl(len(s.features), yellow), "features from path:", cl(cfg.featuresPath, green))
}
