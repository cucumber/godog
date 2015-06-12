package godog

import (
	"log"
	"regexp"
)

type Suite interface {
	Step(exp *regexp.Regexp, h StepHandler)
}

type GodogSuite struct {
	steps map[*regexp.Regexp]StepHandler
}

func (s *GodogSuite) Step(exp *regexp.Regexp, h StepHandler) {
	s.steps[exp] = h
}

func (s *GodogSuite) Run() {
	log.Println("running godoc, num registered steps:", len(s.steps))
}
