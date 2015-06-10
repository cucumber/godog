package godog

import "regexp"

var stepHandlers map[*regexp.Regexp]StepHandler

type StepHandler interface {
	HandleStep(args ...interface{}) error
}

type StepHandlerFunc func(...interface{}) error

func (f StepHandlerFunc) HandleStep(args ...interface{}) error {
	return f(args...)
}

func Step(exp *regexp.Regexp, h StepHandler) {
	stepHandlers[exp] = h
}

func init() {
	stepHandlers = make(map[*regexp.Regexp]StepHandler)
}
