package godog

import (
	"log"
	"regexp"
)

type Suite interface {
	Step(exp *regexp.Regexp, h StepHandler)
}

type suite struct {
	steps map[*regexp.Regexp]StepHandler
}

func New() *suite {
	return &suite{
		steps: make(map[*regexp.Regexp]StepHandler),
	}
}

func (s *suite) Step(exp *regexp.Regexp, h StepHandler) {
	s.steps[exp] = h
}

func (s *suite) Run() {
	log.Println("running godoc, num registered steps:", len(s.steps))
}
